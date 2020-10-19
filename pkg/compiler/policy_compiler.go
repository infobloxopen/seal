package compiler

import (
	"errors"
	"fmt"
	"strings"

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

	for i := len(swaggerTypes) - 1; i >= 0; i-- {
		if swaggerTypes[i] == "" {
			return nil, errors.New("swagger is required for inferring types")
		}

		compiledTypes, err := types.NewTypeFromOpenAPIv3([]byte(swaggerTypes[i]))
		if err != nil {
			return nil, fmt.Errorf("Swagger error: %s at swagger #%d", err.Error(), i)
		}

		for _, cType := range compiledTypes {
			isExists := false
			for e, eType := range cmplr.swaggerTypes {
				if cType.GetGroup() == eType.GetGroup() && cType.GetName() == eType.GetName() {
					cmplr.swaggerTypes[e] = cType
					isExists = true
				}
			}

			if !isExists {
				cmplr.swaggerTypes = append(cmplr.swaggerTypes, cType)
			}
		}
	}

	return cmplr, nil
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
	content, err := rc.cmplr.Compile(packageName, pols)
	if err != nil {
		return "", fmt.Errorf("could not compile package %s: %s\n", packageName, err)
	}

	return content, nil
}
