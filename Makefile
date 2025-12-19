# Makefile for ramjam CLI tool

# Binary name
BINARY_NAME=ramjam

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Build directory
BUILD_DIR=bin

# Version information
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Ldflags for version information
LDFLAGS=-ldflags "-X github.com/michaelmccabe/ramjam/cmd/ramjam/cmd.Version=$(VERSION)"

.PHONY: all build install clean test help tidy run

# Default target
all: clean build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ramjam
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install the binary to $GOPATH/bin or $GOBIN
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOINSTALL) $(LDFLAGS) ./cmd/ramjam
	@echo "Installation complete. Binary installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"
	@echo "Make sure $(shell go env GOPATH)/bin is in your PATH"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies tidied"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

# Run the application
run:
	$(GOCMD) run ./cmd/ramjam

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ramjam
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/ramjam
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/ramjam
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/ramjam
	@echo "Multi-platform build complete"

# Help target
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make install        - Install the binary to GOPATH/bin"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make tidy           - Tidy Go modules"
	@echo "  make deps           - Download dependencies"
	@echo "  make run            - Run the application"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make help           - Show this help message"
