package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func TestNewClient(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		c := NewClient()

		if c.baseURL != DefaultBaseURL {
			t.Errorf("expected baseURL %s, got %s", DefaultBaseURL, c.baseURL)
		}

		if c.userAgent != DefaultUserAgent {
			t.Errorf("expected userAgent %s, got %s", DefaultUserAgent, c.userAgent)
		}

		if c.httpClient == nil {
			t.Error("expected httpClient to be non-nil")
		}

		if c.ctx != context.Background() {
			t.Error("expected ctx to be context.Background()")
		}

		if c.tokenSource != nil {
			t.Error("expected tokenSource to be nil by default")
		}
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{}
		c := NewClient(WithHTTPClient(customClient))

		if c.httpClient != customClient {
			t.Error("expected custom HTTP client to be used")
		}
	})

	t.Run("with custom base URL", func(t *testing.T) {
		customURL := "https://test.example.com"
		c := NewClient(WithBaseURL(customURL))

		if c.baseURL != customURL {
			t.Errorf("expected baseURL %s, got %s", customURL, c.baseURL)
		}
	})

	t.Run("with custom user agent", func(t *testing.T) {
		customUA := "test-app/1.0.0"
		c := NewClient(WithUserAgent(customUA))

		if c.userAgent != customUA {
			t.Errorf("expected userAgent %s, got %s", customUA, c.userAgent)
		}
	})

	t.Run("with context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		c := NewClient(WithContext(ctx))

		if c.ctx != ctx {
			t.Error("expected custom context to be used")
		}
	})

	t.Run("with token source", func(t *testing.T) {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})
		c := NewClient(WithTokenSource(ts))

		if c.tokenSource != ts {
			t.Error("expected token source to be set")
		}

		// When token source is provided without custom HTTP client,
		// an OAuth2 client should be created
		if c.httpClient == http.DefaultClient {
			t.Error("expected OAuth2 HTTP client to be created")
		}
	})

	t.Run("with token source and custom HTTP client", func(t *testing.T) {
		customClient := &http.Client{}
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})
		c := NewClient(
			WithHTTPClient(customClient),
			WithTokenSource(ts),
		)

		if c.httpClient != customClient {
			t.Error("expected custom HTTP client to be preserved")
		}

		if c.tokenSource != ts {
			t.Error("expected token source to be set")
		}
	})

	t.Run("with multiple options", func(t *testing.T) {
		customClient := &http.Client{}
		customURL := "https://test.example.com"
		customUA := "test-app/1.0.0"
		ctx := context.WithValue(context.Background(), "key", "value")

		c := NewClient(
			WithHTTPClient(customClient),
			WithBaseURL(customURL),
			WithUserAgent(customUA),
			WithContext(ctx),
		)

		if c.httpClient != customClient {
			t.Error("expected custom HTTP client to be used")
		}
		if c.baseURL != customURL {
			t.Errorf("expected baseURL %s, got %s", customURL, c.baseURL)
		}
		if c.userAgent != customUA {
			t.Errorf("expected userAgent %s, got %s", customUA, c.userAgent)
		}
		if c.ctx != ctx {
			t.Error("expected custom context to be used")
		}
	})
}

func TestClient_Getters(t *testing.T) {
	customClient := &http.Client{}
	customURL := "https://test.example.com"
	customUA := "test-app/1.0.0"
	ctx := context.WithValue(context.Background(), "key", "value")

	c := NewClient(
		WithHTTPClient(customClient),
		WithBaseURL(customURL),
		WithUserAgent(customUA),
		WithContext(ctx),
	)

	if c.HTTPClient() != customClient {
		t.Error("HTTPClient() should return the configured HTTP client")
	}

	if c.BaseURL() != customURL {
		t.Errorf("BaseURL() expected %s, got %s", customURL, c.BaseURL())
	}

	if c.UserAgent() != customUA {
		t.Errorf("UserAgent() expected %s, got %s", customUA, c.UserAgent())
	}

	if c.Context() != ctx {
		t.Error("Context() should return the configured context")
	}
}

func TestClient_Do(t *testing.T) {
	t.Run("adds User-Agent header", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if ua != DefaultUserAgent {
				t.Errorf("expected User-Agent %s, got %s", DefaultUserAgent, ua)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(WithBaseURL(server.URL))

		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("preserves existing User-Agent header", func(t *testing.T) {
		customUA := "custom-user-agent/1.0"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if ua != customUA {
				t.Errorf("expected User-Agent %s, got %s", customUA, ua)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(WithBaseURL(server.URL))

		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("User-Agent", customUA)

		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("custom user agent from options", func(t *testing.T) {
		customUA := "my-app/2.0.0"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ua := r.Header.Get("User-Agent")
			if ua != customUA {
				t.Errorf("expected User-Agent %s, got %s", customUA, ua)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient(
			WithBaseURL(server.URL),
			WithUserAgent(customUA),
		)

		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("with token source", func(t *testing.T) {
		expectedToken := "test-access-token"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			expectedAuth := "Bearer " + expectedToken
			if auth != expectedAuth {
				t.Errorf("expected Authorization %s, got %s", expectedAuth, auth)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: expectedToken})
		c := NewClient(
			WithBaseURL(server.URL),
			WithTokenSource(ts),
		)

		req, err := http.NewRequest("GET", server.URL+"/test", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}
