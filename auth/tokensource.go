package auth

import (
	"context"
	"sync"

	"golang.org/x/oauth2"
)

// CachedTokenSource is a TokenSource that caches tokens and optionally
// persists them to a file.
//
// It wraps an underlying TokenSource and caches the token in memory.
// If a filename is provided, it also saves the token to disk whenever
// a new token is obtained.
type CachedTokenSource struct {
	src      oauth2.TokenSource
	mu       sync.Mutex
	token    *oauth2.Token
	filename string
}

// NewCachedTokenSource creates a new CachedTokenSource.
//
// Parameters:
//   - src: The underlying TokenSource to wrap
//   - token: The initial token (can be nil)
//   - filename: Optional file path to persist tokens (empty string to disable)
//
// Example:
//
//	ts := auth.NewCachedTokenSource(
//	    config.TokenSource(ctx, token),
//	    token,
//	    "token.json",
//	)
//	newToken, err := ts.Token()
func NewCachedTokenSource(src oauth2.TokenSource, token *oauth2.Token, filename string) *CachedTokenSource {
	return &CachedTokenSource{
		src:      src,
		token:    token,
		filename: filename,
	}
}

// Token returns a valid token, refreshing it if necessary.
//
// This method is safe for concurrent use.
func (s *CachedTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return cached token if it's still valid
	if IsTokenValid(s.token) {
		return s.token, nil
	}

	// Get new token from underlying source
	token, err := s.src.Token()
	if err != nil {
		return nil, err
	}

	// Update cache
	s.token = token

	// Save to file if filename is specified
	if s.filename != "" {
		// Ignore save errors to avoid breaking the flow
		// The application can still use the token even if save fails
		_ = SaveTokenToFile(token, s.filename)
	}

	return token, nil
}

// Invalidate clears the cached token, forcing a refresh on next Token() call.
func (s *CachedTokenSource) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = nil
}

// GetCachedToken returns the currently cached token without refreshing it.
//
// Returns nil if no token is cached or if the cached token is invalid.
func (s *CachedTokenSource) GetCachedToken() *oauth2.Token {
	s.mu.Lock()
	defer s.mu.Unlock()

	if IsTokenValid(s.token) {
		return s.token
	}
	return nil
}

// StaticTokenSource returns a TokenSource that always returns the same token.
//
// This is useful for testing or when using a long-lived access token.
//
// Example:
//
//	ts := auth.StaticTokenSource(&oauth2.Token{
//	    AccessToken: "your-access-token",
//	})
func StaticTokenSource(token *oauth2.Token) oauth2.TokenSource {
	return oauth2.StaticTokenSource(token)
}

// ReuseTokenSourceWithCallback returns a TokenSource that reuses a single token
// and calls a callback function whenever the token is refreshed.
//
// This is useful for applications that need to be notified when tokens change,
// for example to persist them to a database.
type ReuseTokenSourceWithCallback struct {
	src      oauth2.TokenSource
	mu       sync.Mutex
	token    *oauth2.Token
	callback func(*oauth2.Token) error
}

// NewReuseTokenSourceWithCallback creates a new ReuseTokenSourceWithCallback.
//
// The callback function is called whenever a new token is obtained.
// If the callback returns an error, it is logged but does not prevent
// the token from being returned.
//
// Example:
//
//	ts := auth.NewReuseTokenSourceWithCallback(
//	    config.TokenSource(ctx, token),
//	    token,
//	    func(t *oauth2.Token) error {
//	        return saveTokenToDatabase(t)
//	    },
//	)
func NewReuseTokenSourceWithCallback(src oauth2.TokenSource, token *oauth2.Token, callback func(*oauth2.Token) error) *ReuseTokenSourceWithCallback {
	return &ReuseTokenSourceWithCallback{
		src:      src,
		token:    token,
		callback: callback,
	}
}

// Token returns a valid token, refreshing it if necessary.
func (s *ReuseTokenSourceWithCallback) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return cached token if it's still valid
	if IsTokenValid(s.token) {
		return s.token, nil
	}

	// Get new token from underlying source
	token, err := s.src.Token()
	if err != nil {
		return nil, err
	}

	// Update cache
	s.token = token

	// Call callback if provided
	if s.callback != nil {
		// Ignore callback errors to avoid breaking the flow
		_ = s.callback(token)
	}

	return token, nil
}

// TokenSourceWithContext creates a TokenSource from a Config that uses the provided context.
//
// This is a convenience function that combines Config.TokenSource with context.
func TokenSourceWithContext(ctx context.Context, config *Config, token *oauth2.Token) oauth2.TokenSource {
	return config.TokenSource(ctx, token)
}
