package ast

import (
	"bytes"
	"fmt"

	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"
)

// interfaces
type Node interface {
	TokenLiteral() string
	String() string
}

type Policy interface {
	Node
	policyNode()
}

type Statement interface {
	Node
	statementNode()
}

type Action interface {
	Node
	actionNode()
}

type Subject interface {
	Node
	subjectNode()
}

type Condition interface {
	Node
	conditionNode()
	GetTypes() []*Identifier
}

// concrete types
type Policies struct {
	Statements []Statement
}

func (p *Policies) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Policies) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type Identifier struct {
	Token token.Token
	Value string
}

func (slf *Identifier) conditionNode()       {}
func (slf *Identifier) TokenLiteral() string { return slf.Token.Literal }
func (slf *Identifier) String() string {
	switch slf.Token.Type {
	case token.LITERAL:
		return fmt.Sprintf(`"%s"`, slf.Token.Literal)
	}
	return slf.Token.Literal
}
func (slf *Identifier) GetTypes() []*Identifier {
	switch slf.Token.Type {
	case token.TYPE_PATTERN:
		return []*Identifier{slf}
	}
	return []*Identifier{}
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (slf *IntegerLiteral) conditionNode()       {}
func (slf *IntegerLiteral) TokenLiteral() string { return slf.Token.Literal }
func (slf *IntegerLiteral) String() string       { return slf.Token.Literal }
func (slf *IntegerLiteral) GetTypes() []*Identifier {
	return []*Identifier{}
}

type ActionStatement struct {
	Token       token.Token
	Action      *Identifier
	Subject     Subject
	Verb        *Identifier
	TypePattern *Identifier
	WhereClause Condition
}

func (a *ActionStatement) statementNode()       {}
func (a *ActionStatement) TokenLiteral() string { return a.Token.Literal }
func (a *ActionStatement) String() string {
	var out bytes.Buffer
	out.WriteString(a.TokenLiteral() + " ")
	if !types.IsNilInterface(a.Subject) {
		out.WriteString(a.Subject.String() + " ")
	}
	out.WriteString("to " + a.Verb.TokenLiteral() + " ")
	out.WriteString(a.TypePattern.TokenLiteral())
	if !types.IsNilInterface(a.WhereClause) {
		out.WriteString(" " + a.WhereClause.String())
	}
	out.WriteString(";")
	return out.String()
}

type ContextCondition struct {
	Subject Subject
	Where   *WhereClause
}
type ContextAction struct {
	Action      *Identifier
	TypePattern *Identifier
}
type ContextStatement struct {
	Token token.Token

	Contidions []*ContextCondition
	Verb       *Identifier
	Actions    []*ContextAction
}

func (a *ContextStatement) statementNode()       {}
func (a *ContextStatement) TokenLiteral() string { return a.Token.Literal }
func (a *ContextStatement) String() string {
	return "context TODO"
}

type SubjectGroup struct {
	Token token.TokenType
	Group string
}

func (s *SubjectGroup) subjectNode()         {}
func (s *SubjectGroup) TokenLiteral() string { return string(s.Token) }
func (s *SubjectGroup) String() string {
	return fmt.Sprintf("subject group %s", s.Group)
}

type SubjectUser struct {
	Token token.TokenType
	User  string
}

func (s *SubjectUser) subjectNode()         {}
func (s *SubjectUser) TokenLiteral() string { return string(s.Token) }
func (s *SubjectUser) String() string {
	return fmt.Sprintf("subject user %s", s.User)
}

type TypePattern struct {
	Token   token.TokenType
	Pattern string
}

func (t *TypePattern) typePatternNode()     {}
func (t *TypePattern) TokenLiteral() string { return string(t.Token) }
func (t *TypePattern) String() string       { return string(t.Token) }

// WhereClause defines the where clause
type WhereClause struct {
	Token     token.Token // the first token of the condition
	Condition Condition
}

func (slf *WhereClause) conditionNode()       {}
func (slf *WhereClause) TokenLiteral() string { return slf.Token.Literal }
func (slf *WhereClause) String() string {
	if !types.IsNilInterface(slf.Condition) {
		return "where " + slf.Condition.String()
	}
	return ""
}
func (slf *WhereClause) GetTypes() []*Identifier {
	return slf.Condition.GetTypes()
}

type PrefixCondition struct {
	Token    token.Token // the prefix token, e.g. `not`
	Operator string
	Right    Condition
}

func (slf *PrefixCondition) conditionNode()       {}
func (slf *PrefixCondition) TokenLiteral() string { return slf.Token.Literal }
func (slf *PrefixCondition) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(slf.Operator)
	if !types.IsNilInterface(slf.Right) {
		out.WriteString(slf.Right.String())
	}
	out.WriteString(")")
	return out.String()
}
func (slf *PrefixCondition) GetTypes() []*Identifier {
	out := []*Identifier{}
	out = append(out, slf.Right.GetTypes()...)
	return out
}

type InfixCondition struct {
	Token    token.Token // the operator token, e.g. `==`
	Left     Condition
	Operator string
	Right    Condition
}

func (slf *InfixCondition) conditionNode()       {}
func (slf *InfixCondition) TokenLiteral() string { return slf.Token.Literal }
func (slf *InfixCondition) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	if !types.IsNilInterface(slf.Left) {
		out.WriteString(slf.Left.String())
	}
	out.WriteString(" " + slf.Operator + " ")

	if !types.IsNilInterface(slf.Right) {
		out.WriteString(slf.Right.String())
	}
	out.WriteString(")")
	return out.String()
}
func (slf *InfixCondition) GetTypes() []*Identifier {
	out := []*Identifier{}
	out = append(out, slf.Left.GetTypes()...)
	out = append(out, slf.Right.GetTypes()...)
	return out
}
