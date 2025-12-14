.PHONY: help build test lint generate update-openapi clean coverage

# Default target
help:
	@echo "Available targets:"
	@echo "  make build           - Build all packages"
	@echo "  make test            - Run all tests"
	@echo "  make lint            - Run golangci-lint"
	@echo "  make generate        - Generate code from OpenAPI spec"
	@echo "  make update-openapi  - Download latest OpenAPI specification"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make coverage        - Generate test coverage report"

# Build all packages
build:
	@echo "Building all packages..."
	@go build ./...

# Run all tests
test:
	@echo "Running tests..."
	@go test -v -race ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run --timeout=5m

# Generate code from OpenAPI spec
generate:
	@echo "Generating code from OpenAPI specification..."
	@go generate ./tools

# Update OpenAPI specification
update-openapi:
	@echo "Updating OpenAPI specification..."
	@./tools/update-openapi.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@go clean ./...
	@rm -f coverage.txt coverage.html

# Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"
