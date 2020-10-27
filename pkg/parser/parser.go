package parser

import (
	"fmt"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"
	"github.com/sirupsen/logrus"

	"github.com/mb0/glob"
)

type Parser struct {
	l           *lexer.Lexer
	curToken    token.Token
	peekToken   token.Token
	domainTypes map[string]types.Type
	errors      []string

	prefixConditionParseFns map[token.TokenType]prefixConditionParseFn
	infixConditionParseFns  map[token.TokenType]infixConditionParseFn
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

	p.registerPrefixConditionParseFns()
	p.registerInfixConditionParseFns()

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

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
		if !types.IsNilInterface(stmt) {
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

// parseSubject parses the subject clause `subject { group | user } X`
func (p *Parser) parseSubject() ast.Subject {
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
		msg := fmt.Sprintf("expected next token to be user or group, got %s instead", p.curToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	return subject
}

func (p *Parser) validateActionStatement(stmt *ast.ActionStatement) error {

	if types.IsNilInterface(stmt) {
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

		if !types.IsNilInterface(stmt.WhereClause) {
			typs := stmt.WhereClause.GetTypes()
			logrus.WithField("types", typs).Debug("where clause types")
			for _, l := range typs {
				v := !types.IsValidProperty(t, l.Value)                // v == true for invalid property
				v = v && !types.IsValidSubject(p.domainTypes, l.Value) // v == true for invalid subject too (mean jwt)
				v = v && !types.IsValidTag(t, l.Value)                 // v == true for invalid property + subject + tag
				if v {
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

	// action is required
	stmt.Action = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// subject is optional
	if p.peekToken.Type == token.SUBJECT {
		p.nextToken()
		stmt.Subject = p.parseSubject()
	}

	// verb is required
	if !p.expectPeek(token.TO) { //  is required
		return nil
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Verb = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// resource is required
	if !p.expectPeek(token.TYPE_PATTERN) {
		return nil
	}
	stmt.TypePattern = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// where clause is optional
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
