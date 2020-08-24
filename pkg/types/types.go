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

	actions, err := getActionTypes(swagger.Components.Schemas)
	if err != nil {
		return nil, err
	}

	for k, v := range swagger.Components.Schemas {

		extension, err := extractExtension(v)
		if err != nil {
			return nil, fmt.Errorf("model %s has errors: %s", k, err)
		}
		switch extension.Type {
		case "type":
			break
		case "":
			break
		default:
			continue
		}

		if len(extension.Actions) <= 0 {
			return nil, fmt.Errorf("no actions defined")
		}
		if len(extension.Verbs) <= 0 {
			return nil, fmt.Errorf("no verbs defined")
		}
		if len(extension.DefaultAction) <= 0 {
			return nil, fmt.Errorf("no default action defined")
		}

		theseActions := make(map[string]Action)
		for _, s := range extension.Actions {
			theseActions[s] = actions[s]
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
			actions:       theseActions,
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
	var extension swaggerExtension
	err = json.Unmarshal(bs, &extension)
	if err != nil {
		return nil, err
	}

	return &extension, nil
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
	actions       map[string]Action
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

func (s *swaggerType) GetActions() map[string]Action {
	return s.actions
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
	GetActions() map[string]Action
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
