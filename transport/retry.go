package transport

import (
	"context"
	"io"
	"math"
	"net/http"
	"time"
)

// RetryRoundTripper implements automatic retry logic with exponential backoff.
type RetryRoundTripper struct {
	base         http.RoundTripper
	maxRetries   int
	initialDelay time.Duration
}

// NewRetryRoundTripper creates a new retry-enabled RoundTripper.
//
// Parameters:
//   - base: The underlying RoundTripper to wrap
//   - maxRetries: Maximum number of retry attempts (0 means no retries)
//   - initialDelay: Initial delay between retries (doubles with each retry)
//
// Example:
//
//	// Retry up to 3 times with initial delay of 1 second
//	rt := NewRetryRoundTripper(http.DefaultTransport, 3, time.Second)
func NewRetryRoundTripper(base http.RoundTripper, maxRetries int, initialDelay time.Duration) *RetryRoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	if maxRetries < 0 {
		maxRetries = 0
	}

	if initialDelay <= 0 {
		initialDelay = time.Second
	}

	return &RetryRoundTripper{
		base:         base,
		maxRetries:   maxRetries,
		initialDelay: initialDelay,
	}
}

// RoundTrip implements the http.RoundTripper interface with retry logic.
func (rt *RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= rt.maxRetries; attempt++ {
		// Clone the request for retry attempts
		reqClone := cloneRequest(req)

		resp, lastErr = rt.base.RoundTrip(reqClone)

		// If successful and not a retryable status code, return immediately
		if lastErr == nil && !isRetryableStatusCode(resp.StatusCode) {
			return resp, nil
		}

		// If this was the last attempt, break
		if attempt == rt.maxRetries {
			break
		}

		// Clean up response body if present
		if resp != nil && resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		// Calculate delay with exponential backoff
		delay := rt.calculateBackoff(attempt)

		// Wait before retry
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-req.Context().Done():
			// Context cancelled
			return nil, req.Context().Err()
		}
	}

	return resp, lastErr
}

// SetBase sets the base RoundTripper.
func (rt *RetryRoundTripper) SetBase(base http.RoundTripper) {
	rt.base = base
}

// calculateBackoff calculates the backoff delay for a given attempt.
// Uses exponential backoff: initialDelay * 2^attempt.
func (rt *RetryRoundTripper) calculateBackoff(attempt int) time.Duration {
	multiplier := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(rt.initialDelay) * multiplier)

	// Cap at 30 seconds
	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// isRetryableStatusCode determines if an HTTP status code is retryable.
func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests: // 429
		return true
	case http.StatusInternalServerError: // 500
		return true
	case http.StatusBadGateway: // 502
		return true
	case http.StatusServiceUnavailable: // 503
		return true
	case http.StatusGatewayTimeout: // 504
		return true
	default:
		return false
	}
}

// cloneRequest creates a shallow copy of an HTTP request.
// This is necessary for retry attempts to avoid issues with consumed request bodies.
func cloneRequest(req *http.Request) *http.Request {
	// Create a shallow copy
	reqClone := req.Clone(req.Context())

	// Note: Request body handling is tricky for retries.
	// For GET requests (no body), this works fine.
	// For POST/PUT requests with bodies, the caller should use GetBody
	// or the body will be empty on retry.
	if req.GetBody != nil {
		body, _ := req.GetBody()
		reqClone.Body = body
	}

	return reqClone
}

// WithRetryableContext wraps a context to support retries.
// This is a helper function for users who want to control retry behavior.
func WithRetryableContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}
