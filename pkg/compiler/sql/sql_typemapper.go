package sqlcompiler

// http://www.silota.com/docs/recipes/sql-postgres-json-data-types.html
// https://kb.objectrocket.com/postgresql/how-to-query-a-postgres-jsonb-column-1433
// https://www.postgresql.org/message-id/b9341406-f066-ea08-5c0d-1a7404a95df3%40postgrespro.ru

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infobloxopen/seal/pkg/lexer"
)

// JSONB operators
const (
	JSONBObjectOperator = `->`
	JSONBTextOperator   = `->>`
	JSONBExistsOperator = `?`
)

// PropertyMapper contains mapping parameters for Swagger properties
type PropertyMapper struct {
	TypeMapper      *TypeMapper
	SEALProperty    string
	SQLColumn       string
	JSONBOperator   string
	JSONBIntKeyFlag bool
}

// NewPropertyMapper returns new instance of PropertyMapper for specified property 'name'.
// The specified property can be "*" to allow matching of any property name,
// as long as there is no specific match.
// Note that these two statements are equivalent:
//     pmpr := NewPropertyMapper("tags")
//     pmpr := NewPropertyMapper("").WithSEALProperty("tags")
func NewPropertyMapper(name string) *PropertyMapper {
	pmpr := &PropertyMapper{
		SEALProperty:    name,
		SQLColumn:       name,
		JSONBOperator:   JSONBObjectOperator,
		JSONBIntKeyFlag: false,
	}
	return pmpr
}

// WithSEALProperty specifies the property name, overriding the previously set name.
// Property name can be "*" to match any property.
// However a PropertyMapper with a more specific match takes precedence over "*".
func (pmpr *PropertyMapper) WithSEALProperty(name string) *PropertyMapper {
	pmpr.SEALProperty = name
	return pmpr
}

// ToSQLColumn specifies what column name the property should map to.
// If column name specified is "*", then the column name will be whatever
// the input property name is that matched this PropertyMapper.
// For example, if ToSQLColumn("*") is specified and the input property name
// is "tags", then the column name in the converted SQL will be "tags".
func (pmpr *PropertyMapper) ToSQLColumn(name string) *PropertyMapper {
	pmpr.SQLColumn = name
	return pmpr
}

// UseJSONBOperator specifies the JSONB operator to use for converting indexed properties.
// The default is JSONBObjectOperator (ie: "->").
// Only the DialectPostgres SQL dialect supports JSONB conversion.
func (pmpr *PropertyMapper) UseJSONBOperator(op string) *PropertyMapper {
	pmpr.JSONBOperator = op
	return pmpr
}

// UseJSONBIntKeyFlag specifies the JSONB integer index flag to use for converting indexed properties.
// The default is false.
// Only the DialectPostgres SQL dialect supports JSONB conversion.
func (pmpr *PropertyMapper) UseJSONBIntKeyFlag(flag bool) *PropertyMapper {
	pmpr.JSONBIntKeyFlag = flag
	return pmpr
}

// TypeMapper contains mapping parameters for Swagger types
type TypeMapper struct {
	SQLCompiler *SQLCompiler
	SwaggerType string
	SQLTable    string
	PptyMappers map[string]*PropertyMapper
}

// NewTypeMapper returns new instance of NewTypeMapper for specified swagger type 'name'.
// The specified type can be "app.*" to allow matching of any application-specific swagger type name,
// as long as there is no specific match.
// Note that these two statements are equivalent:
//     tmpr := NewTypeMapper("ddi.ipam")
//     tmpr := NewTypeMapper("").WithSwaggerType("ddi.ipam")
func NewTypeMapper(name string) *TypeMapper {
	tmpr := &TypeMapper{
		SwaggerType: name,
		SQLTable:    name,
		PptyMappers: map[string]*PropertyMapper{},
	}
	return tmpr
}

// WithSwaggerType specifies the swagger type name, overriding the previously set name.
// Swagger type name can be "app.*" to match any application-specific swagger type.
// However a TypeMapper with a more specific match takes precedence over "app.*".
func (tmpr *TypeMapper) WithSwaggerType(name string) *TypeMapper {
	tmpr.SwaggerType = name
	return tmpr
}

// ToSQLTable specifies what table name the swagger type should map to.
// If table name specified is "*", then the table name will be whatever
// the input swagger type name is that matched this TypeMapper.
// For example, if ToSQLTable("*") is specified and the input swagger type name
// is "ddi.ipam", then the table name in the converted SQL will be "ipam".
func (tmpr *TypeMapper) ToSQLTable(name string) *TypeMapper {
	tmpr.SQLTable = name
	return tmpr
}

// WithPropertyMapper adds PropertyMapper to this TypeMapper.
// PropertyMapper must be name-unique within TypeMapper.
// When adding multiple PropertyMapper with the same name, the most recent add wins.
func (tmpr *TypeMapper) WithPropertyMapper(pmpr *PropertyMapper) *TypeMapper {
	tmpr.PptyMappers[pmpr.SEALProperty] = pmpr
	pmpr.TypeMapper = tmpr
	return tmpr
}

// ReplaceIdentifier performs type and property SQL mapping on the given SEAL identifier "id".
// "swtype" is the swagger type for this identifier.
// If there is no mapping that matches "swtype" or "id", then original "id" is returned with nil error.
// Always returns original id on error.
//
// For example, TypeMapper("ddi.ipam") will match swtype "ddi.ipam".
// TypeMapper("ddi.*") will also match swtype "ddi.ipam" if there is no TypeMapper("ddi.ipam").
//
// For example, PropertyMapper("tags") will match id "ctx.tags".
// PropertyMapper("*") will also match id "ctx.tags" if there is no PropertyMapper("tags").
func (tmpr *TypeMapper) ReplaceIdentifier(swtype string, id string) (string, error) {
	swParts := lexer.SplitSwaggerType(swtype)
	if swtype != tmpr.SwaggerType {
		savedType := swParts.Type
		swParts.Type = `*`
		if swParts.String() != tmpr.SwaggerType {
			return id, nil
		}
		swParts.Type = savedType
	}

	idParts := lexer.SplitIdentifier(id)
	pmpr, foundPpty := tmpr.PptyMappers[idParts.Field]
	if !foundPpty {
		pmpr, foundPpty = tmpr.PptyMappers[`*`]
		if !foundPpty {
			return id, nil
		}
	}

	if len(idParts.Key) > 0 {
		if tmpr.SQLCompiler.Dialect != DialectPostgres {
			return id, fmt.Errorf("SQL dialect %s does not support JSONB conversion of type/id: %s/%s",
				tmpr.SQLCompiler.Dialect, swtype, id)
		} else if pmpr.JSONBIntKeyFlag {
			_, err := strconv.ParseUint(idParts.Key, 10, 0)
			if err != nil {
				return id, fmt.Errorf("JSONB index key is not unsigned-integer for type/id: %s/%s", swtype, id)
			}
		}
	}

	var newID strings.Builder

	if tmpr.SQLTable == `*` {
		newID.WriteString(swParts.Type)
	} else {
		newID.WriteString(tmpr.SQLTable)
	}

	newID.WriteString(`.`)

	if pmpr.SQLColumn == `*` {
		newID.WriteString(idParts.Field)
	} else {
		newID.WriteString(pmpr.SQLColumn)
	}

	if len(idParts.Key) > 0 {
		newID.WriteString(pmpr.JSONBOperator)

		if !pmpr.JSONBIntKeyFlag {
			newID.WriteString(`'`)
		}

		newID.WriteString(idParts.Key)

		if !pmpr.JSONBIntKeyFlag {
			newID.WriteString(`'`)
		}
	}

	return newID.String(), nil
}
