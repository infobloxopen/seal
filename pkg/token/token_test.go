package token

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLookupOperator(t *testing.T) {
	logrus.StandardLogger().SetLevel(logrus.InfoLevel)

	testcases := []struct {
		name     string
		tok      string
		expected TokenType
	}{
		{
			name:     "invalid empty",
			expected: ILLEGAL,
		},
		{
			name:     "equal to",
			tok:      "==",
			expected: OP_EQUAL_TO,
		},
		{
			name:     "not equal",
			tok:      "!=",
			expected: OP_NOT_EQUAL,
		},
		{
			name:     "less than",
			tok:      "<",
			expected: OP_LESS_THAN,
		},
		{
			name:     "greater than",
			tok:      ">",
			expected: OP_GREATER_THAN,
		},
		{
			name:     "less than or equal to",
			tok:      "<=",
			expected: OP_LESS_EQUAL,
		},
		{
			name:     "greater than or equal to",
			tok:      ">=",
			expected: OP_GREATER_EQUAL,
		},
		{
			name:     "not",
			tok:      "not",
			expected: NOT,
		},
		{
			name:     "and",
			tok:      "and",
			expected: AND,
		},
		{
			name:     "or",
			tok:      "or",
			expected: OR,
		},
		{
			name:     "matches",
			tok:      "=~",
			expected: OP_MATCH,
		},
	}

	for _, tst := range testcases {
		t.Run(tst.name, func(t *testing.T) {
			actual := LookupOperator(tst.tok)
			if actual != tst.expected {
				t.Fatalf("actual does not match expected: actual: %s expected: %s",
					actual, tst.expected)
			}

		})
	}
}
