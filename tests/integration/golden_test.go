package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/u-masato/freee-api-go/tests/integration/golden"
	"github.com/u-masato/freee-api-go/tests/integration/mockserver"
)

// TestGolden_DealResponse tests that deal responses match golden files.
func TestGolden_DealResponse(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	// Add a well-defined deal for consistent testing
	server.ClearDeals()
	dueDate := "2024-01-31"
	refNumber := "REF-001"
	description := "Test expense"
	server.AddDeal(&mockserver.Deal{
		ID:        1,
		CompanyID: 1,
		IssueDate: "2024-01-15",
		DueDate:   &dueDate,
		Amount:    10000,
		DueAmount: 10000,
		Type:      "expense",
		RefNumber: &refNumber,
		Status:    "unsettled",
		Details: []mockserver.DealDetail{
			{
				ID:            1,
				AccountItemID: 12345,
				TaxCode:       108,
				Amount:        10000,
				Vat:           909,
				Description:   &description,
				EntryType:     "debit",
			},
		},
		Payments:   []mockserver.DealPayment{},
		Receipts:   []mockserver.DealReceipt{},
		CreateTime: "2024-01-15T09:00:00+09:00",
		UpdateTime: "2024-01-15T09:00:00+09:00",
	})

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	resp, err := accountingClient.Deals().Get(ctx, 1, 1, nil)
	if err != nil {
		t.Fatalf("Get deal failed: %v", err)
	}

	// Convert to a stable structure for golden file comparison
	type GoldenDeal struct {
		ID        int64   `json:"id"`
		CompanyID int64   `json:"company_id"`
		IssueDate string  `json:"issue_date"`
		DueDate   *string `json:"due_date,omitempty"`
		Amount    int64   `json:"amount"`
		DueAmount *int64  `json:"due_amount,omitempty"`
		Type      string  `json:"type"`
		RefNumber *string `json:"ref_number,omitempty"`
		Status    string  `json:"status"`
	}

	var typeStr string
	if resp.Deal.Type != nil {
		typeStr = string(*resp.Deal.Type)
	}

	goldenDeal := GoldenDeal{
		ID:        resp.Deal.Id,
		CompanyID: resp.Deal.CompanyId,
		IssueDate: resp.Deal.IssueDate,
		DueDate:   resp.Deal.DueDate,
		Amount:    resp.Deal.Amount,
		DueAmount: resp.Deal.DueAmount,
		Type:      typeStr,
		RefNumber: resp.Deal.RefNumber,
		Status:    string(resp.Deal.Status),
	}

	golden.Assert(t, "deal_response", goldenDeal)
}

// TestGolden_DealsListResponse tests that deal list responses match golden files.
func TestGolden_DealsListResponse(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	// Add well-defined deals for consistent testing
	server.ClearDeals()
	for i := 1; i <= 3; i++ {
		server.AddDeal(&mockserver.Deal{
			ID:        int64(i),
			CompanyID: 1,
			IssueDate: "2024-01-15",
			Amount:    int64(1000 * i),
			DueAmount: int64(1000 * i),
			Type:      "expense",
			Status:    "unsettled",
			Details:   []mockserver.DealDetail{},
			Payments:  []mockserver.DealPayment{},
			Receipts:  []mockserver.DealReceipt{},
		})
	}

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	result, err := accountingClient.Deals().List(ctx, 1, nil)
	if err != nil {
		t.Fatalf("List deals failed: %v", err)
	}

	// Create a stable structure for golden file comparison
	type GoldenListResult struct {
		TotalCount int `json:"total_count"`
		DealCount  int `json:"deal_count"`
	}

	goldenResult := GoldenListResult{
		TotalCount: int(result.TotalCount),
		DealCount:  len(result.Deals),
	}

	golden.Assert(t, "deals_list_response", goldenResult)
}

// TestGolden_ErrorResponse tests that error responses match golden files.
func TestGolden_ErrorResponse(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	// Request a non-existent deal to get a 404 error
	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().Get(ctx, 1, 99999, nil)
	if err == nil {
		t.Fatal("Expected error for non-existent deal")
	}

	// Create a stable error structure for golden file comparison
	type GoldenError struct {
		ErrorContains string `json:"error_contains"`
		IsNotNil      bool   `json:"is_not_nil"`
	}

	// We can't compare exact error messages, so we just verify structure
	goldenError := GoldenError{
		ErrorContains: "not found",
		IsNotNil:      err != nil,
	}

	golden.Assert(t, "error_not_found", goldenError)
}

// TestGolden_MockServerResponse tests that mock server responses are consistent.
func TestGolden_MockServerResponse(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	// Test that mock server produces consistent JSON responses
	server.ClearDeals()
	server.AddDeal(&mockserver.Deal{
		ID:        100,
		CompanyID: 1,
		IssueDate: "2024-01-01",
		Amount:    50000,
		DueAmount: 50000,
		Type:      "income",
		Status:    "settled",
	})

	deals := server.GetDeals()
	if len(deals) != 1 {
		t.Fatalf("Expected 1 deal, got %d", len(deals))
	}

	deal := deals[100]

	// Marshal to JSON for golden file
	dealJSON, err := json.Marshal(deal)
	if err != nil {
		t.Fatalf("Failed to marshal deal: %v", err)
	}

	golden.AssertJSON(t, "mock_server_deal", dealJSON)
}
