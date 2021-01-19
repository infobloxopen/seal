package types

import (
	"fmt"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
)

func getPropertyTypes(schema *openapi3.SchemaRef) (map[string]Property, error) {

	properties := make(map[string]Property)

	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("Schema.Value is not set")
	}

	for k, v := range schema.Value.Properties {
		pr := &swaggerProperty{
			name:                        k,
			schema:                      v.Value,
			additionalPropertiesAllowed: false,
		}

		if v.Value.AdditionalPropertiesAllowed != nil && *v.Value.AdditionalPropertiesAllowed {
			pr.additionalPropertiesAllowed = true
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
	name                        string
	schema                      *openapi3.Schema
	additionalPropertiesAllowed bool
}

func (s *swaggerProperty) String() string {
	return s.name + fmt.Sprintf("-property-schema:%+v", s.schema)
}

func (s *swaggerProperty) GetName() string {
	return s.name
}

func (s *swaggerProperty) GetProperty(name string) (SwaggerProperty, bool) {
	// FIXME, get real property
	return nil, false
}

func (s *swaggerProperty) HasAdditionalProperties() bool {
	return s.additionalPropertiesAllowed
}

func (s *swaggerProperty) GetExtensionProp(name string) (string, bool, error) {
	untypedExtValue, ok := s.schema.Extensions[name]
	if !ok {
		return "", false, nil
	}

	// this will panic: interface conversion: interface {} is json.RawMessage, not string
	//return untypedExtValue.(string), true

	// untypedExtValue is of type json.RawMessage
	marshaledBytes, err := json.Marshal(untypedExtValue)
	if err != nil {
		return "", false, fmt.Errorf("cannot json.marshal the value of extension property '%s': %s", err)
	}
	return string(marshaledBytes), true, nil
}
