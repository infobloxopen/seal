package compiler_rego

import (
	"fmt"
	"strconv"
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
	lineNots     int // number of nots per line currently encountered during compileCondition
	swaggerTypes []types.Type
	swaggerMap   map[string]*types.Type // by-name convenience map into swaggerTypes slice
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

	c.swaggerTypes = swaggerTypes

	// Build by-name convenience map into swaggerTypes slice
	c.swaggerMap = make(map[string]*types.Type)
	for i, swt := range c.swaggerTypes {
		c.swaggerMap[swt.String()] = &c.swaggerTypes[i]
	}

	compiled := []string{
		"",
		fmt.Sprintf("package %s", pkgname),
	}

	compiled = append(compiled, c.compileSetDefaults("false", "allow", "deny")...)

	compiled = append(compiled, c.compileBaseVerbs()...)

	var compiledObligations []string

	var lineNum int
	for idx, stmt := range pols.Statements {
		var out string
		var err error
		var stmtObligations []string

		lineNum += 1
		switch stmt.(type) {
		case *ast.ActionStatement:
			out, stmtObligations, err = c.compileStatement(stmt.(*ast.ActionStatement), lineNum)
		case *ast.ContextStatement:
			out, stmtObligations, err = c.compileContextStatement(stmt.(*ast.ContextStatement), &lineNum)
		}

		if err != nil {
			return "", compiler_error.New(err, idx, fmt.Sprintf("%s", stmt))
		}
		compiled = append(compiled, out)
		compiledObligations = append(compiledObligations, stmtObligations...)
	}

	// Add collected obligations to compiled rego outout
	compiled = append(compiled, "")
	compiled = append(compiled, "obligations := [")
	for _, oblige := range compiledObligations {
		compiled = append(compiled, fmt.Sprintf("`%s`,", oblige))
	}
	compiled = append(compiled, "]")

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
func (c *CompilerRego) compileBaseVerbs() []string {
	compiled := []string{}
	compiled = append(compiled, "")
	compiled = append(compiled, "base_verbs := {")
	for _, swt := range c.swaggerTypes {
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
	logger := logrus.WithField("method", "linearizeContext")
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
	logger.WithField("line", line).Debug("linearize")
	return line
}

func (c *CompilerRego) compileContextStatement(stmt *ast.ContextStatement, lineNum *int) (string, []string, error) {
	var err error
	rego := "\n"
	var line []*ast.ActionStatement
	var contextObligations []string

	line = c.linearizeContext(stmt)

	for _, li := range line {
		var cs string
		var stmtObligations []string
		*(lineNum)++
		if cs, stmtObligations, err = c.compileStatement(li, *lineNum); err != nil {
			return "", nil, err
		}

		rego += cs + "\n"
		contextObligations = append(contextObligations, stmtObligations...)
	}

	return rego, contextObligations, err
}

// compileStatement converts the AST statement to a string
func (c *CompilerRego) compileStatement(stmt *ast.ActionStatement, lineNum int) (string, []string, error) {
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
			return "", nil, err
		}
		compiled = append(compiled, sub)
	}

	vrb, err := c.compileVerb(stmt.Verb)
	if err != nil {
		return "", nil, err
	}
	compiled = append(compiled, vrb)

	tp, swtype, err := c.compileTypePattern(stmt.TypePattern)
	if err != nil {
		return "", nil, err
	}
	compiled = append(compiled, tp)

	cnds, whereObligations, err := c.compileWhereClause(swtype, stmt.WhereClause, lineNum)
	if err != nil {
		return "", nil, err
	}
	if cnds != "" {
		compiled = append(compiled, cnds)
	}

	compiled = append(compiled, "}")

	return strings.Join(compiled, "\n"), whereObligations, nil
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
func (c *CompilerRego) compileTypePattern(tp *ast.Identifier) (string, *types.Type, error) {
	logger := logrus.WithField("method", "compileTypePattern")
	if tp == nil {
		return "", nil, compiler_error.ErrEmptyTypePattern
	}

	// TODO: optimize with list of registered types instead of regex
	quoted := strings.ReplaceAll(tp.Value, "*", ".*")
	quoted = strings.ReplaceAll(quoted, "..*", ".*")

	swtype := c.swaggerMap[tp.Value]
	swtypeStr := "nil"
	if swtype != nil {
		swtypeStr = (*swtype).String()
	}

	result := fmt.Sprintf("    re_match(`%s`, input.type)", quoted)
	logger.WithFields(logrus.Fields{
		"tpValue": tp.Value,
		"swtype":  swtypeStr,
		"quoted":  quoted,
		"result":  result,
	}).Debug("compileTypePattern")

	return result, swtype, nil
}

// String satifies stringer interface
func (c *CompilerRego) String() string {
	return fmt.Sprintf("compiler for %s language", Language)
}

func (c *CompilerRego) compileWhereClause(swtype *types.Type, cnds ast.Condition, lineNum int) (string, []string, error) {
	if types.IsNilInterface(cnds) {
		return "", nil, nil
	}

	c.lineNots = 0
	switch s := cnds.(type) {
	case *ast.WhereClause:
		condString, obligations, isObligation, err := c.compileCondition(swtype, s.Condition, 0, lineNum)
		if err != nil {
			return "", nil, err
		}

		if isObligation {
			condString = ""
			obligations = append(obligations, s.Condition.String())
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
		return condString, obligations, nil
	default:
		return "", nil, compiler_error.ErrUnknownWhereClause
	}
}

func (c *CompilerRego) compileCondition(swtype *types.Type, o ast.Condition, lvl, lineNum int) (string, []string, bool, error) {
	logger := logrus.WithField("method", "compileCondition").WithField("lvl", lvl).WithField("condition", o.String())
	if types.IsNilInterface(o) {
		return "", nil, false, nil
	}

	logger.WithField("type", fmt.Sprintf("%#v", o)).Debug("compileCondition")

	switch s := o.(type) {
	case *ast.Identifier:
		switch s.Token.Type {
		case token.LITERAL:
			logger.WithField("result", s.String()).Debug("s.Token.Type==token.LITERAL")
			return s.String(), nil, false, nil
		}

		id := s.Token.Literal
		logger.WithField("id", id).Debug("s.Token.Type!=token.LITERAL")
		var isObligation bool

		if strings.HasPrefix(id, "ctx.") {
			id = strings.Replace(id, "ctx.", "", 1)
			id = strings.Replace(id, "\"]", "", 1)
			id = strings.Replace(id, "[\"", ".", 1)

			// If object-type is known, check property exists and if it is obligation
			if swtype != nil {
				swlogger := logger.WithField("swtype", (*swtype).String()).WithField("id", id)
				// Get the 0th component of the property, in case the condition
				// is something like: ctx.tags["color"] == "blue"
				id0 := strings.Split(id, ".")[0]

				propMap := (*swtype).GetProperties()
				pprop, ok := propMap[id0]
				if !ok {
					return "", nil, false, fmt.Errorf("Unknown property '%s' of type '%s'",
						id0, (*swtype).String())
				}

				x_seal_obligation, ok, err := pprop.GetExtensionProp("x-seal-obligation")
				if err != nil {
					return "", nil, false, fmt.Errorf("type '%s': %s", (*swtype).String(), err)
				} else if ok {
					swlogger.WithField("x_seal_obligation", x_seal_obligation).Debug("x_seal_obligation")
					var err error
					isObligation, err = strconv.ParseBool(x_seal_obligation)
					if err != nil {
						return "", nil, false, fmt.Errorf("Bad bool value '%s' for property '%s' of type '%s'",
							x_seal_obligation, id0, (*swtype).String())
					}
				}
			}

			lid := strings.Split(id, ".")
			id = "input.ctx[i][\"" + strings.Join(lid, "\"][\"") + "\"]"
		}
		if strings.HasPrefix(id, types.SUBJECT+".") {
			id = strings.Replace(id, types.SUBJECT, "seal_subject", 1)
		}

		logger.WithField("id", id).WithField("isObligation", isObligation).Debug("isObligation")
		return id, nil, isObligation, nil

	case *ast.IntegerLiteral:
		id := s.Token.Literal
		return id, nil, false, nil

	case *ast.PrefixCondition:
		rhs, subObligations, subIsObligation, err := c.compileCondition(swtype, s.Right, lvl+1, lineNum)
		if err != nil {
			return "", nil, false, err
		}

		switch s.Token.Type {
		case token.NOT:
			c.lineNots += 1
			ref := fmt.Sprintf("line%d_not%d_cnd", lineNum, c.lineNots)
			return fmt.Sprintf("%snot %s\n}\n%s {\n"+SOME_I+"\n%s\n", spaces(lvl+1), ref, ref, rhs), subObligations, subIsObligation, nil
		}
		return fmt.Sprintf(SOME_I+"\n%s %s", s.Token.Literal, rhs), subObligations, subIsObligation, nil

	case *ast.InfixCondition:
		lhs, subObligations, lhsIsObligation, err := c.compileCondition(swtype, s.Left, lvl+1, lineNum)
		if err != nil {
			return "", nil, false, err
		}
		rhs, rhsObligations, rhsIsObligation, err := c.compileCondition(swtype, s.Right, lvl+1, lineNum)
		if err != nil {
			return "", nil, false, err
		}

		subObligations = append(subObligations, rhsObligations...)

		// if strings.Contains(lhs, SOME_I) && strings.Contains(rhs, SOME_I) {
		// 	lhs = strings.ReplaceAll(lhs, SOME_I+"\n", "")
		// 	rhs = strings.ReplaceAll(rhs, SOME_I+"\n", "")
		// }
		condString := ""
		switch s.Token.Type {
		case token.AND:
			// Assume obligation conditions cannot include AND nor OR,
			// so we can now generate fully-specified obligation condition(s)
			// if either side of the AND/OR has an obligation property
			if lhsIsObligation {
				lhsIsObligation = false
				subObligations = append(subObligations, s.Left.String())
			} else {
				condString = fmt.Sprintf("%s", lhs)
			}
			if rhsIsObligation {
				rhsIsObligation = false
				subObligations = append(subObligations, s.Right.String())
			} else {
				condString = fmt.Sprintf("%s\n%s", condString, rhs)
			}
			condString = strings.Trim(condString, "\n")
		case token.OR:
			return "", nil, false, fmt.Errorf("OR operator not supported yet")
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

		return condString, subObligations, (lhsIsObligation || rhsIsObligation), nil
	default:
		logger.WithField("type", fmt.Sprintf("%#v", o)).Warn("unknown_condition")
		return "", nil, false, compiler_error.ErrUnknownCondition
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
