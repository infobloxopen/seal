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

func NewPolicyCompiler(backend string, swaggerTypes string) (*PolicyCompiler, error) {
	var err error
	cmplr := &PolicyCompiler{}

	if swaggerTypes == "" {
		return nil, errors.New("swagger is required for inferring types")
	}
	cmplr.cmplr, err = New(backend)
	if err != nil {
		return nil, fmt.Errorf("unable to create backend compiler: %s", err)
	}

	cmplr.swaggerTypes, err = types.NewTypeFromOpenAPIv3([]byte(swaggerTypes))
	if err != nil {
		return nil, fmt.Errorf("Swagger error: %s", err.Error())
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
