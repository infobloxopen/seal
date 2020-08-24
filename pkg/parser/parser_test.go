package parser

import (
	"testing"

	"github.com/infobloxopen/seal/pkg/lexer"
	"github.com/infobloxopen/seal/pkg/types"
)

type simpleType struct {
	name          string
	group         string
	actions       []string
	verbs         []string
	defaultAction string
}

func (t simpleType) DefaultAction() string {
	return t.defaultAction
}
func (t simpleType) String() string {
	return t.name
}

func (t simpleType) GetName() string {
	return t.name
}
func (t simpleType) GetGroup() string {
	return t.group
}
func (t simpleType) GetVerbs() []types.Verb {
	var verbs []types.Verb
	for _, s := range t.verbs {
		verbs = append(verbs, simpleVerb(s))
	}
	return verbs
}
func (t simpleType) GetActions() map[string]types.Action {
	actions := make(map[string]types.Action)
	for _, s := range t.actions {
		actions[s] = simpleAction(s)
	}
	return actions
}

type simpleVerb string

func (s simpleVerb) GetName() string {
	return string(s)
}
func (s simpleVerb) String() string {
	return string(s)
}

type simpleAction string

func (s simpleAction) GetName() string {
	return string(s)
}
func (s simpleAction) String() string {
	return string(s)
}

var dnsRequestT = simpleType{
	group:         "dns",
	name:          "request",
	actions:       []string{"allow", "deny"},
	verbs:         []string{"resolve"},
	defaultAction: "deny",
}

var ddiRangeT = simpleType{
	group:         "ddi",
	name:          "ip_range",
	actions:       []string{"allow", "deny"},
	verbs:         []string{"use", "manage"},
	defaultAction: "deny",
}

func TestLetStatements(t *testing.T) {
	input := `
allow subject group foo to resolve dns.request;
allow subject group bar to use ddi.*;
allow subject user foo to manage ddi.*;
`
	l := lexer.New(input)
	tTypes := []types.Type{ddiRangeT, dnsRequestT}
	p := New(l, tTypes)

	policies := p.ParsePolicies()
	checkParserErrors(t, p)
	if policies == nil {
		t.Fatalf("ParsePolicies() returned nil")
	}
	if len(policies.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(policies.Statements))
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
