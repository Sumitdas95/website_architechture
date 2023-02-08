# Install all development tools and build artifacts to the project's `bin` directory.
export GOBIN=$(CURDIR)/bin

# Ensure that Deliveroo's private repos are not checksummed by Go.
export GOPRIVATE=github.com/deliveroo/*

# Binary dependencies
$(GOBIN):
	mkdir -p $(GOBIN)

.PHONY: install-tools
install-tools: $(GOBIN) ## Install all dev tools into bin directory.
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
	@go install github.com/vektra/mockery/v2@v2.14.0

.PHONY: build
build: generate
	go build ./...

.PHONY: generate
generate: install-tools
	go generate ./...

.PHONY: lint
lint: install-tools
	$(GOBIN)/golangci-lint run --config .golangci.yml ./...

.PHONY: test
test:
	go run ./examples/mocked_tests
	go run ./examples/tasks
	go test -race -v ./...

.PHONY: test/db
test/db:
	go test -race -v -tags db ./...
