package compiler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/infobloxopen/seal/pkg/parser"
	"github.com/infobloxopen/seal/pkg/types"
)

type IPolicyCompiler interface {
	Compile(packageName string, policyString string) (string, error)
}

var _ IPolicyCompiler = &PolicyCompiler{}

type PolicyCompiler struct {
	cmplr        Compiler
	swaggerTypes []types.Type
}

func NewPolicyCompiler(backend string, swaggerTypes ...string) (*PolicyCompiler, error) {
	var err error
	cmplr := &PolicyCompiler{}

	if len(swaggerTypes) == 0 {
		return nil, errors.New("swagger is required for inferring types")
	}
	cmplr.cmplr, err = New(backend)
	if err != nil {
		return nil, fmt.Errorf("unable to create backend compiler: %s", err)
	}

	mergedSwagger, err := cmplr.mergeSwaggers(swaggerTypes...)
	if err != nil {
		return nil, err
	}

	cmplr.swaggerTypes, err = types.NewTypeFromOpenAPIv3([]byte(mergedSwagger))
	if err != nil {
		return nil, fmt.Errorf("Swagger error: %s", err.Error())
	}

	return cmplr, nil
}

func (rc *PolicyCompiler) mergeSwaggers(swaggerTypes ...string) (string, error) {
	var rSw *openapi3.Swagger

	for i := len(swaggerTypes) - 1; i >= 0; i-- {
		psw := &openapi3.Swagger{}
		if err := yaml.Unmarshal([]byte(swaggerTypes[i]), psw); err != nil {
			return "", fmt.Errorf("Swagger yaml unmarshal error: %s:\n%s", err.Error(), swaggerTypes[i])
		}
		if rSw == nil {
			rSw = psw
			continue
		}

		for name, schema := range psw.Components.Schemas {
			rSw.Components.Schemas[name] = schema
		}
	}

	str, err := yaml.Marshal(rSw)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func (rc *PolicyCompiler) Compile(packageName string, policyString string) (string, error) {
	l := lexer.New(policyString)
	p := parser.New(l, rc.swaggerTypes)
	pols := p.ParsePolicies()
	polErrors := p.Errors()
	if n := len(polErrors); n > 0 {
		return "", errors.New(strings.Join(polErrors, "\n"))
	}
	if pols == nil {
		return "", fmt.Errorf("unable to find any policies in package %s", packageName)
	}

	// compile policies from AST
	content, err := rc.cmplr.Compile(packageName, pols, rc.swaggerTypes)
	if err != nil {
		return "", fmt.Errorf("could not compile package %s: %s\n", packageName, err)
	}

	return content, nil
}
