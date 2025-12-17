package transport

import (
	"log/slog"
	"net/http"
	"time"
)

// Option is a function that configures a Transport.
type Option func(*Transport)

// WithBaseTransport sets the base HTTP transport.
// If not set, http.DefaultTransport is used.
func WithBaseTransport(base http.RoundTripper) Option {
	return func(t *Transport) {
		t.base = base
	}
}

// WithRateLimit adds rate limiting to the transport.
// Rate is specified as requests per second (e.g., 10 requests per second).
// Burst allows temporary bursts above the rate limit.
func WithRateLimit(rate float64, burst int) Option {
	return func(t *Transport) {
		rt := NewRateLimitRoundTripper(t.base, rate, burst)
		t.base = rt
	}
}

// WithRetry adds retry logic to the transport.
// maxRetries specifies the maximum number of retry attempts.
// initialDelay specifies the initial delay between retries (exponential backoff).
func WithRetry(maxRetries int, initialDelay time.Duration) Option {
	return func(t *Transport) {
		rt := NewRetryRoundTripper(t.base, maxRetries, initialDelay)
		t.base = rt
	}
}

// WithLogging adds request/response logging to the transport.
// logger is the slog.Logger instance to use for logging.
func WithLogging(logger *slog.Logger) Option {
	return func(t *Transport) {
		rt := NewLoggingRoundTripper(t.base, logger)
		t.base = rt
	}
}

// WithUserAgent sets a custom User-Agent header for all requests.
// userAgent is the User-Agent string (e.g., "my-app/1.0.0").
func WithUserAgent(userAgent string) Option {
	return func(t *Transport) {
		rt := NewUserAgentRoundTripper(t.base, userAgent)
		t.base = rt
	}
}

// WithDefaultUserAgent sets the default User-Agent for the freee-api-go library.
// The format is: "freee-api-go/VERSION (+github.com/u-masato/freee-api-go)".
func WithDefaultUserAgent(version string) Option {
	userAgent := "freee-api-go/" + version + " (+github.com/u-masato/freee-api-go)"
	return WithUserAgent(userAgent)
}
