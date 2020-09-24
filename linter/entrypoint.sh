#!/bin/bash
cd "$GITHUB_WORKSPACE"

golangci-lint run --out-format colored-line-number ${INPUT_GOLANGCI_LINT_FLAGS}
