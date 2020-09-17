package compiler_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/infobloxopen/seal/pkg/compiler"
	compiler_rego "github.com/infobloxopen/seal/pkg/compiler/rego"
)

func checkError(t *testing.T, got, expected error) bool {
	if (got != nil) != (expected != nil) {
		t.Fatalf("expecting an error: '%+v', but got: '%+v'", expected, got)
	}
	if (got != nil) && (expected != nil) {
		if got.Error() != expected.Error() {
			t.Fatalf("\nexpected = '%+v'\ngot =      '%+v'", expected, got)
		} else {
			return true
		}
	}
	return false
}

func TestBackend(gt *testing.T) {
	swaggerContent := strings.ReplaceAll(`openapi: "3.0.0"
components:
	schemas:
		allow:
			type: object
			properties:
				log:
					type: boolean
			x-seal-type: action
		products.inventory:
			type: object
			x-seal-actions:
			- allow
			- deny
			x-seal-verbs:
			- inspect
			- use
			- manage
			x-seal-default-action: deny
			properties:
				id:
					type: string
				name:
					type: string
		company.personnel:
			type: object
			x-seal-actions:
			- allow
			- deny
			x-seal-verbs:
			- inspect
			- list
			- manage
			- operate
			x-seal-default-action: deny
			properties:
				id:
					type: string
`, "	", "  ")

	tCases := map[string]struct {
		compilerError  error
		swaggerError   error
		packageName    string
		policyString   string
		swaggerContent string
		result         string
	}{
		"blank swagger": {
			swaggerContent: " ",
			swaggerError:   errors.New("swagger is required for inferring types"),
		},
		"no swagger actions": {
			swaggerContent: "openapi: \"3.0.0\"\ncomponents:\n  schemas:",
			swaggerError:   errors.New("Swagger error: no actions are defined"),
		},
		"missing to errors": {
			packageName:   "products.errors",
			policyString:  `allow;`,
			compilerError: errors.New("expected next token to be to, got ; instead"),
		},
		"missing verb errors": {
			packageName:   "products.errors",
			policyString:  `allow to;`,
			compilerError: errors.New("expected next token to be IDENT, got ; instead"),
		},
		"missing resource errors": {
			packageName:   "products.errors",
			policyString:  `allow to inspect;`,
			compilerError: errors.New("expected next token to be TYPE_PATTERN, got ; instead"),
		},
		"invalid resource format without using family.type errors": {
			packageName:  "products.errors",
			policyString: `allow to inspect fake;`,
			compilerError: errors.New(
				`expected next token to be TYPE_PATTERN, got IDENT instead
expected next token to be to, got ; instead`),
		},
		"invalid resource not registered": {
			packageName:   "products.errors",
			policyString:  `allow to inspect fake.fake;`,
			compilerError: errors.New(`type pattern fake.fake did not match any registered types`),
		},
		/* TODO: subject should be optional and not required
		"simplest statement": {
			packageName: "products.inventory",
			policyString: `allow to inspect products.inventory;`,
			result: `TODO`,
		},
		*/
		"simplest statement with subject": {
			packageName:  "products.inventory",
			policyString: `allow subject group everyone to inspect products.inventory;`,
			result: `
package products.inventory
default allow = false
default deny = false
allow {
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`,
		},
		"products.inventory": {
			packageName: "products.inventory",
			policyString: `
				allow subject group everyone to inspect products.inventory where ctx.id=="bar";
				allow subject group nobody to use products.inventory;
				# WIP
				`,
			result: `
package products.inventory
default allow = false
default deny = false
allow {
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
    input.id == "bar"
}
allow {
    seal_list_contains(input.subject.groups, 'nobody')
    input.verb == 'use'
    re_match('products.inventory', input.type)
}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`,
		},
		/* TODO: where clause with and does not currently work
		   		"products.inventory.2": {
		   			packageName: "products.inventory",
		   			policyString: `
		   				allow subject group everyone to inspect products.inventory where ctx.id=="bar" and ctx.name=="foo";
		   				allow subject group nobody to use products.inventory;
		   				`,
		   			result: `
		   				package products.inventory
		   				default allow = false
		   				default deny = false
		   				allow {
		   					seal_list_contains(input.subject.groups, 'everyone')
		   					input.verb == 'inspect'
		   					re_match('products.inventory', input.type)
		   					ctx.id = "bar"
		   					ctx.name = "foo"
		   				}
		   				allow {
		   					seal_list_contains(input.subject.groups, 'nobody')
		   					input.verb == 'use'
		   					re_match('products.inventory', input.type)
		   				}

		   				# rego functions defined by seal

		   				# seal_list_contains returns true if elem exists in list
		   				seal_list_contains(list, elem) {
		   					list[_] = elem
		   				}
		   `,
		   		},
		*/
		"company.personnel": {
			packageName:  "company.personnel",
			policyString: "allow subject group manager to operate company.*;\nallow subject group users to list company.personnel;",
			result: `
				package company.personnel
				default allow = false
				default deny = false
				allow {
					seal_list_contains(input.subject.groups, 'manager')
					input.verb == 'operate'
					re_match('company.*', input.type)
				}
				allow {
					seal_list_contains(input.subject.groups, 'users')
					input.verb == 'list'
					re_match('company.personnel', input.type)
				}

				# rego functions defined by seal

				# seal_list_contains returns true if elem exists in list
				seal_list_contains(list, elem) {
					list[_] = elem
				}
`,
		},
	}

	for name, tCase := range tCases {
		tCase.result = strings.ReplaceAll(tCase.result, "'", "`")
		tCase.result = strings.ReplaceAll(tCase.result, "\t\t\t\t", "")
		tCase.result = strings.ReplaceAll(tCase.result, "\t", "    ")

		gt.Run(name, func(t *testing.T) {
			var err error
			var cmplr compiler.IPolicyCompiler

			swContent := swaggerContent
			if tCase.swaggerContent != "" {
				swContent = tCase.swaggerContent
			}
			cmplr, err = compiler.NewPolicyCompiler(compiler_rego.Language, strings.Trim(swContent, " "))
			if checkError(t, err, tCase.swaggerError) {
				return
			}

			result, err := cmplr.Compile(tCase.packageName, tCase.policyString)
			if checkError(t, err, tCase.compilerError) {
				return
			}

			if strings.Compare(result, tCase.result) != 0 {
				t.Errorf("Unexpected result\nexpected = '%+v'\ngot =         '%+v'", tCase.result, result)
			}
		})
	}
}
