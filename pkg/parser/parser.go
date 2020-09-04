package parser

import (
	"fmt"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"

	"github.com/mb0/glob"
)

type Parser struct {
	l           *lexer.Lexer
	curToken    token.Token
	peekToken   token.Token
	domainTypes map[string]types.Type
	errors      []string
}

func New(l *lexer.Lexer, domainTypes []types.Type) *Parser {
	p := &Parser{
		l:           l,
		domainTypes: make(map[string]types.Type),
		errors:      []string{},
	}
	for i, t := range domainTypes {
		s := fmt.Sprintf("%s.%s", t.GetGroup(), t.GetName())
		p.domainTypes[s] = domainTypes[i]
	}
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

// parser/parser.go
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParsePolicies() *ast.Policies {
	policies := &ast.Policies{}
	policies.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			policies.Statements = append(policies.Statements, stmt)
		}
		p.nextToken()
	}
	return policies
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.IDENT:
		return p.parseActionStatement()
	default:
		return nil
	}
}

func (p *Parser) parseSubject() ast.Subject {

	// This function parses the subject clause
	// `subject group X to`

	t := p.curToken
	p.nextToken()

	var subject ast.Subject

	switch p.curToken.Type {
	case token.GROUP:
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		subject = &ast.SubjectGroup{
			Token: t.Type,
			Group: p.curToken.Literal,
		}

	case token.USER:
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		subject = &ast.SubjectUser{
			Token: t.Type,
			User:  p.curToken.Literal,
		}
	default:
		msg := fmt.Sprintf("expected next token to be user or group, got %s instead",
			p.curToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.TO) {
		return nil
	}
	return subject
}

func (p *Parser) validateActionStatement(stmt *ast.ActionStatement) error {

	if stmt == nil {
		return nil
	}
	for s, t := range p.domainTypes {
		m, err := glob.Match(stmt.TypePattern.Value, s)
		if err != nil {
			return err
		}
		if !m {
			continue
		}

		if stmt.Verb == nil {
			return fmt.Errorf("verb must be specified for type %s", stmt.TypePattern.Value)
		}
		if v := types.IsValidVerb(t, stmt.Verb.Value); !v {
			return fmt.Errorf("verb %s is not valid for type %s", stmt.Verb, stmt.TypePattern.Value)
		}
		if v := types.IsValidAction(t, stmt.Action.Value); !v {
			return fmt.Errorf("verb %s is not valid for type %s", stmt.Action, stmt.Action.Value)
		}

		if stmt.WhereClause != nil {
			for _, l := range stmt.WhereClause.GetLiterals() {
				if v := types.IsValidProperty(t, l.Value); !v {
					return fmt.Errorf("property %s is not valid for type %s", stmt.WhereClause, l.Value)
				}
			}
		}

		return nil
	}
	return fmt.Errorf("type pattern %v did not match any registered types", stmt.TypePattern.TokenLiteral())
}

func (p *Parser) parseActionStatement() (stmt *ast.ActionStatement) {

	defer func() {
		if err := p.validateActionStatement(stmt); err != nil {
			p.errors = append(p.errors, err.Error())
			stmt = nil
		}
	}()
	stmt = &ast.ActionStatement{
		Token: p.curToken,
	}

	stmt.Action = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if p.peekToken.Type == token.SUBJECT {
		p.nextToken()
		stmt.Subject = p.parseSubject()
	}

	switch p.peekToken.Type {
	case token.IDENT:
		if !p.expectPeek(token.IDENT) {
			return nil
		}

		stmt.Verb = &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
	case token.TYPE_PATTERN:
		if !p.expectPeek(token.TYPE_PATTERN) {
			return nil
		}
		stmt.TypePattern = &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
		return stmt
	default:
		msg := fmt.Sprintf("expected next token to be user or group, got %s instead",
			p.curToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}

	if !p.expectPeek(token.TYPE_PATTERN) {
		return nil
	}
	stmt.TypePattern = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if p.peekToken.Type == token.WHERE {
		p.nextToken()
		stmt.WhereClause = p.parseWhereClause()
	}

	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) parseWhereClause() ast.Conditions {
	selector := &ast.WhereClause{
		Token: p.curToken,
	}
	selector.Conditions = p.parseCondition()

	return selector
}

func (p *Parser) parseCondition() ast.Condition {
	var curOperation ast.Condition
	for !p.isOperationCompleted() {
		switch p.peekToken.Type {
		case token.OPEN_PAREN:
			p.nextToken()
			if curOperation == nil {
				curOperation = p.parseCondition()
			} else {
				curOperation = p.parseBinaryOperation(curOperation)
			}
		default:
			if curOperation == nil {
				curOperation = p.parseUnaryOperation()
			} else {
				curOperation = p.parseBinaryOperation(curOperation)
			}
		}
		if curOperation == nil {
			return nil
		}
	}
	return curOperation
}

func (p *Parser) isOperationCompleted() bool {
	isCompleted := p.peekTokenIs(token.CLOSE_PAREN) || p.peekTokenIs(token.ILLEGAL) || p.peekTokenIs(token.DELIMETER)
	if isCompleted {
		p.nextToken()
	}
	return isCompleted
}

func (p *Parser) parseBinaryOperation(LHS ast.Condition) ast.Condition {
	p.nextToken()
	op := &ast.BinaryCondition{
		Token: p.curToken,
		LHS:   LHS,
	}
	var RHS ast.Condition
	if p.peekTokenIs(token.OPEN_PAREN) {
		p.nextToken()
		RHS = p.parseCondition()
	} else {
		RHS = p.parseUnaryOperation()
	}
	if RHS == nil {
		return nil
	}
	op.RHS = RHS

	return op
}

func (p *Parser) parseUnaryOperation() ast.Condition {
	op := &ast.UnaryCondition{}

	if !p.expectPeek(token.TYPE_PATTERN) {
		return nil
	}

	op.LHS = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if p.peekTokenIs(token.LITERAL) {
		p.nextToken()
		op.Operator = &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		}
	}

	if !p.expectPeek(token.OP_COMPARISON) {
		return nil
	}
	op.Token = p.curToken

	if !p.expectPeek(token.LITERAL) {
		return nil
	}

	op.RHS = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	return op
}
