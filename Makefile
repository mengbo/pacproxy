.PHONY: all build build-all test clean run install fmt lint vet help

# Variables
BINARY_NAME=pacproxy
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/mengbo/pacproxy/cmd.Version=$(VERSION)"

# Default target
all: build

## help: Show this help message
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

## build-all: Build binaries for all platforms
build-all: build-linux build-darwin build-windows
	@echo "All builds complete"

## build-linux: Build for Linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .

## build-darwin: Build for macOS
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .

## build-windows: Build for Windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

## coverage: Run tests with coverage report
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

## run: Build and run the application
run: build
	./$(BINARY_NAME)

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

## lint: Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
		exit 1; \
	fi

## deps: Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

## tidy: Clean up go.mod and go.sum
tidy:
	@echo "Tidying modules..."
	go mod tidy

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed"

## release: Create a release build
release: clean build-all
	@echo "Creating release archives..."
	@mkdir -p dist
	@for file in dist/*; do \
		if [ -f "$$file" ]; then \
			base=$$(basename $$file); \
			tar -czf "dist/$$base.tar.gz" -C dist "$$base"; \
			echo "Created dist/$$base.tar.gz"; \
		fi; \
	done
	@echo "Release builds ready in dist/"
