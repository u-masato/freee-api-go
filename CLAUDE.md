# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

freee-api-go is a Go SDK for the freee Accounting API (フリー会計API). The project implements OAuth2 authentication and provides a type-safe client library generated from OpenAPI specifications.

**Current Status**: Phase 3 complete (see TODO.md for detailed progress)
- ✅ Phase 1: Project foundation
- ✅ Phase 2: OAuth2 authentication
- ✅ Phase 3: HTTP Transport layer
- 🚧 Phase 4+: API client generation, Facade, documentation (pending)

## Commands

### Build & Test
```bash
# Download dependencies
go mod download

# Build all packages
go build ./...

# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run specific package tests
go test -v ./auth/...
go test -v ./transport/...

# Run single test
go test -v -run TestName ./package/
```

### Linting
```bash
# Run linter (requires golangci-lint installed)
golangci-lint run

# With timeout
golangci-lint run --timeout=5m

# Auto-fix issues
golangci-lint run --fix
```

### Code Generation (Phase 4+)
```bash
# Install oapi-codegen (when needed for Phase 4)
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Generate code from OpenAPI spec (future)
go generate ./tools
```

## Architecture

The project follows a layered architecture separating concerns:

```
User Code
    ↓
Facade Layer (accounting/*)        ← User-friendly API (Phase 5)
    ↓
Generated Client (internal/gen)    ← OpenAPI-generated types/client (Phase 4)
    ↓
Transport Layer (transport/*)      ← HTTP middleware (Phase 3) ✅
    ↓
Auth Layer (auth/*)                ← OAuth2 flow (Phase 2) ✅
    ↓
freee API
```

### Package Structure & Responsibilities

**Public Packages** (stable API, semantic versioning):
- `client/` - Main client configuration and options (Phase 5)
- `accounting/` - High-level Facade for accounting operations (Phase 5)
- `auth/` - OAuth2 authentication (complete)
- `transport/` - HTTP transport middleware (complete)

**Internal Packages** (may change without notice):
- `internal/gen/` - OpenAPI-generated code (Phase 4)
- `internal/testutil/` - Test utilities

**Other**:
- `examples/` - Sample code and usage examples
- `tools/` - Code generation scripts

### Transport Layer Design (Phase 3)

The transport layer uses a **composable RoundTripper pattern**:

```go
// Each RoundTripper wraps the next, forming a chain
transport := NewTransport(
    WithRateLimit(10, 5),           // 10 req/sec, burst 5
    WithRetry(3, time.Second),      // 3 retries, exponential backoff
    WithLogging(logger),            // Structured logging with slog
    WithUserAgent("my-app/1.0"),   // Custom User-Agent
)
```

**RoundTripper Components**:
- `RateLimitRoundTripper` - Token bucket rate limiting (golang.org/x/time/rate)
- `RetryRoundTripper` - Exponential backoff for 5xx/429 errors
- `LoggingRoundTripper` - Structured logging with sensitive header masking
- `UserAgentRoundTripper` - User-Agent header management

**Important**: Each RoundTripper has a `SetBase(http.RoundTripper)` method for chaining.

### OAuth2 Implementation (Phase 2)

OAuth2 flow uses `golang.org/x/oauth2` as the foundation:

```go
// Config wraps oauth2.Config
config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)

// Standard OAuth2 flow
authURL := config.AuthCodeURL(state)
token, err := config.Exchange(ctx, code)

// TokenSource for auto-refresh
ts := config.TokenSource(ctx, token)
```

**Key Files**:
- `auth/config.go` - OAuth2 configuration wrapper
- `auth/token.go` - Token validation, file persistence (0600 permissions)
- `auth/tokensource.go` - CachedTokenSource with file backing
- `auth/errors.go` - Authentication-specific error types

**Security**: Token files saved with 0600 permissions, already in .gitignore.

## Development Conventions

### Error Handling

Use custom error types with context:

```go
// auth/errors.go pattern
type AuthError struct {
    Op          string  // Operation name
    Err         error   // Underlying error
    Code        string  // Error code (if applicable)
    Description string  // Human-readable description
}
```

### Testing Strategy

- **Unit tests**: Use `httptest.Server` for mocking HTTP endpoints
- **Table-driven tests**: Prefer subtests with `t.Run()`
- **Coverage target**: 80%+ (currently auth: 23/23 tests pass, transport: 42/42 tests pass)
- **Mock responses**: Use realistic freee API response structures

Example test pattern:
```go
func TestFeature(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate request
        // Return mock response
    }))
    defer server.Close()

    // Test implementation
}
```

### Commit Messages

Follow this format (established in Phase 2 & 3):

```
Brief summary line

Detailed description:
- Feature/change 1
- Feature/change 2

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

### Linter Configuration

Key linters enabled (see .golangci.yml):
- errcheck, govet, staticcheck (core)
- gosec (security)
- gofmt, goimports (formatting)
- revive, stylecheck (style)

**Excluded**:
- `internal/gen/*` - Generated code excluded from all linting
- `*_test.go` - Test files exempt from errcheck, dupl, gosec
- `examples/*` - Examples exempt from errcheck, gosec

## Phase-Specific Guidance

### Phase 4: Generated API Client (Next Phase)

When implementing Phase 4:

1. **OpenAPI Spec**: Download from https://developer.freee.co.jp/ → save to `api/openapi.yaml`
2. **Configure oapi-codegen**: Create `oapi-codegen.yaml` with:
   - Output: `internal/gen/`
   - Package: `gen`
   - Generate: models, client, types
3. **Validation**: Generated code must be version-controlled (not .gitignored)
4. **Error Types**: Create `client/error.go` wrapping freee API errors

### Phase 5: Accounting Facade

Design principles from PLAN.md:
- Hide generated client behind user-friendly Facade
- Implement Iterator/Pager for transparent pagination
- Context-first: All methods take `context.Context`
- Service-based: `DealsService`, `JournalsService`, `PartnersService`

### OAuth2 Examples

The `examples/oauth/` directory contains a complete working example:
- Local callback server on port 8080
- CSRF protection with state parameter
- Token persistence with auto-refresh
- Detailed README with troubleshooting

Reference this when creating other examples.

## Key Design Principles

From PLAN.md section 6:

1. **API Stability**: Generated code stays in `internal/`, only Facade is public
2. **User-Centric**: Hide pagination, error handling complexity
3. **OAuth Separation**: SDK assists auth but doesn't control web flow
4. **Extensibility**: Transport layer handles cross-cutting concerns

## Project-Specific Notes

### freee API Endpoints

OAuth2:
- AuthURL: `https://accounts.secure.freee.co.jp/public_api/authorize`
- TokenURL: `https://accounts.secure.freee.co.jp/public_api/token`

### Sensitive Data

Always masked in logs (LoggingRoundTripper):
- `Authorization` header
- `Cookie` / `Set-Cookie`
- `X-Api-Key` / `Api-Key`

### Dependencies

Core:
- `golang.org/x/oauth2` - OAuth2 implementation
- `golang.org/x/time/rate` - Rate limiting

Planned:
- `github.com/oapi-codegen/oapi-codegen/v2` - Code generation (Phase 4)

### TODO.md Structure

TODO.md tracks all implementation tasks by phase:
- Update progress after completing work
- Mark tasks with [x] when done
- Add commit hashes for reference
- Update "現在のフェーズ" (current phase)

## Reference Documentation

- **PLAN.md**: Complete architectural design and requirements
- **TODO.md**: Detailed implementation task breakdown and progress
- **README.md**: User-facing documentation
- **examples/oauth/README.md**: OAuth2 flow documentation

Check these files for detailed context before making significant changes.
