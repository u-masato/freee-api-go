# Contributing to freee-api-go

Thank you for your interest in contributing to freee-api-go! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Conventions](#coding-conventions)
- [Testing Requirements](#testing-requirements)
- [Documentation Requirements](#documentation-requirements)

## Development Environment Setup

### Prerequisites

- Go 1.21 or later
- Git
- golangci-lint (for linting)
- Make (optional, for using Makefile commands)

### Setup Steps

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/freee-api-go.git
   cd freee-api-go
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Install Development Tools**
   ```bash
   # Install golangci-lint
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

   # Install oapi-codegen (for code generation)
   go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
   ```

4. **Verify Setup**
   ```bash
   # Build all packages
   make build
   # or
   go build ./...

   # Run tests
   make test
   # or
   go test ./...

   # Run linter
   make lint
   # or
   golangci-lint run
   ```

### Available Make Commands

```bash
make help          # Show available commands
make build         # Build all packages
make test          # Run all tests
make coverage      # Generate test coverage report
make lint          # Run linter
make generate      # Generate code from OpenAPI spec
make update-openapi # Update OpenAPI specification
make clean         # Clean build artifacts
```

## Pull Request Process

### Before You Start

1. **Check for Existing Issues**: Search existing issues to see if your idea or bug has already been reported
2. **Create an Issue**: For new features or significant changes, create an issue first to discuss the approach
3. **Claim an Issue**: Comment on the issue to let others know you're working on it

### Creating a Pull Request

1. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make Your Changes**
   - Write code following our [coding conventions](#coding-conventions)
   - Add tests for new functionality
   - Update documentation as needed

3. **Run Tests and Linter**
   ```bash
   # Run tests
   make test

   # Run linter
   make lint

   # Check test coverage
   make coverage
   ```

4. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "Brief description of changes

   Detailed description:
   - Change 1
   - Change 2

   Fixes #123"
   ```

5. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a pull request on GitHub.

### PR Requirements

- All tests must pass
- Code coverage should not decrease (target: 80%+)
- Linter must pass with no errors
- Documentation must be updated for user-facing changes
- PR description should clearly explain the changes and their purpose

## Coding Conventions

### Go Style Guide

We follow the official [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) and [Effective Go](https://go.dev/doc/effective_go) guidelines.

**Key Points**:

- Use `gofmt` and `goimports` for formatting
- Follow Go naming conventions:
  - Use `MixedCaps` or `mixedCaps` (not underscores)
  - Acronyms should be all uppercase (e.g., `APIClient`, not `ApiClient`)
- Keep functions small and focused
- Use meaningful variable names (avoid single-letter names except in short scopes)
- Add comments for exported functions and types

### Package Structure

```
freee-api-go/
├── auth/           # OAuth2 authentication (public)
├── transport/      # HTTP transport middleware (public)
├── client/         # Main client configuration (public)
├── accounting/     # High-level Facade (public)
├── internal/       # Internal packages (may change)
│   ├── gen/        # OpenAPI-generated code
│   └── testutil/   # Test utilities
├── examples/       # Sample code
└── tools/          # Code generation scripts
```

**Guidelines**:
- Public packages must maintain API stability (semantic versioning)
- Use `internal/` for implementation details that may change
- Generated code stays in `internal/gen/` and is excluded from linting

### Error Handling

Use custom error types with context:

```go
type MyError struct {
    Op  string  // Operation name
    Err error   // Underlying error
}

func (e *MyError) Error() string {
    return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *MyError) Unwrap() error {
    return e.Err
}
```

Always provide context in errors:
```go
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}
```

### Context Usage

Always accept `context.Context` as the first parameter for:
- Network operations
- Database operations
- Long-running operations
- API calls

```go
func DoSomething(ctx context.Context, param string) error {
    // Implementation
}
```

## Testing Requirements

### Test Coverage

- Target: 80%+ code coverage
- All public functions must have tests
- Test both success and error cases

### Testing Patterns

**Use table-driven tests with subtests**:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "test",
            want:  "result",
        },
        {
            name:    "invalid input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Feature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Feature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Use httptest.Server for HTTP mocking**:

```go
func TestHTTPClient(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate request
        // Return mock response
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer server.Close()

    // Test implementation
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run specific package
go test -v ./auth/...

# Run single test
go test -v -run TestName ./package/
```

## Documentation Requirements

### Code Documentation

**Exported functions and types must have comments**:

```go
// NewClient creates a new freee API client with the given options.
// It returns an error if the configuration is invalid.
func NewClient(opts ...Option) (*Client, error) {
    // Implementation
}
```

**Package documentation** (in `doc.go` or package file):

```go
// Package auth provides OAuth2 authentication for the freee API.
//
// This package implements the OAuth2 authorization code flow with PKCE
// and provides token management with automatic refresh.
//
// Example usage:
//
//     config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
//     authURL := config.AuthCodeURL(state)
//     // ... redirect user to authURL ...
//     token, err := config.Exchange(ctx, code)
//
package auth
```

### User-Facing Documentation

For user-facing changes, update:

- **README.md**: For major features or usage changes
- **Examples**: Add or update code examples in `examples/`
- **CLAUDE.md**: Update project instructions if architecture changes
- **TODO.md**: Mark completed tasks and update progress

### Example Code

All examples must:
- Be complete and runnable
- Include error handling
- Have comments explaining key steps
- Include a README.md with setup instructions

## Issue and PR Templates

When creating issues or pull requests, please use the provided templates:

- Bug reports: `.github/ISSUE_TEMPLATE/bug_report.md`
- Feature requests: `.github/ISSUE_TEMPLATE/feature_request.md`
- Pull requests: `.github/PULL_REQUEST_TEMPLATE.md`

## Questions?

If you have questions:

1. Check the [README.md](README.md), [PLAN.md](PLAN.md), and [TODO.md](TODO.md)
2. Search existing issues
3. Create a new issue with the question label

## Code of Conduct

Please be respectful and constructive in all interactions. We're all here to build something useful together.

## License

By contributing to freee-api-go, you agree that your contributions will be licensed under the MIT License.
