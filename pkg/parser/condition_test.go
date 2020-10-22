package parser

import (
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
      - inspect
      - read
      - use
      - manage
      - buy
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
      - inspect
      - read
      - use
      - manage
      - sign_in
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
