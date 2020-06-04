package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT        = "IDENT"
	TYPE_PATTERN = "TYPE_PATTERN"
	DELIMETER    = ";"

	// keywords
	WITH    = "with"
	SUBJECT = "subject"
	GROUP   = "group"
	USER    = "user"
	TO      = "to"
)

var keywords = map[string]TokenType{
	"with":    WITH,
	"subject": SUBJECT,
	"user":    USER,
	"group":   GROUP,
	"to":      TO,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	if ident == DELIMETER {
		return DELIMETER
	}
	return IDENT
}
