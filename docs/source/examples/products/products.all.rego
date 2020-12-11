package products.all

default allow = false

default deny = false

allow {
	seal_list_contains(seal_subject.groups, `everyone`)
	input.verb == `inspect`
	re_match(`products.inventory`, input.type)
}

# rego functions defined by seal

# Helper to get the token payload.
seal_subject = payload {
	[header, payload, signature] := io.jwt.decode(input.jwt)
}

# seal_list_contains returns true if elem exists in list
seal_list_contains(list, elem) {
	list[_] = elem
}
