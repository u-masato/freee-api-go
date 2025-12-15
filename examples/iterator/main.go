package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/muno/freee-api-go/accounting"
	"github.com/muno/freee-api-go/auth"
	"github.com/muno/freee-api-go/client"
)

func main() {
	// Load OAuth2 token
	tokenPath := "token.json"
	token, err := auth.LoadTokenFromFile(tokenPath)
	if err != nil {
		log.Fatalf("Failed to load token: %v", err)
	}

	// Get company ID from environment variable
	companyIDStr := os.Getenv("FREEE_COMPANY_ID")
	if companyIDStr == "" {
		log.Fatal("FREEE_COMPANY_ID environment variable is required")
	}

	var companyID int64
	fmt.Sscanf(companyIDStr, "%d", &companyID)

	// Create OAuth2 config
	config := auth.NewConfig(
		os.Getenv("FREEE_CLIENT_ID"),
		os.Getenv("FREEE_CLIENT_SECRET"),
		os.Getenv("FREEE_REDIRECT_URL"),
		[]string{"read", "write"},
	)

	// Create token source with auto-refresh
	ctx := context.Background()
	ts := config.TokenSource(ctx, token)

	// Create freee client
	freeeClient := client.NewClient(
		client.WithTokenSource(ts),
	)

	// Create accounting service
	accountingClient, err := accounting.NewClient(freeeClient)
	if err != nil {
		log.Fatalf("Failed to create accounting client: %v", err)
	}

	// Example 1: Iterate through all expense deals
	fmt.Println("Example 1: Iterate through all expense deals")
	fmt.Println("--------------------------------------------")

	expenseType := "expense"
	opts := &accounting.ListDealsOptions{
		Type: &expenseType,
	}

	iter := accountingClient.Deals().ListIter(ctx, companyID, opts)
	count := 0
	totalAmount := int64(0)

	for iter.Next() {
		deal := iter.Value()
		count++
		totalAmount += deal.Amount
		fmt.Printf("%d. Deal ID: %d, Amount: ¥%d, Date: %s\n",
			count, deal.Id, deal.Amount, deal.IssueDate)
	}

	if err := iter.Err(); err != nil {
		log.Fatalf("Error during iteration: %v", err)
	}

	fmt.Printf("\nTotal: %d deals, ¥%d\n\n", count, totalAmount)

	// Example 2: Iterate with custom page size
	fmt.Println("Example 2: Iterate with custom page size (50 items per page)")
	fmt.Println("-------------------------------------------------------------")

	limit := int64(50)
	opts2 := &accounting.ListDealsOptions{
		Type:  &expenseType,
		Limit: &limit,
	}

	iter2 := accountingClient.Deals().ListIter(ctx, companyID, opts2)
	count2 := 0

	for iter2.Next() {
		deal := iter2.Value()
		count2++
		// Just count, don't print all
		if count2 <= 5 {
			fmt.Printf("%d. Deal ID: %d, Amount: ¥%d\n",
				count2, deal.Id, deal.Amount)
		}
	}

	if err := iter2.Err(); err != nil {
		log.Fatalf("Error during iteration: %v", err)
	}

	if count2 > 5 {
		fmt.Printf("... and %d more deals\n", count2-5)
	}
	fmt.Printf("\nTotal: %d deals\n\n", count2)

	// Example 3: Early termination
	fmt.Println("Example 3: Early termination (stop after 10 deals)")
	fmt.Println("--------------------------------------------------")

	iter3 := accountingClient.Deals().ListIter(ctx, companyID, opts)
	count3 := 0
	maxDeals := 10

	for iter3.Next() {
		deal := iter3.Value()
		count3++
		fmt.Printf("%d. Deal ID: %d, Amount: ¥%d\n",
			count3, deal.Id, deal.Amount)

		if count3 >= maxDeals {
			break // Early termination
		}
	}

	// Check for errors even after early termination
	if err := iter3.Err(); err != nil {
		log.Fatalf("Error during iteration: %v", err)
	}

	fmt.Printf("\nProcessed %d deals (stopped early)\n\n", count3)

	// Example 4: Filter and aggregate
	fmt.Println("Example 4: Filter and aggregate (deals over ¥100,000)")
	fmt.Println("-----------------------------------------------------")

	iter4 := accountingClient.Deals().ListIter(ctx, companyID, opts)
	largeDeals := 0
	largeDealAmount := int64(0)

	for iter4.Next() {
		deal := iter4.Value()
		if deal.Amount > 100000 {
			largeDeals++
			largeDealAmount += deal.Amount
			fmt.Printf("Large deal ID: %d, Amount: ¥%d\n", deal.Id, deal.Amount)
		}
	}

	if err := iter4.Err(); err != nil {
		log.Fatalf("Error during iteration: %v", err)
	}

	fmt.Printf("\nFound %d large deals totaling ¥%d\n", largeDeals, largeDealAmount)
}
