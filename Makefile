# llima-box Makefile
# Build, test, and lint the project

# Variables
BINARY_NAME=llima-box
MAIN_PATH=./cmd/llima-box
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Go proxy configuration with fallback to direct downloads
export GOPROXY=https://proxy.golang.org,direct

# Build targets
.PHONY: help check build all clean test coverage lint fmt fmt-check vet tidy install deps deps-verify

# Default target - show help
.DEFAULT_GOAL := help

# Run all checks and build
all: check build

# Build the binary
build: fmt
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

build-linux-amd64:
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-linux-arm64:
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

build-darwin-amd64:
	@echo "Building for macOS AMD64..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-darwin-arm64:
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS) $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Check formatting
fmt-check:
	@echo "Checking code formatting..."
	@files=$$($(GOFMT) -l . | grep -v '^vendor/'); \
	if [ -n "$$files" ]; then \
		echo "The following files need formatting:"; \
		echo "$$files"; \
		exit 1; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run golangci-lint (requires golangci-lint to be installed)
lint:
	@echo "Running golangci-lint..."
	@which $(GOLINT) > /dev/null || (echo "golangci-lint not found. Install it with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin"; exit 1)
	$(GOLINT) run ./...

# Run all checks (formatting, vetting, linting, tests)
# Build first to apply automatic fixes (fmt, tidy)
check: build vet lint test

# Tidy go.mod and go.sum
tidy:
	@echo "Tidying go.mod and go.sum..."
	$(GOMOD) tidy

# Update dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Verify dependencies
deps-verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Help target (default)
help:
	@echo "llima-box Makefile"
	@echo ""
	@echo "Common targets:"
	@echo "  make help           - Show this help message (default)"
	@echo "  make check          - Run all validations (fmt, vet, lint, test)"
	@echo "  make build          - Build the binary for current platform"
	@echo ""
	@echo "Build targets:"
	@echo "  make build-all      - Build binaries for all platforms"
	@echo "  make install        - Install the binary to GOPATH/bin"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Testing targets:"
	@echo "  make test           - Run tests"
	@echo "  make coverage       - Run tests with coverage report"
	@echo ""
	@echo "Code quality targets:"
	@echo "  make fmt            - Format code"
	@echo "  make fmt-check      - Check if code is formatted"
	@echo "  make vet            - Run go vet"
	@echo "  make lint           - Run golangci-lint"
	@echo ""
	@echo "Dependency targets:"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make deps-verify    - Verify dependencies"
	@echo ""
	@echo "Other targets:"
	@echo "  make all            - Run check and build"
