package ast

import (
	"github.com/infobloxopen/seal/pkg/token"
)

type Node interface {
	TokenLiteral() string
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

type Policy interface {
	Node
	policyNode()
}

type Conditions interface {
	Node
	conditionsNode()
	GetLiterals() []*Identifier
}

type Condition interface {
	Node
	conditionNode()
	getLiterals() []*Identifier
}

type Policies struct {
	Statements []Statement
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

func (p *Policies) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

type ActionStatement struct {
	Token       token.Token
	Action      *Identifier
	Subject     Subject
	Verb        *Identifier
	TypePattern *Identifier
	WhereClause Conditions
}

func (a *ActionStatement) TokenLiteral() string {
	return a.TokenLiteral()
}

func (a *ActionStatement) statementNode() {}

type SubjectGroup struct {
	Token token.TokenType
	Group string
}

func (s *SubjectGroup) TokenLiteral() string {
	return s.TokenLiteral()
}
func (s *SubjectGroup) subjectNode() {}

type SubjectUser struct {
	Token token.TokenType
	User  string
}

func (s *SubjectUser) TokenLiteral() string {
	return s.TokenLiteral()
}
func (s *SubjectUser) subjectNode() {}

type TypePattern struct {
	Token   token.TokenType
	Pattern string
}

func (t *TypePattern) TokenLiteral() string {
	return t.TokenLiteral()
}
func (t *TypePattern) typePatternNode() {}

// TODO: to collapse UnaryCondition and BinaryCondition into just Condition
type UnaryCondition struct {
	Token    token.Token
	LHS      *Identifier
	Operator *Identifier
	RHS      *Identifier
}

func (s *UnaryCondition) TokenLiteral() string {
	return s.TokenLiteral()
}
func (s *UnaryCondition) conditionNode() {}
func (s *UnaryCondition) getLiterals() []*Identifier {
	return []*Identifier{s.LHS}
}

type BinaryCondition struct {
	Token token.Token
	LHS   Condition
	RHS   Condition
}

func (s *BinaryCondition) TokenLiteral() string {
	return s.TokenLiteral()
}
func (s *BinaryCondition) conditionNode() {}
func (s *BinaryCondition) getLiterals() []*Identifier {
	out := []*Identifier{}
	out = append(out, s.LHS.getLiterals()...)
	out = append(out, s.RHS.getLiterals()...)
	return out
}

type WhereClause struct {
	Token      token.Token
	Conditions Condition
}

func (s *WhereClause) TokenLiteral() string {
	return s.TokenLiteral()
}
func (s *WhereClause) conditionsNode() {}
func (s *WhereClause) GetLiterals() []*Identifier {
	return s.Conditions.getLiterals()
}
