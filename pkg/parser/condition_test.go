package parser

import (
	"reflect"
	"testing"

	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/infobloxopen/seal/pkg/types"
	"github.com/sirupsen/logrus"
)

func TestWhereClause(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.InfoLevel)

	typesContent := `
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
      x-seal-actions:
      - allow
      - deny
      x-seal-verbs:
        inspect:
        read:
        use:
        manage:
        buy:
      x-seal-default-action: deny
      properties:
        id:
          type: string
        name:
          type: string
        status:
          type: string
        age:    # months
          type: integer
          format: int32
        is_healthy:
          type: bool
    iam.user:
      type: object
      x-seal-actions:
      - allow
      - deny
      x-seal-verbs:
        inspect:
        read:
        use:
        manage:
        sign_in:
      x-seal-default-action: deny
      properties:
        id:
          type: string
        email:
          type: string
`

	testcases := []struct {
		name     string
		rules    string
		expected string
	}{
		{

			name:     "simple user",
			rules:    `allow subject user cto@acme.com to manage petstore.pet;`,
			expected: `allow subject user cto@acme.com to manage petstore.pet;`,
		},
		{

			name:     "simple group",
			rules:    `allow subject group managers to manage iam.*;`,
			expected: `allow subject group managers to manage iam.*;`,
		},
		{
			name:     "simple where clause compare equal",
			rules:    `allow subject group customers to buy petstore.pet where ctx.status == "available";`,
			expected: `allow subject group customers to buy petstore.pet where (ctx.status == "available");`,
		},
		{
			name:     "simple where clause compare not equal",
			rules:    `allow subject group customers to buy petstore.pet where ctx.status != "available";`,
			expected: `allow subject group customers to buy petstore.pet where (ctx.status != "available");`,
		},
		{
			name:     "simple where clause compare int",
			rules:    `allow subject group customers to buy petstore.pet where ctx.age > 2;`,
			expected: `allow subject group customers to buy petstore.pet where (ctx.age > 2);`,
		},
		{
			name:     "simple where clause compare bool", // TODO: bool needs to be bare word in OPA
			rules:    `allow subject group customers to buy petstore.pet where ctx.is_healthy == "true";`,
			expected: `allow subject group customers to buy petstore.pet where (ctx.is_healthy == "true");`,
		},
		{
			name:     "single where clause and",
			rules:    `allow subject group customers to buy petstore.pet where ctx.status == "available" and ctx.is_healthy == "true";`,
			expected: `allow subject group customers to buy petstore.pet where ((ctx.status == "available") and (ctx.is_healthy == "true"));`,
		},
		{
			name:     "left associative where clause and",
			rules:    `allow subject group customers to buy petstore.pet where ctx.status == "available" and ctx.is_healthy == "true" and ctx.name == "fido";`,
			expected: `allow subject group customers to buy petstore.pet where (((ctx.status == "available") and (ctx.is_healthy == "true")) and (ctx.name == "fido"));`,
		},
		{
			name:     "where clause grouped conditions",
			rules:    `allow subject group customers to buy petstore.pet where not (ctx.status == "available" and ctx.is_healthy == "true");`,
			expected: `allow subject group customers to buy petstore.pet where (not((ctx.status == "available") and (ctx.is_healthy == "true")));`,
		},
		{
			name:     "where clause multiple grouped conditions",
			rules:    `allow subject group customers to buy petstore.pet where not ( (not (ctx.status == "available" and ctx.is_healthy == "true")) and (not (ctx.id == "foo" and ctx.name == "bar")) );`,
			expected: `allow subject group customers to buy petstore.pet where (not((not((ctx.status == "available") and (ctx.is_healthy == "true"))) and (not((ctx.id == "foo") and (ctx.name == "bar")))));`,
		},
		{
			name:     "in-operator array literal",
			rules:    `allow to manage petstore.pet where ctx.status in [ "available", 2 ]`,
			expected: `allow to manage petstore.pet where (ctx.status in ["available",2,]);`,
		},
	}

	typs, err := types.NewTypeFromOpenAPIv3([]byte(typesContent))
	if err != nil {
		t.Fatalf("Swagger types error: %s", err)
	}

	for _, tst := range testcases {
		t.Run(tst.name, func(t *testing.T) {
			lxr := lexer.New(tst.rules)
			psr := New(lxr, typs)

			policies := psr.ParsePolicies()
			checkParserErrors(t, psr)
			if policies == nil {
				t.Fatalf("ParsePolicies() returned nil")
			}

			if policies.String() != tst.expected {
				t.Fatalf("actual does not match expected for: %s\nactual:   %s\nexpected: %s",
					tst.rules, policies, tst.expected)
			}

		})
	}
}

func TestExportedParseCondition(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.InfoLevel)

	testcases := []struct {
		name      string
		condition string
		expected  string
		shouldErr bool
	}{
		{

			name:      "no parens",
			condition: `ctx.age > 65`,
			expected:  `(ctx.age > 65)`,
			shouldErr: false,
		},
		{

			name:      "single parens",
			condition: `(ctx.age > 65)`,
			expected:  `(ctx.age > 65)`,
			shouldErr: false,
		},
		{

			name:      "double parens",
			condition: `((ctx.age > 65))`,
			expected:  `(ctx.age > 65)`,
			shouldErr: false,
		},
		{

			name:      "mismatched parens",
			condition: `(((ctx.age > 65))`,
			expected:  ``,
			shouldErr: true,
		},
	}

	for _, tst := range testcases {
		t.Run(tst.name, func(t *testing.T) {
			astCond, err := ParseCondition(tst.condition)
			if (err == nil) && tst.shouldErr {
				t.Errorf("ParseCondition(`%s`) expected error, but no error returned",
					tst.condition)
			} else if (err != nil) {
				if tst.shouldErr {
					t.Logf("ParseCondition(`%s`) expected error, and error returned: %q",
						tst.condition, err)
				} else {
					t.Errorf("ParseCondition(`%s`) expected no error, but error returned: %q",
						tst.condition, err)
				}
			}

			if astCond == nil {
				if !tst.shouldErr {
					t.Errorf("ParseCondition(`%s`) unexpectedly returned nil", tst.condition)
				}
			} else if astCond.String() != tst.expected {
				t.Errorf("ParseCondition(`%s`) got `%s` expected `%s` %#v\n", tst.condition, astCond.String(), tst.expected, astCond)
			}
		})
	}
}

func TestSplitKeyValueAnnotations(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.InfoLevel)

	testcases := []struct {
		inputStr  string
		remaining string
		annoMap   map[string]string
	}{
		{

			inputStr:  ` a b c d `,
			remaining: ` a b c d `,
			annoMap:   nil,
		},
		{

			inputStr:  ` ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   nil,
		},
		{

			inputStr:  ` , ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   nil,
		},
		{

			inputStr:  `k1; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k1`: ``,
			},
		},
		{

			inputStr:  `k1: v1 ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k1`: `v1`,
			},
		},
		{

			inputStr:  `k1: v1 , ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k1`: `v1`,
			},
		},
		{

			inputStr:  ` , k1: v1 , ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k1`: `v1`,
			},
		},
		{

			inputStr:  `k1: v1 , k2 ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k1`: `v1`,
				`k2`: ``,
			},
		},
		{

			inputStr:  `k1: v1 , k2 : ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k1`: `v1`,
				`k2`: ``,
			},
		},
		{

			inputStr:  `k1: v1 , k2 : v2 ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k1`: `v1`,
				`k2`: `v2`,
			},
		},
		{

			inputStr:  `k 1: v 1 , k 2 : v 2 ; a b c d `,
			remaining: ` a b c d `,
			annoMap:   map[string]string{
				`k 1`: `v 1`,
				`k 2`: `v 2`,
			},
		},
	}

	for idx, tst := range testcases {
		remainingActual, annoMapActual := SplitKeyValueAnnotations(tst.inputStr)

		if remainingActual != tst.remaining {
			t.Errorf("Test#%d: FAIL (remain): input=`%s` expected=`%s` actual=`%s`\n",
				idx, tst.inputStr, tst.remaining, remainingActual)
		} else {
			t.Logf("Test#%d: pass (remain): input=`%s` remaining=`%s`\n",
				idx, tst.inputStr, remainingActual)
		}

		if !reflect.DeepEqual(annoMapActual, tst.annoMap) {
			t.Errorf("Test#%d: FAIL (annMap): input=`%s` expected=`%s` actual=`%s`\n",
				idx, tst.inputStr, tst.annoMap, annoMapActual)
		} else {
			t.Logf("Test#%d: pass (annMap): input=`%s` annoMap=`%s`\n",
				idx, tst.inputStr, annoMapActual)
		}
	}
}
