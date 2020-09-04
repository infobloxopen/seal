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
