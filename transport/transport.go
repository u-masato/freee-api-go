// Package transport provides HTTP transport functionality for the freee API client.
//
// This package implements various RoundTripper wrappers that add functionality
// such as rate limiting, retry logic, logging, and user agent management.
//
// RoundTrippers can be chained together to compose multiple behaviors:
//
//	transport := NewTransport(
//		WithRateLimit(10, 1),
//		WithRetry(3, time.Second),
//		WithLogging(logger),
//		WithUserAgent("my-app/1.0"),
//	)
package transport

import (
	"net/http"
)

// Transport is a configurable HTTP transport for the freee API client.
type Transport struct {
	base http.RoundTripper
}

// NewTransport creates a new Transport with the given options.
// If no base transport is provided via options, http.DefaultTransport is used.
func NewTransport(opts ...Option) *Transport {
	t := &Transport{
		base: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// RoundTrip implements the http.RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.base.RoundTrip(req)
}

// Client creates an http.Client using this transport.
func (t *Transport) Client() *http.Client {
	return &http.Client{
		Transport: t,
	}
}

// ChainRoundTrippers creates a RoundTripper that chains multiple RoundTrippers together.
// The RoundTrippers are applied in the order they are provided.
// The last RoundTripper in the chain should be the actual transport.
func ChainRoundTrippers(roundTrippers ...http.RoundTripper) http.RoundTripper {
	if len(roundTrippers) == 0 {
		return http.DefaultTransport
	}

	if len(roundTrippers) == 1 {
		return roundTrippers[0]
	}

	// Build chain from right to left
	result := roundTrippers[len(roundTrippers)-1]
	for i := len(roundTrippers) - 2; i >= 0; i-- {
		result = wrapRoundTripper(roundTrippers[i], result)
	}

	return result
}

// wrapRoundTripper wraps an outer RoundTripper around an inner one.
type roundTripperWrapper struct {
	outer http.RoundTripper
	inner http.RoundTripper
}

func wrapRoundTripper(outer, inner http.RoundTripper) http.RoundTripper {
	// If outer is a wrapper type that can accept a base, inject inner
	if w, ok := outer.(interface{ SetBase(http.RoundTripper) }); ok {
		w.SetBase(inner)
		return outer
	}
	// Otherwise, create a simple wrapper
	return &roundTripperWrapper{outer: outer, inner: inner}
}

func (w *roundTripperWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	// This is a fallback; in practice, each RoundTripper should handle its own wrapping
	return w.inner.RoundTrip(req)
}
