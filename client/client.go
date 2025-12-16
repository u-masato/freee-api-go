package client

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	// DefaultBaseURL is the default base URL for the freee API.
	DefaultBaseURL = "https://api.freee.co.jp"

	// DefaultUserAgent is the default User-Agent header value.
	// This should be overridden with the actual version during build.
	DefaultUserAgent = "freee-api-go/dev (+github.com/u-masato/freee-api-go)"
)

// Client is the main client for interacting with the freee API.
//
// A Client should be created using NewClient with appropriate options:
//
//	// Create client with token source
//	client := client.NewClient(
//	    client.WithTokenSource(tokenSource),
//	)
//
//	// Create client with custom configuration
//	client := client.NewClient(
//	    client.WithHTTPClient(customHTTPClient),
//	    client.WithBaseURL("https://api.freee.co.jp"),
//	    client.WithUserAgent("my-app/1.0.0"),
//	)
type Client struct {
	// httpClient is the HTTP client used for making requests to the freee API.
	httpClient *http.Client

	// baseURL is the base URL for the freee API.
	// Defaults to DefaultBaseURL.
	baseURL string

	// tokenSource provides OAuth2 tokens for authentication.
	// If set, it will be used to automatically add Authorization headers to requests.
	tokenSource oauth2.TokenSource

	// userAgent is the User-Agent header value sent with requests.
	// Defaults to DefaultUserAgent.
	userAgent string

	// ctx is the context used for token refresh and HTTP requests.
	// Defaults to context.Background() if not provided.
	ctx context.Context
}

// NewClient creates a new freee API client with the given options.
//
// If no options are provided, the client will use sensible defaults:
//   - HTTP client: http.DefaultClient
//   - Base URL: https://api.freee.co.jp
//   - User agent: freee-api-go/dev (+github.com/u-masato/freee-api-go)
//   - Context: context.Background()
//
// Example with token source:
//
//	config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
//	token, err := config.Exchange(ctx, code)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	tokenSource := config.TokenSource(ctx, token)
//	client := client.NewClient(client.WithTokenSource(tokenSource))
//
// Example with custom HTTP client:
//
//	transport := transport.NewTransport(
//	    transport.WithRateLimit(10, 5),
//	    transport.WithRetry(3, time.Second),
//	)
//	httpClient := &http.Client{Transport: transport}
//	client := client.NewClient(
//	    client.WithHTTPClient(httpClient),
//	    client.WithTokenSource(tokenSource),
//	)
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
		ctx:        context.Background(),
	}

	for _, opt := range opts {
		opt(c)
	}

	// If a token source is provided but no custom HTTP client,
	// create an OAuth2-aware HTTP client.
	if c.tokenSource != nil && c.httpClient == http.DefaultClient {
		c.httpClient = oauth2.NewClient(c.ctx, c.tokenSource)
	}

	return c
}

// HTTPClient returns the underlying HTTP client.
//
// This can be useful for advanced use cases where direct access
// to the HTTP client is needed.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// BaseURL returns the base URL for the freee API.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// UserAgent returns the User-Agent header value.
func (c *Client) UserAgent() string {
	return c.userAgent
}

// Context returns the context used by this client.
func (c *Client) Context() context.Context {
	return c.ctx
}

// Do executes an HTTP request using the configured client.
//
// This method:
//  1. Adds the User-Agent header if not already set
//  2. Executes the request using the configured HTTP client (which may include
//     OAuth2 token handling, rate limiting, retries, etc.)
//
// Example:
//
//	req, err := http.NewRequestWithContext(ctx, "GET", client.BaseURL()+"/api/1/companies", nil)
//	if err != nil {
//	    return err
//	}
//	resp, err := client.Do(req)
//	if err != nil {
//	    return err
//	}
//	defer resp.Body.Close()
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Add User-Agent header if not already set
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return c.httpClient.Do(req)
}
