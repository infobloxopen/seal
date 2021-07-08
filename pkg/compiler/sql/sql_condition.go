package sqlcompiler

// https://www.postgresql.org/docs/9.3/functions-matching.html
// https://dba.stackexchange.com/questions/10694/pattern-matching-with-like-similar-to-or-regular-expressions-in-postgresql

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/parser"
	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"
)

// SQL dialects
const (
	_ int = iota
	DialectUnknown
	DialectPostgres
)

// CompileCondition compiles the given input condition string into an SQL condition string.
// Optional colNameReplacer can be specified to adjust the column names in the SQL condition.
func CompileCondition(dialect int, singleCondition string, colNameReplacer *strings.Replacer) (string, error) {
	logger := logrus.WithField("method", "CompileCondition")
	ast, err := parser.ParseCondition(singleCondition)
	if err != nil {
		return "", err
	} else if ast == nil {
		return "", fmt.Errorf("Unknown error parsing condition: %s", singleCondition)
	}

	singleWhere, err := astConditionToSQL(dialect, ast, colNameReplacer, 0)
	if err != nil {
		return "", err
	}
	logger.WithField("where", singleWhere).Trace("single_where_clause")

	return singleWhere, nil
}

func astConditionToSQL(dialect int, o ast.Condition, colNameReplacer *strings.Replacer, lvl int) (string, error) {
	logger := logrus.WithField("method", "astConditionToSQL").WithField("lvl", lvl).WithField("condition", o.String())
	if types.IsNilInterface(o) {
		return "", nil
	}

	logger.WithField("type", fmt.Sprintf("%#v", o)).Trace("astConditionToSQL")

	sqlReplacer := strings.NewReplacer(
		`'`, `''`,
	)

	switch s := o.(type) {
	case *ast.Identifier:
		switch s.Token.Type {
		case token.LITERAL:
			result := s.String()

			// If double-quoted string literal:
			//   Escape any single-quotes
			//   Replace begin/end double-quotes with single-quotes
			if strings.HasPrefix(result, `"`) && strings.HasSuffix(result, `"`) {
				result = sqlReplacer.Replace(result)
				result = `'` + result[1:len(result)-1] + `'`
			}

			logger.WithField("result", result).Trace("s.Token.Type==token.LITERAL")
			return result, nil
		}

		id := s.Token.Literal
		if strings.ContainsAny(id, `["']`) {
			return "", fmt.Errorf("map/array indexing not supported yet: %s", id)
		}

		if colNameReplacer != nil {
			id = colNameReplacer.Replace(id)
		}

		logger.WithField("id", id).Trace("s.Token.Type!=token.LITERAL")
		return id, nil

	case *ast.IntegerLiteral:
		id := s.Token.Literal
		return id, nil

	case *ast.PrefixCondition:
		rhs, err := astConditionToSQL(dialect, s.Right, colNameReplacer, lvl+1)
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
		lhs, err := astConditionToSQL(dialect, s.Left, colNameReplacer, lvl+1)
		if err != nil {
			return "", err
		}

		rhs, err := astConditionToSQL(dialect, s.Right, colNameReplacer, lvl+1)
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
