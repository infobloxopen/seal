package petstore.all

#deny to deliver petstore.order where "boss" in subject.groups;
test_in {
	in := {
		"type": "petstore.order",
		"verb": "deliver",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"jti": "petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["boss", "everyone"],
		}),
	}

	deny with input as in
}

test_in_negative {
	in := {
		"type": "petstore.order",
		"verb": "deliver",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"jti": "petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["employee", "everyone"],
		}),
	}

	not deny with input as in
}

#allow subject group not_operator_precedence to buy petstore.pet where not ctx.neutered and ctx.potty_trained;
test_not_operator_precedence_positive {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({"groups": ["not_operator_precedence"]}),
		"ctx": [{"neutered": false, "potty_trained": true}],
	}

	allow with input as in
}

test_not_operator_precedence_negative1 {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({"groups": ["not_operator_precedence"]}),
		"ctx": [{"neutered": false, "potty_trained": false}],
	}

	not allow with input as in
}

test_not_operator_precedence_negative2 {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({"groups": ["not_operator_precedence"]}),
		"ctx": [{"neutered": true, "potty_trained": false}],
	}

	not allow with input as in
}

test_not_operator_precedence_negative3 {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({"groups": ["not_operator_precedence"]}),
		"ctx": [{"neutered": true, "potty_trained": true}],
	}

	not allow with input as in
}

#deny subject group regexp to use petstore.* where subject.jti =~ "@petstore.swagger.io$";
test_regexp {
	in := {
		"type": "petstore.pet",
		"verb": "get",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"jti": "@petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["regexp", "test"],
		}),
	}

	deny with input as in
}

test_regexp_negative {
	in := {
		"type": "petstore.pet",
		"verb": "watch",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"jti": "just test regexp params",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["regexp", "test"],
		}),
	}

	not deny with input as in
}

# deny subject group everyone to buy petstore.pet where ctx.age <= 2 and ctx.name == "specificPetName";
test_ctx_usage_multiply {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone"],
		}),
		"ctx": [{"age": 1, "name": "specificPetName"}],
	}

	deny with input as in
}

# deny ... where ctx.age <= 2 and ctx.name == "specificPetName";
# pet with age==1 and name==NotSpecificPetName is allowed for buy verb
# also, pet with age==3 and name==specificPetName is allowed too
# test is not fully relative to logic, but just to demo how multipy ctx will work
test_ctx_usage_negative_multi_ctx {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone"],
		}),
		"ctx": [
			{"age": 1, "name": "NotSpecificPetName"},
			{"age": 3, "name": "specificPetName"},
		],
	}

	not deny with input as in
}

test_ctx_usage_negative {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone"],
		}),
		"ctx": [{"age": 3, "name": "specificPetName"}],
	}

	not deny with input as in
}

#deny subject group everyone to buy petstore.pet
#    where ctx.tags["endangered"] == "true";
test_use_tags {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		}),
		"ctx": [{"tags": {"endangered": "true"}}],
	}

	deny with input as in
}

test_use_tags_negative {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		}),
		"ctx": [{"tags": {"endangered": "not_true"}}],
	}

	not deny with input as in
}

test_use_tags_negative_missing_endangered_tag {
	in := {
		"type": "petstore.pet",
		"verb": "buy",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		}),
		"ctx": [{"tags": {"trash": "not_true"}}],
	}

	not deny with input as in
}

# deny subject group everyone to use petstore.* where subject.iss != "petstore.swagger.io";
test_use_petstore_jwt {
	in := {
		"type": "petstore.pet",
		"verb": "list",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		}),
	}

	deny with input as in
}

test_use_petstore_jwt_negative {
	in := {
		"type": "petstore.pet",
		"verb": "update",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		}),
	}

	not deny with input as in
}

# deny subject group banned to manage petstore.pet;
test_banned_deny {
	in := {
		"type": "petstore.pet",
		"verb": "create",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "banned"],
		}),
	}

	deny with input as in
}

test_banned_deny_negative {
	in := {
		"type": "petstore.pet",
		"verb": "delete",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone"],
		}),
	}

	not deny with input as in
}

# allow subject group everyone to inspect petstore.pet;
test_inspect {
	in := {
		"type": "petstore.pet",
		"verb": "list",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "inspector-gadget@disney.com",
			"groups": ["everyone"],
		}),
	}

	allow with input as in
}

test_inspect_negative {
	in := {
		"type": "petstore.pet",
		"verb": "get",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "inspector-gadget@disney.com",
			"groups": ["everyone"],
		}),
	}

	not allow with input as in
}

# allow subject group customers to read petstore.pet;
test_read {
	in := {
		"type": "petstore.pet",
		"verb": "watch",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "customers"],
		}),
	}

	allow with input as in
}

test_read_negative {
	in := {
		"type": "petstore.pet",
		"verb": "get",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "customers"],
		}),
	}

	deny with input as in
}

# allow subject user cto@petstore.swagger.io to manage petstore.pet;
test_manage_cto {
	in := {
		"type": "petstore.pet",
		"verb": "watch",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "cto@petstore.swagger.io",
			"groups": ["everyone", "ctos"],
		}),
	}

	allow with input as in
}

#deny to cancel petstore.order where ctx.status == "delivered";
test_blank_subject {
	in := {
		"type": "petstore.order",
		"verb": "deliver",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		}),
		"ctx": [{"status": "delivered"}],
	}

	deny with input as in
}

# allow subject group employ33s to oper4te petstore.stor3
#       where ctx.addre55 == "1234 Main St." and ctx.t4gs["0"] == "zer0";
test_alphanumeric_identifiers {
	in := {
		"type": "petstore.stor3",
		"verb": "sw33p",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["ex3cut1ves", "employ33s"],
		}),
		"ctx": [{"t4gs": {"0": "zer0"}, "addre55": "1234 Main St."}],
	}

	allow with input as in
}

# sealtest_jwt_encode_sign returns HMAC signed jwt from claims for testing purposes
sealtest_jwt_encode_sign(claims) = jwt {
	jwt = io.jwt.encode_sign({
		"typ": "JWT",
		"alg": "HS256",
	}, claims, {
		"kty": "oct",
		# k from https://tools.ietf.org/html/rfc7517#appendix-A.3
		"k": "AyM1SysPpbyDfgZld3umj1qzKObwVMkoqQ-EstJQLr_T-1qS0gZH75aKtMN3Yj0iPS4hcgUuTwjAzZr1Z9CAowg",
	})
}
