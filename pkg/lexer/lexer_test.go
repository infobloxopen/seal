package lexer

import (
	"testing"

	"github.com/infobloxopen/seal/pkg/token"
)

func TestNextToken(t *testing.T) {

	input := `
	allow subject group foo to manage ddi.*;
	allow subject user cto@acme.com to manage products.inventory;
	allow subject user cto@acme.com to manage products.inventory where ctx.tag["foo"] == "bar";
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

		{token.IDENT, "allow"},
		{token.SUBJECT, "subject"},
		{token.USER, "user"},
		{token.IDENT, "cto@acme.com"},
		{token.TO, "to"},
		{token.IDENT, "manage"},
		{token.TYPE_PATTERN, "products.inventory"},
		{token.DELIMETER, ";"},

		{token.IDENT, "allow"},
		{token.SUBJECT, "subject"},
		{token.USER, "user"},
		{token.IDENT, "cto@acme.com"},
		{token.TO, "to"},
		{token.IDENT, "manage"},
		{token.TYPE_PATTERN, "products.inventory"},
		{token.WHERE, "where"},
		{token.TYPE_PATTERN, "ctx.tag"},
		{token.LITERAL, "foo"},
		{token.OP_COMPARISON, "=="},
		{token.LITERAL, "bar"},
		{token.DELIMETER, ";"},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] %q - tokentype wrong. expected=%q, got %q",
				i, tt.expectedLiteral, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
