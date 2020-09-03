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
	properties    []string
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

func (t simpleType) GetProperties() map[string]types.Property {
	properties := make(map[string]types.Property)
	for _, s := range t.properties {
		properties[s] = simpleProperty(s)
	}
	return properties
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
func (s simpleAction) GetProperty(name string) (types.ActionProperty, bool) {
	// FIXME, get real property
	return nil, false
}

type simpleProperty string

func (s simpleProperty) GetName() string {
	return string(s)
}
func (s simpleProperty) String() string {
	return string(s)
}

func (s simpleProperty) GetProperty(name string) (types.SwaggerProperty, bool) {
	// FIXME, get real property
	return nil, false
}

var dnsRequestT = simpleType{
	group:         "dns",
	name:          "request",
	actions:       []string{"allow", "deny"},
	verbs:         []string{"resolve"},
	defaultAction: "deny",
	properties:    []string{"name"},
}

var ddiRangeT = simpleType{
	group:         "ddi",
	name:          "ip_range",
	actions:       []string{"allow", "deny"},
	verbs:         []string{"use", "manage"},
	defaultAction: "deny",
	properties:    []string{"id", "name"},
}

func TestLetStatements(t *testing.T) {
	input := `
allow subject group foo to resolve dns.request where ctx.name == "bar";
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
