# Security Policy

## Supported Versions

This section tells you which versions of freee-api-go are currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

**Note**: As this project is in pre-1.0 development, we support only the latest release. Once we reach 1.0, we will provide clearer long-term support policies.

## Reporting a Vulnerability

We take the security of freee-api-go seriously. If you discover a security vulnerability, please report it responsibly.

### How to Report

**Please DO NOT open a public GitHub issue for security vulnerabilities.**

Instead, please report security vulnerabilities through one of the following methods:

1. **GitHub Security Advisories** (Preferred)
   - Go to https://github.com/u-masato/freee-api-go/security/advisories
   - Click "Report a vulnerability"
   - Provide detailed information about the vulnerability

2. **Email** (Alternative)
   - Send an email to: [Security contact to be added]
   - Use the subject line: `[SECURITY] freee-api-go vulnerability report`

### What to Include in Your Report

Please include as much of the following information as possible:

- **Description**: A clear description of the vulnerability
- **Impact**: What could an attacker do with this vulnerability?
- **Affected versions**: Which versions of freee-api-go are affected?
- **Steps to reproduce**: Detailed steps to reproduce the vulnerability
- **Proof of concept**: If possible, provide code that demonstrates the issue
- **Suggested fix**: If you have ideas on how to fix it, please share

Example report structure:

```
Subject: [SECURITY] OAuth2 token exposure in logs

Description:
OAuth2 access tokens are being logged in plaintext when debug logging is enabled.

Impact:
Attackers with access to application logs could steal user access tokens and
impersonate users to access their freee accounting data.

Affected versions:
All versions up to and including v0.2.0

Steps to reproduce:
1. Enable debug logging
2. Perform OAuth2 authentication
3. Check logs - access token visible in plaintext

Proof of concept:
[Code sample or log excerpt]

Suggested fix:
Mask sensitive headers in LoggingRoundTripper (already partially implemented
for Authorization header, but may need additional coverage).
```

### Response Timeline

- **Initial response**: Within 48 hours, we'll acknowledge receipt of your report
- **Status update**: Within 5 business days, we'll provide an initial assessment
- **Fix timeline**: We aim to release security fixes within 30 days for critical issues

### What to Expect

1. **Acknowledgment**: We'll confirm we received your report
2. **Investigation**: We'll investigate and validate the vulnerability
3. **Fix development**: We'll develop and test a fix
4. **Coordinated disclosure**: We'll coordinate the release of the fix and public disclosure
5. **Credit**: With your permission, we'll credit you in the security advisory

### Security Advisory Process

When a security vulnerability is confirmed:

1. We'll create a private security advisory on GitHub
2. We'll develop a fix in a private repository fork
3. We'll release a patched version
4. We'll publish the security advisory with details
5. We'll notify users through GitHub releases and README

## Security Best Practices

When using freee-api-go, follow these security best practices:

### 1. Token Management

**DO**:
- Store OAuth2 tokens securely with appropriate file permissions (0600)
- Use the built-in `TokenSource` with automatic refresh
- Invalidate tokens when users log out or revoke access
- Use environment variables or secure vaults for client credentials

**DON'T**:
- Commit tokens to version control
- Store tokens in publicly readable files
- Share tokens between different users or applications
- Log tokens in plaintext

Example:
```go
// GOOD: Use file-based token storage with secure permissions
tokenSource, err := auth.NewCachedTokenSource(ctx, config, token, "token.json")

// BAD: Don't hardcode tokens
token := &oauth2.Token{AccessToken: "hardcoded-token"} // NEVER DO THIS
```

### 2. Client Credentials

**DO**:
- Store `client_id` and `client_secret` in environment variables or secure vaults
- Use different credentials for development, staging, and production
- Rotate credentials periodically
- Restrict OAuth2 scopes to minimum required permissions

**DON'T**:
- Commit credentials to version control
- Share production credentials with developers
- Use production credentials in automated tests
- Grant excessive OAuth2 scopes

Example:
```go
// GOOD: Use environment variables
clientID := os.Getenv("FREEE_CLIENT_ID")
clientSecret := os.Getenv("FREEE_CLIENT_SECRET")

config := auth.NewConfig(
    clientID,
    clientSecret,
    redirectURL,
    []string{"read", "write"}, // Only request needed scopes
)
```

### 3. Secure Communication

**DO**:
- Always use HTTPS endpoints (freee API requires HTTPS)
- Validate TLS certificates (default in Go's http.Client)
- Use the latest stable version of freee-api-go
- Keep dependencies up to date

**DON'T**:
- Disable TLS certificate verification
- Use custom HTTP transports without understanding security implications
- Ignore TLS/SSL warnings

### 4. Input Validation

**DO**:
- Validate all user input before passing to API calls
- Sanitize data displayed in logs
- Use parameterized queries if building any SQL/database layers
- Implement rate limiting to prevent abuse

**DON'T**:
- Trust user input without validation
- Log sensitive user data (personal info, financial data)
- Expose detailed error messages to end users

### 5. Logging

**DO**:
- Use the built-in `LoggingRoundTripper` which masks sensitive headers
- Review logs regularly for suspicious activity
- Implement log rotation and retention policies
- Monitor for unusual API usage patterns

**DON'T**:
- Log full request/response bodies containing sensitive data
- Store logs in publicly accessible locations
- Keep logs indefinitely without review

The `LoggingRoundTripper` automatically masks these sensitive headers:
- `Authorization`
- `Cookie` / `Set-Cookie`
- `X-Api-Key` / `Api-Key`

### 6. Error Handling

**DO**:
- Handle errors gracefully without exposing internal details
- Log errors with context for debugging
- Implement appropriate retry logic for transient failures
- Use structured error types for better error analysis

**DON'T**:
- Return raw error messages to end users
- Ignore errors silently
- Expose stack traces in production

Example:
```go
// GOOD: Handle errors with context
deal, err := dealsService.Get(ctx, companyID, dealID)
if err != nil {
    log.Printf("failed to fetch deal %d: %v", dealID, err)
    return fmt.Errorf("failed to retrieve transaction: %w", err)
}

// BAD: Expose raw errors to users
deal, err := dealsService.Get(ctx, companyID, dealID)
if err != nil {
    return err // Don't expose internal error details
}
```

### 7. Dependency Management

**DO**:
- Regularly update dependencies using `go get -u`
- Monitor security advisories for dependencies
- Use `go mod verify` to ensure dependency integrity
- Review dependency changes before updating

**DON'T**:
- Use outdated dependencies with known vulnerabilities
- Ignore security updates
- Add unnecessary dependencies

Run periodically:
```bash
# Update dependencies
go get -u ./...
go mod tidy

# Check for known vulnerabilities (requires govulncheck)
govulncheck ./...
```

### 8. Rate Limiting

**DO**:
- Use the built-in `RateLimitRoundTripper` to respect API limits
- Implement exponential backoff for retries
- Monitor API usage to avoid hitting limits
- Handle rate limit errors gracefully

**DON'T**:
- Ignore rate limits
- Implement aggressive retry without backoff
- Make unnecessary API calls

Example:
```go
// GOOD: Configure rate limiting
transport := transport.NewTransport(
    transport.WithRateLimit(10, 5),  // 10 req/sec, burst 5
    transport.WithRetry(3, time.Second),
)
```

## Known Security Considerations

### OAuth2 Redirect URI

- Always validate the `state` parameter to prevent CSRF attacks
- Use HTTPS for redirect URIs in production
- Never expose authorization codes or tokens in URLs

### Token Storage

- Tokens are stored with 0600 permissions (owner read/write only)
- Token files should be added to `.gitignore` (already included)
- Consider using OS keychain/credential managers for production

### Concurrent Access

- `CachedTokenSource` is safe for concurrent use
- HTTP clients are safe for concurrent requests
- Be careful with shared state in custom middleware

## Security Updates

Subscribe to security updates:

1. **Watch this repository** on GitHub
2. **Enable security alerts** in your repository settings if using freee-api-go as a dependency
3. **Follow releases** to stay informed about security patches

## Security Contacts

- **Security Issues**: Use GitHub Security Advisories (preferred)
- **General Security Questions**: Open a discussion in GitHub Discussions with the "security" tag

## Acknowledgments

We appreciate the security research community's efforts to responsibly disclose vulnerabilities. Contributors who report valid security issues will be acknowledged in the security advisory (with their permission).

## Additional Resources

- [freee API Security Documentation](https://developer.freee.co.jp/)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [Go Security Policy](https://go.dev/security)
- [OWASP API Security Project](https://owasp.org/www-project-api-security/)

## License

This security policy is part of the freee-api-go project and is licensed under the MIT License.
