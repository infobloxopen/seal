package sqlcompiler

// http://www.silota.com/docs/recipes/sql-postgres-json-data-types.html
// https://kb.objectrocket.com/postgresql/how-to-query-a-postgres-jsonb-column-1433
// https://www.postgresql.org/message-id/b9341406-f066-ea08-5c0d-1a7404a95df3%40postgrespro.ru

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/seal/pkg/lexer"
)

// JSONB operators
const (
	JSONBObjectOperator = `->`
	JSONBTextOperator   = `->>`
	JSONBExistsOperator = `?`
)

// JSONBReplacer contains JSONB conversion parameters
type JSONBReplacer struct {
	JSONBOperator string
	IsNumericKey  bool
}

// JSONBOption is option function for JSONB conversion parameters
type JSONBOption func(jsonb *JSONBReplacer)

// WithJSONBOperator is an option to specify JSONB operator
func WithJSONBOperator(op string) JSONBOption {
	return func(jsonb *JSONBReplacer) {
		jsonb.JSONBOperator = op
	}
}

// WithIsNumericKey is an option to specify whether JSONB index key is numeric or not
func WithIsNumericKey(isNumericKey bool) JSONBOption {
	return func(jsonb *JSONBReplacer) {
		jsonb.IsNumericKey = isNumericKey
	}
}

// NewJSONBReplacer returns an SQLReplacerFn that replaces indexed-identifiers with Postgres JSONB
func NewJSONBReplacer(jsonbOpts ...JSONBOption) SQLReplacerFn {
	jsonb := &JSONBReplacer{
		JSONBOperator: JSONBObjectOperator,
		IsNumericKey:  false,
	}

	for _, opt := range jsonbOpts {
		opt(jsonb)
	}

	return func(sqlc *SQLCompiler, idParts *lexer.IdentifierParts, id string) (string, error) {
		if sqlc.Dialect != DialectPostgres {
			return "", fmt.Errorf("Dialect %s does not support JSONB conversion of id: %s", sqlc.Dialect, id)
		}

		if len(idParts.Key) <= 0 {
			return id, nil
		}

		var result strings.Builder

		if len(idParts.Table) > 0 {
			result.WriteString(idParts.Table)
			result.WriteString(`.`)
		}

		result.WriteString(idParts.Field)
		result.WriteString(jsonb.JSONBOperator)

		if !jsonb.IsNumericKey {
			result.WriteString(`'`)
		}

		result.WriteString(idParts.Key)

		if !jsonb.IsNumericKey {
			result.WriteString(`'`)
		}

		return result.String(), nil
	}
}
