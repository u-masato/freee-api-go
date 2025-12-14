// Package client provides the main client interface for the freee API.
//
// # Overview
//
// The client package is the primary entry point for using the freee-api-go SDK.
// It provides a flexible, configurable Client type that handles HTTP communication,
// authentication, and configuration management.
//
// # Basic Usage
//
// Create a client with default settings:
//
//	import "github.com/u-masato/freee-api-go/client"
//
//	client := client.NewClient()
//
// # Authentication
//
// The most common use case is to create a client with OAuth2 authentication:
//
//	import (
//	    "github.com/u-masato/freee-api-go/auth"
//	    "github.com/u-masato/freee-api-go/client"
//	)
//
//	// Configure OAuth2
//	config := auth.NewConfig(clientID, clientSecret, redirectURL, scopes)
//
//	// Exchange authorization code for token
//	token, err := config.Exchange(ctx, code)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create token source for automatic refresh
//	tokenSource := config.TokenSource(ctx, token)
//
//	// Create client with token source
//	client := client.NewClient(client.WithTokenSource(tokenSource))
//
// # Advanced Configuration
//
// Customize the client with additional options:
//
//	import (
//	    "time"
//	    "github.com/u-masato/freee-api-go/client"
//	    "github.com/u-masato/freee-api-go/transport"
//	)
//
//	// Create custom transport with rate limiting and retries
//	transport := transport.NewTransport(
//	    transport.WithRateLimit(10, 5),        // 10 req/sec, burst 5
//	    transport.WithRetry(3, time.Second),   // 3 retries, exponential backoff
//	    transport.WithLogging(logger),         // Structured logging
//	)
//
//	// Create OAuth2 client with custom transport
//	httpClient := &http.Client{
//	    Transport: &oauth2.Transport{
//	        Source: tokenSource,
//	        Base:   transport,
//	    },
//	}
//
//	// Create client with all customizations
//	client := client.NewClient(
//	    client.WithHTTPClient(httpClient),
//	    client.WithUserAgent("my-app/1.0.0"),
//	)
//
// # Making Requests
//
// Use the client's Do method to make authenticated requests:
//
//	req, err := http.NewRequestWithContext(ctx, "GET", client.BaseURL()+"/api/1/companies", nil)
//	if err != nil {
//	    return err
//	}
//
//	resp, err := client.Do(req)
//	if err != nil {
//	    return err
//	}
//	defer resp.Body.Close()
//
//	// Process response...
//
// # Architecture
//
// The client package is designed to be the foundation layer that higher-level
// packages (like accounting/) build upon. It handles:
//
//   - HTTP client configuration and lifecycle
//   - Base URL management
//   - OAuth2 token source integration
//   - User-Agent header management
//   - Request execution with automatic header injection
//
// Higher-level packages provide user-friendly facades that hide the complexity
// of the underlying API and provide type-safe, idiomatic Go interfaces.
//
// # Design Principles
//
// The client package follows these design principles:
//
//  1. Flexibility: Uses functional options pattern for easy customization
//  2. Sensible Defaults: Works out of the box with minimal configuration
//  3. Composability: Integrates cleanly with auth and transport packages
//  4. Testability: Easy to mock for testing (use WithBaseURL for test servers)
//  5. Explicitness: Clear, documented behavior with no hidden magic
//
// # Thread Safety
//
// A Client is safe for concurrent use by multiple goroutines. The underlying
// http.Client handles concurrent requests safely.
//
// # Context Handling
//
// The Client respects context cancellation and timeouts. When using WithTokenSource,
// the context provided to WithContext (or context.Background() by default) is used
// for token refresh operations. For individual requests, use http.NewRequestWithContext
// to provide request-specific contexts.
package client
