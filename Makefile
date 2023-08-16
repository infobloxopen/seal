.PHONY: default
default: build

.PHONY: build
build: seal

.PHONY: seal
seal:
	@go build

.PHONY: test
test:
	@go test -v ./...

.PHONY: demo
demo: petstore

.PHONY: petstore
petstore: dir=docs/source/examples/petstore
petstore: seal
	./seal compile \
		-s $(dir)/petstore.jwt.swagger \
		-s $(dir)/petstore.tags.swagger \
		-s $(dir)/petstore.all.swagger \
		-f $(dir)/petstore.all.seal \
		-o $(dir)/petstore.all.rego.compiled

	cat $(dir)/petstore.all.rego.compiled
	cp $(dir)/petstore.all.rego.compiled $(dir)/petstore.all.rego
	# beware that check-rego.sh reformats the compiled rego files...
	./docs/source/examples/check-rego.sh $(dir)
	git diff --exit-code $(dir)
	@echo "### petstore example passed REGO OPA tests"

.PHONY: bench
bench: bench-petstore
bench-petstore:
bench-petstore: dir=docs/source/examples/petstore
bench-petstore: seal
	./seal compile \
		-s $(dir)/petstore.jwt.swagger \
		-s $(dir)/petstore.tags.swagger \
		-s $(dir)/petstore.all.swagger \
		-f $(dir)/petstore.all.seal \
		> $(dir)/petstore.all.rego.compiled
	cat $(dir)/petstore.all.rego.compiled

	cp $(dir)/petstore.all.rego.compiled $(dir)/petstore.all.rego
	# beware that bench-rego.sh reformats the compiled rego files...
	./docs/source/examples/bench-rego.sh $(dir)
	git diff --exit-code $(dir)
	@echo "### petstore benchmarks successfully completed"
