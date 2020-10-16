package petstore.all

species_protected := {
	"lion": true,
	"tiger": true,
	"cheetah": true,
}

# assume this is generated from seal rule:
#   deny to provision petstore.pet where ctx.species in $company.list["name=species_protected"];
bench_deny_in_map_species {
	input.type == `petstore.pet`
	input.verb == `provision`
	species_protected[input.species]
}

# positive test
test_bench_deny_in_map_species {
	in := {
		"type": "petstore.pet",
		"verb": "provision",
		"species": "cheetah",
	}

	bench_deny_in_map_species with input as in
}

test_bench_deny_in_map_species_2nd {
	in := {
		"type": "petstore.pet",
		"verb": "provision",
		"species": "cheetah",
	}

	bench_deny_in_map_species with input as in
}

# negative test
test_bench_deny_in_map_species_negative {
	in := {
		"type": "petstore.pet",
		"verb": "provision",
		"species": "dog",
	}

	not bench_deny_in_map_species with input as in
}

test_bench_deny_in_map_species_negative_2nd {
	in := {
		"type": "petstore.pet",
		"verb": "provision",
		"species": "dog",
	}

	not bench_deny_in_map_species with input as in
}
