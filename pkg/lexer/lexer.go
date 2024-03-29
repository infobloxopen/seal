package lexer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/infobloxopen/seal/pkg/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '#':
		tok = newToken(token.DELIMETER, l.ch)
		tok.Literal = l.readComment()
		tok.Type = token.COMMENT
		return tok
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case ';':
		tok = newToken(token.DELIMETER, l.ch)
	case '(':
		tok = newToken(token.OPEN_PAREN, l.ch)
	case ')':
		tok = newToken(token.CLOSE_PAREN, l.ch)
	case '{':
		tok = newToken(token.OPEN_BLOCK, l.ch)
	case '}':
		tok = newToken(token.CLOSE_BLOCK, l.ch)
	case '[':
		tok = newToken(token.OPEN_SQ, l.ch)
	case ']':
		tok = newToken(token.CLOSE_SQ, l.ch)
	case '"':
		tok.Literal = l.readLiteral()
		tok.Type = token.LITERAL
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			if isTypePattern(tok.Literal) {
				tok.Type = token.TYPE_PATTERN
			}
/* TODO REMOVE ME NOT NEEDED?
			if isLongOperator(tok.Literal) {
				tok.Type = token.LookupOperator(tok.Literal)
			}
*/
			return tok
		}
		if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		}
		if isOperator(l.ch) {
			tok.Literal = l.readOperator()
			tok.Type = token.LookupOperator(tok.Literal)
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ch)
	}
	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isIdentifierChar(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readLiteral() string {
	l.readChar()
	start := l.position
	for isLiteralChar(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readComment() string {
	l.readChar()
	start := l.position
	for !isNewline(l.ch) && l.ch != 0 {
		l.readChar()
	}
	end := l.position
	if '\r' == l.input[end-1] {
		end -= 1
	}
	return l.input[start:end]
}

func isIdentifierChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '.' || ch == '*' || ch == '@' || ch == '[' || ch == ']' || ch == '"'
}

// IndexedIdentifierChars are chars any indexed-identifier would have
const IndexedIdentifierChars = `["]`

// IsIndexedIdentifier returns true if id is:
//   table.field["key"]
//   table.field[key]
//   field["key"]
//   field[key]
// IsIndexedIdentifier returns false if id is:
//   table.field
//   field
func IsIndexedIdentifier(id string) bool {
	return strings.ContainsAny(id, IndexedIdentifierChars)
}

// IdentifierParts holds components of splitted identifiers:
// Examples of unsplitted identitiers:
//   table.field["key"]
//   table.field[key]
//   table.field
//   field["key"]
//   field[key]
//   field
type IdentifierParts struct {
	Table string // component before dot (empty if no dot)
	Field string // component after dot
	Key   string // index key (empty if no key)
}

// SplitIdentifier splits id into IdentifierParts
func SplitIdentifier(id string) *IdentifierParts {
	idParts := IdentifierParts{}
	splitID := strings.SplitN(id, `.`, 2)

	idParts.Field = splitID[0]
	if len(splitID) > 1 {
		idParts.Table = splitID[0]
		idParts.Field = splitID[1]
	}

	keyIdx := strings.IndexAny(idParts.Field, IndexedIdentifierChars)
	if keyIdx > 0 {
		fieldAndKey := idParts.Field
		idParts.Field = fieldAndKey[:keyIdx]
		idParts.Key = fieldAndKey[keyIdx:]
		idParts.Key = strings.TrimLeft(idParts.Key, IndexedIdentifierChars)
		idParts.Key = strings.TrimRight(idParts.Key, IndexedIdentifierChars)
	}

	return &idParts
}

// SwaggerTypeParts holds components of splitted swagger-types:
// Examples of unsplitted swagger-types:
//   app.type
//   type
type SwaggerTypeParts struct {
	App  string // component before dot (empty if no dot)
	Type string // component after dot
}

// SplitSwaggerType splits swagger-type into SwaggerTypeParts
func SplitSwaggerType(swtype string) *SwaggerTypeParts {
	swParts := SwaggerTypeParts{}
	splitted := strings.SplitN(swtype, `.`, 2)

	swParts.Type = splitted[0]
	if len(splitted) > 1 {
		swParts.App = splitted[0]
		swParts.Type = splitted[1]
	}

	return &swParts
}

// String implements fmt.Stringer interface
func (swParts SwaggerTypeParts) String() string {
	if len(swParts.App) <= 0 {
		return swParts.Type
	}
	return fmt.Sprintf("%s.%s", swParts.App, swParts.Type)
}

func isLiteralChar(ch byte) bool {
	return ch != '"'
}

func isOperator(ch byte) bool {
	return ch == '=' || ch == '!' || ch == '<' || ch == '>' || ch == '~'
}

/* TODO REMOVE ME NOT NEEDED?
func isLongOperator(s string) bool {
	switch s {
	case token.OP_IN:
		return true
	}

	return false
}
*/

func (l *Lexer) readOperator() string {
	start := l.position
	for isOperator(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func isTypePattern(s string) bool {
	if !isLetter(s[0]) {
		return false
	}
	return typePatternRegex.MatchString(s)

}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isNewline(ch byte) bool {
	return '\n' == ch
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

var (
	typePatternRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\.([a-zA-Z_][a-zA-Z0-9_]*|[*]+)?(\[\"[a-zA-Z0-9_]*\"\])?$`)
)
