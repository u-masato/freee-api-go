package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/internal/gen"
)

func TestDealsService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListDealsOptions
		mockStatus int
		mockBody   string
		wantErr    bool
		wantCount  int
		wantTotal  int64
	}{
		{
			name:       "successful list with no options",
			companyID:  1,
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: `{
				"deals": [
					{"id": 1, "company_id": 1, "amount": 10000, "issue_date": "2024-01-15"},
					{"id": 2, "company_id": 1, "amount": 20000, "issue_date": "2024-01-16"}
				],
				"meta": {"total_count": 2}
			}`,
			wantErr:   false,
			wantCount: 2,
			wantTotal: 2,
		},
		{
			name:      "successful list with options",
			companyID: 1,
			opts: &ListDealsOptions{
				Type:   stringPtr("expense"),
				Limit:  int64Ptr(10),
				Offset: int64Ptr(0),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"deals": [
					{"id": 3, "company_id": 1, "amount": 5000, "issue_date": "2024-01-17"}
				],
				"meta": {"total_count": 1}
			}`,
			wantErr:   false,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:       "empty result",
			companyID:  1,
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: `{
				"deals": [],
				"meta": {"total_count": 0}
			}`,
			wantErr:   false,
			wantCount: 0,
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/1/deals" {
					t.Errorf("expected path /api/1/deals, got %s", r.URL.Path)
				}

				// Verify query parameters
				query := r.URL.Query()
				if query.Get("company_id") != "1" {
					t.Errorf("expected company_id=1, got %s", query.Get("company_id"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockBody))
			}))
			defer server.Close()

			baseClient := client.NewClient(client.WithBaseURL(server.URL))
			accountingClient, err := NewClient(baseClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			dealsService := accountingClient.Deals()
			result, err := dealsService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.Deals) != tt.wantCount {
					t.Errorf("List() got %d deals, want %d", len(result.Deals), tt.wantCount)
				}
				if result.TotalCount != tt.wantTotal {
					t.Errorf("List() got total count %d, want %d", result.TotalCount, tt.wantTotal)
				}
			}
		})
	}
}

func TestDealsService_Get(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		dealID     int64
		opts       *GetDealOptions
		mockStatus int
		mockBody   string
		wantErr    bool
		wantDealID int64
	}{
		{
			name:       "successful get",
			companyID:  1,
			dealID:     123,
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: `{
				"deal": {
					"id": 123,
					"company_id": 1,
					"amount": 10000,
					"issue_date": "2024-01-15"
				}
			}`,
			wantErr:    false,
			wantDealID: 123,
		},
		{
			name:      "successful get with accruals",
			companyID: 1,
			dealID:    456,
			opts: &GetDealOptions{
				Accruals: stringPtr("with"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"deal": {
					"id": 456,
					"company_id": 1,
					"amount": 20000,
					"issue_date": "2024-01-16"
				}
			}`,
			wantErr:    false,
			wantDealID: 456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/deals/123"
				if tt.dealID == 456 {
					expectedPath = "/api/1/deals/456"
				}
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockBody))
			}))
			defer server.Close()

			baseClient := client.NewClient(client.WithBaseURL(server.URL))
			accountingClient, err := NewClient(baseClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			dealsService := accountingClient.Deals()
			deal, err := dealsService.Get(context.Background(), tt.companyID, tt.dealID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if deal == nil {
					t.Error("Get() returned nil deal")
					return
				}
				if deal.Deal.Id != tt.wantDealID {
					t.Errorf("Get() got deal ID %d, want %d", deal.Deal.Id, tt.wantDealID)
				}
			}
		})
	}
}

func TestDealsService_Create(t *testing.T) {
	// Note: We skip full struct initialization for brevity in tests
	// The actual usage would include properly populated Details slices
	mockStatus := http.StatusCreated
	mockBody := `{
		"deal": {
			"id": 999,
			"company_id": 1,
			"amount": 10000,
			"issue_date": "2024-01-15"
		}
	}`
	wantDealID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/deals" {
			t.Errorf("expected path /api/1/deals, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(mockStatus)
		w.Write([]byte(mockBody))
	}))
	defer server.Close()

	baseClient := client.NewClient(client.WithBaseURL(server.URL))
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test with minimal params (Details would be populated in real usage)
	params := gen.DealCreateParams{
		CompanyId: 1,
		IssueDate: "2024-01-15",
		Type:      "expense",
		Details:   nil, // Simplified for testing
	}

	dealsService := accountingClient.Deals()
	deal, err := dealsService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}

	if deal == nil {
		t.Error("Create() returned nil deal")
		return
	}
	if deal.Deal.Id != wantDealID {
		t.Errorf("Create() got deal ID %d, want %d", deal.Deal.Id, wantDealID)
	}
}

func TestDealsService_Update(t *testing.T) {
	dealID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"deal": {
			"id": 123,
			"company_id": 1,
			"amount": 15000,
			"issue_date": "2024-01-20"
		}
	}`
	wantDealID := int64(123)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/1/deals/123"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(mockStatus)
		w.Write([]byte(mockBody))
	}))
	defer server.Close()

	baseClient := client.NewClient(client.WithBaseURL(server.URL))
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test with minimal params (Details would be populated in real usage)
	params := gen.DealUpdateParams{
		CompanyId: 1,
		IssueDate: "2024-01-20",
		Type:      "expense",
		Details:   nil, // Simplified for testing
	}

	dealsService := accountingClient.Deals()
	deal, err := dealsService.Update(context.Background(), dealID, params)

	if err != nil {
		t.Errorf("Update() unexpected error = %v", err)
		return
	}

	if deal == nil {
		t.Error("Update() returned nil deal")
		return
	}
	if deal.Deal.Id != wantDealID {
		t.Errorf("Update() got deal ID %d, want %d", deal.Deal.Id, wantDealID)
	}
}

func TestDealsService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		dealID     int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			companyID:  1,
			dealID:     123,
			mockStatus: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "delete with 200 OK",
			companyID:  1,
			dealID:     456,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "delete not found",
			companyID:  1,
			dealID:     999,
			mockStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE request, got %s", r.Method)
				}

				w.WriteHeader(tt.mockStatus)
			}))
			defer server.Close()

			baseClient := client.NewClient(client.WithBaseURL(server.URL))
			accountingClient, err := NewClient(baseClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			dealsService := accountingClient.Deals()
			err = dealsService.Delete(context.Background(), tt.companyID, tt.dealID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestDealsService_ListIter(t *testing.T) {
	tests := []struct {
		name        string
		companyID   int64
		opts        *ListDealsOptions
		mockPages   []string
		wantErr     bool
		wantCount   int
		wantFetches int
	}{
		{
			name:      "single page iteration",
			companyID: 1,
			opts: &ListDealsOptions{
				Type: stringPtr("expense"),
			},
			mockPages: []string{
				`{
					"deals": [
						{"id": 1, "company_id": 1, "amount": 10000, "issue_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "amount": 20000, "issue_date": "2024-01-16"}
					],
					"meta": {"total_count": 2}
				}`,
			},
			wantErr:     false,
			wantCount:   2,
			wantFetches: 1,
		},
		{
			name:      "multiple page iteration",
			companyID: 1,
			opts: &ListDealsOptions{
				Type:  stringPtr("expense"),
				Limit: int64Ptr(2),
			},
			mockPages: []string{
				// Page 1 (offset=0, limit=2)
				`{
					"deals": [
						{"id": 1, "company_id": 1, "amount": 10000, "issue_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "amount": 20000, "issue_date": "2024-01-16"}
					],
					"meta": {"total_count": 5}
				}`,
				// Page 2 (offset=2, limit=2)
				`{
					"deals": [
						{"id": 3, "company_id": 1, "amount": 30000, "issue_date": "2024-01-17"},
						{"id": 4, "company_id": 1, "amount": 40000, "issue_date": "2024-01-18"}
					],
					"meta": {"total_count": 5}
				}`,
				// Page 3 (offset=4, limit=2)
				`{
					"deals": [
						{"id": 5, "company_id": 1, "amount": 50000, "issue_date": "2024-01-19"}
					],
					"meta": {"total_count": 5}
				}`,
			},
			wantErr:     false,
			wantCount:   5,
			wantFetches: 3,
		},
		{
			name:      "empty result",
			companyID: 1,
			opts:      nil,
			mockPages: []string{
				`{
					"deals": [],
					"meta": {"total_count": 0}
				}`,
			},
			wantErr:     false,
			wantCount:   0,
			wantFetches: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if fetchCount < len(tt.mockPages) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(tt.mockPages[fetchCount]))
					fetchCount++
				} else {
					// Return empty page if we've exhausted mock pages
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"deals": [], "meta": {"total_count": 0}}`))
				}
			}))
			defer server.Close()

			// Create client
			baseClient := client.NewClient(
				client.WithBaseURL(server.URL),
				client.WithHTTPClient(server.Client()),
			)
			accountingClient, err := NewClient(baseClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			dealsService := accountingClient.Deals()

			// Create iterator
			iter := dealsService.ListIter(context.Background(), tt.companyID, tt.opts)

			// Iterate through results
			var deals []gen.Deal
			for iter.Next() {
				deals = append(deals, iter.Value())
			}

			// Check error
			if err := iter.Err(); (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify count
			if len(deals) != tt.wantCount {
				t.Errorf("ListIter() got %d deals, want %d", len(deals), tt.wantCount)
			}

			// Verify fetch count
			if fetchCount != tt.wantFetches {
				t.Errorf("ListIter() made %d fetches, want %d", fetchCount, tt.wantFetches)
			}

			// Verify deal IDs are sequential
			for i, deal := range deals {
				expectedID := int64(i + 1)
				if deal.Id != expectedID {
					t.Errorf("deal %d: expected ID %d, got %d", i, expectedID, deal.Id)
				}
			}
		})
	}
}
