package products.all

# allow subject group everyone to inspect products.inventory;
test_inspect {
	in := {
		"type": "products.inventory",
		"verb": "inspect",
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
		"type": "products.inventory",
		"verb": "read",
		"jwt": sealtest_jwt_encode_sign({
			"iss": "not_petstore.swagger.io",
			"sub": "inspector-gadget@disney.com",
			"groups": ["everyone"],
		}),
	}

	not allow with input as in
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
