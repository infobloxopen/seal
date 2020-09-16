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
	./seal compile -s $(dir)/petstore.all.swagger -f $(dir)/petstore.all.seal | tee $(dir)/petstore.all.rego
	./docs/source/examples/check-rego.sh $(dir)
	git diff --exit-code $(dir)
	@echo "### petstore example passed REGO OPA tests"
