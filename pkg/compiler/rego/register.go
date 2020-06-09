package compiler_rego

import (
	"github.com/infobloxopen/seal/pkg/compiler"
)

// const...
const (
	Language = "rego"
)

func init() {
	compiler.Register(Language, New)
}
