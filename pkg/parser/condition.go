package parser

// parsing conditions as Pratt parser - Top Down Operator Precedence

import (
	"fmt"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/token"
)

type (
	prefixConditionParseFn func() ast.Condition
	infixConditionParseFn  func(ast.Condition) ast.Condition
)

func (p *Parser) registerPrefixConditionParseFns() {
	if p.prefixConditionParseFns == nil {
		p.prefixConditionParseFns = make(map[token.TokenType]prefixConditionParseFn)
	}

	p.registerPrefixCondition(token.TYPE_PATTERN, p.parseIdentifier)
	p.registerPrefixCondition(token.LITERAL, p.parseIdentifier)

	p.registerPrefixCondition(token.NOT, p.parsePrefixCondition)
	p.registerPrefixCondition(token.OPEN_PAREN, p.parseGroupedCondition)
}

func (p *Parser) registerPrefixCondition(tokenType token.TokenType, fn prefixConditionParseFn) {
	p.prefixConditionParseFns[tokenType] = fn
}

func (p *Parser) registerInfixConditionParseFns() {
	if p.infixConditionParseFns == nil {
		p.infixConditionParseFns = make(map[token.TokenType]infixConditionParseFn)
	}

	p.registerInfixCondition(token.OP_EQUAL_TO, p.parseInfixCondition)
	p.registerInfixCondition(token.OP_NOT_EQUAL, p.parseInfixCondition)
	p.registerInfixCondition(token.OP_LESS_THAN, p.parseInfixCondition)
	p.registerInfixCondition(token.OP_GREATER_THAN, p.parseInfixCondition)
	p.registerInfixCondition(token.OP_LESS_EQUAL, p.parseInfixCondition)
	p.registerInfixCondition(token.OP_GREATER_EQUAL, p.parseInfixCondition)

	p.registerInfixCondition(token.AND, p.parseInfixCondition)
	// TODO GH-42: p.registerInfixCondition(token.OR, p.parseInfixCondition)
}

func (p *Parser) registerInfixCondition(tokenType token.TokenType, fn infixConditionParseFn) {
	p.infixConditionParseFns[tokenType] = fn
}

// parsers
func (p *Parser) parseIdentifier() ast.Condition {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseLiteral() ast.Condition {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseWhereClause() *ast.WhereClause {
	wc := &ast.WhereClause{Token: p.curToken}
	p.nextToken()
	wc.Condition = p.parseCondition(PRECEDENCE_LOWEST)
	return wc
}

func (p *Parser) parseCondition(precedence int) ast.Condition {
	prefix := p.prefixConditionParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixConditionParseFnError(p.curToken.Type)
		return nil
	}
	leftCnd := prefix()
	for !p.peekTokenIs(token.DELIMETER) && precedence < p.peekPrecedence() {
		infix := p.infixConditionParseFns[p.peekToken.Type]
		if infix == nil {
			return leftCnd
		}
		p.nextToken()
		leftCnd = infix(leftCnd)
	}
	return leftCnd
}

func (p *Parser) noPrefixConditionParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix condition parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parsePrefixCondition() ast.Condition {
	condition := &ast.PrefixCondition{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	switch condition.Token.Type {
	case token.NOT:
		condition.Right = p.parseCondition(PRECEDENCE_NOT)
	default:
		condition.Right = p.parseCondition(PRECEDENCE_PREFIX)
	}
	return condition
}

func (p *Parser) parseInfixCondition(left ast.Condition) ast.Condition {
	condition := &ast.InfixCondition{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	condition.Right = p.parseCondition(precedence)
	return condition
}

func (p *Parser) parseGroupedCondition() ast.Condition {
	p.nextToken()

	cnd := p.parseCondition(PRECEDENCE_LOWEST)
	if !p.expectPeek(token.CLOSE_PAREN) {
		return nil
	}

	return cnd
}

// precedences
const (
	_ int = iota
	PRECEDENCE_LOWEST
	PRECEDENCE_OR          // logical or
	PRECEDENCE_AND         // logical and
	PRECEDENCE_NOT         // logical not
	PRECEDENCE_EQUALS      // ==
	PRECEDENCE_LESSGREATER // > or <
	PRECEDENCE_SUM         // +
	PRECEDENCE_PRODUCT     // *
	PRECEDENCE_PREFIX      // -X or !X
	PRECEDENCE_CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.OR:               PRECEDENCE_OR,
	token.AND:              PRECEDENCE_AND,
	token.NOT:              PRECEDENCE_NOT,
	token.OP_EQUAL_TO:      PRECEDENCE_EQUALS,
	token.OP_NOT_EQUAL:     PRECEDENCE_EQUALS,
	token.OP_LESS_THAN:     PRECEDENCE_LESSGREATER,
	token.OP_GREATER_THAN:  PRECEDENCE_LESSGREATER,
	token.OP_LESS_EQUAL:    PRECEDENCE_LESSGREATER,
	token.OP_GREATER_EQUAL: PRECEDENCE_LESSGREATER,
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return PRECEDENCE_LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return PRECEDENCE_LOWEST
}
