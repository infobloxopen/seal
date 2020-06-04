package types

type Type interface {
	GetGroup() string
	GetName() string
	GetVerbs() []Verb
	GetActions() []Action
	String() string
	DefaultAction() string
}

type Verb interface {
	GetName() string
	String() string
}

type Action interface {
	GetName() string
	String() string
}

func IsValidVerb(t Type, verb string) bool {
	for _, v := range t.GetVerbs() {
		if verb == v.GetName() {
			return true
		}
	}
	return false
}

func IsValidAction(t Type, action string) bool {
	for _, a := range t.GetActions() {
		if action == a.GetName() {
			return true
		}
	}
	return false
}
