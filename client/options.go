package client

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// Option is a function that configures a Client.
// Options are passed to NewClient to customize the client's behavior.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client for the freee API client.
//
// This is useful when you want to customize the HTTP transport layer,
// for example to add rate limiting, retry logic, or custom TLS configuration.
//
// Example:
//
//	transport := transport.NewTransport(
//	    transport.WithRateLimit(10, 5),
//	    transport.WithRetry(3, time.Second),
//	    transport.WithLogging(logger),
//	)
//	httpClient := &http.Client{Transport: transport}
//	client := client.NewClient(client.WithHTTPClient(httpClient))
//
// Note: If you provide both WithHTTPClient and WithTokenSource, the token source
// will not be automatically integrated. You should create an OAuth2 client yourself:
//
//	oauthClient := oauth2.NewClient(ctx, tokenSource)
//	// Apply additional transport configuration to oauthClient if needed
//	client := client.NewClient(client.WithHTTPClient(oauthClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets a custom base URL for the freee API.
//
// The default base URL is "https://api.freee.co.jp".
// This option is primarily useful for testing with mock servers.
//
// Example:
//
//	// Use with a test server
//	server := httptest.NewServer(handler)
//	defer server.Close()
//	client := client.NewClient(client.WithBaseURL(server.URL))
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithTokenSource sets an OAuth2 token source for authentication.
//
// The token source will automatically refresh expired tokens and add
// the Authorization header to all requests.
//
// Example:
//
//	config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
//	token, err := config.Exchange(ctx, code)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	tokenSource := config.TokenSource(ctx, token)
//	client := client.NewClient(client.WithTokenSource(tokenSource))
//
// If you also need custom transport configuration, create the OAuth2 client first:
//
//	transport := transport.NewTransport(transport.WithRateLimit(10, 5))
//	httpClient := &http.Client{Transport: transport}
//	oauthClient := &http.Client{
//	    Transport: &oauth2.Transport{
//	        Source: tokenSource,
//	        Base:   transport,
//	    },
//	}
//	client := client.NewClient(client.WithHTTPClient(oauthClient))
func WithTokenSource(tokenSource oauth2.TokenSource) Option {
	return func(c *Client) {
		c.tokenSource = tokenSource
	}
}

// WithUserAgent sets a custom User-Agent header for all requests.
//
// The default User-Agent is "freee-api-go/dev (+github.com/u-masato/freee-api-go)".
//
// Example:
//
//	client := client.NewClient(client.WithUserAgent("my-app/1.0.0"))
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithContext sets the context used for token refresh and HTTP requests.
//
// The default context is context.Background().
//
// This is particularly important when using WithTokenSource, as the context
// will be used for token refresh operations.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	client := client.NewClient(
//	    client.WithContext(ctx),
//	    client.WithTokenSource(tokenSource),
//	)
func WithContext(ctx context.Context) Option {
	return func(c *Client) {
		c.ctx = ctx
	}
}
