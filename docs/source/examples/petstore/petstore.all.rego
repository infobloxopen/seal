package petstore.all

default allow = false

deny {
	contains(input.subject.groups, `banned`)
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
	allow = false
}

allow {
	contains(input.subject.groups, `operators`)
	input.verb == `use`
	re_match(`petstore.*`, input.type)
}

allow {
	contains(input.subject.groups, `managers`)
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

allow {
	input.subject.email == `cto@petstore.swagger.io`
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

allow {
	contains(input.subject.groups, `everyone`)
	input.verb == `inspect`
	re_match(`petstore.pet`, input.type)
}

allow {
	contains(input.subject.groups, `customers`)
	input.verb == `read`
	re_match(`petstore.pet`, input.type)
}

# functions

# contains returns true if elem exists in list
contains(list, elem) {
	list[_] = elem
}
