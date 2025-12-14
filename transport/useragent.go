package transport

import (
	"net/http"
)

// UserAgentRoundTripper adds a User-Agent header to all HTTP requests.
type UserAgentRoundTripper struct {
	base      http.RoundTripper
	userAgent string
}

// NewUserAgentRoundTripper creates a new User-Agent RoundTripper.
//
// The User-Agent header helps identify the client making the request.
// If a request already has a User-Agent header, it will be preserved
// and the custom user agent will be appended.
//
// Example:
//
//	rt := NewUserAgentRoundTripper(http.DefaultTransport, "my-app/1.0.0")
func NewUserAgentRoundTripper(base http.RoundTripper, userAgent string) *UserAgentRoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return &UserAgentRoundTripper{
		base:      base,
		userAgent: userAgent,
	}
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *UserAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqClone := req.Clone(req.Context())

	// Set or append User-Agent header
	existingUA := reqClone.Header.Get("User-Agent")
	if existingUA == "" {
		reqClone.Header.Set("User-Agent", rt.userAgent)
	} else {
		// Append our user agent to the existing one
		reqClone.Header.Set("User-Agent", existingUA+" "+rt.userAgent)
	}

	return rt.base.RoundTrip(reqClone)
}

// SetBase sets the base RoundTripper.
func (rt *UserAgentRoundTripper) SetBase(base http.RoundTripper) {
	rt.base = base
}

// DefaultUserAgent returns the default user agent string for this library.
// Format: "freee-api-go/VERSION (+github.com/muno/freee-api-go)"
func DefaultUserAgent(version string) string {
	if version == "" {
		version = "dev"
	}
	return "freee-api-go/" + version + " (+github.com/muno/freee-api-go)"
}
