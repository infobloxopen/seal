package lexer

import (
	"reflect"
	"testing"

	"github.com/infobloxopen/seal/pkg/token"
)

func TestNextToken(t *testing.T) {

	input := `
	allow subject group managers to manage petstore.*;
	allow subject user cto@petstore.swagger.io to manage petstore.*;
	allow subject group customers to buy petstore.pet where ctx.tag["color"] == "purple";
	allow subject group everyone to read petstore.pet;
        deny subject group everyone to buy petstore.pet where ctx.age < 2;
	=== !! << >> == != < > <= >= =~ ==~ =~~
		not and or
	context {} to test {where ctx.age}
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
		{token.TYPE_PATTERN, "ctx.tag[\"color\"]"},
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

		{token.IDENT, "deny"},
		{token.SUBJECT, "subject"},
		{token.GROUP, "group"},
		{token.IDENT, "everyone"},
		{token.TO, "to"},
		{token.IDENT, "buy"},
		{token.TYPE_PATTERN, "petstore.pet"},
		{token.WHERE, "where"},
		{token.TYPE_PATTERN, "ctx.age"},
		{token.OP_LESS_THAN, "<"},
		{token.INT, "2"},
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

		{token.OP_MATCH, "=~"},
		{token.ILLEGAL, "==~"},
		{token.ILLEGAL, "=~~"},

		{token.NOT, "not"},
		{token.AND, "and"},
		{token.OR, "or"},
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
func TestContextToken(t *testing.T) {

	input := `
	context {} to test {where ctx.age};
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.CONTEXT, "context"},

		{token.OPEN_BLOCK, "{"},
		{token.CLOSE_BLOCK, "}"},
		{token.TO, "to"},
		{token.IDENT, "test"},
		{token.OPEN_BLOCK, "{"},

		{token.WHERE, "where"},

		{token.TYPE_PATTERN, "ctx.age"},
		{token.CLOSE_BLOCK, "}"},
		{token.DELIMETER, ";"},
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

func TestIsIndexedIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    `table.field["key"]`,
			expected: true,
		},
		{
			input:    `table.field[key]`,
			expected: true,
		},
		{
			input:    `table.field`,
			expected: false,
		},
		{
			input:    `field["key"]`,
			expected: true,
		},
		{
			input:    `field[key]`,
			expected: true,
		},
		{
			input:    `field`,
			expected: false,
		},
	}

	for idx, tst := range tests {
		actual := IsIndexedIdentifier(tst.input)
		if tst.expected != actual {
			t.Errorf("Test#%d: failure: input=%s expected=%v actual=%v\n",
				idx, tst.input, tst.expected, actual)
		} else {
			t.Logf("Test#%d: success: actual=%v\n", idx, actual)
		}
	}
}

func TestSplitIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected *IdentifierParts
	}{
		{
			input:    `table.field["key"]`,
			expected: &IdentifierParts{
				Table: `table`,
				Field: `field`,
				Key:   `key`,
			},
		},
		{
			input:    `table.field[key]`,
			expected: &IdentifierParts{
				Table: `table`,
				Field: `field`,
				Key:   `key`,
			},
		},
		{
			input:    `table.field`,
			expected: &IdentifierParts{
				Table: `table`,
				Field: `field`,
				Key:   ``,
			},
		},
		{
			input:    `field["key"]`,
			expected: &IdentifierParts{
				Table: ``,
				Field: `field`,
				Key:   `key`,
			},
		},
		{
			input:    `field[key]`,
			expected: &IdentifierParts{
				Table: ``,
				Field: `field`,
				Key:   `key`,
			},
		},
		{
			input:    `field`,
			expected: &IdentifierParts{
				Table: ``,
				Field: `field`,
				Key:   ``,
			},
		},
	}

	for idx, tst := range tests {
		actual := SplitIdentifier(tst.input)
		if !reflect.DeepEqual(actual, tst.expected) {
			t.Errorf("Test#%d: failure: input=%s expected=%+v actual=%+v\n",
				idx, tst.input, tst.expected, actual)
		} else {
			t.Logf("Test#%d: success: actual=%+v\n", idx, actual)
		}
	}
}
