package petstore.all

default allow = false

default deny = false

deny {
	seal_list_contains(seal_subject.groups, `everyone`)
	input.verb == `use`
	re_match(`petstore.*`, input.type)
	seal_subject.iss != "petstore.swagger.io"
}

deny {
	seal_list_contains(seal_subject.groups, `everyone`)
	input.verb == `buy`
	re_match(`petstore.pet`, input.type)
	input.age <= 2
}

deny {
	seal_list_contains(seal_subject.groups, `banned`)
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

deny {
	seal_list_contains(seal_subject.groups, `managers`)
	input.verb == `sell`
	re_match(`petstore.pet`, input.type)
	input.status != "available"
}

deny {
	seal_list_contains(seal_subject.groups, `fussy`)
	input.verb == `buy`
	re_match(`petstore.pet`, input.type)
	not line5_not1_cnd
}

line5_not1_cnd {
	input.neutered

	not line5_not2_cnd
}

line5_not2_cnd {
	input.potty_trained
}

allow {
	seal_list_contains(seal_subject.groups, `fussy`)
	input.verb == `buy`
	re_match(`petstore.pet`, input.type)
	not line6_not1_cnd
}

line6_not1_cnd {
	input.neutered
	input.potty_trained
}

allow {
	seal_list_contains(seal_subject.groups, `operators`)
	input.verb == `use`
	re_match(`petstore.*`, input.type)
}

allow {
	seal_list_contains(seal_subject.groups, `managers`)
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

allow {
	input.subject.email == `cto@petstore.swagger.io`
	input.verb == `manage`
	re_match(`petstore.*`, input.type)
}

allow {
	seal_list_contains(seal_subject.groups, `everyone`)
	input.verb == `inspect`
	re_match(`petstore.pet`, input.type)
}

allow {
	seal_list_contains(seal_subject.groups, `customers`)
	input.verb == `read`
	re_match(`petstore.pet`, input.type)
}

allow {
	seal_list_contains(seal_subject.groups, `customers`)
	input.verb == `buy`
	re_match(`petstore.pet`, input.type)
	input.status == "available"
}

allow {
	seal_list_contains(seal_subject.groups, `breeders_maltese`)
	input.verb == `buy`
	re_match(`petstore.pet`, input.type)
	input.status == "reserved"
	input.breed == "maltese"
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
