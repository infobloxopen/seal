package types

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func getActionTypes(schemas map[string]*openapi3.SchemaRef) (map[string]Action, error) {

	actions := make(map[string]Action)

	// look for explicit action declarations
	for k, v := range schemas {

		extension, err := extractExtension(v)
		if err != nil {
			return nil, fmt.Errorf("model %s has errors: %s", k, err)
		}
		if extension.Type != TYPE_ACTION {
			continue
		}
		actions[k] = &swaggerAction{
			name:   k,
			schema: schemas[k],
		}
	}

	// look for implicit action declarations
	for _, v := range schemas {

		// no need to check error second time
		extension, _ := extractExtension(v)
		if extension.Type == TYPE_ACTION {
			continue
		}
		for _, s := range extension.Actions {
			_, ok := actions[s]
			if ok {
				continue // we already have this action
			}
			// we create a dummy action
			actions[s] = &swaggerAction{
				name: s,
			}
		}
	}

	// if len(actions) == 0 {
	// 	return nil, fmt.Errorf("no actions are defined")
	// }
	return actions, nil
}

type swaggerAction struct {
	name   string
	schema *openapi3.SchemaRef
}

func (s *swaggerAction) String() string {
	return s.name
}

func (s *swaggerAction) GetName() string {
	return s.name
}

func (s *swaggerAction) GetProperty(name string) (ActionProperty, bool) {
	// FIXME, get real property
	return nil, false
}

type ActionProperty interface {
	GetName() string
	String() string
}

type actionProperty struct {
	name   string
	schema *openapi3.SchemaRef
}

func (aa *actionProperty) GetName() string {
	return aa.name
}

func (aa *actionProperty) String() string {
	return aa.name
}
