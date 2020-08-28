
.PHONY: default

default:
	go build

test:
	@go test -v ./...

