package transport

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewLoggingRoundTripper(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	rt := NewLoggingRoundTripper(nil, logger)

	if rt == nil {
		t.Fatal("NewLoggingRoundTripper returned nil")
	}

	if rt.logger == nil {
		t.Fatal("logger is nil")
	}

	if rt.base == nil {
		t.Fatal("base is nil")
	}
}

func TestLoggingRoundTripperRequest(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := NewLoggingRoundTripper(http.DefaultTransport, logger)
	client := &http.Client{Transport: rt}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuffer.String()

	// Check that request was logged
	if !strings.Contains(logOutput, "HTTP request") {
		t.Error("Log should contain 'HTTP request'")
	}

	// Check that response was logged
	if !strings.Contains(logOutput, "HTTP response") {
		t.Error("Log should contain 'HTTP response'")
	}

	// Check that method was logged
	if !strings.Contains(logOutput, "method=GET") {
		t.Error("Log should contain method=GET")
	}

	// Check that URL was logged
	if !strings.Contains(logOutput, server.URL) {
		t.Errorf("Log should contain URL %s", server.URL)
	}

	// Check that status code was logged
	if !strings.Contains(logOutput, "status_code=200") {
		t.Error("Log should contain status_code=200")
	}
}

func TestLoggingRoundTripperError(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	// Create a server that immediately closes connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Force connection close
		panic("connection closed")
	}))
	server.Close() // Close immediately to cause error

	rt := NewLoggingRoundTripper(http.DefaultTransport, logger)
	client := &http.Client{Transport: rt}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	_, err := client.Do(req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	logOutput := logBuffer.String()

	// Check that error was logged
	if !strings.Contains(logOutput, "HTTP request failed") {
		t.Error("Log should contain 'HTTP request failed'")
	}

	// Check that error details were logged
	if !strings.Contains(logOutput, "error=") {
		t.Error("Log should contain error details")
	}
}

func TestMaskSensitiveHeaders(t *testing.T) {
	headers := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bearer secret-token"},
		"Cookie":        []string{"session=abc123"},
		"X-Api-Key":     []string{"api-key-value"},
		"User-Agent":    []string{"test-client/1.0"},
	}

	masked := maskSensitiveHeaders(headers)

	tests := []struct {
		header     string
		shouldMask bool
	}{
		{"Content-Type", false},
		{"Authorization", true},
		{"Cookie", true},
		{"X-Api-Key", true},
		{"User-Agent", false},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			value := masked.Get(tt.header)

			if tt.shouldMask {
				if value != "[REDACTED]" {
					t.Errorf("%s = %q, want [REDACTED]", tt.header, value)
				}
			} else {
				if value == "[REDACTED]" {
					t.Errorf("%s should not be masked", tt.header)
				}
			}
		})
	}
}

func TestLoggingSensitiveHeaders(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := NewLoggingRoundTripper(http.DefaultTransport, logger)
	client := &http.Client{Transport: rt}

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Authorization", "Bearer secret-token-12345")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	logOutput := logBuffer.String()

	// Should NOT contain the actual token
	if strings.Contains(logOutput, "secret-token-12345") {
		t.Error("Log should not contain the actual authorization token")
	}

	// Should contain the redacted marker
	if !strings.Contains(logOutput, "[REDACTED]") {
		t.Error("Log should contain [REDACTED] for authorization header")
	}
}

func TestSetBaseLogging(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	rt := NewLoggingRoundTripper(nil, logger)

	newBase := &http.Transport{}
	rt.SetBase(newBase)

	if rt.base != newBase {
		t.Error("SetBase did not update base transport")
	}
}

func TestCaptureResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response body"))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	body, err := CaptureResponseBody(resp)
	if err != nil {
		t.Fatalf("CaptureResponseBody failed: %v", err)
	}

	expected := "test response body"
	if string(body) != expected {
		t.Errorf("Body = %q, want %q", string(body), expected)
	}

	// Body should still be readable from response
	body2 := make([]byte, len(expected))
	n, err := resp.Body.Read(body2)
	if err != nil {
		t.Fatalf("Reading body failed: %v", err)
	}

	if n != len(expected) {
		t.Errorf("Read %d bytes, want %d", n, len(expected))
	}

	if string(body2) != expected {
		t.Errorf("Body after capture = %q, want %q", string(body2), expected)
	}
}
