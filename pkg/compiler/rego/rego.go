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

const (
	SOME_I = "some.i"
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
func (c *CompilerRego) Compile(pkgname string, pols *ast.Policies, swaggerTypes []types.Type) (string, error) {
	if pols == nil {
		return "", compiler_error.ErrEmptyPolicies
	}

	compiled := []string{
		"",
		fmt.Sprintf("package %s", pkgname),
	}

	compiled = append(compiled, c.compileSetDefaults("false", "allow", "deny")...)

	compiled = append(compiled, c.compileBaseVerbs(swaggerTypes)...)

	var lineNum int
	for idx, stmt := range pols.Statements {
		var out string
		var err error

		lineNum += 1
		switch stmt.(type) {
		case *ast.ActionStatement:
			out, err = c.compileStatement(stmt.(*ast.ActionStatement), lineNum)
		case *ast.ContextStatement:
			out, err = c.compileContextStatement(stmt.(*ast.ContextStatement), &lineNum)
		}

		if err != nil {
			return "", compiler_error.New(err, idx, fmt.Sprintf("%s", stmt))
		}
		compiled = append(compiled, out)
	}

	compiled = append(compiled, CompiledRegoHelpers)

	return c.prettify(strings.Join(compiled, "\n")), nil
}

func (c *CompilerRego) isOpenBracket(sym byte) bool {
	return sym == '{' || sym == '[' || sym == '('
}
func (c *CompilerRego) isCloseBracket(openingBkt byte, sym byte) bool {
	return (openingBkt == '{' && sym == '}') ||
		(openingBkt == '[' && sym == ']') ||
		(openingBkt == '(' && sym == ')')
}

func (c *CompilerRego) prettify(rego string) string {
	rego = strings.Trim(rego, " \t")
	rego = strings.ReplaceAll(rego, "\r", "\n")
	for strings.Contains(rego, "\n\n\n") {
		rego = strings.ReplaceAll(rego, "\n\n\n", "\n\n")
	}

	indent := 0
        var bktStack []byte
	list := strings.Split(rego, "\n")
	for i := 0; i < len(list); i++ {
		list[i] = strings.Trim(list[i], " \t")
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
		if len(bktStack) > 0 {
			// If closing bracket matches opening bracket,
			// pop closing bracket off stack, and decrease indent
			if c.isCloseBracket(bktStack[len(bktStack)-1], list[i][0]) {
				bktStack = bktStack[0:len(bktStack)-1]
				indent--
			}
		}
		list[i] = strings.Repeat("    ", indent) + list[i]

		if c.isOpenBracket(list[i][len(list[i])-1]) {
			// Push opening bracket onto stack, and increase indent
			bktStack = append(bktStack, list[i][len(list[i])-1])
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
	compiled = append(compiled, "")
	for _, id := range ids {
		compiled = append(compiled, fmt.Sprintf("default %s = %s", id, val))
	}

	return compiled
}

// compileBaseVerbs defines base_verb mappings
func (c *CompilerRego) compileBaseVerbs(swaggerTypes []types.Type) []string {
	compiled := []string{}
	compiled = append(compiled, "")
	compiled = append(compiled, "base_verbs := {")
	for _, swt := range swaggerTypes {
		seal_verbs := swt.GetVerbs()
		if len(seal_verbs) > 0 {
			compiled = append(compiled, fmt.Sprintf("\"%s\": {", swt.String()))
			for _, sv := range seal_verbs {
				compiled = append(compiled, fmt.Sprintf("\"%s\": [", sv.GetName()))
				for _, bv := range sv.GetBaseVerbs() {
					compiled = append(compiled, fmt.Sprintf("\"%s\",", bv))
				}
				compiled = append(compiled, "],")
			}
			compiled = append(compiled, "},")
		}
	}

	compiled = append(compiled, "}")
	return compiled
}

// ast.ContextStatement contains tree-like data
// 'upper' part named 'Conditions' is the list of objects, that contain subjects and\or conditions
// 'lower' part named 'ActionRules' contains action, that also might contain context
// it should be exploded to the list of ActionStatement
func (c *CompilerRego) linearizeContext(stmt *ast.ContextStatement) []*ast.ActionStatement {
	line := []*ast.ActionStatement{}

	// range for each condition and action
	for _, cond := range stmt.Conditions {
		for _, act := range stmt.ActionRules {
			if types.IsNilInterface(act.Context) {
				// non-context ActionRule, it should be mapped to the single ActionStatement
				// initializing it with default values
				cAction := &ast.ActionStatement{
					Token:       act.Action.Token, // token is taken from ActionRule
					Action:      act.Action,       // and Action (allow, deny, etc) too.
					Verb:        stmt.Verb,        // By default Verb (to operate\read\...) is taken from context record
					TypePattern: stmt.TypePattern, // and TypePattern (petstore.pet, as an example) too.
					Subject:     cond.Subject,     // Subject is taken from condition
					WhereClause: cond.Where,       // and WhereClause too
				}

				if !types.IsNilInterface(act.Verb) { // If Verb is defined for action - context's verb should be replaced
					cAction.Verb = act.Verb
				}
				if !types.IsNilInterface(act.Subject) { // and subject
					cAction.Subject = act.Subject
				}
				if !types.IsNilInterface(act.TypePattern) { // and type
					cAction.TypePattern = act.TypePattern
				}

				if !types.IsNilInterface(act.Where) { // and Where, but it's a little harder
					if types.IsNilInterface(cond.Where) {
						// if no Where in context - just use Where from action
						cAction.WhereClause = act.Where
					} else {
						// if Where defined in context and in ActionRule
						// I should use both like (Where1) and (Where2)
						cAction.WhereClause = &ast.WhereClause{
							Token: act.Where.Token,
							Condition: &ast.InfixCondition{
								Token:    token.Token{Type: token.AND, Literal: token.AND},
								Left:     act.Where.Condition,
								Operator: token.AND,
								Right:    cond.Where.Condition,
							},
						}
					}
				}

				// And append generated ActionStatement to the list
				line = append(line, cAction)
			} else {
				// in case of context in action it also should be exploded to list of ActionStatement
				ctx := act.Context

				for _, icond := range stmt.Conditions { // add conditions defined for 'parent' context
					// but only in case it's not blank, mean parent does not looks like context {}...
					if !types.IsNilInterface(icond.Subject) || !types.IsNilInterface(icond.Where) {
						ctx.Conditions = append(ctx.Conditions, icond)
					}
				}

				// expand nested context and add resulting []ActionStatement to the current list
				line = append(line, c.linearizeContext(ctx)...)
			}
		}
	}
	return line
}

func (c *CompilerRego) compileContextStatement(stmt *ast.ContextStatement, lineNum *int) (string, error) {
	var err error
	rego := "\n"
	var line []*ast.ActionStatement

	line = c.linearizeContext(stmt)

	for _, li := range line {
		var cs string
		*(lineNum)++
		if cs, err = c.compileStatement(li, *lineNum); err != nil {
			return "", err
		}

		rego += cs + "\n"
	}

	return rego, err
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

	if !types.IsNilInterface(stmt.Subject) {
		sub, err := c.compileSubject(stmt.Subject)
		if err != nil {
			return "", err
		}
		compiled = append(compiled, sub)
	}

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

	regoStr := fmt.Sprintf("    seal_list_contains(base_verbs[input.type][`%s`], input.verb)",
		vrb.Value)
	return regoStr, nil
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
		condString, err := c.compileCondition(s.Condition, 0, lineNum)
		if err != nil {
			return "", err
		}

		// some.i is added everywhere it might be needed
		// and now extra some.i should be removed
		arr := strings.Split(condString, "{")
		for i := 0; i < len(arr); i++ {
			if !strings.Contains(arr[i], "[i]") {
				// some.i is not needed, in case block does not contain [i]
				arr[i] = strings.ReplaceAll(arr[i], SOME_I+"\n", "")
			} else {
				// first some.i is replaced with 'some i'
				// and all other some.i strings are removed from block
				arr[i] = strings.Replace(arr[i], SOME_I, "some i", 1)
				arr[i] = strings.ReplaceAll(arr[i], SOME_I+"\n", "")
			}
		}
		condString = strings.Join(arr, "{")

		// add blank line before 'some i'
		condString = strings.ReplaceAll(condString, "some i", "\nsome i")
		// and remove it in case 'some i' in the beginning of the block
		condString = strings.ReplaceAll(condString, "{\n\nsome i", "{\nsome i")
		return condString, nil
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
			id = strings.Replace(id, "ctx.", "", 1)
			id = strings.Replace(id, "\"]", "", 1)
			id = strings.Replace(id, "[\"", ".", 1)
			lid := strings.Split(id, ".")
			id = "input.ctx[i][\"" + strings.Join(lid, "\"][\"") + "\"]"
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
			return fmt.Sprintf("%snot %s\n}\n%s {\n"+SOME_I+"\n%s\n", spaces(lvl+1), ref, ref, rhs), nil
		}
		return fmt.Sprintf(SOME_I+"\n%s %s", s.Token.Literal, rhs), nil

	case *ast.InfixCondition:
		lhs, err := c.compileCondition(s.Left, lvl+1, lineNum)
		if err != nil {
			return "", err
		}
		rhs, err := c.compileCondition(s.Right, lvl+1, lineNum)
		if err != nil {
			return "", err
		}

		// if strings.Contains(lhs, SOME_I) && strings.Contains(rhs, SOME_I) {
		// 	lhs = strings.ReplaceAll(lhs, SOME_I+"\n", "")
		// 	rhs = strings.ReplaceAll(rhs, SOME_I+"\n", "")
		// }
		condString := ""
		switch s.Token.Type {
		case token.AND:
			condString = fmt.Sprintf("%s\n%s", lhs, rhs)
		case token.OR:
			condString = fmt.Sprintf("# TODO: support or: %s or %s", lhs, rhs)
		case token.OP_MATCH:
			condString = fmt.Sprintf("re_match(`%s`, %s)", strings.Trim(rhs, "\""), lhs)
		case token.OP_IN:
			condString = fmt.Sprintf("seal_list_contains(%s, `%s`)", rhs, strings.Trim(lhs, "\""))
		default:
			condString = fmt.Sprintf("%s %s %s", lhs, s.Token.Literal, rhs)
		}

		// brPos := strings.Index(lhs, "}")
		// if brPos == -1 {
		// 	brPos = len(lhs)
		// }
		// if strings.Contains(lhs[0:brPos], "ctx[i]") {
		// 	condString = SOME_I + "\n" + condString
		// }
		if strings.Contains(lhs, "ctx[i]") {
			condString = SOME_I + "\n" + condString
		}
		return condString, nil
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
