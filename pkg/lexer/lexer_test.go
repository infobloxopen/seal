package lexer

import (
	"testing"

	"github.com/infobloxopen/seal/pkg/token"
)

func TestNextToken(t *testing.T) {

	input := `
	allow subject group managers to manage petstore.*;
	allow subject user cto@petstore.swagger.io to manage petstore.*;
	allow subject group customers to buy petstore.pet where ctx.tag["color"] == "purple";
	allow subject group everyone to read petstore.pet;
	=== !! << >> == != < > <= >=
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "allow"},
		{token.SUBJECT, "subject"},
		{token.GROUP, "group"},
		{token.IDENT, "managers"},
		{token.TO, "to"},
		{token.IDENT, "manage"},
		{token.TYPE_PATTERN, "petstore.*"},
		{token.DELIMETER, ";"},

		{token.IDENT, "allow"},
		{token.SUBJECT, "subject"},
		{token.USER, "user"},
		{token.IDENT, "cto@petstore.swagger.io"},
		{token.TO, "to"},
		{token.IDENT, "manage"},
		{token.TYPE_PATTERN, "petstore.*"},
		{token.DELIMETER, ";"},

		{token.IDENT, "allow"},
		{token.SUBJECT, "subject"},
		{token.GROUP, "group"},
		{token.IDENT, "customers"},
		{token.TO, "to"},
		{token.IDENT, "buy"},
		{token.TYPE_PATTERN, "petstore.pet"},
		{token.WHERE, "where"},
		{token.TYPE_PATTERN, "ctx.tag"},
		{token.LITERAL, "color"},
		{token.OP_EQUAL_TO, "=="},
		{token.LITERAL, "purple"},
		{token.DELIMETER, ";"},

		{token.IDENT, "allow"},
		{token.SUBJECT, "subject"},
		{token.GROUP, "group"},
		{token.IDENT, "everyone"},
		{token.TO, "to"},
		{token.IDENT, "read"},
		{token.TYPE_PATTERN, "petstore.pet"},
		{token.DELIMETER, ";"},

		{token.ILLEGAL, "==="},
		{token.ILLEGAL, "!!"},
		{token.ILLEGAL, "<<"},
		{token.ILLEGAL, ">>"},
		{token.OP_EQUAL_TO, "=="},
		{token.OP_NOT_EQUAL, "!="},
		{token.OP_LESS_THAN, "<"},
		{token.OP_GREATER_THAN, ">"},
		{token.OP_LESS_EQUAL, "<="},
		{token.OP_GREATER_EQUAL, ">="},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] %q - tokentype wrong. expected=%q, got %#v",
				i, tt.expectedLiteral, tt.expectedType, tok)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextTokenComment(t *testing.T) {

	type expected struct {
		typ     token.TokenType
		literal string
	}

	tests := []struct {
		name     string
		input    string
		expected []expected
	}{
		{
			name: "first comment",
			input: `
	# this is the first comment
	allow subject group managers to manage petstore.*;
	`,
			expected: []expected{
				{token.COMMENT, " this is the first comment"},

				{token.IDENT, "allow"},
				{token.SUBJECT, "subject"},
				{token.GROUP, "group"},
				{token.IDENT, "managers"},
				{token.TO, "to"},
				{token.IDENT, "manage"},
				{token.TYPE_PATTERN, "petstore.*"},
				{token.DELIMETER, ";"},
			},
		},
		{
			name: "infix comment",
			input: `
	allow subject group managers to manage petstore.*;
	# this is another comment
	allow subject group everyone to read petstore.pet;
	`,
			expected: []expected{
				{token.IDENT, "allow"},
				{token.SUBJECT, "subject"},
				{token.GROUP, "group"},
				{token.IDENT, "managers"},
				{token.TO, "to"},
				{token.IDENT, "manage"},
				{token.TYPE_PATTERN, "petstore.*"},
				{token.DELIMETER, ";"},

				{token.COMMENT, " this is another comment"},

				{token.IDENT, "allow"},
				{token.SUBJECT, "subject"},
				{token.GROUP, "group"},
				{token.IDENT, "everyone"},
				{token.TO, "to"},
				{token.IDENT, "read"},
				{token.TYPE_PATTERN, "petstore.pet"},
				{token.DELIMETER, ";"},
			},
		},
		{
			name: "end comment",
			input: `
	allow subject group everyone to read petstore.pet;
	# this is the last comment
	`,
			expected: []expected{
				{token.IDENT, "allow"},
				{token.SUBJECT, "subject"},
				{token.GROUP, "group"},
				{token.IDENT, "everyone"},
				{token.TO, "to"},
				{token.IDENT, "read"},
				{token.TYPE_PATTERN, "petstore.pet"},
				{token.DELIMETER, ";"},

				{token.COMMENT, " this is the last comment"},
			},
		},
		{
			name: "carriage returns",
			input: `
	# comment with carriage return` + "\r" + `
	allow subject group managers to manage petstore.*;` + "\n" +
				"#\r\n" +
				"allow subject group everyone to read petstore.pet;",
			expected: []expected{
				{token.COMMENT, " comment with carriage return"},

				{token.IDENT, "allow"},
				{token.SUBJECT, "subject"},
				{token.GROUP, "group"},
				{token.IDENT, "managers"},
				{token.TO, "to"},
				{token.IDENT, "manage"},
				{token.TYPE_PATTERN, "petstore.*"},
				{token.DELIMETER, ";"},

				{token.COMMENT, ""},

				{token.IDENT, "allow"},
				{token.SUBJECT, "subject"},
				{token.GROUP, "group"},
				{token.IDENT, "everyone"},
				{token.TO, "to"},
				{token.IDENT, "read"},
				{token.TYPE_PATTERN, "petstore.pet"},
				{token.DELIMETER, ";"},
			},
		},
	}

	for _, tst := range tests {
		l := New(tst.input)
		for i, tt := range tst.expected {
			tok := l.NextToken()

			if tok.Type != tt.typ {
				t.Fatalf("tests[%q][%d] %q - tokentype wrong. expected=%q, got %#v",
					tst.name, i, tt.literal, tt.typ, tok)
			}

			if tok.Literal != tt.literal {
				t.Fatalf("tests[%q][%d] - literal wrong. expected=%q, got=%q",
					tst.name, i, tt.literal, tok.Literal)
			}
		}
	}
}
