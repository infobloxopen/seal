package petstore.all

# deny subject group everyone to use petstore.* where subject.iss != "petstore.com";
test_use_petstore_jwt {
	in := {
		"type": "petstore.pet",
		"verb": "use",
		"subject": {
			"email": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		},
		# jwt: {"iss": "not_petstore.com"}
		"jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJub3RfcGV0c3RvcmUuY29tIn0.iEsWURIWWbd4LV4UAyU7SPufo9pD5qcGLnCmqxxQExo",
	}

	deny with input as in
}

test_use_petstore_jwt_negative {
	in := {
		"type": "petstore.pet",
		"verb": "use",
		"subject": {
			"email": "wiley-e-coyote@acme.com",
			"groups": ["everyone", "test"],
		},
		# jwt: {"iss": "petstore.com"}
		"jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJwZXRzdG9yZS5jb20ifQ.0OBS3OsSdj5y_Tigcr0pYjuH0uuVnmUcryHNJvvEuy0",
	}

	not deny with input as in
}

# deny subject group banned to manage petstore.pet;
test_banned_deny {
	in := {
		"type": "petstore.pet",
		"verb": "manage",
		"subject": {
			"email": "wiley-e-coyote@acme.com",
			"groups": ["banned", "everyone"],
		},
	}

	deny with input as in
}

test_banned_deny_negative {
	in := {
		"type": "petstore.pet",
		"verb": "manage",
		"subject": {
			"email": "road-runner@acme.com",
			"groups": ["everyone"],
		},
	}

	not deny with input as in
}

# allow subject group everyone to inspect petstore.pet;
test_inspect {
	in := {
		"type": "petstore.pet",
		"verb": "inspect",
		"subject": {
			"email": "inspector-gadget@disney.com",
			"groups": ["everyone"],
		},
	}

	allow with input as in
}

test_inspect_negative {
	in := {
		"type": "petstore.pet",
		"verb": "read",
		"subject": {
			"email": "inspector-gadget@disney.com",
			"groups": ["everyone"],
		},
	}

	not allow with input as in
}

# allow subject group customers to read petstore.pet;
test_read {
	in := {
		"type": "petstore.pet",
		"verb": "read",
		"subject": {
			"email": "doc-mcstuffin@disney.com",
			"groups": ["customers", "everyone"],
		},
	}

	allow with input as in
}

test_read_negative {
	in := {
		"type": "petstore.pet",
		"verb": "use",
		"subject": {
			"email": "doc-mcstuffin@disney.com",
			"groups": ["customers", "everyone"],
		},
	}

	not allow with input as in
}

# allow subject user cto@petstore.swagger.io to manage petstore.pet;
test_manage_cto {
	in := {
		"type": "petstore.pet",
		"verb": "manage",
		"subject": {
			"email": "cto@petstore.swagger.io",
			"groups": ["ctos", "everyone"],
		},
	}

	allow with input as in
}
