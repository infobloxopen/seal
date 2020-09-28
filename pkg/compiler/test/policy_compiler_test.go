package compiler_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/infobloxopen/seal/pkg/compiler"
	compiler_rego "github.com/infobloxopen/seal/pkg/compiler/rego"
	"github.com/sirupsen/logrus"
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
	logrus.StandardLogger().SetLevel(logrus.InfoLevel)

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
				neutered:
					type: boolean
				potty_trained:
					type: boolean
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
		"statement with and": {
			packageName:  "products.inventory",
			policyString: `allow subject group everyone to inspect products.inventory where ctx.id=="bar" and ctx.name=="foo";`,
			result: `
package products.inventory
default allow = false
default deny = false
allow {
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
input.id == "bar"
input.name == "foo"
}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`,
		},
		"statement-with-not": {
			packageName:  "products.inventory",
			policyString: `allow subject group everyone to inspect products.inventory where not ctx.neutered and not ctx.potty_trained;`,
			result: `
package products.inventory
default allow = false
default deny = false
allow {
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
not line1_not1_cnd
}
line1_not1_cnd {
input.neutered

not line1_not2_cnd
}
line1_not2_cnd {
input.potty_trained

}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`,
		},
		"precedence-with-not": {
			packageName:  "products.inventory",
			policyString: `allow subject group everyone to inspect products.inventory where not ctx.id == "bar" and not ctx.name == "foo";`,
			result: `
package products.inventory
default allow = false
default deny = false
allow {
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
not line1_not1_cnd
}
line1_not1_cnd {
input.id == "bar"

not line1_not2_cnd
}
line1_not2_cnd {
input.name == "foo"

}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`,
		},
		"grouping-with-parens": {
			packageName:  "products.inventory",
			policyString: `allow subject group everyone to inspect products.inventory where not (ctx.id == "bar" and ctx.name == "foo");`,
			result: `
package products.inventory
default allow = false
default deny = false
allow {
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
not line1_not1_cnd
}
line1_not1_cnd {
input.id == "bar"
input.name == "foo"

}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`,
		},
		"grouping-with-not-and-parens": {
			packageName:  "products.inventory",
			policyString: `allow subject group everyone to inspect products.inventory where not ( (not (ctx.id == "bar" and ctx.name == "foo")) and (not (ctx.neutered and ctx.potty_trained)) ));`,
			result: `
package products.inventory
default allow = false
default deny = false
allow {
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
not line1_not3_cnd
}
line1_not3_cnd {
not line1_not1_cnd
}
line1_not1_cnd {
input.id == "bar"
input.name == "foo"

not line1_not2_cnd
}
line1_not2_cnd {
input.neutered
input.potty_trained


}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
    list[_] = elem
}
`,
		},
		"multiple statements": {
			packageName: "products.inventory",
			policyString: `
				allow subject group everyone to inspect products.inventory where ctx.id=="bar";
				allow subject group everyone to inspect products.inventory where ctx.id!="bar";
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
    seal_list_contains(input.subject.groups, 'everyone')
    input.verb == 'inspect'
    re_match('products.inventory', input.type)
input.id != "bar"
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
