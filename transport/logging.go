package transport

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// LoggingRoundTripper implements request/response logging.
type LoggingRoundTripper struct {
	base   http.RoundTripper
	logger *slog.Logger
}

// NewLoggingRoundTripper creates a new logging-enabled RoundTripper.
//
// The logger will record:
//   - Request method, URL, and headers (with sensitive data masked)
//   - Response status code and duration
//   - Errors if any
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	rt := NewLoggingRoundTripper(http.DefaultTransport, logger)
func NewLoggingRoundTripper(base http.RoundTripper, logger *slog.Logger) *LoggingRoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &LoggingRoundTripper{
		base:   base,
		logger: logger,
	}
}

// RoundTrip implements the http.RoundTripper interface with logging.
func (rt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Log request
	rt.logRequest(req)

	// Execute request
	resp, err := rt.base.RoundTrip(req)

	duration := time.Since(start)

	// Log response
	if err != nil {
		rt.logError(req, err, duration)
	} else {
		rt.logResponse(req, resp, duration)
	}

	return resp, err
}

// SetBase sets the base RoundTripper.
func (rt *LoggingRoundTripper) SetBase(base http.RoundTripper) {
	rt.base = base
}

// logRequest logs the outgoing HTTP request.
func (rt *LoggingRoundTripper) logRequest(req *http.Request) {
	rt.logger.Info("HTTP request",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("proto", req.Proto),
		slog.Any("headers", maskSensitiveHeaders(req.Header)),
	)
}

// logResponse logs the HTTP response.
func (rt *LoggingRoundTripper) logResponse(req *http.Request, resp *http.Response, duration time.Duration) {
	rt.logger.Info("HTTP response",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Int("status_code", resp.StatusCode),
		slog.String("status", resp.Status),
		slog.Duration("duration", duration),
		slog.Int64("content_length", resp.ContentLength),
	)
}

// logError logs HTTP request errors.
func (rt *LoggingRoundTripper) logError(req *http.Request, err error, duration time.Duration) {
	rt.logger.Error("HTTP request failed",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("error", err.Error()),
		slog.Duration("duration", duration),
	)
}

// maskSensitiveHeaders creates a copy of headers with sensitive values masked.
func maskSensitiveHeaders(headers http.Header) http.Header {
	masked := make(http.Header)

	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"set-cookie":    true,
		"x-api-key":     true,
		"api-key":       true,
	}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if sensitiveHeaders[lowerKey] {
			masked[key] = []string{"[REDACTED]"}
		} else {
			masked[key] = values
		}
	}

	return masked
}

// CaptureResponseBody is a helper to capture response body for logging.
// Use with caution as it reads the entire response body into memory.
func CaptureResponseBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, nil
	}

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Close the original body
	resp.Body.Close()

	// Replace body with a new reader so it can be read again
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}
