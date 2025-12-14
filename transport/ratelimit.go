package transport

import (
	"net/http"

	"golang.org/x/time/rate"
)

// RateLimitRoundTripper implements rate limiting for HTTP requests.
// It uses a token bucket algorithm to limit the request rate.
type RateLimitRoundTripper struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

// NewRateLimitRoundTripper creates a new rate-limited RoundTripper.
//
// Parameters:
//   - base: The underlying RoundTripper to wrap
//   - requestsPerSecond: The maximum number of requests per second (e.g., 10.0)
//   - burst: The maximum burst size (number of requests that can be made at once)
//
// Example:
//
//	// Allow 10 requests per second with a burst of 5
//	rt := NewRateLimitRoundTripper(http.DefaultTransport, 10.0, 5)
func NewRateLimitRoundTripper(base http.RoundTripper, requestsPerSecond float64, burst int) *RateLimitRoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return &RateLimitRoundTripper{
		base:    base,
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
	}
}

// RoundTrip implements the http.RoundTripper interface.
// It waits for rate limit permission before executing the request.
func (rt *RateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Wait for rate limiter to allow the request
	if err := rt.limiter.Wait(req.Context()); err != nil {
		return nil, err
	}

	return rt.base.RoundTrip(req)
}

// SetBase sets the base RoundTripper.
func (rt *RateLimitRoundTripper) SetBase(base http.RoundTripper) {
	rt.base = base
}
