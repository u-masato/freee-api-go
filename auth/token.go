package auth

import (
	"encoding/json"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// IsTokenValid checks if the token is valid and not expired.
//
// A token is considered valid if:
//   - The AccessToken is not empty
//   - The token has not expired (with a 5-minute buffer)
//
// Returns true if the token is valid, false otherwise.
func IsTokenValid(token *oauth2.Token) bool {
	if token == nil || token.AccessToken == "" {
		return false
	}
	// Consider token invalid if it expires in less than 5 minutes
	return token.Valid() && time.Until(token.Expiry) > 5*time.Minute
}

// NeedsRefresh checks if the token needs to be refreshed.
//
// A token needs refresh if:
//   - It is invalid or expired
//   - It will expire within 5 minutes
//
// Returns true if the token should be refreshed.
func NeedsRefresh(token *oauth2.Token) bool {
	return !IsTokenValid(token)
}

// HasRefreshToken checks if the token has a refresh token.
func HasRefreshToken(token *oauth2.Token) bool {
	return token != nil && token.RefreshToken != ""
}

// SaveTokenToFile saves the OAuth2 token to a JSON file.
//
// This is useful for persisting tokens between application restarts.
// The file should be kept secure as it contains sensitive credentials.
//
// Example:
//
//	err := auth.SaveTokenToFile(token, "token.json")
//	if err != nil {
//	    // Handle error
//	}
func SaveTokenToFile(token *oauth2.Token, filename string) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return &AuthError{
			Op:  "SaveTokenToFile",
			Err: err,
		}
	}

	// Write with restricted permissions (0600 = rw-------)
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return &AuthError{
			Op:  "SaveTokenToFile",
			Err: err,
		}
	}

	return nil
}

// LoadTokenFromFile loads an OAuth2 token from a JSON file.
//
// This is useful for restoring tokens between application restarts.
//
// Example:
//
//	token, err := auth.LoadTokenFromFile("token.json")
//	if err != nil {
//	    // Handle error or initiate new OAuth2 flow
//	}
//	if auth.NeedsRefresh(token) {
//	    token, err = config.TokenSource(ctx, token).Token()
//	}
func LoadTokenFromFile(filename string) (*oauth2.Token, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, &AuthError{
			Op:  "LoadTokenFromFile",
			Err: err,
		}
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, &AuthError{
			Op:  "LoadTokenFromFile",
			Err: err,
		}
	}

	return &token, nil
}

// TokenInfo provides information about a token.
type TokenInfo struct {
	// Valid indicates if the token is currently valid.
	Valid bool

	// Expired indicates if the token has expired.
	Expired bool

	// ExpiresIn is the duration until the token expires.
	// If negative, the token has already expired.
	ExpiresIn time.Duration

	// HasRefreshToken indicates if a refresh token is available.
	HasRefreshToken bool

	// NeedsRefresh indicates if the token should be refreshed.
	NeedsRefresh bool
}

// GetTokenInfo returns detailed information about the token.
func GetTokenInfo(token *oauth2.Token) TokenInfo {
	if token == nil {
		return TokenInfo{
			Valid:           false,
			Expired:         true,
			ExpiresIn:       0,
			HasRefreshToken: false,
			NeedsRefresh:    true,
		}
	}

	expiresIn := time.Until(token.Expiry)
	expired := !token.Expiry.IsZero() && expiresIn <= 0

	return TokenInfo{
		Valid:           IsTokenValid(token),
		Expired:         expired,
		ExpiresIn:       expiresIn,
		HasRefreshToken: HasRefreshToken(token),
		NeedsRefresh:    NeedsRefresh(token),
	}
}
