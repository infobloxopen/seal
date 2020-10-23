package petstore.all

# deny subject group everyone to use petstore.* where subject.iss != "petstore.swagger.io";
test_use_petstore_jwt {
	in := {
		"type": "petstore.pet",
		"verb": "use",
		# TODO: GH-82 remove "subject"
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
		"verb": "use",
		# TODO: GH-82 remove "subject"
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
		"verb": "manage",
		# TODO: GH-82 replace "subject" with "jwt": sealtest_jwt_encode_sign(...)
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
		"verb": "manage",
		# TODO: GH-82 replace "subject" with "jwt": sealtest_jwt_encode_sign(...)
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
		"verb": "inspect",
		# TODO: GH-82 replace "subject" with "jwt": sealtest_jwt_encode_sign(...)
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
		"verb": "read",
		# TODO: GH-82 replace "subject" with "jwt": sealtest_jwt_encode_sign(...)
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
		"verb": "read",
		# TODO: GH-82 replace "subject" with "jwt": sealtest_jwt_encode_sign(...)
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
		"verb": "use",
		# TODO: GH-82 replace "subject" with "jwt": sealtest_jwt_encode_sign(...)
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "customers"],
		}),
	}

	not allow with input as in
}

# allow subject user cto@petstore.swagger.io to manage petstore.pet;
test_manage_cto {
	in := {
		"type": "petstore.pet",
		"verb": "manage",
		# TODO: GH-82 replace "subject" with "jwt": sealtest_jwt_encode_sign(...)
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "cto@petstore.swagger.io",
			"groups": ["everyone", "ctos"],
		}),
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
