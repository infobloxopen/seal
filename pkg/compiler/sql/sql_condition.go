package sqlcompiler

// https://www.postgresql.org/docs/9.3/functions-matching.html
// https://dba.stackexchange.com/questions/10694/pattern-matching-with-like-similar-to-or-regular-expressions-in-postgresql

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

// SQLReplacerFn is function to perform optional string replacement.
// The original string should always be returned if not modified, even on error.
type SQLReplacerFn func(sqlc *SQLCompiler, idParts *lexer.IdentifierParts, src string) (string, error)

// SQLCompiler contains SQL conversion parameters
type SQLCompiler struct {
	Logger              *logrus.Logger
	Dialect             SQLDialectEnum
	IdentifierReplacers []SQLReplacerFn
	LiteralReplacers    []SQLReplacerFn
}

// NewSQLCompiler returns new instance of SQLCompiler
func NewSQLCompiler(sqlOpts ...SQLOption) *SQLCompiler {
	sqlc := &SQLCompiler{
		Logger: logrus.StandardLogger(),
	}

	for _, opt := range sqlOpts {
		opt(sqlc)
	}

	return sqlc
}

// CompileCondition compiles the given input condition string into an SQL condition string.
func (sqlc *SQLCompiler) CompileCondition(annotatedCondition string) (string, error) {
	logger := sqlc.Logger.WithField("method", "CompileCondition")

	singleCondition, _ := parser.SplitKeyValueAnnotations(annotatedCondition)
	ast, err := parser.ParseCondition(singleCondition)
	if err != nil {
		return "", err
	} else if ast == nil {
		return "", fmt.Errorf("Unknown error parsing condition: %s", singleCondition)
	}

	singleWhere, err := sqlc.astConditionToSQL(0, ast)
	if err != nil {
		return "", err
	}
	logger.WithField("where", singleWhere).Trace("single_where_clause")

	return singleWhere, nil
}

func (sqlc *SQLCompiler) astConditionToSQL(lvl int, o ast.Condition) (string, error) {
	logger := sqlc.Logger.WithField("method", "astConditionToSQL").WithField("lvl", lvl).WithField("condition", o.String())
	if types.IsNilInterface(o) {
		return "", nil
	}

	logger.WithField("type", fmt.Sprintf("%#v", o)).Trace("astConditionToSQL")

	squoteReplacer := strings.NewReplacer(
		`'`, `''`,
	)

	switch s := o.(type) {
	case *ast.Identifier:
		switch s.Token.Type {
		case token.LITERAL:
			var err error
			literal := s.String()

			// If double-quoted string literal:
			//   Escape any single-quotes
			//   Replace begin/end double-quotes with single-quotes
			if strings.HasPrefix(literal, `"`) && strings.HasSuffix(literal, `"`) {
				literal, err = sqlc.applyLiteralReplacers(literal[1 : len(literal)-1])
				if err != nil {
					return "", err
				}
				literal = squoteReplacer.Replace(literal)
				literal = `'` + literal + `'`
			} else {
				// This should never happen as this is invalid SEAL,
				// but SQL doesn't support unquoted literals,
				// so we'll return error
				return "", fmt.Errorf("Cannot SQL-convert unquoted literal-string: '%s'", literal)
			}

			logger.WithField("literal", literal).Trace("s.Token.Type==token.LITERAL")
			return literal, nil
		}

		id, err := sqlc.applyIdentifierReplacers(s.Token.Literal)
		if err != nil {
			return "", err
		}

		if lexer.IsIndexedIdentifier(id) {
			return "", fmt.Errorf("Do not know how to SQL-convert indexed-identifier: %s", id)
		}

		logger.WithField("id", id).Trace("s.Token.Type!=token.LITERAL")
		return id, nil

	case *ast.IntegerLiteral:
		literal, err := sqlc.applyLiteralReplacers(s.Token.Literal)
		if err != nil {
			return "", err
		}
		return literal, nil

	case *ast.PrefixCondition:
		rhs, err := sqlc.astConditionToSQL(lvl+1, s.Right)
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
		lhs, err := sqlc.astConditionToSQL(lvl+1, s.Left)
		if err != nil {
			return "", err
		}

		rhs, err := sqlc.astConditionToSQL(lvl+1, s.Right)
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
			result = fmt.Sprintf("(%s ~ %s)", lhs, rhs)
		case token.OP_IN:
			//TODO: select * from permissions where id in ('tag-manage', 'tag-view');
			return "", fmt.Errorf("IN operator not supported yet: %s", o)
		default:
			result = fmt.Sprintf("%s %s %s", lhs, s.Token.Literal, rhs)
		}

		return result, nil

	default:
		logger.WithField("type", fmt.Sprintf("%#v", o)).Warn("unknown_condition")
		return "", fmt.Errorf("unknown_condition")
	}
}

// Always returns original id on error
func (sqlc *SQLCompiler) applyIdentifierReplacers(id string) (string, error) {
	newId := id
	for nth, replacerFn := range sqlc.IdentifierReplacers {
		var err error
		idParts := lexer.SplitIdentifier(newId)
		newId, err = replacerFn(sqlc, idParts, newId)
		if err != nil {
			return id, fmt.Errorf("Replacer %d on identifier '%s' failed: %s", nth, id, err)
		}
	}
	return newId, nil
}

// Always returns original literal on error
func (sqlc *SQLCompiler) applyLiteralReplacers(literal string) (string, error) {
	newLiteral := literal
	for nth, replacerFn := range sqlc.LiteralReplacers {
		var err error
		newLiteral, err = replacerFn(sqlc, nil, newLiteral)
		if err != nil {
			return literal, fmt.Errorf("Replacer %d on literal '%s' failed: %s", nth, literal, err)
		}
	}
	return newLiteral, nil
}
