.PHONY: default
default: seal

.PHONY: seal
seal:
	@go build

.PHONY: test
test:
	@go test -v ./...

.PHONY: example
example: seal
	./seal compile -s docs/source/examples/simple/products.inventory.swagger -f docs/source/examples/simple/products.inventory.seal

.PHONY: petstore
petstore: seal
	./seal compile -s docs/source/examples/petstore/petstore.all.swagger -f docs/source/examples/petstore/petstore.all.seal
	@echo "TODO: redirect above output to docs/source/examples/petstore/petstore.all.rego"
	@echo "TODO: fix compiler so it generates valid rego - example is in docs/source/examples/petstore/petstore.all.rego"
	./docs/source/examples/check-rego.sh docs/source/examples/petstore
