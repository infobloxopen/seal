package sqlcompiler

// https://www.postgresql.org/docs/9.3/functions-matching.html
// https://dba.stackexchange.com/questions/10694/pattern-matching-with-like-similar-to-or-regular-expressions-in-postgresql
// https://dataschool.com/how-to-teach-people-sql/how-regex-works-in-sql/
// https://lerner.co.il/2016/03/01/regexps-in-postgresql/
// https://www.educba.com/sql-regexp/

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/infobloxopen/seal/pkg/parser"
	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"
)

// SQLStringLiteralReplacer is string Replacer for converting to SQL string literals
var SQLStringLiteralReplacer = strings.NewReplacer(
	`'`, `''`, // escape single-quotes
)

// SQLCompiler contains SQL conversion parameters
type SQLCompiler struct {
	Logger      *logrus.Logger
	Dialect     SQLDialectEnum
	TypeMappers map[string]*TypeMapper
}

// NewSQLCompiler returns new instance of SQLCompiler.
func NewSQLCompiler() *SQLCompiler {
	sqlc := &SQLCompiler{
		Logger:      logrus.StandardLogger(),
		Dialect:     DialectUnknown,
		TypeMappers: map[string]*TypeMapper{},
	}
	return sqlc
}

// WithLogger specifies the Logrus logger for this compiler.
// Default is logrus.StandardLogger().
func (sqlc *SQLCompiler) WithLogger(logger *logrus.Logger) *SQLCompiler {
	sqlc.Logger = logger
	return sqlc
}

// WithDialect specifies the SQL dialect for this compiler.
// Default is DialectUnknown.
func (sqlc *SQLCompiler) WithDialect(dialect SQLDialectEnum) *SQLCompiler {
	sqlc.Dialect = dialect
	return sqlc
}

// WithTypeMapper adds TypeMapper to this compiler.
// TypeMapper must be name-unique within compiler.
// When adding multiple TypeMapper with the same name, the most recent add wins.
func (sqlc *SQLCompiler) WithTypeMapper(tmpr *TypeMapper) *SQLCompiler {
	sqlc.TypeMappers[tmpr.SwaggerType] = tmpr
	tmpr.SQLCompiler = sqlc
	return sqlc
}

// CompileCondition compiles the given SEAL annotated condition string into an SQL condition string.
// Internally calls ReplaceIdentifier to perform type and property SQL mapping on SEAL identifiers.
func (sqlc *SQLCompiler) CompileCondition(annotatedCondition string) (string, error) {
	logger := sqlc.Logger.WithField("method", "CompileCondition")

	// Extract type annotation and SEAL condition string
	singleCondition, annotationsMap := parser.SplitKeyValueAnnotations(annotatedCondition)
	swtype := annotationsMap["type"]

	// Parse SEAL condition string into AST
	ast, err := parser.ParseCondition(singleCondition)
	if err != nil {
		return "", err
	} else if ast == nil {
		return "", fmt.Errorf("Unknown error parsing condition: %s", singleCondition)
	}

	// Compile AST into SQL
	singleWhere, err := sqlc.astConditionToSQL(0, swtype, ast)
	if err != nil {
		return "", err
	}
	logger.WithField("where", singleWhere).Trace("single_where_clause")

	return singleWhere, nil
}

// astConditionToSQL recursively walks a parsed AST condition tree and compiles into an SQL condition string.
// Internally calls ReplaceIdentifier to perform type and property SQL mapping on SEAL identifiers.
func (sqlc *SQLCompiler) astConditionToSQL(lvl int, swtype string, o ast.Condition) (string, error) {
	logger := sqlc.Logger.WithField("method", "astConditionToSQL").WithField("lvl", lvl).WithField("astcondition", o.String())
	if types.IsNilInterface(o) {
		return "", nil
	}

	logger.WithField("type", fmt.Sprintf("%#v", o)).Trace("astConditionToSQL")

	switch s := o.(type) {
	case *ast.Identifier:
		switch s.Token.Type {
		case token.LITERAL:
			literal := s.String()

			// If double-quoted string literal:
			//   Escape any single-quotes
			//   Replace begin/end double-quotes with single-quotes
			if strings.HasPrefix(literal, `"`) && strings.HasSuffix(literal, `"`) {
				literal = SQLStringLiteralReplacer.Replace(literal[1 : len(literal)-1])
				literal = `'` + literal + `'`
			} else {
				// Unquoted string literals should never happen as they are invalid SEAL,
				// but SQL doesn't support unquoted literals, so we'll return error
				return "", fmt.Errorf("Cannot SQL-convert unquoted literal-string: '%s'", literal)
			}

			logger.WithField("literal", literal).Trace("s.Token.Type==token.LITERAL")
			return literal, nil
		}

		// Map type/property of identifier into SQL table/column
		id, err := sqlc.ReplaceIdentifier(swtype, s.Token.Literal)
		if err != nil {
			return "", err
		}

		if lexer.IsIndexedIdentifier(id) {
			return "", fmt.Errorf("Do not know how to SQL-convert indexed-identifier: %s", id)
		}

		logger.WithField("id", id).Trace("s.Token.Type!=token.LITERAL")
		return id, nil

	case *ast.IntegerLiteral:
		literal := s.Token.Literal
		return literal, nil

	case *ast.ArrayLiteral:
		return sqlc.astArrayLiteralToSQL(s)

	case *ast.PrefixCondition:
		rhs, err := sqlc.astConditionToSQL(lvl+1, swtype, s.Right)
		if err != nil {
			return "", err
		}

		switch s.Token.Type {
		case token.NOT:
			return fmt.Sprintf("(NOT %s)", rhs), nil
		}

		logger.WithField("token_type", s.Token.Type).Warn("unknown_prefix_condition")
		return fmt.Sprintf("(%s %s)", s.Token.Literal, rhs), nil

	case *ast.InfixCondition:
		lhs, err := sqlc.astConditionToSQL(lvl+1, swtype, s.Left)
		if err != nil {
			return "", err
		}

		rhs, err := sqlc.astConditionToSQL(lvl+1, swtype, s.Right)
		if err != nil {
			return "", err
		}

		result := ""
		switch s.Token.Type {
		case token.AND:
			result = fmt.Sprintf("(%s AND %s)", lhs, rhs)
		case token.OR:
			result = fmt.Sprintf("(%s OR %s)", lhs, rhs)
		case token.OP_EQUAL_TO:
			result = fmt.Sprintf("(%s = %s)", lhs, rhs)
		case token.OP_MATCH:
			if sqlc.Dialect != DialectPostgres {
				return "", fmt.Errorf("SQL dialect %s does not know how to convert regexp-match: %s %s %s",
					sqlc.Dialect, s.Left, token.OP_MATCH, s.Right)
			}
			result = fmt.Sprintf("(%s ~ %s)", lhs, rhs)
		case token.OP_IN:
			if _, ok := s.Right.(*ast.ArrayLiteral); ok {
				result = fmt.Sprintf("(%s IN %s)", lhs, rhs)
			} else {
				// TODO: maybe seal: `"boss" in subject.groups`
				// would compile into sql: `('boss' IN (SELECT groups FROM subject))`
				return "", fmt.Errorf("SQL-conversion of IN operator not supported yet: %s", o)
			}
		default:
			result = fmt.Sprintf("%s %s %s", lhs, s.Token.Literal, rhs)
		}

		return result, nil

	default:
		logger.WithField("type", fmt.Sprintf("%#v", o)).Warn("unknown_condition")
		return "", fmt.Errorf("unknown_condition")
	}
}

func (sqlc *SQLCompiler) astArrayLiteralToSQL(arrLit *ast.ArrayLiteral) (string, error) {
	//logger := sqlc.Logger.WithField("method", "astArrayLiteralToSQL").WithField("arrLit", arrLit.String())
	var bldr strings.Builder
	bldr.WriteString(`(`)
	notEmpty := false
	for _, it := range arrLit.Items {
		if notEmpty {
			bldr.WriteString(`,`)
		}
		literal := it.String()
		if strings.HasPrefix(literal, `"`) && strings.HasSuffix(literal, `"`) {
			literal = SQLStringLiteralReplacer.Replace(literal[1 : len(literal)-1])
			literal = `'` + literal + `'`
		}
		bldr.WriteString(literal)
		notEmpty = true
	}
	bldr.WriteString(`)`)
	return bldr.String(), nil
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
func (sqlc *SQLCompiler) ReplaceIdentifier(swtype, id string) (string, error) {
	tmpr, foundType := sqlc.TypeMappers[swtype]
	if !foundType {
		swParts := lexer.SplitSwaggerType(swtype)
		swParts.Type = "*"
		tmpr, foundType = sqlc.TypeMappers[swParts.String()]
	}

	newID := id
	var err error
	if foundType {
		newID, err = tmpr.ReplaceIdentifier(swtype, newID)
		if err != nil {
			return id, fmt.Errorf("ReplaceIdentifier '%s' failed: %s", id, err)
		}
	}

	return newID, nil
}
