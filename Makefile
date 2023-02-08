# Go related variables.
GOBASE=$(shell pwd)
export GOBIN=$(GOBASE)/bin

# Configuration for the web server.
WEB_SERVER_BINARY=$(GOBIN)/web

# --silent drops the need to prepend `@` to suppress command output.
MAKEFLAGS += --silent

.PHONY: start
start: start-containers install start-server ## Start the server.

.PHONY: start-server
start-server: install
	$(WEB_SERVER_BINARY)

install: tools ## Install binaries to $(GOBIN)
	go install -mod=vendor ./cmd/services/...

golangci_lint = $(GOBIN)/golangci-lint
$(golangci_lint):
	# https://golangci-lint.run/usage/install/#linux-and-windows (and macOS)
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOBIN} v1.50.1

# Install tools used for development and CI. Try to pin to a specific version
# wherever possible to make CI stable reproducible.
tools: $(golangci_lint)
	@go install mvdan.cc/gofumpt@v0.4.0
	@go install github.com/incu6us/goimports-reviser/v3@v3.3.0

.PHONY: clean
clean: ## Clean build files and artifacts.
	go clean ./...
	rm -rfv $(GOBIN)

.PHONY: start-containers
start-containers: dependencies ## Start the containers that run dependencies like PostgreSQL
	docker-compose up -d --remove-orphans

.PHONY: stop-containers
stop-containers: dependencies ## Stop the containers that run dependencies like PostgreSQL
	docker-compose stop

.PHONY: rm-containers
rm-containers: stop-containers ## Remove the containers
	docker-compose rm -f

.PHONY: generate
generate: ## Generate mocks, florence features and other code.
	@go generate -mod=vendor ./...
	@$(MAKE) fmt

.PHONY: lint
lint: tools ## Runs the linter
	$(GOBIN)/golangci-lint run ./...

.PHONY: test
test: ## Runs the tests
	APP_ENV=test go test -race -mod=vendor ./...

.PHONY: help
help:
	grep -E '^[/a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dependencies:
	command -v docker &> /dev/null || { echo "Docker isn't installed. Please install it." && exit 1; }
	docker info > /dev/null || { echo "Docker isn't running. Please start it and try again." && exit 1; }

.PHONY: fmt
fmt: tools ## Format code.
	@for f in $(shell find . -type f -name '*.go' -not -path "./vendor/*"); do $(GOBIN)/goimports-reviser -output="file" -project-name="github.com/deliveroo/" $${f}; done
	@$(GOBIN)/gofumpt -w $(shell find . -type f -name '*.go' -not -path "./vendor/*") 2> /dev/null # We need to run this twice as not all changes to pb files will be fixed in a single run (https://github.com/mvdan/gofumpt/issues/132).
	@$(GOBIN)/gofumpt -w $(shell find . -type f -name '*.go' -not -path "./vendor/*") 2> /dev/null
