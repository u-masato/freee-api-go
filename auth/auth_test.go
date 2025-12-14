package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// Test token validation
func TestIsTokenValid(t *testing.T) {
	tests := []struct {
		name  string
		token *oauth2.Token
		want  bool
	}{
		{
			name:  "nil token",
			token: nil,
			want:  false,
		},
		{
			name: "empty access token",
			token: &oauth2.Token{
				AccessToken: "",
			},
			want: false,
		},
		{
			name: "expired token",
			token: &oauth2.Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(-1 * time.Hour),
			},
			want: false,
		},
		{
			name: "token expiring soon (4 minutes)",
			token: &oauth2.Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(4 * time.Minute),
			},
			want: false,
		},
		{
			name: "valid token (10 minutes remaining)",
			token: &oauth2.Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(10 * time.Minute),
			},
			want: true,
		},
		{
			name: "valid token (1 hour remaining)",
			token: &oauth2.Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(1 * time.Hour),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTokenValid(tt.token)
			if got != tt.want {
				t.Errorf("IsTokenValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNeedsRefresh(t *testing.T) {
	validToken := &oauth2.Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	invalidToken := &oauth2.Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(-1 * time.Hour),
	}

	if NeedsRefresh(validToken) {
		t.Error("NeedsRefresh should return false for valid token")
	}

	if !NeedsRefresh(invalidToken) {
		t.Error("NeedsRefresh should return true for invalid token")
	}
}

func TestHasRefreshToken(t *testing.T) {
	tests := []struct {
		name  string
		token *oauth2.Token
		want  bool
	}{
		{
			name:  "nil token",
			token: nil,
			want:  false,
		},
		{
			name: "no refresh token",
			token: &oauth2.Token{
				AccessToken: "test-token",
			},
			want: false,
		},
		{
			name: "has refresh token",
			token: &oauth2.Token{
				AccessToken:  "test-token",
				RefreshToken: "test-refresh-token",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasRefreshToken(tt.token)
			if got != tt.want {
				t.Errorf("HasRefreshToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveAndLoadTokenFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "token.json")

	// Create a test token
	originalToken := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour).Round(time.Second),
	}

	// Save token to file
	err := SaveTokenToFile(originalToken, filename)
	if err != nil {
		t.Fatalf("SaveTokenToFile failed: %v", err)
	}

	// Check that file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatal("Token file was not created")
	}

	// Load token from file
	loadedToken, err := LoadTokenFromFile(filename)
	if err != nil {
		t.Fatalf("LoadTokenFromFile failed: %v", err)
	}

	// Compare tokens
	if loadedToken.AccessToken != originalToken.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loadedToken.AccessToken, originalToken.AccessToken)
	}

	if loadedToken.RefreshToken != originalToken.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loadedToken.RefreshToken, originalToken.RefreshToken)
	}

	if loadedToken.TokenType != originalToken.TokenType {
		t.Errorf("TokenType = %q, want %q", loadedToken.TokenType, originalToken.TokenType)
	}

	// Expiry comparison (allowing for small differences due to rounding)
	expiryDiff := loadedToken.Expiry.Sub(originalToken.Expiry).Abs()
	if expiryDiff > time.Second {
		t.Errorf("Expiry difference too large: %v", expiryDiff)
	}
}

func TestLoadTokenFromFile_NotFound(t *testing.T) {
	_, err := LoadTokenFromFile("/nonexistent/path/token.json")
	if err == nil {
		t.Error("LoadTokenFromFile should fail for nonexistent file")
	}

	if !IsAuthError(err) {
		t.Error("Error should be an AuthError")
	}
}

func TestSaveTokenToFile_InvalidPath(t *testing.T) {
	token := &oauth2.Token{
		AccessToken: "test-token",
	}

	err := SaveTokenToFile(token, "/nonexistent/path/token.json")
	if err == nil {
		t.Error("SaveTokenToFile should fail for invalid path")
	}

	if !IsAuthError(err) {
		t.Error("Error should be an AuthError")
	}
}

func TestGetTokenInfo(t *testing.T) {
	tests := []struct {
		name  string
		token *oauth2.Token
		want  TokenInfo
	}{
		{
			name:  "nil token",
			token: nil,
			want: TokenInfo{
				Valid:           false,
				Expired:         true,
				ExpiresIn:       0,
				HasRefreshToken: false,
				NeedsRefresh:    true,
			},
		},
		{
			name: "valid token with refresh token",
			token: &oauth2.Token{
				AccessToken:  "test-token",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(1 * time.Hour),
			},
			want: TokenInfo{
				Valid:           true,
				Expired:         false,
				HasRefreshToken: true,
				NeedsRefresh:    false,
			},
		},
		{
			name: "expired token",
			token: &oauth2.Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(-1 * time.Hour),
			},
			want: TokenInfo{
				Valid:           false,
				Expired:         true,
				HasRefreshToken: false,
				NeedsRefresh:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTokenInfo(tt.token)

			if got.Valid != tt.want.Valid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.want.Valid)
			}

			if got.Expired != tt.want.Expired {
				t.Errorf("Expired = %v, want %v", got.Expired, tt.want.Expired)
			}

			if got.HasRefreshToken != tt.want.HasRefreshToken {
				t.Errorf("HasRefreshToken = %v, want %v", got.HasRefreshToken, tt.want.HasRefreshToken)
			}

			if got.NeedsRefresh != tt.want.NeedsRefresh {
				t.Errorf("NeedsRefresh = %v, want %v", got.NeedsRefresh, tt.want.NeedsRefresh)
			}
		})
	}
}

// Test CachedTokenSource
func TestCachedTokenSource(t *testing.T) {
	callCount := 0
	validToken := &oauth2.Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	// Create a mock token source
	mockSource := mockTokenSource{
		tokenFunc: func() (*oauth2.Token, error) {
			callCount++
			return validToken, nil
		},
	}

	ts := NewCachedTokenSource(&mockSource, validToken, "")

	// First call should use cached token
	token1, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() failed: %v", err)
	}

	if callCount != 0 {
		t.Errorf("Token source was called %d times, want 0 (should use cache)", callCount)
	}

	// Invalidate cache
	ts.Invalidate()

	// Second call should fetch new token
	token2, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Token source was called %d times, want 1", callCount)
	}

	if token1.AccessToken != token2.AccessToken {
		t.Error("Tokens should be the same")
	}
}

func TestCachedTokenSource_AutoRefresh(t *testing.T) {
	expiredToken := &oauth2.Token{
		AccessToken: "expired-token",
		Expiry:      time.Now().Add(-1 * time.Hour),
	}

	newToken := &oauth2.Token{
		AccessToken: "new-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	callCount := 0
	mockSource := mockTokenSource{
		tokenFunc: func() (*oauth2.Token, error) {
			callCount++
			return newToken, nil
		},
	}

	ts := NewCachedTokenSource(&mockSource, expiredToken, "")

	// Should automatically refresh expired token
	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Token source was called %d times, want 1", callCount)
	}

	if token.AccessToken != "new-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "new-token")
	}
}

func TestCachedTokenSource_WithFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "token.json")

	newToken := &oauth2.Token{
		AccessToken: "new-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	mockSource := mockTokenSource{
		tokenFunc: func() (*oauth2.Token, error) {
			return newToken, nil
		},
	}

	ts := NewCachedTokenSource(&mockSource, nil, filename)

	// Get token (should save to file)
	_, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() failed: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Token file was not created")
	}

	// Load token from file and verify
	loadedToken, err := LoadTokenFromFile(filename)
	if err != nil {
		t.Fatalf("LoadTokenFromFile failed: %v", err)
	}

	if loadedToken.AccessToken != "new-token" {
		t.Errorf("AccessToken = %q, want %q", loadedToken.AccessToken, "new-token")
	}
}

func TestReuseTokenSourceWithCallback(t *testing.T) {
	callbackCalled := false
	var callbackToken *oauth2.Token

	callback := func(token *oauth2.Token) error {
		callbackCalled = true
		callbackToken = token
		return nil
	}

	newToken := &oauth2.Token{
		AccessToken: "new-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	mockSource := mockTokenSource{
		tokenFunc: func() (*oauth2.Token, error) {
			return newToken, nil
		},
	}

	ts := NewReuseTokenSourceWithCallback(&mockSource, nil, callback)

	// Get token
	token, err := ts.Token()
	if err != nil {
		t.Fatalf("Token() failed: %v", err)
	}

	if !callbackCalled {
		t.Error("Callback was not called")
	}

	if callbackToken == nil {
		t.Error("Callback token is nil")
	}

	if callbackToken.AccessToken != token.AccessToken {
		t.Error("Callback received different token")
	}
}

// Test error types
func TestAuthError(t *testing.T) {
	err := &AuthError{
		Op:          "Exchange",
		Err:         errors.New("network error"),
		Code:        "invalid_grant",
		Description: "Invalid authorization code",
	}

	errStr := err.Error()
	if errStr != "Exchange: invalid_grant: Invalid authorization code" {
		t.Errorf("Error() = %q, want %q", errStr, "Exchange: invalid_grant: Invalid authorization code")
	}

	// Test Unwrap
	unwrapped := err.Unwrap()
	if unwrapped.Error() != "network error" {
		t.Errorf("Unwrap() = %q, want %q", unwrapped.Error(), "network error")
	}
}

func TestIsAuthError(t *testing.T) {
	authErr := &AuthError{
		Op:  "Test",
		Err: errors.New("test error"),
	}

	if !IsAuthError(authErr) {
		t.Error("IsAuthError should return true for AuthError")
	}

	regularErr := errors.New("regular error")
	if IsAuthError(regularErr) {
		t.Error("IsAuthError should return false for regular error")
	}
}

func TestIsTokenExpiredError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrInvalidToken",
			err:  ErrInvalidToken,
			want: true,
		},
		{
			name: "AuthError with invalid_token code",
			err: &AuthError{
				Op:   "Test",
				Code: "invalid_token",
			},
			want: true,
		},
		{
			name: "AuthError with token_expired code",
			err: &AuthError{
				Op:   "Test",
				Code: "token_expired",
			},
			want: true,
		},
		{
			name: "Other error",
			err:  errors.New("other error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTokenExpiredError(tt.err)
			if got != tt.want {
				t.Errorf("IsTokenExpiredError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInvalidGrantError(t *testing.T) {
	invalidGrantErr := &AuthError{
		Op:   "Exchange",
		Code: "invalid_grant",
	}

	if !IsInvalidGrantError(invalidGrantErr) {
		t.Error("IsInvalidGrantError should return true for invalid_grant error")
	}

	otherErr := errors.New("other error")
	if IsInvalidGrantError(otherErr) {
		t.Error("IsInvalidGrantError should return false for other error")
	}
}

// Mock token source for testing
type mockTokenSource struct {
	tokenFunc func() (*oauth2.Token, error)
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return m.tokenFunc()
}

// Test full OAuth2 flow with mock server
func TestFullOAuth2Flow(t *testing.T) {
	// Create a mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "test-access-token",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"refresh_token": "test-refresh-token",
			})
		}
	}))
	defer server.Close()

	// Create config
	config := NewConfig(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)
	config.oauth2Config.Endpoint.TokenURL = server.URL + "/token"

	// Generate auth URL
	authURL := config.AuthCodeURL("test-state")
	if authURL == "" {
		t.Fatal("AuthCodeURL returned empty string")
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := config.Exchange(ctx, "test-code")
	if err != nil {
		t.Fatalf("Exchange failed: %v", err)
	}

	// Verify token
	if token.AccessToken != "test-access-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "test-access-token")
	}

	// Create token source
	ts := config.TokenSource(ctx, token)
	if ts == nil {
		t.Fatal("TokenSource returned nil")
	}

	// Create HTTP client
	client := config.Client(ctx, token)
	if client == nil {
		t.Fatal("Client returned nil")
	}
}
