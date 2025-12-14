package auth

import (
	"errors"
	"fmt"
)

// Common authentication errors.
var (
	// ErrInvalidToken is returned when the access token is invalid or expired.
	ErrInvalidToken = errors.New("invalid or expired access token")

	// ErrInvalidRefreshToken is returned when the refresh token is invalid or expired.
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

	// ErrInvalidAuthorizationCode is returned when the authorization code is invalid.
	ErrInvalidAuthorizationCode = errors.New("invalid authorization code")

	// ErrMissingClientCredentials is returned when client ID or secret is missing.
	ErrMissingClientCredentials = errors.New("missing client ID or client secret")

	// ErrInvalidRedirectURL is returned when the redirect URL doesn't match.
	ErrInvalidRedirectURL = errors.New("redirect URL mismatch")

	// ErrInvalidScope is returned when requested scope is invalid.
	ErrInvalidScope = errors.New("invalid scope")

	// ErrStateMismatch is returned when the state parameter doesn't match.
	ErrStateMismatch = errors.New("state parameter mismatch")
)

// AuthError represents an authentication error with additional context.
type AuthError struct {
	// Op is the operation that failed (e.g., "Exchange", "RefreshToken").
	Op string

	// Err is the underlying error.
	Err error

	// Code is the OAuth2 error code if available (e.g., "invalid_grant").
	Code string

	// Description is the human-readable error description.
	Description string
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s: %s", e.Op, e.Code, e.Description)
	}
	if e.Code != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Code, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// IsAuthError checks if the error is an AuthError.
func IsAuthError(err error) bool {
	var authErr *AuthError
	return errors.As(err, &authErr)
}

// IsTokenExpiredError checks if the error indicates an expired token.
func IsTokenExpiredError(err error) bool {
	if errors.Is(err, ErrInvalidToken) {
		return true
	}
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == "invalid_token" || authErr.Code == "token_expired"
	}
	return false
}

// IsInvalidGrantError checks if the error is an invalid_grant error.
func IsInvalidGrantError(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == "invalid_grant"
	}
	return false
}
