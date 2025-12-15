package integration

import (
	"context"
	"testing"

	"github.com/muno/freee-api-go/accounting"
	"github.com/muno/freee-api-go/tests/integration/mockserver"
)

// TestPagination_DealsListIter tests the pagination iterator for deals.
func TestPagination_DealsListIter(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()
	companyID := int64(1)

	// Add 25 deals to test pagination (default page size is 20)
	server.ClearDeals()
	for i := 0; i < 25; i++ {
		server.AddDeal(&mockserver.Deal{
			CompanyID: companyID,
			IssueDate: "2024-01-15",
			Amount:    int64(1000 * (i + 1)),
			Type:      "expense",
			Status:    "unsettled",
		})
	}

	t.Run("DefaultPageSize", func(t *testing.T) {
		iter := accountingClient.Deals().ListIter(ctx, companyID, nil)

		count := 0
		for iter.Next() {
			deal := iter.Value()
			if deal.Id == 0 {
				t.Error("Deal should have an ID")
			}
			count++
		}

		if err := iter.Err(); err != nil {
			t.Fatalf("Iterator error: %v", err)
		}

		if count != 25 {
			t.Errorf("Expected to iterate through 25 deals, got %d", count)
		}
	})

	t.Run("CustomPageSize", func(t *testing.T) {
		limit := int64(5)
		opts := &accounting.ListDealsOptions{
			Limit: &limit,
		}
		iter := accountingClient.Deals().ListIter(ctx, companyID, opts)

		count := 0
		for iter.Next() {
			count++
		}

		if err := iter.Err(); err != nil {
			t.Fatalf("Iterator error: %v", err)
		}

		if count != 25 {
			t.Errorf("Expected to iterate through 25 deals with custom page size, got %d", count)
		}
	})

	t.Run("WithFilter", func(t *testing.T) {
		// Clear and add mixed deals
		server.ClearDeals()
		for i := 0; i < 15; i++ {
			server.AddDeal(&mockserver.Deal{
				CompanyID: companyID,
				IssueDate: "2024-01-15",
				Amount:    int64(1000 * (i + 1)),
				Type:      "expense",
				Status:    "unsettled",
			})
		}
		for i := 0; i < 10; i++ {
			server.AddDeal(&mockserver.Deal{
				CompanyID: companyID,
				IssueDate: "2024-01-15",
				Amount:    int64(2000 * (i + 1)),
				Type:      "income",
				Status:    "settled",
			})
		}

		typeFilter := "expense"
		opts := &accounting.ListDealsOptions{
			Type: &typeFilter,
		}
		iter := accountingClient.Deals().ListIter(ctx, companyID, opts)

		count := 0
		for iter.Next() {
			count++
		}

		if err := iter.Err(); err != nil {
			t.Fatalf("Iterator error: %v", err)
		}

		if count != 15 {
			t.Errorf("Expected to iterate through 15 expense deals, got %d", count)
		}
	})

	t.Run("EmptyResult", func(t *testing.T) {
		server.ClearDeals()

		iter := accountingClient.Deals().ListIter(ctx, companyID, nil)

		count := 0
		for iter.Next() {
			count++
		}

		if err := iter.Err(); err != nil {
			t.Fatalf("Iterator error: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected 0 deals, got %d", count)
		}
	})
}

// TestPagination_ManualPaging tests manual pagination using offset/limit.
func TestPagination_ManualPaging(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()
	companyID := int64(1)

	// Add 50 deals
	server.ClearDeals()
	for i := 0; i < 50; i++ {
		server.AddDeal(&mockserver.Deal{
			CompanyID: companyID,
			IssueDate: "2024-01-15",
			Amount:    int64(1000 * (i + 1)),
			Type:      "expense",
			Status:    "unsettled",
		})
	}

	t.Run("FirstPage", func(t *testing.T) {
		limit := int64(10)
		offset := int64(0)
		opts := &accounting.ListDealsOptions{
			Limit:  &limit,
			Offset: &offset,
		}

		result, err := accountingClient.Deals().List(ctx, companyID, opts)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if result.TotalCount != 50 {
			t.Errorf("Expected total count 50, got %d", result.TotalCount)
		}
		if len(result.Deals) != 10 {
			t.Errorf("Expected 10 deals on first page, got %d", len(result.Deals))
		}
	})

	t.Run("SecondPage", func(t *testing.T) {
		limit := int64(10)
		offset := int64(10)
		opts := &accounting.ListDealsOptions{
			Limit:  &limit,
			Offset: &offset,
		}

		result, err := accountingClient.Deals().List(ctx, companyID, opts)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if result.TotalCount != 50 {
			t.Errorf("Expected total count 50, got %d", result.TotalCount)
		}
		if len(result.Deals) != 10 {
			t.Errorf("Expected 10 deals on second page, got %d", len(result.Deals))
		}
	})

	t.Run("LastPage", func(t *testing.T) {
		limit := int64(10)
		offset := int64(45)
		opts := &accounting.ListDealsOptions{
			Limit:  &limit,
			Offset: &offset,
		}

		result, err := accountingClient.Deals().List(ctx, companyID, opts)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if result.TotalCount != 50 {
			t.Errorf("Expected total count 50, got %d", result.TotalCount)
		}
		if len(result.Deals) != 5 {
			t.Errorf("Expected 5 deals on last page, got %d", len(result.Deals))
		}
	})

	t.Run("BeyondLastPage", func(t *testing.T) {
		limit := int64(10)
		offset := int64(100)
		opts := &accounting.ListDealsOptions{
			Limit:  &limit,
			Offset: &offset,
		}

		result, err := accountingClient.Deals().List(ctx, companyID, opts)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if result.TotalCount != 50 {
			t.Errorf("Expected total count 50, got %d", result.TotalCount)
		}
		if len(result.Deals) != 0 {
			t.Errorf("Expected 0 deals beyond last page, got %d", len(result.Deals))
		}
	})
}

// TestPagination_LargeDataset tests pagination with a large dataset.
func TestPagination_LargeDataset(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()
	companyID := int64(1)

	// Add 500 deals to simulate a larger dataset
	server.ClearDeals()
	for i := 0; i < 500; i++ {
		server.AddDeal(&mockserver.Deal{
			CompanyID: companyID,
			IssueDate: "2024-01-15",
			Amount:    int64(1000 * (i + 1)),
			Type:      "expense",
			Status:    "unsettled",
		})
	}

	t.Run("IterateAll", func(t *testing.T) {
		limit := int64(100) // Max page size
		opts := &accounting.ListDealsOptions{
			Limit: &limit,
		}
		iter := accountingClient.Deals().ListIter(ctx, companyID, opts)

		count := 0
		for iter.Next() {
			count++
		}

		if err := iter.Err(); err != nil {
			t.Fatalf("Iterator error: %v", err)
		}

		if count != 500 {
			t.Errorf("Expected to iterate through 500 deals, got %d", count)
		}
	})
}
