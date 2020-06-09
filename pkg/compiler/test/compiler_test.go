package compiler_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/compiler"
	"github.com/infobloxopen/seal/pkg/compiler/error"
	"github.com/infobloxopen/seal/pkg/compiler/rego"
)

func TestCompiler(t *testing.T) {
	tests := []struct {
		name     string
		language string
		pkg      string
		pols     *ast.Policies
		expected string
		err1     error
		err2     error
	}{
		{
			name: "validate constructor requires non-empty language",
			err1: compiler_error.ErrEmptyLanguage,
		},
		{
			name:     "validate constructor fails on unregistered language",
			language: "doesnotexist",
			err1:     fmt.Errorf("invalid compiler language: doesnotexist"),
		},
		{
			name:     "validate compiler is constructed",
			language: compiler_rego.Language,
			err2:     compiler_error.ErrEmptyPolicies,
		},
	}

	for idx, tst := range tests {
		c, err := compiler.New(tst.language)
		if tst.err1 == nil && err != nil || tst.err1 != nil && err == nil {
			t.Fatalf("did not expect error creating backend for ts #%d tst:%s.\n  expected: '%s'  actual: '%s'",
				idx+1, tst.name, tst.err1, err)
		}
		if err != nil {
			continue
		}

		actual, err := c.Compile(tst.pkg, tst.pols)
		if tst.err2 == nil && err != nil || tst.err2 != nil && err == nil {
			t.Fatalf("expected error state not returned for tst #%d tst:%s.\n  expected: %s  actual: %s",
				idx+1, tst.name, tst.err2, err)
		}

		if tst.expected != actual {
			t.Fatalf("expected output not returned for tst #%d %s.\n  EXPECTED: %s\n  ACTUAL: %s\n",
				idx, tst.name, tst.expected, actual)
		}
	}
}

func TestLanguages(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "validate list of languages",
			expected: []string{
				compiler_rego.Language,
			},
		},
	}

	for idx, tst := range tests {
		actual := compiler.Languages()
		if !reflect.DeepEqual(tst.expected, actual) {
			t.Fatalf("expected output not returned for tst #%d %s.\nEXPECTED: %#v\n  ACTUAL: %#v\n",
				idx, tst.name, tst.expected, actual)
		}

		if len(tst.expected) > 0 {
			t.Logf("%s - supported languages: %#v", tst.name, tst.expected)
		}
	}
}
