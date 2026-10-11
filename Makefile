MODULE_NAME := github.com/smford/golang-term2svg
GO          := go
BINARY_NAME := term2svg

.PHONY: all test test-race test-cover fmt lint tidy vet build cross-build clean pre-commit help

all: test

## test: Run unit tests across all packages
test:
	@echo "==> Running module tests..."
	$(GO) test -v ./...

## test-race: Run tests with data race detector
test-race:
	@echo "==> Running tests with race detector..."
	$(GO) test -v -race ./...

## test-cover: Run tests with code coverage analysis
test-cover:
	@echo "==> Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

## fmt: Format all Go files with gofmt -s
fmt:
	@echo "==> Formatting code..."
	gofmt -s -w .

## vet: Run go vet
vet:
	@echo "==> Vetting code..."
	$(GO) vet ./...

## lint: Run golangci-lint
lint:
	@echo "==> Running linters..."
	golangci-lint run ./...

## tidy: Tidy and verify go.mod / go.sum dependencies
tidy:
	@echo "==> Tidying module dependencies..."
	$(GO) mod tidy
	$(GO) mod verify

## build: Build CLI binary for host platform
build:
	@echo "==> Building $(BINARY_NAME)..."
	CGO_ENABLED=0 $(GO) build -trimpath -o bin/$(BINARY_NAME) ./cmd/term2svg

## cross-build: Cross-compile static binaries across matrix
cross-build:
	@echo "==> Cross-compiling for Linux, macOS, and Windows..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/term2svg
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -trimpath -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/term2svg
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/term2svg
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/term2svg
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/term2svg

## pre-commit: Run all pre-commit hooks locally
pre-commit:
	@echo "==> Running pre-commit hooks..."
	pre-commit run --all-files

## clean: Clean test cache and build artifacts
clean:
	@echo "==> Cleaning build artifacts and test cache..."
	$(GO) clean -testcache
	rm -rf bin/ dist/ coverage.out
	@echo "==> Cleaned."

## help: Display this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
