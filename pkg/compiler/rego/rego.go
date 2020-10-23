package compiler_rego

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/compiler"
	compiler_error "github.com/infobloxopen/seal/pkg/compiler/error"
	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"
	"github.com/sirupsen/logrus"
)

// CompilerRego defines the compiler rego backend
type CompilerRego struct {
	lineNots int // number of nots per line currently encountered during compileCondition
}

// New creates a new compiler
func New() (compiler.Compiler, error) {
	return &CompilerRego{}, nil
}

// CompilerRegoOption defines options
type CompilerRegoOption func(c *CompilerRego)

// Compile converts the AST policies to a string
func (c *CompilerRego) Compile(pkgname string, pols *ast.Policies) (string, error) {
	if pols == nil {
		return "", compiler_error.ErrEmptyPolicies
	}

	compiled := []string{
		"",
		fmt.Sprintf("package %s", pkgname),
	}

	compiled = append(compiled, c.compileSetDefaults("false", "allow", "deny")...)

	var lineNum int
	for idx, stmt := range pols.Statements {
		lineNum += 1
		switch stmt.(type) {
		case *ast.ActionStatement:
			out, err := c.compileStatement(stmt.(*ast.ActionStatement), lineNum)
			if err != nil {
				return "", compiler_error.New(err, idx, fmt.Sprintf("%s", stmt))
			}
			compiled = append(compiled, out)
		}
	}

	compiled = append(compiled, CompiledRegoHelpers)

	return c.prettify(strings.Join(compiled, "\n")), nil
}

func (c *CompilerRego) isOpenBracket(sym byte) bool {
	return sym == '{' || sym == '[' || sym == '('
}
func (c *CompilerRego) isCloseBracket(br byte, sym byte) bool {
	return (br == '{' && sym == '}') ||
		(br == '[' && sym == ']') ||
		(br == '(' && sym == ')')
}

func (c *CompilerRego) prettify(rego string) string {
	rego = strings.Trim(rego, " 	")
	rego = strings.ReplaceAll(rego, "\r", "\n")
	for strings.Contains(rego, "\n\n\n") {
		rego = strings.ReplaceAll(rego, "\n\n\n", "\n\n")
	}

	indent := 0
	var bOpen byte
	list := strings.Split(rego, "\n")
	for i := 0; i < len(list); i++ {
		list[i] = strings.Trim(list[i], " 	")
	}

	for i := 0; i < len(list); i++ {
		// replace \n} with }
		if i < len(list)-1 && list[i+1] == "}" && list[i] == "" {
			list = append(list[0:i], list[i+1:]...)
			i--
			continue
		}

		if len(list[i]) == 0 {
			continue
		}
		if c.isCloseBracket(bOpen, list[i][0]) {
			bOpen = 0
			indent--
		}
		list[i] = strings.Repeat("    ", indent) + list[i]

		if c.isOpenBracket(list[i][len(list[i])-1]) {
			bOpen = list[i][len(list[i])-1]
			indent++
		}

		// add newline after }
		if list[i] == "}" && i < len(list)-1 && list[i+1] != "" {
			list[i] += "\n"
		}
	}

	return strings.Join(list, "\n")
}

// compileSetDefaults sets all defaults of ids in the arguments to the value
func (c *CompilerRego) compileSetDefaults(val string, ids ...string) []string {
	compiled := []string{}
	for _, id := range ids {
		compiled = append(compiled, fmt.Sprintf("default %s = %s", id, val))
	}

	return compiled
}

// compileStatement converts the AST statement to a string
func (c *CompilerRego) compileStatement(stmt *ast.ActionStatement, lineNum int) (string, error) {
	compiled := []string{}
	action := stmt.Token.Literal
	switch action {
	case "allow":
		compiled = append(compiled, "allow {")
	case "deny":
		compiled = append(compiled, "deny {")
	}

	sub, err := c.compileSubject(stmt.Subject)
	if err != nil {
		return "", err
	}
	compiled = append(compiled, sub)

	vrb, err := c.compileVerb(stmt.Verb)
	if err != nil {
		return "", err
	}
	compiled = append(compiled, vrb)

	tp, err := c.compileTypePattern(stmt.TypePattern)
	if err != nil {
		return "", err
	}
	compiled = append(compiled, tp)

	cnds, err := c.compileWhereClause(stmt.WhereClause, lineNum)
	if err != nil {
		return "", err
	}
	if cnds != "" {
		compiled = append(compiled, cnds)
	}

	compiled = append(compiled, "}")

	return strings.Join(compiled, "\n"), nil
}

// compileSubject converts the AST subject to a string
func (c *CompilerRego) compileSubject(sub ast.Subject) (string, error) {
	if types.IsNilInterface(sub) {
		return "", compiler_error.ErrEmptySubject
	}

	switch t := sub.(type) {
	case *ast.SubjectGroup:
		return fmt.Sprintf("    seal_list_contains(seal_subject.groups, `%s`)", t.Group), nil
	case *ast.SubjectUser:
		return fmt.Sprintf("    seal_subject.sub == `%s`", t.User), nil
	}

	return "", compiler_error.ErrInvalidSubject
}

// compileVerb converts the AST verb to a string
func (c *CompilerRego) compileVerb(vrb *ast.Identifier) (string, error) {
	if vrb == nil {
		return "", compiler_error.ErrEmptyVerb
	}

	return fmt.Sprintf("    input.verb == `%s`", vrb.Value), nil
}

// compileTypePattern converts the AST type pattern to a string
func (c *CompilerRego) compileTypePattern(tp *ast.Identifier) (string, error) {
	if tp == nil {
		return "", compiler_error.ErrEmptyTypePattern
	}

	// TODO: optimize with list of registered types instead of regex
	quoted := strings.ReplaceAll(tp.Value, "*", ".*")
	quoted = strings.ReplaceAll(quoted, "..*", ".*")
	return fmt.Sprintf("    re_match(`%s`, input.type)", quoted), nil
}

// String satifies stringer interface
func (c *CompilerRego) String() string {
	return fmt.Sprintf("compiler for %s language", Language)
}

func (c *CompilerRego) compileWhereClause(cnds ast.Condition, lineNum int) (string, error) {
	if types.IsNilInterface(cnds) {
		return "", nil
	}

	c.lineNots = 0
	switch s := cnds.(type) {
	case *ast.WhereClause:
		return c.compileCondition(s.Condition, 0, lineNum)
	default:
		return "", compiler_error.ErrUnknownWhereClause
	}
}

func (c *CompilerRego) compileCondition(o ast.Condition, lvl, lineNum int) (string, error) {
	if types.IsNilInterface(o) {
		return "", nil
	}

	logrus.WithFields(logrus.Fields{
		"level": lvl,
		"type":  fmt.Sprintf("%#v", o),
	}).Debug("compileCondition: start of function")

	switch s := o.(type) {
	case *ast.Identifier:
		switch s.Token.Type {
		case token.LITERAL:
			return s.String(), nil
		}

		id := s.Token.Literal
		if strings.HasPrefix(id, "ctx.") {
			id = strings.Replace(id, "ctx", "input", 1)
		}
		if strings.HasPrefix(id, types.SUBJECT+".") {
			id = strings.Replace(id, types.SUBJECT, "seal_subject", 1)
		}

		return id, nil

	case *ast.IntegerLiteral:
		id := s.Token.Literal
		return id, nil

	case *ast.PrefixCondition:
		rhs, err := c.compileCondition(s.Right, lvl+1, lineNum)
		if err != nil {
			return "", err
		}

		switch s.Token.Type {
		case token.NOT:
			c.lineNots += 1
			ref := fmt.Sprintf("line%d_not%d_cnd", lineNum, c.lineNots)
			return fmt.Sprintf("%snot %s\n}\n%s {\n%s\n", spaces(lvl+1), ref, ref, rhs), nil
		}
		return fmt.Sprintf("%s %s", s.Token.Literal, rhs), nil

	case *ast.InfixCondition:
		lhs, err := c.compileCondition(s.Left, lvl+1, lineNum)
		if err != nil {
			return "", err
		}
		rhs, err := c.compileCondition(s.Right, lvl+1, lineNum)
		if err != nil {
			return "", err
		}

		switch s.Token.Type {
		case token.AND:
			return fmt.Sprintf("%s\n%s", lhs, rhs), nil
		case token.OR:
			return fmt.Sprintf("# TODO: support or: %s or %s", lhs, rhs), nil
		}
		return fmt.Sprintf("%s%s %s %s", spaces(lvl+1), lhs, s.Token.Literal, rhs), nil

	default:
		logrus.WithFields(logrus.Fields{
			"level": lvl,
			"type":  fmt.Sprintf("%#v", o),
		}).Warn("compileCondition: unknown type")
		return "", compiler_error.ErrUnknownCondition
	}
}

func spaces(lvl int) string {
	out := ""
	for i := 0; i < lvl; i++ {
		out += "    "
	}
	return out
}

const (
	CompiledRegoHelpers = `
# rego functions defined by seal

# Helper to get the token payload.
seal_subject = payload {
    [header, payload, signature] := io.jwt.decode(input.jwt)
}

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`
)
