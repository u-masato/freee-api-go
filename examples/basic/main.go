// Package main demonstrates basic usage of the freee API client.
//
// This example shows how to:
//   - Create an API client with an access token
//   - Retrieve transaction data (deals) from the freee API
//   - Use filtering options to narrow down results
//   - Handle errors and contexts properly
//
// Usage:
//
//	# Set environment variables
//	export FREEE_ACCESS_TOKEN="your-access-token"
//	export FREEE_COMPANY_ID="123456"
//
//	# Run the example
//	go run main.go
//
// Prerequisites:
// You need a valid access token. You can obtain one by running the oauth example:
//
//	cd ../oauth
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/muno/freee-api-go/accounting"
	"github.com/muno/freee-api-go/client"
	"golang.org/x/oauth2"
)

func main() {
	// Get credentials from environment variables
	accessToken := os.Getenv("FREEE_ACCESS_TOKEN")
	companyIDStr := os.Getenv("FREEE_COMPANY_ID")

	if accessToken == "" {
		log.Fatal("FREEE_ACCESS_TOKEN must be set. Run examples/oauth first to obtain a token.")
	}
	if companyIDStr == "" {
		log.Fatal("FREEE_COMPANY_ID must be set")
	}

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid FREEE_COMPANY_ID: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a static token source from the access token
	token := &oauth2.Token{
		AccessToken: accessToken,
	}
	tokenSource := oauth2.StaticTokenSource(token)

	// Create the freee API client
	c := client.NewClient(
		client.WithTokenSource(tokenSource),
		client.WithContext(ctx),
	)

	// Create the accounting client
	accountingClient, err := accounting.NewClient(c)
	if err != nil {
		log.Fatalf("Failed to create accounting client: %v", err)
	}

	fmt.Println("=== freee API Basic Example ===")

	// Example 1: List all deals
	fmt.Println("1. Fetching recent deals...")
	if err := listRecentDeals(ctx, accountingClient, companyID); err != nil {
		log.Printf("Error listing recent deals: %v", err)
	}

	fmt.Println()

	// Example 2: List deals with filters
	fmt.Println("2. Fetching expense deals only...")
	if err := listExpenseDeals(ctx, accountingClient, companyID); err != nil {
		log.Printf("Error listing expense deals: %v", err)
	}

	fmt.Println()

	// Example 3: Get a specific deal (if any exists)
	fmt.Println("3. Fetching a specific deal...")
	if err := getSpecificDeal(ctx, accountingClient, companyID); err != nil {
		log.Printf("Error getting specific deal: %v", err)
	}

	fmt.Println("\n=== Example completed successfully ===")
}

// listRecentDeals demonstrates fetching recent deals with pagination
func listRecentDeals(ctx context.Context, client *accounting.Client, companyID int64) error {
	// Set up options to fetch the first 5 deals
	limit := int64(5)
	opts := &accounting.ListDealsOptions{
		Limit: &limit,
	}

	// Fetch deals
	result, err := client.Deals().List(ctx, companyID, opts)
	if err != nil {
		return fmt.Errorf("failed to list deals: %w", err)
	}

	fmt.Printf("   Found %d deals (total: %d)\n", len(result.Deals), result.TotalCount)
	fmt.Println()

	// Display deal information
	for i, deal := range result.Deals {
		fmt.Printf("   Deal #%d:\n", i+1)
		fmt.Printf("     ID: %d\n", deal.Id)
		fmt.Printf("     Issue Date: %s\n", deal.IssueDate)
		if deal.Type != nil {
			fmt.Printf("     Type: %s\n", *deal.Type)
		}
		fmt.Printf("     Status: %s\n", deal.Status)

		if deal.PartnerId != nil {
			fmt.Printf("     Partner ID: %d\n", *deal.PartnerId)
		}
		if deal.PartnerCode != nil {
			fmt.Printf("     Partner Code: %s\n", *deal.PartnerCode)
		}

		fmt.Printf("     Amount: %d\n", deal.Amount)

		fmt.Println()
	}

	return nil
}

// listExpenseDeals demonstrates filtering deals by type
func listExpenseDeals(ctx context.Context, client *accounting.Client, companyID int64) error {
	// Set up options to filter expense deals
	dealType := "expense"
	limit := int64(3)
	opts := &accounting.ListDealsOptions{
		Type:  &dealType,
		Limit: &limit,
	}

	// Fetch expense deals
	result, err := client.Deals().List(ctx, companyID, opts)
	if err != nil {
		return fmt.Errorf("failed to list expense deals: %w", err)
	}

	fmt.Printf("   Found %d expense deals\n", len(result.Deals))
	fmt.Println()

	// Display deal information
	for i, deal := range result.Deals {
		fmt.Printf("   Expense Deal #%d:\n", i+1)
		fmt.Printf("     ID: %d\n", deal.Id)
		fmt.Printf("     Issue Date: %s\n", deal.IssueDate)

		if deal.PartnerId != nil {
			fmt.Printf("     Partner ID: %d\n", *deal.PartnerId)
		}
		if deal.PartnerCode != nil {
			fmt.Printf("     Partner Code: %s\n", *deal.PartnerCode)
		}

		fmt.Printf("     Amount: %d\n", deal.Amount)

		fmt.Println()
	}

	return nil
}

// getSpecificDeal demonstrates fetching a single deal by ID
func getSpecificDeal(ctx context.Context, client *accounting.Client, companyID int64) error {
	// First, get a deal ID from the list
	limit := int64(1)
	opts := &accounting.ListDealsOptions{
		Limit: &limit,
	}

	result, err := client.Deals().List(ctx, companyID, opts)
	if err != nil {
		return fmt.Errorf("failed to list deals: %w", err)
	}

	if len(result.Deals) == 0 {
		fmt.Println("   No deals found to fetch")
		return nil
	}

	dealID := result.Deals[0].Id

	// Fetch the specific deal
	resp, err := client.Deals().Get(ctx, companyID, dealID, nil)
	if err != nil {
		return fmt.Errorf("failed to get deal: %w", err)
	}

	deal := &resp.Deal
	fmt.Printf("   Deal Details:\n")
	fmt.Printf("     ID: %d\n", deal.Id)
	fmt.Printf("     Issue Date: %s\n", deal.IssueDate)
	if deal.Type != nil {
		fmt.Printf("     Type: %s\n", *deal.Type)
	}
	fmt.Printf("     Status: %s\n", deal.Status)

	if deal.PartnerId != nil {
		fmt.Printf("     Partner ID: %d\n", *deal.PartnerId)
	}
	if deal.PartnerCode != nil {
		fmt.Printf("     Partner Code: %s\n", *deal.PartnerCode)
	}

	fmt.Printf("     Amount: %d\n", deal.Amount)

	if deal.DueDate != nil {
		fmt.Printf("     Due Date: %s\n", *deal.DueDate)
	}

	// Display details if available
	if deal.Details != nil && len(*deal.Details) > 0 {
		fmt.Printf("     Details: %d item(s)\n", len(*deal.Details))
		for i, detail := range *deal.Details {
			fmt.Printf("       Detail #%d:\n", i+1)
			fmt.Printf("         Account Item ID: %d\n", detail.AccountItemId)
			fmt.Printf("         Amount: %d\n", detail.Amount)
			fmt.Printf("         Entry Side: %s\n", detail.EntrySide)
			if detail.TaxCode != 0 {
				fmt.Printf("         Tax Code: %d\n", detail.TaxCode)
			}
		}
	}

	return nil
}
