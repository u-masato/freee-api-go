// Package main demonstrates advanced usage of the freee API client.
//
// This example shows how to:
//   - Configure a custom transport with rate limiting and retry logic
//   - Use structured logging for debugging
//   - Handle pagination efficiently
//   - Implement proper error handling with type assertions
//   - Use token auto-refresh with OAuth2
//
// Usage:
//
//	# Set environment variables
//	export FREEE_CLIENT_ID="your-client-id"
//	export FREEE_CLIENT_SECRET="your-client-secret"
//	export FREEE_COMPANY_ID="123456"
//
//	# Ensure token.json exists (run examples/oauth first)
//
//	# Run the example
//	go run main.go
//
// Prerequisites:
// You need a valid access token stored in token.json.
// You can obtain one by running the oauth example:
//
//	cd ../oauth
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/u-masato/freee-api-go/accounting"
	"github.com/u-masato/freee-api-go/auth"
	"github.com/u-masato/freee-api-go/client"
	"github.com/u-masato/freee-api-go/transport"
	"golang.org/x/oauth2"
)

const (
	// tokenFile is the path to the saved OAuth2 token
	tokenFile = "../oauth/token.json"
)

func main() {
	// Get company ID from environment variable
	companyIDStr := os.Getenv("FREEE_COMPANY_ID")
	if companyIDStr == "" {
		log.Fatal("FREEE_COMPANY_ID must be set")
	}

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid FREEE_COMPANY_ID: %v", err)
	}

	// Set up structured logger for debugging
	// In production, you might want to use JSON format:
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== freee API Advanced Example ===")
	fmt.Println()

	// Step 1: Load OAuth2 token
	fmt.Println("1. Loading OAuth2 token...")
	token, err := loadToken()
	if err != nil {
		log.Fatalf("Failed to load token: %v", err)
	}
	fmt.Printf("   Token loaded (expires: %s)\n\n", token.Expiry.Format(time.RFC3339))

	// Step 2: Create custom transport with rate limiting and retry
	fmt.Println("2. Creating custom transport...")
	customTransport := createCustomTransport(logger)
	fmt.Println("   Transport configured with:")
	fmt.Println("   - Rate limit: 3 requests/second, burst 5")
	fmt.Println("   - Retry: 3 attempts with exponential backoff")
	fmt.Println("   - Structured logging enabled")
	fmt.Println()

	// Step 3: Create OAuth2 token source with auto-refresh
	fmt.Println("3. Setting up OAuth2 with auto-refresh...")
	tokenSource := createTokenSource(token)
	fmt.Println("   Token source created with auto-refresh capability")
	fmt.Println()

	// Step 4: Create the freee API client
	fmt.Println("4. Creating freee API client...")
	freeeClient := createClient(tokenSource, customTransport)
	fmt.Println("   Client created successfully")
	fmt.Println()

	// Create accounting client
	accountingClient, err := accounting.NewClient(freeeClient)
	if err != nil {
		log.Fatalf("Failed to create accounting client: %v", err)
	}

	// Create context with timeout for all API operations
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Step 5: Demonstrate various advanced features
	fmt.Println("5. Demonstrating advanced features...")
	fmt.Println()

	// Example A: Pagination with Iterator
	fmt.Println("   A. Fetching deals with pagination (Iterator pattern)...")
	if err := demonstratePagination(ctx, accountingClient, companyID); err != nil {
		log.Printf("   Pagination demo error: %v", err)
	}
	fmt.Println()

	// Example B: Structured error handling
	fmt.Println("   B. Demonstrating structured error handling...")
	demonstrateErrorHandling(ctx, accountingClient)
	fmt.Println()

	// Example C: Multiple service usage
	fmt.Println("   C. Using multiple accounting services...")
	if err := demonstrateMultipleServices(ctx, accountingClient, companyID); err != nil {
		log.Printf("   Multiple services demo error: %v", err)
	}
	fmt.Println()

	// Example D: Context cancellation
	fmt.Println("   D. Demonstrating context cancellation...")
	demonstrateContextCancellation(accountingClient, companyID)
	fmt.Println()

	fmt.Println("=== Advanced example completed ===")
}

// loadToken loads the OAuth2 token from file.
func loadToken() (*oauth2.Token, error) {
	token, err := auth.LoadTokenFromFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load token from %s: %w (run examples/oauth first)", tokenFile, err)
	}
	return token, nil
}

// createCustomTransport creates a transport with rate limiting, retry, and logging.
func createCustomTransport(logger *slog.Logger) *transport.Transport {
	return transport.NewTransport(
		// Rate limiting: 3 requests per second with burst of 5
		// freee API has rate limits, so this helps avoid 429 errors
		transport.WithRateLimit(3, 5),

		// Retry logic: 3 retries with exponential backoff starting at 1 second
		// Automatically retries on 5xx errors and 429 (rate limited)
		transport.WithRetry(3, time.Second),

		// Structured logging: logs all requests and responses
		// Sensitive headers (Authorization, Cookie) are automatically masked
		transport.WithLogging(logger),

		// Custom User-Agent
		transport.WithUserAgent("freee-api-go-advanced-example/1.0"),
	)
}

// createTokenSource creates an OAuth2 token source with auto-refresh capability.
func createTokenSource(token *oauth2.Token) oauth2.TokenSource {
	// Get OAuth2 credentials from environment
	clientID := os.Getenv("FREEE_CLIENT_ID")
	clientSecret := os.Getenv("FREEE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Println("   Warning: FREEE_CLIENT_ID or FREEE_CLIENT_SECRET not set")
		log.Println("   Token auto-refresh will not work")
		// Return static token source
		return oauth2.StaticTokenSource(token)
	}

	// Create OAuth2 config for token refresh
	config := auth.NewConfig(
		clientID,
		clientSecret,
		"http://localhost:8080/callback", // Not used for refresh, but required
		[]string{"read", "write"},
	)

	// TokenSource with auto-refresh
	// When the token expires, it will automatically refresh using the refresh token
	ctx := context.Background()
	return config.TokenSource(ctx, token)
}

// createClient creates the freee API client with custom transport and token source.
func createClient(tokenSource oauth2.TokenSource, customTransport *transport.Transport) *client.Client {
	// Create HTTP client with custom transport and OAuth2
	// The OAuth2 transport wraps our custom transport to add Authorization headers
	oauthTransport := &oauth2.Transport{
		Source: tokenSource,
		Base:   customTransport,
	}

	// Create an http.Client with the OAuth2 transport
	httpClient := &http.Client{
		Transport: oauthTransport,
	}

	return client.NewClient(
		client.WithHTTPClient(httpClient),
		client.WithUserAgent("freee-api-go-advanced-example/1.0"),
	)
}

// demonstratePagination shows how to use the Iterator pattern for pagination.
func demonstratePagination(ctx context.Context, client *accounting.Client, companyID int64) error {
	// Use the iterator to transparently handle pagination
	// The iterator automatically fetches new pages as needed
	expenseType := "expense"
	limit := int64(10) // Fetch 10 items per page

	opts := &accounting.ListDealsOptions{
		Type:  &expenseType,
		Limit: &limit,
	}

	iter := client.Deals().ListIter(ctx, companyID, opts)

	count := 0
	totalAmount := int64(0)
	maxItems := 25 // Stop after processing 25 items to keep the demo short

	fmt.Println("      Processing deals with iterator...")

	for iter.Next() {
		deal := iter.Value()
		count++
		totalAmount += deal.Amount

		// Print first 5 items
		if count <= 5 {
			fmt.Printf("      [%d] Deal ID: %d, Amount: ¥%d, Date: %s\n",
				count, deal.Id, deal.Amount, deal.IssueDate)
		}

		// Early termination after maxItems
		if count >= maxItems {
			fmt.Printf("      ... (stopped after %d items for demo)\n", maxItems)
			break
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("iterator error: %w", err)
	}

	fmt.Printf("      Summary: Processed %d deals, total amount: ¥%d\n", count, totalAmount)
	return nil
}

// demonstrateErrorHandling shows how to handle different error types.
func demonstrateErrorHandling(ctx context.Context, client *accounting.Client) {
	// Try to get a deal with an invalid company ID to trigger an error
	invalidCompanyID := int64(-1)
	limit := int64(1)
	opts := &accounting.ListDealsOptions{
		Limit: &limit,
	}

	_, err := client.Deals().List(ctx, invalidCompanyID, opts)
	if err != nil {
		fmt.Printf("      Expected error occurred: %v\n", err)
		fmt.Println("      Error handling strategies:")
		fmt.Println("      - Check error type with errors.Is() or errors.As()")
		fmt.Println("      - Log error details for debugging")
		fmt.Println("      - Return user-friendly error messages")
	}
}

// demonstrateMultipleServices shows how to use multiple accounting services together.
func demonstrateMultipleServices(ctx context.Context, client *accounting.Client, companyID int64) error {
	fmt.Println("      Fetching data from multiple services...")

	// Fetch partners
	partnerLimit := int64(3)
	partnerOpts := &accounting.ListPartnersOptions{
		Limit: &partnerLimit,
	}
	partnersResult, err := client.Partners().List(ctx, companyID, partnerOpts)
	if err != nil {
		return fmt.Errorf("failed to list partners: %w", err)
	}
	fmt.Printf("      Partners: Found %d partners\n", partnersResult.Count)

	// Fetch account items
	// Note: AccountItems API doesn't support limit parameter
	accountItemsResult, err := client.AccountItems().List(ctx, companyID, nil)
	if err != nil {
		return fmt.Errorf("failed to list account items: %w", err)
	}
	fmt.Printf("      Account Items: Found %d items\n", accountItemsResult.Count)

	// Fetch tags
	tagsResult, err := client.Tags().List(ctx, companyID, nil)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}
	fmt.Printf("      Tags: Found %d tags\n", tagsResult.Count)

	return nil
}

// demonstrateContextCancellation shows how context cancellation works.
func demonstrateContextCancellation(client *accounting.Client, companyID int64) {
	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to make a request with cancelled context
	limit := int64(1)
	opts := &accounting.ListDealsOptions{
		Limit: &limit,
	}

	_, err := client.Deals().List(ctx, companyID, opts)
	if err != nil {
		fmt.Printf("      Expected cancellation error: %v\n", err)
		fmt.Println("      Context cancellation is useful for:")
		fmt.Println("      - Implementing request timeouts")
		fmt.Println("      - Graceful shutdown")
		fmt.Println("      - User-initiated cancellation in UIs")
	}
}
