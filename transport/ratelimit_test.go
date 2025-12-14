package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRateLimitRoundTripper(t *testing.T) {
	rt := NewRateLimitRoundTripper(nil, 10.0, 5)

	if rt == nil {
		t.Fatal("NewRateLimitRoundTripper returned nil")
	}

	if rt.limiter == nil {
		t.Fatal("limiter is nil")
	}

	if rt.base == nil {
		t.Fatal("base is nil")
	}
}

func TestRateLimitRoundTripper(t *testing.T) {
	// Create a test server
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create rate limiter: 5 requests per second, burst of 2
	rt := NewRateLimitRoundTripper(http.DefaultTransport, 5.0, 2)

	client := &http.Client{Transport: rt}

	// Make 3 requests - first 2 should be instant (burst), 3rd should be delayed
	start := time.Now()

	for i := 0; i < 3; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		resp.Body.Close()
	}

	duration := time.Since(start)

	// First 2 requests should be fast (burst)
	// 3rd request should wait ~200ms (1/5 second)
	// Allow some tolerance
	if duration < 100*time.Millisecond {
		t.Errorf("Rate limiting not working: completed too quickly (%v)", duration)
	}

	if requestCount != 3 {
		t.Errorf("Request count = %d, want 3", requestCount)
	}
}

func TestRateLimitContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Very low rate to ensure blocking
	rt := NewRateLimitRoundTripper(http.DefaultTransport, 0.1, 1)

	client := &http.Client{Transport: rt}

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// First request should succeed
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	resp.Body.Close()

	// Second request should fail due to context timeout
	req2, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Do(req2)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

func TestSetBaseRateLimit(t *testing.T) {
	rt := NewRateLimitRoundTripper(nil, 10.0, 5)

	newBase := &http.Transport{}
	rt.SetBase(newBase)

	if rt.base != newBase {
		t.Error("SetBase did not update base transport")
	}
}
