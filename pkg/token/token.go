package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	LITERAL = "LITERAL"

	IDENT         = "IDENT"
	TYPE_PATTERN  = "TYPE_PATTERN"
	DELIMETER     = ";"
	OP_COMPARISON = "=="
	OPEN_PAREN    = "("
	CLOSE_PAREN   = ")"

	// keywords
	WITH    = "with"
	SUBJECT = "subject"
	GROUP   = "group"
	USER    = "user"
	TO      = "to"
	WHERE   = "where"
	AND     = "and"
	OR      = "or"
)

var keywords = map[string]TokenType{
	"with":    WITH,
	"subject": SUBJECT,
	"user":    USER,
	"group":   GROUP,
	"to":      TO,
	"where":   WHERE,
	"and":     AND,
	"or":      OR,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	if ident == DELIMETER {
		return DELIMETER
	}
	if ident == OP_COMPARISON {
		return OP_COMPARISON
	}
	return IDENT
}
