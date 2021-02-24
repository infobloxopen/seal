package token

import (
	"github.com/sirupsen/logrus"
)

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
	INT          = "INT" // 1343456
	TYPE_PATTERN = "TYPE_PATTERN"
	DELIMETER    = ";"

	// Operators
	OP_EQUAL_TO      = "=="
	OP_NOT_EQUAL     = "!="
	OP_LESS_THAN     = "<"
	OP_GREATER_THAN  = ">"
	OP_LESS_EQUAL    = "<="
	OP_GREATER_EQUAL = ">="
	OP_MATCH         = "=~"
	OP_IN            = "in"
	OPEN_PAREN       = "("
	CLOSE_PAREN      = ")"
	OPEN_BLOCK       = "{"
	CLOSE_BLOCK      = "}"

	// keywords
	WITH    = "with"
	SUBJECT = "subject"
	GROUP   = "group"
	USER    = "user"
	TO      = "to"
	WHERE   = "where"
	NOT     = "not"
	AND     = "and"
	OR      = "or"
	CONTEXT = "context"
)

var keywords = map[string]TokenType{
	"with":    WITH,
	"subject": SUBJECT,
	"user":    USER,
	"group":   GROUP,
	"to":      TO,
	"where":   WHERE,
	"not":     NOT,
	"and":     AND,
	"or":      OR,
	"context": CONTEXT,
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
	logrus.WithField("op", op).Trace("LookupOperatorComparison")
	switch op {
	default:
		return ILLEGAL

	case OP_EQUAL_TO:
	case OP_NOT_EQUAL:
	case OP_LESS_THAN:
	case OP_GREATER_THAN:
	case OP_LESS_EQUAL:
	case OP_GREATER_EQUAL:
	case OP_MATCH:
	case OP_IN:
	}

	return TokenType(op)
}

func LookupOperatorLogical(op string) TokenType {
	logrus.WithField("op", op).Trace("LookupOperatorLogical")
	switch op {
	default:
		return ILLEGAL

	case NOT:
	case AND:
	case OR:
	}

	return TokenType(op)
}

func LookupOperator(op string) TokenType {
	logrus.WithField("op", op).Trace("LookupOperator")
	type funcLookupOperator func(string) TokenType

	for _, f := range []funcLookupOperator{
		LookupOperatorComparison,
		LookupOperatorLogical,
	} {
		if typ := f(op); typ != ILLEGAL {
			return typ
		}
	}

	return ILLEGAL
}
