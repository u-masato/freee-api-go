package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/u-masato/freee-api-go/accounting"
	"github.com/u-masato/freee-api-go/client"
	"github.com/u-masato/freee-api-go/internal/gen"
	"github.com/u-masato/freee-api-go/tests/integration/mockserver"
	"golang.org/x/oauth2"
)

// TestDealsService_CRUD tests the complete lifecycle of deals (Create, Read, Update, Delete).
func TestDealsService_CRUD(t *testing.T) {
	// Set up mock server
	server := mockserver.NewServer()
	defer server.Close()

	// Create client
	accountingClient := createTestAccountingClient(t, server)

	ctx := context.Background()
	companyID := int64(1)

	// Test Create
	t.Run("Create", func(t *testing.T) {
		description := "Test deal description"
		accountItemID := int64(12345)
		taxCode := int64(108)
		params := gen.DealCreateParams{
			CompanyId: companyID,
			IssueDate: "2024-01-15",
			Type:      gen.DealCreateParamsTypeExpense,
			Details: []struct {
				AccountItemCode *string `json:"account_item_code,omitempty"`
				AccountItemId   *int64  `json:"account_item_id,omitempty"`
				Amount          int64   `json:"amount"`
				Description     *string `json:"description,omitempty"`
				ItemCode        *string `json:"item_code,omitempty"`
				ItemId          *int64  `json:"item_id,omitempty"`
				SectionCode     *string `json:"section_code,omitempty"`
				SectionId       *int64  `json:"section_id,omitempty"`
				Segment1TagCode *string `json:"segment_1_tag_code,omitempty"`
				Segment1TagId   *int64  `json:"segment_1_tag_id,omitempty"`
				Segment2TagCode *string `json:"segment_2_tag_code,omitempty"`
				Segment2TagId   *int64  `json:"segment_2_tag_id,omitempty"`
				Segment3TagCode *string `json:"segment_3_tag_code,omitempty"`
				Segment3TagId   *int64  `json:"segment_3_tag_id,omitempty"`
				TagIds          *[]int64 `json:"tag_ids,omitempty"`
				TaxCode         int64   `json:"tax_code"`
				Vat             *int64  `json:"vat,omitempty"`
			}{
				{
					AccountItemId: &accountItemID,
					TaxCode:       taxCode,
					Amount:        10000,
					Description:   &description,
				},
			},
		}

		resp, err := accountingClient.Deals().Create(ctx, params)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Create returned nil response")
		}
		if resp.Deal.Id == 0 {
			t.Error("Created deal should have an ID")
		}
		if resp.Deal.Amount != 10000 {
			t.Errorf("Expected amount 10000, got %d", resp.Deal.Amount)
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		// Add some deals to the mock server
		server.ClearDeals()
		for i := 0; i < 5; i++ {
			server.AddDeal(&mockserver.Deal{
				CompanyID: companyID,
				IssueDate: "2024-01-15",
				Amount:    int64(1000 * (i + 1)),
				Type:      "expense",
				Status:    "unsettled",
				Details:   []mockserver.DealDetail{},
				Payments:  []mockserver.DealPayment{},
				Receipts:  []mockserver.DealReceipt{},
			})
		}

		result, err := accountingClient.Deals().List(ctx, companyID, nil)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if result.TotalCount != 5 {
			t.Errorf("Expected total count 5, got %d", result.TotalCount)
		}
		if len(result.Deals) != 5 {
			t.Errorf("Expected 5 deals, got %d", len(result.Deals))
		}
	})

	// Test List with filters
	t.Run("List_WithFilters", func(t *testing.T) {
		server.ClearDeals()
		// Add expense deals
		for i := 0; i < 3; i++ {
			server.AddDeal(&mockserver.Deal{
				CompanyID: companyID,
				IssueDate: "2024-01-15",
				Amount:    int64(1000 * (i + 1)),
				Type:      "expense",
				Status:    "unsettled",
			})
		}
		// Add income deals
		for i := 0; i < 2; i++ {
			server.AddDeal(&mockserver.Deal{
				CompanyID: companyID,
				IssueDate: "2024-01-15",
				Amount:    int64(2000 * (i + 1)),
				Type:      "income",
				Status:    "settled",
			})
		}

		// Filter by type
		typeFilter := "expense"
		opts := &accounting.ListDealsOptions{
			Type: &typeFilter,
		}
		result, err := accountingClient.Deals().List(ctx, companyID, opts)
		if err != nil {
			t.Fatalf("List with filters failed: %v", err)
		}

		if result.TotalCount != 3 {
			t.Errorf("Expected 3 expense deals, got %d", result.TotalCount)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		server.ClearDeals()
		server.AddDeal(&mockserver.Deal{
			ID:        100,
			CompanyID: companyID,
			IssueDate: "2024-01-20",
			Amount:    50000,
			Type:      "expense",
			Status:    "unsettled",
			Details:   []mockserver.DealDetail{},
		})

		resp, err := accountingClient.Deals().Get(ctx, companyID, 100, nil)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if resp.Deal.Id != 100 {
			t.Errorf("Expected deal ID 100, got %d", resp.Deal.Id)
		}
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		server.ClearDeals()
		server.AddDeal(&mockserver.Deal{
			ID:        200,
			CompanyID: companyID,
			IssueDate: "2024-01-15",
			Amount:    10000,
			Type:      "expense",
			Status:    "unsettled",
			Details: []mockserver.DealDetail{
				{Amount: 10000},
			},
		})

		updateDesc := "Updated description"
		accountItemID := int64(12345)
		taxCode := int64(108)
		params := gen.DealUpdateParams{
			CompanyId: companyID,
			IssueDate: "2024-01-25",
			Type:      gen.DealUpdateParamsTypeExpense,
			Details: []struct {
				AccountItemId  int64   `json:"account_item_id"`
				Amount         int64   `json:"amount"`
				Description    *string `json:"description,omitempty"`
				Id             *int64  `json:"id,omitempty"`
				ItemId         *int64  `json:"item_id,omitempty"`
				SectionId      *int64  `json:"section_id,omitempty"`
				Segment1TagId  *int64  `json:"segment_1_tag_id,omitempty"`
				Segment2TagId  *int64  `json:"segment_2_tag_id,omitempty"`
				Segment3TagId  *int64  `json:"segment_3_tag_id,omitempty"`
				TagIds         *[]int64 `json:"tag_ids,omitempty"`
				TaxCode        int64   `json:"tax_code"`
				Vat            *int64  `json:"vat,omitempty"`
			}{
				{
					AccountItemId: accountItemID,
					TaxCode:       taxCode,
					Amount:        15000,
					Description:   &updateDesc,
				},
			},
		}

		resp, err := accountingClient.Deals().Update(ctx, 200, params)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		if resp.Deal.IssueDate != "2024-01-25" {
			t.Errorf("Expected issue_date '2024-01-25', got %s", resp.Deal.IssueDate)
		}
		if resp.Deal.Amount != 15000 {
			t.Errorf("Expected amount 15000, got %d", resp.Deal.Amount)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		server.ClearDeals()
		server.AddDeal(&mockserver.Deal{
			ID:        300,
			CompanyID: companyID,
			IssueDate: "2024-01-15",
			Amount:    10000,
			Type:      "expense",
			Status:    "unsettled",
		})

		err := accountingClient.Deals().Delete(ctx, companyID, 300)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify it's deleted
		_, err = accountingClient.Deals().Get(ctx, companyID, 300, nil)
		if err == nil {
			t.Error("Expected error when getting deleted deal")
		}
	})
}

// TestDealsService_NotFound tests 404 error handling.
func TestDealsService_NotFound(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().Get(ctx, 1, 99999, nil)
	if err == nil {
		t.Error("Expected error for non-existent deal")
	}
}

// createTestAccountingClient creates an accounting client for testing.
func createTestAccountingClient(t *testing.T, server *mockserver.Server) *accounting.Client {
	t.Helper()

	// Create a token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Create a static token source
	ts := oauth2.StaticTokenSource(token)

	// Create the base client
	baseClient := client.NewClient(
		client.WithBaseURL(server.URL),
		client.WithHTTPClient(&http.Client{
			Transport: &oauth2.Transport{
				Source: ts,
				Base:   http.DefaultTransport,
			},
		}),
	)

	// Create the accounting client
	accountingClient, err := accounting.NewClient(baseClient)
	if err != nil {
		t.Fatalf("Failed to create accounting client: %v", err)
	}

	return accountingClient
}
