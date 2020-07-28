package cmd

import (
	"github.com/infobloxopen/seal/pkg/types"
)

type exampleType struct {
	name          string
	group         string
	actions       []string
	verbs         []string
	defaultAction string
}

func (t exampleType) DefaultAction() string {
	return t.defaultAction
}
func (t exampleType) String() string {
	return t.name
}

func (t exampleType) GetName() string {
	return t.name
}
func (t exampleType) GetGroup() string {
	return t.group
}
func (t exampleType) GetVerbs() []types.Verb {
	var verbs []types.Verb
	for _, s := range t.verbs {
		verbs = append(verbs, exampleVerb(s))
	}
	return verbs
}
func (t exampleType) GetActions() []types.Action {
	var actions []types.Action
	for _, s := range t.actions {
		actions = append(actions, exampleAction(s))
	}
	return actions
}

type exampleVerb string

func (s exampleVerb) GetName() string {
	return string(s)
}
func (s exampleVerb) String() string {
	return string(s)
}

type exampleAction string

func (s exampleAction) GetName() string {
	return string(s)
}
func (s exampleAction) String() string {
	return string(s)
}

var dnsRequestT = exampleType{
	group:         "dns",
	name:          "request",
	actions:       []string{"allow", "deny"},
	verbs:         []string{"resolve"},
	defaultAction: "deny",
}

var exampleProductsInventoryT = exampleType{
	group:         "products",
	name:          "inventory",
	actions:       []string{"allow", "deny"},
	verbs:         []string{"audit", "use", "manage"},
	defaultAction: "deny",
}
