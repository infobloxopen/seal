package compiler_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/infobloxopen/seal/pkg/compiler"
	compiler_rego "github.com/infobloxopen/seal/pkg/compiler/rego"
	"github.com/infobloxopen/seal/pkg/types"
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

	tCases := map[string]struct {
		compilerError  error
		swaggerError   error
		packageName    string
		policyString   string
		swaggerContent []string
		result         string
	}{
		"blank-swagger": {
			swaggerContent: []string{" "},
			swaggerError:   errors.New("Swagger error: no schemas found"),
		},
		"no-swagger-actions": {
			swaggerContent: []string{"openapi: \"3.0.0\"\ncomponents:\n  schemas:"},
			swaggerError:   errors.New("Swagger error: no schemas found"),
		},
		"missing-to-errors": {
			packageName:    "products.errors",
			swaggerContent: []string{"company"},
			policyString:   `allow;`,
			compilerError:  errors.New("expected next token to be to, got ; instead"),
		},
		"missing-verb-errors": {
			packageName:    "products.errors",
			swaggerContent: []string{"company"},
			policyString:   `allow to;`,
			compilerError:  errors.New("expected next token to be IDENT, got ; instead"),
		},
		"missing-resource-errors": {
			packageName:    "products.errors",
			swaggerContent: []string{"company"},
			policyString:   `allow to inspect;`,
			compilerError:  errors.New("expected next token to be TYPE_PATTERN, got ; instead"),
		},
		"invalid-resource-format-without-using-family.type-errors": {
			packageName:    "products.errors",
			swaggerContent: []string{"company"},
			policyString:   `allow to inspect fake;`,
			compilerError: errors.New(
				`expected next token to be TYPE_PATTERN, got IDENT instead
expected next token to be to, got ; instead`),
		},
		"invalid-resource-not-registered": {
			packageName:    "products.errors",
			swaggerContent: []string{"company"},
			policyString:   `allow to inspect fake.fake;`,
			compilerError:  errors.New(`type pattern fake.fake did not match any registered types`),
		},
		/* TODO: subject should be optional and not required
		"simplest statement": {
			packageName: "products.inventory",
			policyString: `allow to inspect products.inventory;`,
			result: `TODO`,
		},
		*/
		"simplest-statement-with-subject": {
			packageName:    "products.inventory",
			swaggerContent: []string{"company"},
			policyString:   `allow subject group everyone to inspect products.inventory;`,
			result: `
package products.inventory

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"statement-with-and": {
			packageName:    "products.inventory",
			swaggerContent: []string{"company"},
			policyString:   `allow subject group everyone to inspect products.inventory where ctx.id=="bar" and ctx.name=="foo";`,
			result: `
package products.inventory

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)

    some i
    input.ctx[i]["id"] == "bar"
    input.ctx[i]["name"] == "foo"
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"statement-with-not": {
			packageName:    "products.inventory",
			swaggerContent: []string{"company"},
			policyString:   `allow subject group everyone to inspect products.inventory where not ctx.neutered and not ctx.potty_trained;`,
			result: `
package products.inventory

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)
    not line1_not1_cnd
}

line1_not1_cnd {
    some i
    input.ctx[i]["neutered"]

    not line1_not2_cnd
}

line1_not2_cnd {
    some i
    input.ctx[i]["potty_trained"]
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"precedence-with-not": {
			packageName:    "products.inventory",
			swaggerContent: []string{"company"},
			policyString:   `allow subject group everyone to inspect products.inventory where not ctx.id == "bar" and not ctx.name == "foo";`,
			result: `
package products.inventory

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)
    not line1_not1_cnd
}

line1_not1_cnd {
    some i
    input.ctx[i]["id"] == "bar"

    not line1_not2_cnd
}

line1_not2_cnd {
    some i
    input.ctx[i]["name"] == "foo"
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"grouping-with-parens": {
			packageName:    "products.inventory",
			swaggerContent: []string{"company"},
			policyString:   `allow subject group everyone to inspect products.inventory where not (ctx.id == "bar" and ctx.name == "foo");`,
			result: `
package products.inventory

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)
    not line1_not1_cnd
}

line1_not1_cnd {
    some i
    input.ctx[i]["id"] == "bar"
    input.ctx[i]["name"] == "foo"
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"grouping-with-not-and-parens": {
			packageName:    "products.inventory",
			swaggerContent: []string{"company"},
			policyString:   `allow subject group everyone to inspect products.inventory where not ( (not (ctx.id == "bar" and ctx.name == "foo")) and (not (ctx.neutered and ctx.potty_trained)) ));`,
			result: `
package products.inventory

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)
    not line1_not3_cnd
}

line1_not3_cnd {
    not line1_not1_cnd
}

line1_not1_cnd {
    some i
    input.ctx[i]["id"] == "bar"
    input.ctx[i]["name"] == "foo"

    not line1_not2_cnd
}

line1_not2_cnd {
    some i
    input.ctx[i]["neutered"]
    input.ctx[i]["potty_trained"]
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"multiple-statements": {
			packageName:    "products.inventory",
			swaggerContent: []string{"company"},
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

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)

    some i
    input.ctx[i]["id"] == "bar"
}

allow {
    seal_list_contains(seal_subject.groups, 'everyone')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('products.inventory', input.type)

    some i
    input.ctx[i]["id"] != "bar"
}

allow {
    seal_list_contains(seal_subject.groups, 'nobody')
    seal_list_contains(base_verbs[input.type]['use'], input.verb)
    re_match('products.inventory', input.type)
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"company.personnel": {
			packageName:    "company.personnel",
			swaggerContent: []string{"company"},
			policyString:   "allow subject group manager to operate company.*;\nallow subject group users to inspect company.personnel;",
			result: `
package company.personnel

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
    seal_list_contains(seal_subject.groups, 'manager')
    seal_list_contains(base_verbs[input.type]['operate'], input.verb)
    re_match('company.*', input.type)
}

allow {
    seal_list_contains(seal_subject.groups, 'users')
    seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
    re_match('company.personnel', input.type)
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"tags": {
			packageName:    "petstore",
			swaggerContent: []string{"tags", "sw-with-tag"},
			policyString:   "allow subject group patissiers to manage petstore.* where ctx.tags[\"department\"] == \"bakery\"",
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "petstore.pet": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
	seal_list_contains(seal_subject.groups, 'patissiers')
	seal_list_contains(base_verbs[input.type]['manage'], input.verb)
	re_match('petstore.*', input.type)

	some i
	input.ctx[i]["tags"]["department"] == "bakery"
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"matches": {
			packageName:    "petstore",
			swaggerContent: []string{"sw1"},
			policyString:   "allow subject group patissiers to manage petstore.* where ctx.name =~ \"someValue\"",
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "petstore.pet": {
        "emptyvrb1": [
        ],
        "emptyvrb2": [
        ],
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
	seal_list_contains(seal_subject.groups, 'patissiers')
	seal_list_contains(base_verbs[input.type]['manage'], input.verb)
	re_match('petstore.*', input.type)

	some i
	re_match('someValue', input.ctx[i]["name"])
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"blank-subject": {
			packageName:    "petstore",
			swaggerContent: []string{"sw1"},
			policyString:   "allow to manage petstore.* where ctx.name =~ \"someValue\"",
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "petstore.pet": {
        "emptyvrb1": [
        ],
        "emptyvrb2": [
        ],
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
	seal_list_contains(base_verbs[input.type]['manage'], input.verb)
	re_match('petstore.*', input.type)

	some i
	re_match('someValue', input.ctx[i]["name"])
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"context": {
			packageName:    "petstore",
			swaggerContent: []string{"sw1"},
			policyString:   `context { where ctx.name=="name"; } to use { allow petstore.*; }`,
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "petstore.pet": {
        "emptyvrb1": [
        ],
        "emptyvrb2": [
        ],
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
	seal_list_contains(base_verbs[input.type]['use'], input.verb)
	re_match('petstore.*', input.type)

	some i
	input.ctx[i]["name"] == "name"
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"context-2": {
			packageName:    "petstore",
			swaggerContent: []string{"global", "company", "sw1"},
			policyString:   `context { where subject.sub=="name"; } to use { allow petstore.*; deny products.*;}`,
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "petstore.pet": {
        "emptyvrb1": [
        ],
        "emptyvrb2": [
        ],
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
	seal_list_contains(base_verbs[input.type]['use'], input.verb)
	re_match('petstore.*', input.type)
	seal_subject.sub == "name"
}

deny {
	seal_list_contains(base_verbs[input.type]['use'], input.verb)
	re_match('products.*', input.type)
	seal_subject.sub == "name"
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"context-nested": {
			packageName:    "petstore",
			swaggerContent: []string{"global", "company", "sw1"},
			policyString: `
context { 
	where subject.sub=="name"; 
} to use { 
	context {} petstore.* {allow to manage;}
	context {where subject.sub=="name2";} to inspect products.* {deny;}
}`,
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "company.personnel": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "operate": [
            "turn-on",
            "turn-off",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "petstore.pet": {
        "emptyvrb1": [
        ],
        "emptyvrb2": [
        ],
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
    "products.inventory": {
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

allow {
	seal_list_contains(base_verbs[input.type]['manage'], input.verb)
	re_match('petstore.*', input.type)
}

allow {
	seal_list_contains(base_verbs[input.type]['manage'], input.verb)
	re_match('petstore.*', input.type)
	seal_subject.sub == "name"
}

deny {
	seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
	re_match('products.*', input.type)
	seal_subject.sub == "name2"
}

deny {
	seal_list_contains(base_verbs[input.type]['inspect'], input.verb)
	re_match('products.*', input.type)
	seal_subject.sub == "name"
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"in-operator": {
			packageName:    "petstore",
			swaggerContent: []string{"global", "sw1"},
			policyString:   `deny to manage petstore.pet where "banned" in subject.sub;`,
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "petstore.pet": {
        "emptyvrb1": [
        ],
        "emptyvrb2": [
        ],
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

deny {
	seal_list_contains(base_verbs[input.type]['manage'], input.verb)
	re_match('petstore.pet', input.type)
	seal_list_contains(seal_subject.sub, 'banned')
}
` + compiler_rego.CompiledRegoHelpers,
		},
		"not-in-operator": {
			packageName:    "petstore",
			swaggerContent: []string{"global", "sw1"},
			policyString:   `deny to manage petstore.pet where not "banned" in subject.sub;`,
			result: `
package petstore

default allow = false
default deny = false

base_verbs := {
    "petstore.pet": {
        "emptyvrb1": [
        ],
        "emptyvrb2": [
        ],
        "inspect": [
            "list",
            "watch",
        ],
        "manage": [
            "create",
            "delete",
        ],
        "use": [
            "update",
            "get",
        ],
    },
}

deny {
	seal_list_contains(base_verbs[input.type]['manage'], input.verb)
	re_match('petstore.pet', input.type)
	not line1_not1_cnd
}

line1_not1_cnd {
	seal_list_contains(seal_subject.sub, 'banned')
}
` + compiler_rego.CompiledRegoHelpers,
		},
	}

	for name, tCase := range tCases {
		tCase.result = strings.ReplaceAll(tCase.result, "'", "`")
		tCase.result = strings.ReplaceAll(tCase.result, "	", "    ")

		gt.Run(name, func(t *testing.T) {
			var err error
			var cmplr compiler.IPolicyCompiler

			swContent := []string{}
			for _, swc := range tCase.swaggerContent {
				tc := swc
				if _, ok := swaggers[tc]; ok {
					tc = swaggers[tc]
				}

				swContent = append(
					swContent,
					strings.ReplaceAll(tc, "	", "  "),
				)
			}

			cmplr, err = compiler.NewPolicyCompiler(compiler_rego.Language, swContent...)
			if checkError(t, err, tCase.swaggerError) {
				return
			}

			result, err := cmplr.Compile(tCase.packageName, tCase.policyString)
			if checkError(t, err, tCase.compilerError) {
				return
			}

			lGot := strings.Split(result, "\n")
			lExp := strings.Split(tCase.result, "\n")
			if strings.Compare(result, tCase.result) != 0 {
				eString := fmt.Sprintf("Unexpected result\n    | %-50s | %-50s\n", "got", "expected")
				i := 0
				out := make(map[int]bool)
				sLen := len(lGot)
				if sLen < len(lExp) {
					sLen = len(lExp)
				}
				for ; i < sLen; i++ {
					if strings.Compare(getArrItem(lGot, i), getArrItem(lExp, i)) != 0 {
						for k := i - 1; k < i+1; k++ {
							if _, ok := out[k]; !ok {
								eString += fmt.Sprintf("%3d | %-50.50s | %-50.50s\n", k+1, getArrItem(lGot, k), getArrItem(lExp, k))
								out[k] = true
							}
						}
					}
				}
				t.Errorf(eString)
			}
		})
	}
}

func getArrItem(arr []string, i int) string {
	if i < len(arr) {
		return arr[i]
	}
	return ""
}

func TestManySwaggers(gt *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.InfoLevel)

	tCases := map[string]struct {
		swaggers     []string
		swaggerError error
		properties   map[string][]string
	}{
		"empty": {
			swaggers:     []string{},
			swaggerError: errors.New("swagger is required for inferring types"),
		},
		"global-sw1-sw2": { // sw2.petstore.pet will be overwritten by sw1, so no 'test' field expected
			swaggers: []string{"global", "sw1", "sw2"},
			properties: map[string][]string{
				"petstore.pet":    {"id", "name"},
				"unknown.subject": {"aud", "exp", "iss", "sub"},
			},
		},
		"global-sw2-sw1": {
			swaggers: []string{"global", "sw2", "sw1"},
			properties: map[string][]string{
				"petstore.pet":    {"id", "name", "test"},
				"unknown.subject": {"aud", "exp", "iss", "sub"},
			},
		},
		"sw1": {
			swaggers: []string{"sw1"},
			properties: map[string][]string{
				"petstore.pet": {"id", "name"},
			},
		},
		"sw2": {
			swaggers: []string{"sw2"},
			properties: map[string][]string{
				"petstore.pet": {"id", "name", "test"},
			},
		},
		"global": {
			swaggers: []string{"global"},
			properties: map[string][]string{
				"unknown.subject": {"aud", "exp", "iss", "sub"},
			},
		},
	}

	for name, tCase := range tCases {
		gt.Run(name, func(t *testing.T) {
			var err error
			var cmplr compiler.IPolicyCompiler

			tSwaggers := []string{}
			for _, idx := range tCase.swaggers {
				tSwaggers = append(
					tSwaggers,
					strings.ReplaceAll(swaggers[idx], "	", "  "),
				)
			}

			cmplr, err = compiler.NewPolicyCompiler(compiler_rego.Language, tSwaggers...)
			if checkError(t, err, tCase.swaggerError) {
				return
			}

			rSwTypes := reflect.ValueOf(cmplr).Elem().FieldByName("swaggerTypes")
			swTypes := reflect.NewAt(rSwTypes.Type(), unsafe.Pointer(rSwTypes.UnsafeAddr())).Elem().Interface().([]types.Type)

			for _, el := range swTypes {
				expKey := el.GetGroup() + "." + el.GetName()
				gotPList := el.GetProperties()

				expList, ok := tCase.properties[expKey]
				if !ok {
					t.Errorf("Unexpected type: %s", expKey)
					return
				}

				for _, pn := range expList {
					_, ok := gotPList[pn]
					if !ok {
						t.Errorf("Property '%s' not found", pn)
					}
				}

				for gotpn := range gotPList {
					pFound := false
					for _, exppn := range expList {
						if exppn == gotpn {
							pFound = true
							break
						}
					}
					if !pFound {
						t.Errorf("Unexpected property '%s'", gotpn)
					}
				}
			}
		})
	}
}

var swaggers = map[string]string{
	"global": `
openapi: "3.0.0"
components:
	schemas:
		subject:
			type: object
			properties:
				iss:
					type: string
				sub:
					type: string
				aud:
					type: string
				exp:
					type: integer
					format: int32
			x-seal-type: none
`,
	"sw1": `
openapi: "3.0.0"
components:
	schemas:
		allow:
			type: object
			properties:
				log:
					type: boolean
			x-seal-type: action
		petstore.pet:
			type: object
			properties:
				id: 
					type: string
				name:
					type: string
			x-seal-actions:
			- allow
			- deny
			x-seal-verbs:
                          inspect:   [ "list", "watch" ]
                          use:       [ "update", "get" ]
                          manage:    [ "create", "delete" ]
                          emptyvrb1: []
                          emptyvrb2:
			x-seal-default-action: deny 
`,
	"sw2": `
openapi: "3.0.0"
components:
	schemas:
		allow:
			type: object
			properties:
				log:
					type: boolean
			x-seal-type: action
		petstore.pet:
			type: object
			properties:
				id: 
					type: string
				name:
					type: string
				test:
					type: string
			x-seal-actions:
			- allow
			- deny
			x-seal-verbs:
                          inspect:   [ "list", "watch" ]
                          use:       [ "update", "get" ]
                          manage:    [ "create", "delete" ]
			x-seal-default-action: deny 
`,
	"tags": `
openapi: "3.0.0"
components:
  schemas:
    tag:
      type: object
      additionalProperties: true
      x-seal-type: none
`,
	"sw-with-tag": `
openapi: "3.0.0"
components:
	schemas:
		allow:
			type: object
			properties:
				log:
					type: boolean
			x-seal-type: action
		petstore.pet:
			type: object
			properties:
				id: 
					type: string
				name:
					type: string
				test:
					type: string
				tags:
					$ref: '#/components/schemas/tag'
			x-seal-actions:
			- allow
			- deny
			x-seal-verbs:
                          inspect:   [ "list", "watch" ]
                          use:       [ "update", "get" ]
                          manage:    [ "create", "delete" ]
			x-seal-default-action: deny 
`,
	"company": `
openapi: "3.0.0"
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
                          inspect:   [ "list", "watch" ]
                          use:       [ "update", "get" ]
                          manage:    [ "create", "delete" ]
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
                          inspect:   [ "list", "watch" ]
                          use:       [ "update", "get" ]
                          manage:    [ "create", "delete" ]
                          operate:   [ "turn-on", "turn-off" ]
			x-seal-default-action: deny
			properties:
				id:
					type: string
`,
}
