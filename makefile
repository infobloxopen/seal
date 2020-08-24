
.PHONY: default

default: seal

.PHONY: seal
seal:
	@go build

test:
	go test -v ./...

example: seal
	./seal compile -s docs/source/examples/simple/products.inventory.swagger -f docs/source/examples/simple/products.inventory.seal
