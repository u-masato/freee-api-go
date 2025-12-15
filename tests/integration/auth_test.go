package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/muno/freee-api-go/auth"
	"golang.org/x/oauth2"
)

// TestOAuth2Flow tests the complete OAuth2 authorization code flow.
func TestOAuth2Flow(t *testing.T) {
	// Create a mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/authorize":
			// Validate authorization request
			query := r.URL.Query()
			if query.Get("client_id") == "" {
				t.Error("Missing client_id in authorization request")
			}
			if query.Get("redirect_uri") == "" {
				t.Error("Missing redirect_uri in authorization request")
			}
			if query.Get("response_type") != "code" {
				t.Errorf("Expected response_type=code, got %s", query.Get("response_type"))
			}
			state := query.Get("state")
			if state == "" {
				t.Error("Missing state parameter")
			}
			// In a real test, we would redirect to the callback URL
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Authorization page"))

		case "/token":
			// Handle token exchange
			if err := r.ParseForm(); err != nil {
				t.Errorf("Failed to parse form: %v", err)
				return
			}

			grantType := r.FormValue("grant_type")
			switch grantType {
			case "authorization_code":
				code := r.FormValue("code")
				if code == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "invalid_request", "error_description": "Missing code"}`))
					return
				}
				// Return a mock token
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"access_token": "test-access-token-123",
					"token_type": "Bearer",
					"expires_in": 86400,
					"refresh_token": "test-refresh-token-456",
					"scope": "read write",
					"created_at": 1700000000
				}`))

			case "refresh_token":
				refreshToken := r.FormValue("refresh_token")
				if refreshToken == "" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "invalid_request", "error_description": "Missing refresh_token"}`))
					return
				}
				// Return a new mock token
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"access_token": "test-access-token-refreshed-789",
					"token_type": "Bearer",
					"expires_in": 86400,
					"refresh_token": "test-refresh-token-new-012",
					"scope": "read write",
					"created_at": 1700000001
				}`))

			default:
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "unsupported_grant_type"}`))
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test AuthCodeURL generation
	t.Run("AuthCodeURL", func(t *testing.T) {
		config := createTestConfig(server.URL)
		state := "test-csrf-state"
		url := config.AuthCodeURL(state)

		if url == "" {
			t.Error("AuthCodeURL returned empty string")
		}
		// Verify the URL contains expected parameters
		if !containsParam(url, "client_id", "test-client-id") {
			t.Error("URL missing client_id")
		}
		if !containsParam(url, "response_type", "code") {
			t.Error("URL missing response_type=code")
		}
		if !containsParam(url, "state", state) {
			t.Error("URL missing state parameter")
		}
	})

	// Test token exchange
	t.Run("Exchange", func(t *testing.T) {
		config := createTestConfig(server.URL)
		ctx := context.Background()

		token, err := config.Exchange(ctx, "test-auth-code")
		if err != nil {
			t.Fatalf("Exchange failed: %v", err)
		}

		if token.AccessToken != "test-access-token-123" {
			t.Errorf("Expected access_token 'test-access-token-123', got %s", token.AccessToken)
		}
		if token.RefreshToken != "test-refresh-token-456" {
			t.Errorf("Expected refresh_token 'test-refresh-token-456', got %s", token.RefreshToken)
		}
		if token.TokenType != "Bearer" {
			t.Errorf("Expected token_type 'Bearer', got %s", token.TokenType)
		}
	})

	// Test token refresh via TokenSource
	t.Run("TokenSource_Refresh", func(t *testing.T) {
		config := createTestConfig(server.URL)
		ctx := context.Background()

		// Create an expired token
		expiredToken := &oauth2.Token{
			AccessToken:  "expired-token",
			TokenType:    "Bearer",
			RefreshToken: "test-refresh-token-456",
			Expiry:       time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		}

		ts := config.TokenSource(ctx, expiredToken)
		newToken, err := ts.Token()
		if err != nil {
			t.Fatalf("TokenSource.Token() failed: %v", err)
		}

		// The token should have been refreshed
		if newToken.AccessToken == "expired-token" {
			t.Error("Expected token to be refreshed, but got the old token")
		}
		if newToken.AccessToken != "test-access-token-refreshed-789" {
			t.Errorf("Expected refreshed access_token, got %s", newToken.AccessToken)
		}
	})

	// Test HTTP client creation
	t.Run("Client", func(t *testing.T) {
		config := createTestConfig(server.URL)
		ctx := context.Background()

		token := &oauth2.Token{
			AccessToken:  "valid-access-token",
			TokenType:    "Bearer",
			RefreshToken: "test-refresh-token",
			Expiry:       time.Now().Add(1 * time.Hour), // Valid for 1 hour
		}

		httpClient := config.Client(ctx, token)
		if httpClient == nil {
			t.Error("Client returned nil")
		}

		// Verify the client adds Authorization header
		// Note: The actual Authorization header is added by the oauth2 transport
		// when making requests, so we can't easily test it without a mock API
		if httpClient.Transport == nil {
			t.Error("Expected client to have a custom transport")
		}
	})
}

// TestOAuth2ErrorHandling tests error handling in the OAuth2 flow.
func TestOAuth2ErrorHandling(t *testing.T) {
	t.Run("InvalidCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"error": "invalid_grant",
				"error_description": "The authorization code has expired or is invalid"
			}`))
		}))
		defer server.Close()

		config := createTestConfig(server.URL)
		ctx := context.Background()

		_, err := config.Exchange(ctx, "invalid-code")
		if err == nil {
			t.Error("Expected error for invalid code, got nil")
		}
	})

	t.Run("InvalidRefreshToken", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"error": "invalid_grant",
				"error_description": "The refresh token is invalid or has been revoked"
			}`))
		}))
		defer server.Close()

		config := createTestConfig(server.URL)
		ctx := context.Background()

		expiredToken := &oauth2.Token{
			AccessToken:  "expired-token",
			TokenType:    "Bearer",
			RefreshToken: "invalid-refresh-token",
			Expiry:       time.Now().Add(-1 * time.Hour),
		}

		ts := config.TokenSource(ctx, expiredToken)
		_, err := ts.Token()
		if err == nil {
			t.Error("Expected error for invalid refresh token, got nil")
		}
	})

	t.Run("NetworkError", func(t *testing.T) {
		// Use a server that's already closed
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		serverURL := server.URL
		server.Close()

		config := createTestConfig(serverURL)
		ctx := context.Background()

		_, err := config.Exchange(ctx, "test-code")
		if err == nil {
			t.Error("Expected network error, got nil")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := createTestConfig(server.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := config.Exchange(ctx, "test-code")
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

// createTestConfig creates an OAuth2 config for testing with custom endpoints.
func createTestConfig(serverURL string) *auth.Config {
	return auth.NewConfigWithEndpoint(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8080/callback",
		[]string{"read", "write"},
		serverURL+"/authorize",
		serverURL+"/token",
	)
}

// containsParam checks if a URL contains a query parameter with a specific value.
func containsParam(url, key, value string) bool {
	return containsString(url, key+"="+value)
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

// findSubstring searches for a substring in a string.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
