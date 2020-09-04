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
			t.Fatalf("\nexpected =\t'%+v'\ngot =\t'%+v'", expected, got)
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
		"policies errors": {
			packageName: "products.errors",
			policyString: `
				allow subject group everyone to inspect fake;
				`,
			compilerError: errors.New("expected next token to be TYPE_PATTERN, got IDENT instead\nexpected next token to be user or group, got IDENT instead"),
		},
		"products.inventory": {
			packageName: "products.inventory",
			policyString: `
				allow subject group everyone to inspect products.inventory where ctx.id=="bar";
				allow subject group nobody to use products.inventory;
				`,
			result: `
				package products.inventory
				allow = true {
					'everyone' in input.subject.groups
					input.verb == 'inspect'
					re_match('products.inventory', input.type)
				} where ctx.id = "bar"
				allow = true {
					'nobody' in input.subject.groups
					input.verb == 'use'
					re_match('products.inventory', input.type)
				}`,
		},
		"products.inventory.2": {
			packageName: "products.inventory",
			policyString: `
				allow subject group everyone to inspect products.inventory where ctx.id=="bar" and (ctx.name=="foo" or ctx.name=="bar2");
				allow subject group nobody to use products.inventory;
				`,
			result: `
				package products.inventory
				allow = true {
					'everyone' in input.subject.groups
					input.verb == 'inspect'
					re_match('products.inventory', input.type)
				} where {
					ctx.id = "bar" and {
					ctx.name = "foo" or ctx.name = "bar2"
				}
				}
				allow = true {
					'nobody' in input.subject.groups
					input.verb == 'use'
					re_match('products.inventory', input.type)
				}`,
		},
		"company.personnel": {
			packageName:  "company.personnel",
			policyString: "allow subject group manager to operate company.*;\nallow subject group users to list company.personnel;",
			result: `
				package company.personnel
				allow = true {
					'manager' in input.subject.groups
					input.verb == 'operate'
					re_match('company..*', input.type)
				}
				allow = true {
					'users' in input.subject.groups
					input.verb == 'list'
					re_match('company.personnel', input.type)
				}`,
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
			if checkError(t, err, tCase.swaggerError) == true {
				return
			}

			result, err := cmplr.Compile(tCase.packageName, tCase.policyString)
			if checkError(t, err, tCase.compilerError) == true {
				return
			}

			if strings.Compare(result, tCase.result) != 0 {
				t.Errorf("Unexpected result\nexpected =\t'%+v'\ngot =\t'%+v'", tCase.result, result)
			}
		})
	}
}
