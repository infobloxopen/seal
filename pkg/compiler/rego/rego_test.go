package compiler_rego

import (
	"testing"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/compiler/error"
	"github.com/infobloxopen/seal/pkg/token"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		pols     *ast.Policies
		expected string
		err      error
	}{
		{
			name: "validate error for empty policy",
			err:  compiler_error.ErrEmptyPolicies,
		},
		{
			name: "validate error for empty subject",
			pkg:  "foo",
			err:  compiler_error.ErrEmptySubject,
			pols: &ast.Policies{
				Statements: []ast.Statement{
					&ast.ActionStatement{
						Token: token.Token{Type: "IDENT", Literal: "allow"},
					},
				},
			},
		},
		{
			name: "validate error for empty verb",
			pkg:  "foo",
			err:  compiler_error.ErrEmptyVerb,
			pols: &ast.Policies{
				Statements: []ast.Statement{
					&ast.ActionStatement{
						Token: token.Token{Type: "IDENT", Literal: "allow"},
						Action: &ast.Identifier{
							Token: token.Token{Type: "IDENT", Literal: "allow"},
							Value: "allow",
						},
						Subject: &ast.SubjectGroup{Token: "subject", Group: "foo"},
					},
				},
			},
		},
		{
			name: "validate error for empty type-pattern",
			pkg:  "foo",
			err:  compiler_error.ErrEmptyTypePattern,
			pols: &ast.Policies{
				Statements: []ast.Statement{
					&ast.ActionStatement{
						Token: token.Token{Type: "IDENT", Literal: "allow"},
						Action: &ast.Identifier{
							Token: token.Token{Type: "IDENT", Literal: "allow"},
							Value: "allow",
						},
						Subject: &ast.SubjectGroup{Token: "subject", Group: "foo"},
						TypePattern: &ast.Identifier{
							Token: token.Token{Type: "TYPE_PATTERN", Literal: "dns.request"},
							Value: "dns.request",
						},
					},
				},
			},
		},
		{
			name: "validate policy: allow subject group foo to resolve dns.request;",
			pkg:  "foo",
			pols: &ast.Policies{
				Statements: []ast.Statement{
					&ast.ActionStatement{
						Token: token.Token{Type: "IDENT", Literal: "allow"},
						Action: &ast.Identifier{
							Token: token.Token{Type: "IDENT", Literal: "allow"},
							Value: "allow",
						},
						Subject: &ast.SubjectGroup{Token: "subject", Group: "foo"},
						Verb: &ast.Identifier{
							Token: token.Token{Type: "IDENT", Literal: "resolve"},
							Value: "resolve",
						},
						TypePattern: &ast.Identifier{
							Token: token.Token{Type: "TYPE_PATTERN", Literal: "dns.request"},
							Value: "dns.request",

							// TODO: MatchedTypes optimization emit map of matched types
						},
					},
				},
			},
			expected: `
package foo
allow = true {
    ` + "`foo`" + ` in input.subject.groups
    input.verb == ` + "`resolve`" + `
    re_match(` + "`dns.request`" + `, input.type)
}`,
		},
	}

	c, err := New()
	if err != nil {
		t.Fatalf("did not expect error creating backend - error: %s", err)
	}
	for idx, tst := range tests {
		actual, err := c.Compile(tst.pkg, tst.pols)
		if tst.err == nil && err != nil || tst.err != nil && err == nil {
			t.Fatalf("expected error state not returned for tst #%d tst:%s.\n  expected: %s  actual: %s",
				idx+1, tst.name, tst.err, err)
		}

		if tst.expected != actual {
			t.Fatalf("expected output not returned for tst #%d %s.\n  EXPECTED: %s\n  ACTUAL: %s\n",
				idx, tst.name, tst.expected, actual)
		}

		if len(tst.expected) > 0 {
			t.Logf("%s", tst.name)
			t.Logf("%s language output generated:\n%s", Language, tst.expected)
		}
	}
}
