package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

var errIgnoreAction = fmt.Errorf("ingoring action types")

// NewTypeFromOpenAPIv3 parses an Open API v3 spec and creates
// types for registration in seal parser.
func NewTypeFromOpenAPIv3(spec []byte) ([]Type, error) {

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("could not load swagger yaml: %s", err)
	}
	var types []Type

	for k, v := range swagger.Components.Schemas {

		extension, err := extractExtension(v)
		if err != nil {
			if err == errIgnoreAction {
				// FIXME
				continue
			}
			return nil, fmt.Errorf("model %s has errors: %s", k, err)
		}
		parts := strings.SplitN(k, ".", 2)
		name := parts[0]
		group := "unknown"
		if len(parts) > 1 {
			group = parts[0]
			name = parts[1]
		}
		types = append(types, &swaggerType{
			group:         group,
			name:          name,
			schema:        swagger.Components.Schemas[k],
			actions:       extension.Actions,
			verbs:         extension.Verbs,
			defaultAction: extension.DefaultAction,
		})
	}
	if len(types) <= 0 {
		return nil, fmt.Errorf("no schemas found")
	}
	return types, nil
}

func extractExtension(schema *openapi3.SchemaRef) (*swaggerExtension, error) {
	bs, err := json.Marshal(schema.Value.Extensions)
	if err != nil {
		return nil, err
	}
	var extensions swaggerExtension
	err = json.Unmarshal(bs, &extensions)
	if err != nil {
		return nil, err
	}
	if extensions.Type == "action" {
		// FIXME process actions types as top level items
		return nil, errIgnoreAction
	}
	if len(extensions.Actions) <= 0 {
		return nil, fmt.Errorf("no actions defined")
	}
	if len(extensions.Verbs) <= 0 {
		return nil, fmt.Errorf("no verbs defined")
	}
	if len(extensions.DefaultAction) <= 0 {
		return nil, fmt.Errorf("no default action defined")
	}

	return &extensions, nil
}

type swaggerExtension struct {
	Type          string   `json:"x-seal-type"`
	Actions       []string `json:"x-seal-actions"`
	Verbs         []string `json:"x-seal-verbs"`
	DefaultAction string   `json:"x-seal-default-action"`
}

type swaggerType struct {
	group         string
	name          string
	verbs         []string
	actions       []string
	defaultAction string
	schema        *openapi3.SchemaRef
}

func (s *swaggerType) DefaultAction() string {
	return s.defaultAction
}

func (s *swaggerType) String() string {
	return fmt.Sprintf("%s.%s", s.group, s.name)
}

func (s *swaggerType) GetName() string {
	return s.name
}

func (s *swaggerType) GetGroup() string {
	return s.group
}

func (s *swaggerType) GetVerbs() []Verb {
	var verbs []Verb
	for _, s := range s.verbs {
		verbs = append(verbs, swaggerVerb(s))
	}
	return verbs
}

func (s *swaggerType) GetActions() []Action {
	var actions []Action
	for _, s := range s.actions {
		actions = append(actions, swaggerAction(s))
	}
	return actions
}

type swaggerAction string

func (sa swaggerAction) GetName() string {
	return string(sa)
}

func (sa swaggerAction) String() string {
	return string(sa)
}

type swaggerVerb string

func (sv swaggerVerb) GetName() string {
	return string(sv)
}

func (sv swaggerVerb) String() string {
	return string(sv)
}

type Type interface {
	GetGroup() string
	GetName() string
	GetVerbs() []Verb
	GetActions() []Action
	String() string
	DefaultAction() string
}

type Verb interface {
	GetName() string
	String() string
}

type Action interface {
	GetName() string
	String() string
}

func IsValidVerb(t Type, verb string) bool {
	for _, v := range t.GetVerbs() {
		if verb == v.GetName() {
			return true
		}
	}
	return false
}

func IsValidAction(t Type, action string) bool {
	for _, a := range t.GetActions() {
		if action == a.GetName() {
			return true
		}
	}
	return false
}
