package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTransport(t *testing.T) {
	tr := NewTransport()

	if tr == nil {
		t.Fatal("NewTransport returned nil")
	}

	if tr.base == nil {
		t.Fatal("base transport is nil")
	}
}

func TestTransportWithOptions(t *testing.T) {
	customBase := &http.Transport{}

	tr := NewTransport(
		WithBaseTransport(customBase),
		WithUserAgent("test-app/1.0"),
	)

	if tr == nil {
		t.Fatal("NewTransport returned nil")
	}

	// Verify that options were applied by making a test request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != "test-app/1.0" {
			t.Errorf("User-Agent = %q, want %q", ua, "test-app/1.0")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

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

func TestChainRoundTrippers(t *testing.T) {
	tests := []struct {
		name string
		rts  []http.RoundTripper
		want int // number of RoundTrippers
	}{
		{
			name: "empty",
			rts:  []http.RoundTripper{},
			want: 0,
		},
		{
			name: "single",
			rts:  []http.RoundTripper{http.DefaultTransport},
			want: 1,
		},
		{
			name: "multiple",
			rts: []http.RoundTripper{
				NewUserAgentRoundTripper(nil, "test/1.0"),
				http.DefaultTransport,
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChainRoundTrippers(tt.rts...)

			if tt.want == 0 {
				if result != http.DefaultTransport {
					t.Error("Expected default transport for empty chain")
				}
			} else if result == nil {
				t.Error("ChainRoundTrippers returned nil")
			}
		})
	}
}

func TestClient(t *testing.T) {
	tr := NewTransport()
	client := tr.Client()

	if client == nil {
		t.Fatal("Client returned nil")
	}

	if client.Transport != tr {
		t.Error("Client transport is not the Transport instance")
	}
}
