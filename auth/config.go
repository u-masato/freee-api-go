// Package auth provides OAuth2 authentication functionality for the freee API.
//
// This package implements the OAuth2 Authorization Code Grant flow
// required to access the freee API. It handles:
//   - Authorization URL generation
//   - Access token acquisition
//   - Refresh token management
//   - TokenSource implementation for automatic token refresh
package auth

import (
	"context"

	"golang.org/x/oauth2"
)

const (
	// AuthURL is the freee OAuth2 authorization endpoint.
	AuthURL = "https://accounts.secure.freee.co.jp/public_api/authorize"

	// TokenURL is the freee OAuth2 token endpoint.
	TokenURL = "https://accounts.secure.freee.co.jp/public_api/token"
)

// Config holds OAuth2 configuration for the freee API.
//
// Example usage:
//
//	config := auth.NewConfig(
//	    "your-client-id",
//	    "your-client-secret",
//	    "http://localhost:8080/callback",
//	    []string{"read", "write"},
//	)
type Config struct {
	oauth2Config *oauth2.Config
}

// NewConfig creates a new OAuth2 configuration for the freee API.
//
// Parameters:
//   - clientID: OAuth2 client ID obtained from freee developer portal
//   - clientSecret: OAuth2 client secret obtained from freee developer portal
//   - redirectURL: Callback URL registered in freee application settings
//   - scopes: List of permission scopes (e.g., []string{"read", "write"})
//
// Returns a Config instance that can be used to generate authorization URLs
// and exchange authorization codes for access tokens.
func NewConfig(clientID, clientSecret, redirectURL string, scopes []string) *Config {
	return &Config{
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  AuthURL,
				TokenURL: TokenURL,
			},
		},
	}
}

// AuthCodeURL generates the authorization URL for the OAuth2 flow.
//
// The state parameter should be a random string to prevent CSRF attacks.
// The user should be redirected to this URL to authorize the application.
//
// Example:
//
//	state := generateRandomState() // Your CSRF token generation
//	url := config.AuthCodeURL(state)
//	// Redirect user to url
func (c *Config) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return c.oauth2Config.AuthCodeURL(state, opts...)
}

// Exchange exchanges an authorization code for an access token.
//
// This should be called in the OAuth2 callback handler after the user
// has authorized the application. The code parameter is obtained from
// the callback URL query parameter.
//
// Example:
//
//	token, err := config.Exchange(ctx, code)
//	if err != nil {
//	    // Handle error
//	}
//	// Use token.AccessToken to make API requests
func (c *Config) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return c.oauth2Config.Exchange(ctx, code, opts...)
}

// TokenSource creates a TokenSource that automatically refreshes tokens.
//
// The returned TokenSource will automatically refresh the access token
// when it expires, using the refresh token if available.
//
// Example:
//
//	ts := config.TokenSource(ctx, token)
//	client := oauth2.NewClient(ctx, ts)
//	// Use client for HTTP requests
func (c *Config) TokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return c.oauth2Config.TokenSource(ctx, token)
}

// Client creates an HTTP client using the provided token.
//
// The returned client will automatically include the access token in
// the Authorization header and refresh it when necessary.
//
// Example:
//
//	httpClient := config.Client(ctx, token)
//	resp, err := httpClient.Get("https://api.freee.co.jp/...")
func (c *Config) Client(ctx context.Context, token *oauth2.Token) *oauth2.HTTPClient {
	return c.oauth2Config.Client(ctx, token)
}
