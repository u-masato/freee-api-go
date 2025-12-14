package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)

	if config == nil {
		t.Fatal("NewConfig returned nil")
	}

	if config.oauth2Config == nil {
		t.Fatal("oauth2Config is nil")
	}

	if config.oauth2Config.ClientID != "test-client-id" {
		t.Errorf("ClientID = %q, want %q", config.oauth2Config.ClientID, "test-client-id")
	}

	if config.oauth2Config.ClientSecret != "test-client-secret" {
		t.Errorf("ClientSecret = %q, want %q", config.oauth2Config.ClientSecret, "test-client-secret")
	}

	if config.oauth2Config.RedirectURL != "http://localhost:8080/callback" {
		t.Errorf("RedirectURL = %q, want %q", config.oauth2Config.RedirectURL, "http://localhost:8080/callback")
	}

	if len(config.oauth2Config.Scopes) != 2 {
		t.Errorf("len(Scopes) = %d, want 2", len(config.oauth2Config.Scopes))
	}

	if config.oauth2Config.Endpoint.AuthURL != AuthURL {
		t.Errorf("AuthURL = %q, want %q", config.oauth2Config.Endpoint.AuthURL, AuthURL)
	}

	if config.oauth2Config.Endpoint.TokenURL != TokenURL {
		t.Errorf("TokenURL = %q, want %q", config.oauth2Config.Endpoint.TokenURL, TokenURL)
	}
}

func TestAuthCodeURL(t *testing.T) {
	config := NewConfig(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)

	state := "test-state-123"
	authURL := config.AuthCodeURL(state)

	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("Failed to parse auth URL: %v", err)
	}

	// Check that the URL points to the correct endpoint
	expectedHost := "accounts.secure.freee.co.jp"
	if u.Host != expectedHost {
		t.Errorf("Host = %q, want %q", u.Host, expectedHost)
	}

	// Check query parameters
	q := u.Query()

	if clientID := q.Get("client_id"); clientID != "test-client-id" {
		t.Errorf("client_id = %q, want %q", clientID, "test-client-id")
	}

	if redirectURI := q.Get("redirect_uri"); redirectURI != "http://localhost:8080/callback" {
		t.Errorf("redirect_uri = %q, want %q", redirectURI, "http://localhost:8080/callback")
	}

	if responseType := q.Get("response_type"); responseType != "code" {
		t.Errorf("response_type = %q, want %q", responseType, "code")
	}

	if stateParam := q.Get("state"); stateParam != state {
		t.Errorf("state = %q, want %q", stateParam, state)
	}

	scope := q.Get("scope")
	if !strings.Contains(scope, "read") || !strings.Contains(scope, "write") {
		t.Errorf("scope = %q, want it to contain 'read' and 'write'", scope)
	}
}

func TestExchange(t *testing.T) {
	// Create a mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("Unexpected request path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if r.Method != "POST" {
			t.Errorf("Unexpected request method: %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Validate request parameters
		if code := r.FormValue("code"); code != "test-auth-code" {
			t.Errorf("code = %q, want %q", code, "test-auth-code")
		}

		if grantType := r.FormValue("grant_type"); grantType != "authorization_code" {
			t.Errorf("grant_type = %q, want %q", grantType, "authorization_code")
		}

		// Return a mock token response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "test-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "test-refresh-token",
		})
	}))
	defer server.Close()

	// Create config with mock server URL
	config := NewConfig(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)
	// Override token endpoint to point to mock server
	config.oauth2Config.Endpoint.TokenURL = server.URL + "/token"

	// Test token exchange
	ctx := context.Background()
	token, err := config.Exchange(ctx, "test-auth-code")
	if err != nil {
		t.Fatalf("Exchange failed: %v", err)
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "test-access-token")
	}

	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("RefreshToken = %q, want %q", token.RefreshToken, "test-refresh-token")
	}

	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", token.TokenType, "Bearer")
	}
}

func TestExchangeError(t *testing.T) {
	// Create a mock OAuth2 server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":             "invalid_grant",
			"error_description": "Invalid authorization code",
		})
	}))
	defer server.Close()

	config := NewConfig(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)
	config.oauth2Config.Endpoint.TokenURL = server.URL + "/token"

	ctx := context.Background()
	_, err := config.Exchange(ctx, "invalid-code")
	if err == nil {
		t.Fatal("Exchange should have failed with invalid code")
	}
}

func TestTokenSource(t *testing.T) {
	config := NewConfig(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)

	ctx := context.Background()

	// This is a simple test to ensure TokenSource doesn't panic
	// Actual token refresh would require a mock server
	ts := config.TokenSource(ctx, nil)
	if ts == nil {
		t.Error("TokenSource returned nil")
	}
}

func TestClient(t *testing.T) {
	config := NewConfig(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
	)

	ctx := context.Background()

	// This is a simple test to ensure Client doesn't panic
	client := config.Client(ctx, nil)
	if client == nil {
		t.Error("Client returned nil")
	}
}
