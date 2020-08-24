package types

import (
	"testing"
)

func TestNewTypeFromOpenAPIv3(t *testing.T) {

	types, err := NewTypeFromOpenAPIv3(exampleSwagger)
	if err != nil {
		t.Fatalf("could not load swagger yaml: %s", err)
	}
	if expected, actual := 1, len(types); expected != actual {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
	for _, st := range types {
		t.Logf("got type: %v", st)
		for _, ac := range st.GetActions() {
			t.Logf("got action: %#v", ac)
			if attrs, exists := ac.GetProperty(ac.GetName()); attrs != nil && exists {
			    t.Logf("  got type schema for action: %#v", attrs)
			} else {
			    t.Logf("    TODO: get type schema for action")
			}
		}

	}
}

var exampleSwagger = []byte(`
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
      - audit 
      - use
      - manage
      x-seal-default-action: deny
`)
