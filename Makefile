# Terraform Helper

# Variables
BINARY_NAME=terraform-helper
BUILD_DIR=bin
MAIN_PATH=./cmd/terraform-helper
GO_FILES=$(shell find . -name '*.go' -not -path './test/*' -not -path './bin/*')
OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)

# Version information (can be overridden via build flags)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=gofmt

# Build flags with version information
VERSION_PKG=github.com/giovannirossini/terraform-helper/internal/version
LDFLAGS=-ldflags "-s -w -X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).BuildTime=$(BUILD_TIME)"

.PHONY: all build clean test run tidy vet fmt-check deploy help

all: clean tidy test build

## build: Build the binary (use OUTPUT=path/to/binary to override output path)
build:
	@echo "Building $(BINARY_NAME)..."
	@OUTPUT_DIR="$$(dirname $(OUTPUT))"; \
	if [ "$$OUTPUT_DIR" != "." ]; then \
		mkdir -p "$$OUTPUT_DIR"; \
	fi
	$(GOBUILD) $(LDFLAGS) -o $(OUTPUT) $(MAIN_PATH)

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

## test: Run unit tests with coverage
test:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v ./...
	@echo ""
	@echo "Coverage summary:"
	@$(GOTEST) -coverprofile=coverage.out ./... > /dev/null 2>&1 && \
		go tool cover -func=coverage.out | tail -1 || true
	@rm -f coverage.out

## tidy: Clean up go.mod and go.sum
tidy:
	@echo "Tidying up modules..."
	$(GOMOD) tidy

## vet: Run go vet on all packages
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

## fmt-check: Check if code is properly formatted
fmt-check:
	@echo "Checking code formatting..."
	@if [ "$$($(GOFMT) -l . | wc -l)" -gt 0 ]; then \
		echo "Code is not formatted. Run 'go fmt ./...'"; \
		$(GOFMT) -d .; \
		exit 1; \
	fi
	@echo "Code is properly formatted."

## deploy: Install binary to /usr/local/bin (requires sudo)
deploy: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(OUTPUT) /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) installed successfully."

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' | column -t -s ':'
