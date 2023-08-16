package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/sirupsen/logrus"
)

var errIgnoreAction = fmt.Errorf("ignoring action types")

const (
	TYPE_ACTION  = "action"
	TYPE_NONE    = "none"
	TYPE_DEFAULT = "type"

	SUBJECT = "subject"
)

// NewTypeFromOpenAPIv3 parses an Open API v3 spec and creates
// types for registration in seal parser.
func NewTypeFromOpenAPIv3(spec []byte) ([]Type, error) {
	logger := logrus.WithField("method", "NewTypeFromOpenAPIv3")

	swagger, err := openapi3.NewLoader().LoadFromData(spec)
	if err != nil {
		return nil, fmt.Errorf("could not load swagger yaml: %s", err)
	}
	var types []Type

	if swagger.Components == nil {
		return nil, fmt.Errorf("no schemas found")
	}
	actions, err := getActionTypes(swagger.Components.Schemas)
	if err != nil {
		return nil, err
	}

	for k, v := range swagger.Components.Schemas {
		slogger := logger.WithField("swagger_schema_name", k)
		extension, err := extractExtension(v)
		slogger.WithFields(logrus.Fields{
			"schema":    fmt.Sprintf("%+v", *v.Value),
			"extension": fmt.Sprintf("%+v", extension),
		}).Trace("schema")
		if err != nil {
			return nil, fmt.Errorf("swagger model %s has errors: %s", k, err)
		}

		switch extension.Type {
		case TYPE_DEFAULT:
		case "":
			extension.Type = TYPE_DEFAULT
		case TYPE_NONE:
			break
		default:
			slogger.WithField("extension_type", extension.Type).Trace("ignoring_extension_type")
			continue
		}

		properties, err := getPropertyTypes(v)
		if err != nil {
			return nil, err
		}

		if extension.Type != TYPE_NONE {
			if len(extension.Actions) <= 0 {
				return nil, fmt.Errorf("no actions defined")
			}
			if len(extension.Verbs) <= 0 {
				return nil, fmt.Errorf("no verbs defined")
			}
			if len(extension.DefaultAction) <= 0 {
				return nil, fmt.Errorf("no default action defined")
			}
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
			properties:    properties,
		})
	}
	if len(types) <= 0 {
		return nil, fmt.Errorf("no schemas found")
	}

	// Return sorted list of types so unit-tests can rely on deterministic order
	sort.SliceStable(types, func(i, j int) bool { return types[i].String() < types[j].String() })

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

type BaseVerbs []string

type swaggerExtension struct {
	Type          string               `json:"x-seal-type"`
	Actions       []string             `json:"x-seal-actions"`
	Verbs         map[string]BaseVerbs `json:"x-seal-verbs"`
	DefaultAction string               `json:"x-seal-default-action"`
	Properties    []string             `json:"properties"`
}

type swaggerType struct {
	group         string
	name          string
	verbs         map[string]BaseVerbs
	actions       map[string]Action
	defaultAction string
	properties    map[string]Property
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
	sorted_keys := make([]string, len(s.verbs))
	i := 0
	for k := range s.verbs {
		sorted_keys[i] = k
		i++
	}
	sort.Strings(sorted_keys)

	// Return sorted list of seal-verbs so unit-tests can rely on deterministic order
	var verbs []Verb
	for _, vb := range sorted_keys {
		bv := s.verbs[vb]
		sv := swaggerVerb{
			name:      vb,
			baseVerbs: bv,
		}
		verbs = append(verbs, sv)
	}
	return verbs
}

func (s *swaggerType) GetActions() map[string]Action {
	return s.actions
}

type swaggerVerb struct {
	name      string
	baseVerbs BaseVerbs
}

func (sv swaggerVerb) GetName() string {
	return sv.name
}

func (sv swaggerVerb) String() string {
	return fmt.Sprintf("%s: %v", sv.name, sv.baseVerbs)
}

func (sv swaggerVerb) GetBaseVerbs() BaseVerbs {
	return sv.baseVerbs
}

func (s *swaggerType) GetProperties() map[string]Property {
	return s.properties
}

type Type interface {
	GetGroup() string
	GetName() string
	GetVerbs() []Verb
	GetActions() map[string]Action
	String() string
	DefaultAction() string
	GetProperties() map[string]Property
}

type Verb interface {
	GetName() string
	String() string
	GetBaseVerbs() BaseVerbs
}

type Action interface {
	GetName() string
	String() string
	GetProperty(name string) (ActionProperty, bool)
}

type Property interface {
	GetName() string
	String() string
	GetProperty(name string) (SwaggerProperty, bool)
	HasAdditionalProperties() bool
	GetExtensionProp(name string) (string, bool, error)
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

func IsValidProperty(t Type, property string) bool {
	for _, a := range t.GetProperties() {
		if property == "ctx."+a.GetName() {
			return true
		}
	}
	return false
}

func IsValidSubject(t map[string]Type, property string) bool {
	if !strings.HasPrefix(property, SUBJECT) {
		return false
	}

	tp, ok := t["unknown."+SUBJECT]
	if !ok {
		return false
	}

	for _, pName := range tp.GetProperties() {
		if property == SUBJECT+"."+pName.GetName() {
			return true
		}
	}
	return false
}

func IsValidTag(t Type, property string) bool {
	for _, a := range t.GetProperties() {
		if strings.HasPrefix(property, "ctx."+a.GetName()+"[\"") && a.HasAdditionalProperties() {
			return true
		}
	}

	return false
}
