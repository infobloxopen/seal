package parser

// parsing conditions as Pratt parser - Top Down Operator Precedence

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/infobloxopen/seal/pkg/ast"
	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/infobloxopen/seal/pkg/token"
	"github.com/infobloxopen/seal/pkg/types"
	"github.com/sirupsen/logrus"
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
	p.registerPrefixCondition(token.INT, p.parseIntegerLiteral)

	p.registerPrefixCondition(token.NOT, p.parsePrefixCondition)
	p.registerPrefixCondition(token.OPEN_PAREN, p.parseGroupedCondition)
	p.registerPrefixCondition(token.OPEN_SQ, p.parseArrayLiteral)
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
	p.registerInfixCondition(token.OP_MATCH, p.parseInfixCondition)
	p.registerInfixCondition(token.OP_IN, p.parseInfixCondition)

	p.registerInfixCondition(token.AND, p.parseInfixCondition)
	p.registerInfixCondition(token.OR, p.parseInfixCondition)
}

func (p *Parser) registerInfixCondition(tokenType token.TokenType, fn infixConditionParseFn) {
	p.infixConditionParseFns[tokenType] = fn
}

// parsers
func (p *Parser) parseIdentifier() ast.Condition {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Condition {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseArrayLiteral() ast.Condition {
	arrLit := &ast.ArrayLiteral{Token: p.curToken}

	p.nextToken()
	for !p.curTokenIs(token.CLOSE_SQ) {
		if p.curTokenIs(token.DELIMETER) {
			msg := fmt.Sprintf("unexpected end of array literal %q",
				p.curToken.Literal)
			p.errors = append(p.errors, msg)
		} else if p.curTokenIs(token.INT) {
			itemLit := p.parseIntegerLiteral()
			arrLit.Items = append(arrLit.Items, itemLit)
		} else if p.curTokenIs(token.LITERAL) {
			itemLit := p.parseLiteral()
			arrLit.Items = append(arrLit.Items, itemLit)
		} else {
			msg := fmt.Sprintf("unexpected %q in array literal, only integer or string literals currently supported",
				p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}

		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	return arrLit
}

func (p *Parser) parseLiteral() ast.Condition {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseWhereClause() *ast.WhereClause {
	logger := logrus.WithField("method", "parseWhereClause")
	//logger.WithField("curToken", p.curToken).Trace("begin_parse_where_clause")
	wc := &ast.WhereClause{Token: p.curToken}
	p.nextToken()
	wc.Condition = p.parseCondition(PRECEDENCE_LOWEST)
	logger.WithField("where_clause", wc.String()).Trace("parsed_where_clause")
	return wc
}

func (p *Parser) parseCondition(precedence int) ast.Condition {
	logger := logrus.WithField("method", "parseCondition")
	logger = logger.WithField("precedence", precedence)
	//logger.WithField("curToken", p.curToken).Trace("begin_parse_condition")
	prefix := p.prefixConditionParseFns[p.curToken.Type]
	if prefix == nil {
		msg := fmt.Sprintf("no prefix condition parse function for %s found", p.curToken.Type)
		p.errors = append(p.errors, msg)
		logger.WithField("error_msg", msg).Trace("parse_condition_error")
		return nil
	}
	leftCnd := prefix()
	if leftCnd != nil {
		logger.WithField("leftCnd", leftCnd.String()).Trace("parsed_prefix_condition")
	}
	logger.WithField("peekToken", p.peekToken).WithField("peekPrecedence", p.peekPrecedence()).Trace("first_infix_precedence")
	for !p.peekTokenIs(token.DELIMETER) && precedence < p.peekPrecedence() {
		infix := p.infixConditionParseFns[p.peekToken.Type]
		if infix == nil {
			break
		} else {
			p.nextToken()
			logger.WithField("peekToken", p.peekToken).WithField("peekPrecedence", p.peekPrecedence()).Trace("next_infix_precedence")
			leftCnd = infix(leftCnd)
			if leftCnd != nil {
				logger.WithField("leftCnd", leftCnd.String()).Trace("parsed_infix_condition")
			}
		}
	}
	if leftCnd != nil {
		logger.WithField("leftCnd", leftCnd.String()).Trace("parsed_final_condition")
	}
	return leftCnd
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
	logger := logrus.WithField("method", "parseInfixCondition")
	condition := &ast.InfixCondition{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	condition.Right = p.parseCondition(precedence)
	logger.WithField("condition", fmt.Sprintf("%#v", condition)).Trace("infix-condition")

	// TODO GH-42
	if condition.Token.Type == token.OR {
		msg := fmt.Sprintf("OR-operator not supported yet for condition '%s'", condition)
		p.errors = append(p.errors, msg)
		return nil
	}

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
	token.OP_MATCH:         PRECEDENCE_LESSGREATER,
	token.OP_IN:            PRECEDENCE_LESSGREATER,
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

// ParseCondition parses the input condition string
func ParseCondition(inputStr string) (ast.Condition, error) {
	logger := logrus.WithField("method", "ParseCondition")
	lxr := lexer.New(inputStr)
	emptySwaggerTypes := []types.Type{}
	prsr := New(lxr, emptySwaggerTypes)
	ast := prsr.parseCondition(PRECEDENCE_LOWEST)
	logger.WithField("condition", inputStr).WithField("ast", fmt.Sprintf("%#v", ast)).Debug("ParseCondition")
	if len(prsr.errors) > 0 {
		return nil, errors.New(strings.Join(prsr.errors, "\n"))
	}
	return ast, nil
}

// SplitKeyValueAnnotations splits and returns the optional prefix annotations map,
// and the remaining portion of the string.  Example of annotated string:
// `k1:v1, k2:v2; remaining portion of string`
func SplitKeyValueAnnotations(inputStr string) (string, map[string]string) {
	var annoMap map[string]string

	semicolonSplitted := strings.SplitN(inputStr, `;`, 2)
	remainingStr := semicolonSplitted[0]
	if len(semicolonSplitted) <= 1 {
		return remainingStr, annoMap
	}
	remainingStr = semicolonSplitted[1]

	commaSplitted := strings.Split(semicolonSplitted[0], `,`)
	if len(commaSplitted) <= 0 {
		return remainingStr, annoMap
	}

	for _, kvPair := range commaSplitted {
		kvPair = strings.TrimSpace(kvPair)
		if len(kvPair) <= 0 {
			continue
		}

		colonSplitted := strings.SplitN(kvPair, `:`, 2)
		key := strings.TrimSpace(colonSplitted[0])
		if len(key) <= 0 {
			continue
		}

		var val string
		if len(colonSplitted) > 1 {
			val = strings.TrimSpace(colonSplitted[1])
		}

		if annoMap == nil {
			annoMap = map[string]string{}
		}
		annoMap[key] = val
	}

	return remainingStr, annoMap
}
