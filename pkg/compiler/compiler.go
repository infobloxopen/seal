package compiler

import (
	"fmt"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/compiler/error"
)

// compiler defines the interface for the language specific compiler backends
type Compiler interface {
	// Compile compiles a set of policies into a string
	Compile(pkgname string, pols *ast.Policies) (string, error)
}

// New creates a new compiler
func New(language string) (Compiler, error) {
	if len(language) <= 0 {
		return nil, compiler_error.ErrEmptyLanguage
	}

	cnst := constructor(language)
	if cnst == nil {
		return nil, fmt.Errorf("invalid compiler language: %s", language)
	}

	cmp, err := cnst()
	if err != nil {
		return nil, fmt.Errorf("unable to create compiler: %s", err)
	}

	return cmp, nil
}
