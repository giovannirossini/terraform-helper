.PHONY: install build clean help

# Binary name
BINARY_NAME=terraform-helper

# Default target
all: install build

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build without cache
build:
	@echo "Building $(BINARY_NAME) without cache..."
	go build -a -o $(BINARY_NAME) .

deploy:
	@echo "Moving to /usr/local/bin"
	sudo mv $(BINARY_NAME) /usr/local/bin

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean -cache

# Help target
help:
	@echo "Available targets:"
	@echo "  make install  - Install/update dependencies"
	@echo "  make build    - Build the binary without cache"
	@echo "  make clean    - Remove build artifacts and clean cache"
	@echo "  make all      - Install dependencies and build (default)"
	@echo "  make help     - Show this help message"
