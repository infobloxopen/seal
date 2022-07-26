package compiler_rego

import (
	"testing"

	"github.com/infobloxopen/seal/pkg/ast"
	compiler_error "github.com/infobloxopen/seal/pkg/compiler/error"
	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"
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
							Token: token.Token{Type: "TYPE_PATTERN", Literal: "petstore.pet"},
							Value: "petstore.pet",
						},
					},
				},
			},
		},
		{
			name: "validate policy: allow subject group foo to manage petstore.pet;",
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
							Token: token.Token{Type: "IDENT", Literal: "manage"},
							Value: "manage",
						},
						TypePattern: &ast.Identifier{
							Token: token.Token{Type: "TYPE_PATTERN", Literal: "petstore.pet"},
							Value: "petstore.pet",

							// TODO: MatchedTypes optimization emit map of matched types
						},
					},
				},
			},
			expected: `
package foo

default allow = false
default deny = false

base_verbs := {
}

allow {
    seal_list_contains(seal_subject.groups, ` + "`foo`" + `)
    seal_list_contains(base_verbs[input.type][` + "`manage`" + `], input.verb)
    re_match(` + "`petstore.pet`" + `, input.type)
}

obligations := {
}` + "\n" + CompiledRegoHelpers,
		},
		{
			name: "validate policy: allow subject user foo to manage petstore.pet;",
			pkg:  "foo",
			pols: &ast.Policies{
				Statements: []ast.Statement{
					&ast.ActionStatement{
						Token: token.Token{Type: "IDENT", Literal: "allow"},
						Action: &ast.Identifier{
							Token: token.Token{Type: "IDENT", Literal: "allow"},
							Value: "allow",
						},
						Subject: &ast.SubjectUser{Token: "subject", User: "foo"},
						Verb: &ast.Identifier{
							Token: token.Token{Type: "IDENT", Literal: "manage"},
							Value: "manage",
						},
						TypePattern: &ast.Identifier{
							Token: token.Token{Type: "TYPE_PATTERN", Literal: "petstore.pet"},
							Value: "petstore.pet",

							// TODO: MatchedTypes optimization emit map of matched types
						},
					},
				},
			},
			expected: `
package foo

default allow = false
default deny = false

base_verbs := {
}

allow {
    seal_subject.sub == ` + "`foo`" + `
    seal_list_contains(base_verbs[input.type][` + "`manage`" + `], input.verb)
    re_match(` + "`petstore.pet`" + `, input.type)
}

obligations := {
}` + "\n" + CompiledRegoHelpers,
		},
		{
			name: `in-operator: allow subject user foo to manage petstore.pet where ctx.age in [1,"2"];`,
			pkg:  "foo",
			pols: &ast.Policies{
				Statements: []ast.Statement{
					&ast.ActionStatement{
						Token: token.Token{Type: token.IDENT, Literal: "allow"},
						Action: &ast.Identifier{
							Token: token.Token{Type: token.IDENT, Literal: "allow"},
							Value: "allow",
						},
						Subject: &ast.SubjectUser{Token: token.SUBJECT, User: "foo"},
						Verb: &ast.Identifier{
							Token: token.Token{Type: token.IDENT, Literal: "manage"},
							Value: "manage",
						},
						TypePattern: &ast.Identifier{
							Token: token.Token{Type: token.TYPE_PATTERN, Literal: "petstore.pet"},
							Value: "petstore.pet",
						},
						WhereClause: &ast.WhereClause{
							Token: token.Token{Type: token.WHERE, Literal: "where"},
							Condition: &ast.InfixCondition{
								Token: token.Token{Type: token.OP_IN, Literal: "in"},
								Left: &ast.Identifier{
									Token: token.Token{Type: token.IDENT, Literal: "ctx.age"},
									Value: "ctx.age",
								},
								Operator: "in",
								Right: &ast.ArrayLiteral{
									Token: token.Token{Type: token.OPEN_SQ, Literal: "["},
									Items: []ast.Condition{
										&ast.IntegerLiteral{
											Token: token.Token{Type: token.INT, Literal: "1"},
											Value: 1,
										},
										&ast.Identifier{
											Token: token.Token{Type: token.LITERAL, Literal: "2"},
											Value: "2",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: `
package foo

default allow = false
default deny = false

base_verbs := {
}

allow {
    seal_subject.sub == ` + "`foo`" + `
    seal_list_contains(base_verbs[input.type][` + "`manage`" + `], input.verb)
    re_match(` + "`petstore.pet`" + `, input.type)

    some i
    seal_list_contains([1,"2",], input.ctx[i]["age"])
}

obligations := {
}` + "\n" + CompiledRegoHelpers,
		},
	}

	c, err := New()
	if err != nil {
		t.Fatalf("did not expect error creating backend - error: %s", err)
	}
	var emptySwaggerTypes []types.Type
	for idx, tst := range tests {
		actual, err := c.Compile(tst.pkg, tst.pols, emptySwaggerTypes)
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

func TestWithInputName(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		pols     *ast.Policies
		expected string
		err      error
	}{
		{
			name: "allow subject user foo to manage petstore.pet where ctx.age == 13;",
			pkg:  "foo",
			pols: &ast.Policies{
				Statements: []ast.Statement{
					&ast.ActionStatement{
						Token: token.Token{Type: token.IDENT, Literal: "allow"},
						Action: &ast.Identifier{
							Token: token.Token{Type: token.IDENT, Literal: "allow"},
							Value: "allow",
						},
						Subject: &ast.SubjectUser{Token: token.SUBJECT, User: "foo"},
						Verb: &ast.Identifier{
							Token: token.Token{Type: token.IDENT, Literal: "manage"},
							Value: "manage",
						},
						TypePattern: &ast.Identifier{
							Token: token.Token{Type: token.TYPE_PATTERN, Literal: "petstore.pet"},
							Value: "petstore.pet",
						},
						WhereClause: &ast.WhereClause{
							Token: token.Token{Type: token.WHERE, Literal: "where"},
							Condition: &ast.InfixCondition{
								Token: token.Token{Type: token.OP_EQUAL_TO, Literal: "=="},
								Left: &ast.Identifier{
									Token: token.Token{Type: token.IDENT, Literal: "ctx.age"},
									Value: "ctx.age",
								},
								Operator: "==",
								Right: &ast.IntegerLiteral{
									Token: token.Token{Type: token.INT, Literal: "13"},
									Value: 13,
								},
							},
						},
					},
				},
			},
			expected: `
package foo

default allow = false
default deny = false

base_verbs := {
}

allow {
    seal_subject.sub == ` + "`foo`" + `
    seal_list_contains(base_verbs[abac_input.type][` + "`manage`" + `], abac_input.verb)
    re_match(` + "`petstore.pet`" + `, abac_input.type)

    some i
    abac_input.ctx[i]["age"] == 13
}

obligations := {
}` + "\n" + CompiledRegoHelpers,
		},
	}

	c, err := New()
	if err != nil {
		t.Fatalf("did not expect error creating backend - error: %s", err)
	}
	(c.(*CompilerRego)).WithInputName("abac_input")
	var emptySwaggerTypes []types.Type
	for idx, tst := range tests {
		actual, err := c.Compile(tst.pkg, tst.pols, emptySwaggerTypes)
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

func TestCleanupSomeI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:     "",
			expected:  "",
		},
		{
			input:     "\n",
			expected:  "\n",
		},
		{
			input:     SOME_I,
			expected:  "",
		},
		{
			input:     " sdkfvnj [ i ] svkjsnd ",
			expected:  " sdkfvnj [ i ] svkjsnd ",
		},
		{
			input:     "[i]",
			expected:  "some i\n[i]",
		},
		{
			input:     "[i]\n" + SOME_I,
			expected:  "some i\n[i]\n",
		},
		{
			input:     SOME_I + "\n[i]\n" + SOME_I,
			expected:  "some i\n[i]\n",
		},
	}

	for idx, tst := range tests {
		actual := cleanupSomeI(tst.input)
		if tst.expected != actual {
			t.Errorf("tst #%d: input=%#v expected=%#v actual=%#v\n",
				idx, tst.input, tst.expected, actual)
		}
	}
}
