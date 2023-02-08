# Ensure that we use vendored binaries before consulting the system.
GOBIN=$(shell pwd)/bin
export PATH := $(GOBIN):$(PATH)

# Use Go modules.
export GO111MODULE := on

all: install lint test

.PHONY: install
install: ## Install the library.
	@go install ./...

.PHONY: lint
lint: ## Lint the project with golangci-lint.
	@$(GOBIN)/golangci-lint run ./...

.PHONY: setup
setup:  ## Download dependencies.
	git submodule update --init --recursive
	@GOBIN=$(GOBIN) go mod download
	@GOBIN=$(GOBIN) curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOBIN} v1.46.2

.PHONY: test
test:  ## Run tests.
	@go test ./...

.PHONY: help
help:
	@grep -E '^[/a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
