package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewUserAgentRoundTripper(t *testing.T) {
	rt := NewUserAgentRoundTripper(nil, "test-app/1.0")

	if rt == nil {
		t.Fatal("NewUserAgentRoundTripper returned nil")
	}

	if rt.userAgent != "test-app/1.0" {
		t.Errorf("userAgent = %q, want %q", rt.userAgent, "test-app/1.0")
	}

	if rt.base == nil {
		t.Fatal("base is nil")
	}
}

func TestUserAgentRoundTripper(t *testing.T) {
	expectedUA := "test-client/2.0.0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != expectedUA {
			t.Errorf("User-Agent = %q, want %q", ua, expectedUA)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := NewUserAgentRoundTripper(http.DefaultTransport, expectedUA)
	client := &http.Client{Transport: rt}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestUserAgentAppend(t *testing.T) {
	customUA := "freee-api-go/0.1.0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")

		// Should contain both user agents
		if !strings.Contains(ua, "custom-ua/1.0") {
			t.Errorf("User-Agent should contain 'custom-ua/1.0', got %q", ua)
		}

		if !strings.Contains(ua, customUA) {
			t.Errorf("User-Agent should contain %q, got %q", customUA, ua)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := NewUserAgentRoundTripper(http.DefaultTransport, customUA)
	client := &http.Client{Transport: rt}

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("User-Agent", "custom-ua/1.0")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestUserAgentEmpty(t *testing.T) {
	expectedUA := "my-app/1.0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != expectedUA {
			t.Errorf("User-Agent = %q, want %q", ua, expectedUA)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := NewUserAgentRoundTripper(http.DefaultTransport, expectedUA)
	client := &http.Client{Transport: rt}

	// Don't set any User-Agent
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
}

func TestDefaultUserAgent(t *testing.T) {
	tests := []struct {
		version  string
		expected string
	}{
		{"0.1.0", "freee-api-go/0.1.0 (+github.com/muno/freee-api-go)"},
		{"1.0.0", "freee-api-go/1.0.0 (+github.com/muno/freee-api-go)"},
		{"", "freee-api-go/dev (+github.com/muno/freee-api-go)"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := DefaultUserAgent(tt.version)
			if result != tt.expected {
				t.Errorf("DefaultUserAgent(%q) = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}

func TestSetBaseUserAgent(t *testing.T) {
	rt := NewUserAgentRoundTripper(nil, "test/1.0")

	newBase := &http.Transport{}
	rt.SetBase(newBase)

	if rt.base != newBase {
		t.Error("SetBase did not update base transport")
	}
}

func TestWithDefaultUserAgent(t *testing.T) {
	version := "0.1.0"
	expectedUA := "freee-api-go/0.1.0 (+github.com/muno/freee-api-go)"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != expectedUA {
			t.Errorf("User-Agent = %q, want %q", ua, expectedUA)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tr := NewTransport(WithDefaultUserAgent(version))
	client := tr.Client()

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
