package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRetryRoundTripper(t *testing.T) {
	rt := NewRetryRoundTripper(nil, 3, time.Second)

	if rt == nil {
		t.Fatal("NewRetryRoundTripper returned nil")
	}

	if rt.maxRetries != 3 {
		t.Errorf("maxRetries = %d, want 3", rt.maxRetries)
	}

	if rt.initialDelay != time.Second {
		t.Errorf("initialDelay = %v, want %v", rt.initialDelay, time.Second)
	}

	if rt.base == nil {
		t.Fatal("base is nil")
	}
}

func TestRetryRoundTripperSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := NewRetryRoundTripper(http.DefaultTransport, 3, 100*time.Millisecond)
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

func TestRetryRoundTripperRetries(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// First 2 attempts fail with 503
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			// 3rd attempt succeeds
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	rt := NewRetryRoundTripper(http.DefaultTransport, 3, 50*time.Millisecond)
	client := &http.Client{Transport: rt}

	start := time.Now()
	resp, err := client.Get(server.URL)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if attemptCount != 3 {
		t.Errorf("Attempt count = %d, want 3", attemptCount)
	}

	// Should have retried with delays
	// 1st attempt: immediate
	// 2nd attempt: +50ms delay
	// 3rd attempt: +100ms delay
	// Total: ~150ms minimum
	if duration < 100*time.Millisecond {
		t.Errorf("Duration too short: %v (expected delays from retries)", duration)
	}
}

func TestRetryRoundTripperMaxRetries(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Only allow 2 retries (3 total attempts)
	rt := NewRetryRoundTripper(http.DefaultTransport, 2, 10*time.Millisecond)
	client := &http.Client{Transport: rt}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should return the failed response after max retries
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}

	// Should have tried 3 times (1 initial + 2 retries)
	if attemptCount != 3 {
		t.Errorf("Attempt count = %d, want 3", attemptCount)
	}
}

func TestRetryRoundTripperNonRetryableStatus(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusBadRequest) // 400 is not retryable
	}))
	defer server.Close()

	rt := NewRetryRoundTripper(http.DefaultTransport, 3, 10*time.Millisecond)
	client := &http.Client{Transport: rt}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	// Should not retry for 400
	if attemptCount != 1 {
		t.Errorf("Attempt count = %d, want 1 (no retries for 400)", attemptCount)
	}
}

func TestIsRetryableStatusCode(t *testing.T) {
	tests := []struct {
		code      int
		retryable bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusNotFound, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.code), func(t *testing.T) {
			result := isRetryableStatusCode(tt.code)
			if result != tt.retryable {
				t.Errorf("isRetryableStatusCode(%d) = %v, want %v", tt.code, result, tt.retryable)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	rt := NewRetryRoundTripper(nil, 3, time.Second)

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},  // 2^0 = 1
		{1, 2 * time.Second},  // 2^1 = 2
		{2, 4 * time.Second},  // 2^2 = 4
		{3, 8 * time.Second},  // 2^3 = 8
		{4, 16 * time.Second}, // 2^4 = 16
		{5, 30 * time.Second}, // 2^5 = 32, but capped at 30
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.attempt)), func(t *testing.T) {
			result := rt.calculateBackoff(tt.attempt)
			if result != tt.want {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, result, tt.want)
			}
		})
	}
}

func TestSetBaseRetry(t *testing.T) {
	rt := NewRetryRoundTripper(nil, 3, time.Second)

	newBase := &http.Transport{}
	rt.SetBase(newBase)

	if rt.base != newBase {
		t.Error("SetBase did not update base transport")
	}
}

// mockRoundTripper is a test helper that returns errors.
type mockRoundTripper struct {
	err error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, m.err
}

func TestRetryRoundTripperTransportError(t *testing.T) {
	mockErr := errors.New("network error")
	mock := &mockRoundTripper{err: mockErr}

	rt := NewRetryRoundTripper(mock, 2, 10*time.Millisecond)
	client := &http.Client{Transport: rt}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := client.Do(req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Error should be the mock error
	if !errors.Is(err, mockErr) {
		t.Errorf("Error = %v, want %v", err, mockErr)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	rt := NewRetryRoundTripper(http.DefaultTransport, 5, 200*time.Millisecond)
	client := &http.Client{Transport: rt}

	// Context that cancels after 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)

	start := time.Now()
	_, err := client.Do(req)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("Expected error due to context cancellation")
	}

	// Should cancel before second retry (which would be at ~200ms)
	if duration > 150*time.Millisecond {
		t.Errorf("Took too long (%v), context should have cancelled", duration)
	}
}
