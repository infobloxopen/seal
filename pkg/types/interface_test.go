package types

import (
	"errors"
	"testing"
)

func TestIsNilInterface(t *testing.T) {
	var nilError error

	tests := []struct {
		title    string
		ptr      interface{}
		expected bool
	}{
		{
			title:    "nil constant should be nil interface",
			ptr:      nil,
			expected: true,
		},
		{
			title:    "nil error variable should be nil interface",
			ptr:      nilError,
			expected: true,
		},
		{
			title:    "typed nil variable should be nil interface",
			ptr:      error(nil),
			expected: true,
		},
		{
			title:    "constructed error variable should not be nil interface",
			ptr:      errors.New("hello"),
			expected: false,
		},
	}

	for _, test := range tests {
		actual := IsNilInterface(test.ptr)
		if actual != test.expected {
			t.Fatalf("%s: actual: %v  does not match  expected: %v",
				test.title, actual, test.expected)
		}
	}
}
