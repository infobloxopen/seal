package types

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

func getPropertyTypes(schema *openapi3.SchemaRef) (map[string]Property, error) {

	properties := make(map[string]Property)

	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("Schema.Value is not set")
	}

	for k, v := range schema.Value.Properties {
		pr := &swaggerProperty{
			name:                 k,
			schema:               v.Value.Properties,
			additionalProperties: false,
		}

		if v.Value.AdditionalPropertiesAllowed != nil && *v.Value.AdditionalPropertiesAllowed {
			pr.additionalProperties = true
		}

		properties[k] = pr
	}

	if len(properties) == 0 && !(*schema.Value.AdditionalPropertiesAllowed) {
		return nil, fmt.Errorf("no properties are defined")
	}
	return properties, nil
}

type SwaggerProperty interface {
	GetName() string
	String() string
}

type swaggerProperty struct {
	name                 string
	schema               map[string]*openapi3.SchemaRef
	additionalProperties bool
}

func (s *swaggerProperty) String() string {
	return s.name
}

func (s *swaggerProperty) GetName() string {
	return s.name
}

func (s *swaggerProperty) GetProperty(name string) (SwaggerProperty, bool) {
	// FIXME, get real property
	return nil, false
}

func (s *swaggerProperty) HasAdditionalProperties() bool {
	return s.additionalProperties
}
