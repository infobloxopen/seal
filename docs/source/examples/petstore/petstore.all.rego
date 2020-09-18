package petstore.all

default allow = false

default deny = false

deny {
	seal_list_contains(input.subject.groups, `banned`)
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

deny {
	seal_list_contains(input.subject.groups, `managers`)
	input.verb == `sell`
	re_match(`petstore.pet`, input.type)
	input.status != "available"
}

allow {
	seal_list_contains(input.subject.groups, `operators`)
	input.verb == `use`
	re_match(`petstore.*`, input.type)
}

allow {
	seal_list_contains(input.subject.groups, `managers`)
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

allow {
	input.subject.email == `cto@petstore.swagger.io`
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

allow {
	seal_list_contains(input.subject.groups, `everyone`)
	input.verb == `inspect`
	re_match(`petstore.pet`, input.type)
}

allow {
	seal_list_contains(input.subject.groups, `customers`)
	input.verb == `read`
	re_match(`petstore.pet`, input.type)
}

allow {
	seal_list_contains(input.subject.groups, `customers`)
	input.verb == `buy`
	re_match(`petstore.pet`, input.type)
	input.status == "available"
}

# rego functions defined by seal

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
	list[_] = elem
}
