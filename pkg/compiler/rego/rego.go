package compiler_rego

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/compiler"
	"github.com/infobloxopen/seal/pkg/compiler/error"
)

// CompilerRego defines the compiler rego backend
type CompilerRego struct{}

// New creates a new compiler
func New() (compiler.Compiler, error) {
	return &CompilerRego{}, nil
}

// Compile converts the AST policies to a string
func (c *CompilerRego) Compile(pkgname string, pols *ast.Policies) (string, error) {
	if pols == nil {
		return "", compiler_error.ErrEmptyPolicies
	}

	compiled := []string{
		"",
		fmt.Sprintf("package %s", pkgname),
	}
	for idx, stmt := range pols.Statements {
		switch stmt.(type) {
		case *ast.ActionStatement:
			out, err := c.compileStatement(stmt.(*ast.ActionStatement))
			if err != nil {
				return "", compiler_error.New(err, idx, fmt.Sprintf("%s", stmt))
			}
			compiled = append(compiled, out)
		}
	}

	return strings.Join(compiled, "\n"), nil
}

// compileStatement converts the AST statement to a string
func (c *CompilerRego) compileStatement(stmt *ast.ActionStatement) (string, error) {
	compiled := []string{fmt.Sprintf("%s = true {", stmt.Token.Literal)}

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

	compiled = append(compiled, "}")
	return strings.Join(compiled, "\n"), nil
}

// compileSubject converts the AST subject to a string
func (c *CompilerRego) compileSubject(sub ast.Subject) (string, error) {
	if sub == nil {
		return "", compiler_error.ErrEmptySubject
	}

	switch t := sub.(type) {
	case *ast.SubjectGroup:
		return fmt.Sprintf("    `%s` in input.subject.groups", t.Group), nil
	case *ast.SubjectUser:
		return fmt.Sprintf("    input.subject.user == `%s`", t.User), nil
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
	return fmt.Sprintf("    re_match(`%s`, input.type)", quoted), nil
}

// String satifies stringer interface
func (c *CompilerRego) String() string {
	return fmt.Sprintf("compiler for %s language", Language)
}
