MODULE_NAME := github.com/smford/golang-term2svg
GO          := go

.PHONY: all test test-race fmt lint tidy vet clean help

all: test

## test: Run unit tests across all packages
test:
	@echo "==> Running module tests..."
	$(GO) test -v ./...

## test-race: Run tests with data race detector
test-race:
	@echo "==> Running tests with race detector..."
	$(GO) test -v -race ./...

## fmt: Format all Go files
fmt:
	@echo "==> Formatting code..."
	$(GO) fmt ./...

## vet: Run go vet
vet:
	@echo "==> Vetting code..."
	$(GO) vet ./...

## tidy: Tidy and verify go.mod / go.sum dependencies
tidy:
	@echo "==> Tidying module dependencies..."
	$(GO) mod tidy

## clean: Clean test cache
clean:
	@echo "==> Cleaning test cache..."
	$(GO) clean -testcache
	@echo "==> Cleaned."

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
