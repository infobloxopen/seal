package lexer

import (
	"regexp"

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
	case ';':
		tok = newToken(token.DELIMETER, l.ch)
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

func isIdentifierChar(ch byte) bool {
	return isLetter(ch) || ch == '.' || ch == '*' || ch == '@'
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

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

var (
	typePatternRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\.([a-zA-Z_][a-zA-Z0-9_]*|[*]+)$`)
)
