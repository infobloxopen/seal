PROJECT_ROOT            := github.com/infobloxopen/seal
#lint
LINTER_IMAGE             ?= reviewdog/action-golangci-lint:v1.0.5
GOLANGCI_CONFIG_PATH     ?= linter/golangci-lint.config.yaml
GOLANGCI_LINT_FLAGS      ?= --config ${GOLANGCI_CONFIG_PATH}
LINTER_IMAGE_ENNTRYPOINT = ./linter/entrypoint.sh

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
	./seal compile -s $(dir)/petstore.all.swagger -f $(dir)/petstore.all.seal | tee $(dir)/petstore.all.rego.compiled
	cp $(dir)/petstore.all.rego.compiled $(dir)/petstore.all.rego
	# beware that check-rego.sh reformats the compiled rego files...
	./docs/source/examples/check-rego.sh $(dir)
	git diff --exit-code $(dir)
	@echo "### petstore example passed REGO OPA tests"

.PHONY: lint
lint: 
	@docker run --rm -v $(shell pwd):/go/src/${PROJECT_ROOT} -w /go/src/${PROJECT_ROOT} -e GITHUB_WORKSPACE='/go/src/${PROJECT_ROOT}' \
		-e INPUT_GITHUB_TOKEN='${GitHub_PAT}' \
		-e INPUT_GOLANGCI_LINT_FLAGS='${GOLANGCI_LINT_FLAGS}' \
		--entrypoint='${LINTER_IMAGE_ENNTRYPOINT}' \
		${LINTER_IMAGE} 
