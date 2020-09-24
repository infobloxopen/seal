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

	COMMENT      = "#"
	IDENT        = "IDENT"
	TYPE_PATTERN = "TYPE_PATTERN"
	DELIMETER    = ";"

	// Operators
	OP_EQUAL_TO      = "=="
	OP_NOT_EQUAL     = "!="
	OP_LESS_THAN     = "<"
	OP_GREATER_THAN  = ">"
	OP_LESS_EQUAL    = "<="
	OP_GREATER_EQUAL = ">="
	OPEN_PAREN       = "("
	CLOSE_PAREN      = ")"

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
	return IDENT
}

func LookupOperatorComparison(op string) TokenType {
	switch op {
	default:
		return ILLEGAL

	case OP_EQUAL_TO:
	case OP_NOT_EQUAL:
	case OP_LESS_THAN:
	case OP_GREATER_THAN:
	case OP_LESS_EQUAL:
	case OP_GREATER_EQUAL:
	}

	return TokenType(op)
}

func LookupOperator(op string) TokenType {
	return LookupOperatorComparison(op)
}
