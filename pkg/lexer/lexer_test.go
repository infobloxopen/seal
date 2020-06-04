package lexer

import (
	"testing"

	"github.com/infobloxopen/seal/pkg/token"
)

func TestNextToken(t *testing.T) {

	input := `
	allow subject group foo to manage ddi.*;
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "allow"},
		{token.SUBJECT, "subject"},
		{token.GROUP, "group"},
		{token.IDENT, "foo"},
		{token.TO, "to"},
		{token.IDENT, "manage"},
		{token.TYPE_PATTERN, "ddi.*"},
		{token.DELIMETER, ";"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got %q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
