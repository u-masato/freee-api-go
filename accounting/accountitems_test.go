package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muno/freee-api-go/client"
)

func TestAccountItemsService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListAccountItemsOptions
		mockStatus int
		mockBody   string
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "successful list with no options",
			companyID:  1,
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: `{
				"account_items": [
					{"id": 1, "name": "現金", "account_category": "現金・預金", "account_category_id": 1, "available": true, "tax_code": 0, "searchable": 2, "categories": []},
					{"id": 2, "name": "売上高", "account_category": "売上高", "account_category_id": 10, "available": true, "tax_code": 1, "searchable": 2, "categories": []}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with base date filter",
			companyID: 1,
			opts: &ListAccountItemsOptions{
				BaseDate: stringPtr("2024-01-01"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"account_items": [
					{"id": 1, "name": "現金", "account_category": "現金・預金", "account_category_id": 1, "available": true, "tax_code": 0, "searchable": 2, "categories": []}
				]
			}`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "successful list with date range filter",
			companyID: 1,
			opts: &ListAccountItemsOptions{
				StartUpdateDate: stringPtr("2024-01-01"),
				EndUpdateDate:   stringPtr("2024-01-31"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"account_items": [
					{"id": 3, "name": "仕入高", "account_category": "仕入高", "account_category_id": 20, "available": true, "tax_code": 2, "searchable": 2, "categories": []}
				]
			}`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:       "empty result",
			companyID:  1,
			opts:       nil,
			mockStatus: http.StatusOK,
			mockBody: `{
				"account_items": []
			}`,
			wantErr:   false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/1/account_items" {
					t.Errorf("expected path /api/1/account_items, got %s", r.URL.Path)
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

			accountItemsService := accountingClient.AccountItems()
			result, err := accountItemsService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.AccountItems.AccountItems) != tt.wantCount {
					t.Errorf("List() got %d account items, want %d", len(result.AccountItems.AccountItems), tt.wantCount)
				}
				if result.Count != tt.wantCount {
					t.Errorf("List() got count %d, want %d", result.Count, tt.wantCount)
				}
			}
		})
	}
}

func TestAccountItemsService_Get(t *testing.T) {
	tests := []struct {
		name              string
		companyID         int64
		accountItemID     int64
		mockStatus        int
		mockBody          string
		wantErr           bool
		wantAccountItemID int64
	}{
		{
			name:          "successful get",
			companyID:     1,
			accountItemID: 123,
			mockStatus:    http.StatusOK,
			mockBody: `{
				"account_item": {
					"id": 123,
					"company_id": 1,
					"name": "現金",
					"account_category": "現金・預金",
					"account_category_id": 1,
					"available": true,
					"tax_code": 0,
					"searchable": 2
				}
			}`,
			wantErr:           false,
			wantAccountItemID: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/account_items/123"
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

			accountItemsService := accountingClient.AccountItems()
			item, err := accountItemsService.Get(context.Background(), tt.companyID, tt.accountItemID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if item == nil {
					t.Error("Get() returned nil item")
					return
				}
				if item.AccountItem.Id != tt.wantAccountItemID {
					t.Errorf("Get() got account item ID %d, want %d", item.AccountItem.Id, tt.wantAccountItemID)
				}
			}
		})
	}
}

func TestAccountItemsService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		companyID     int64
		accountItemID int64
		mockStatus    int
		wantErr       bool
	}{
		{
			name:          "successful delete",
			companyID:     1,
			accountItemID: 123,
			mockStatus:    http.StatusNoContent,
			wantErr:       false,
		},
		{
			name:          "delete with 200 OK",
			companyID:     1,
			accountItemID: 456,
			mockStatus:    http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "delete not found",
			companyID:     1,
			accountItemID: 999,
			mockStatus:    http.StatusNotFound,
			wantErr:       true,
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

			accountItemsService := accountingClient.AccountItems()
			err = accountItemsService.Delete(context.Background(), tt.companyID, tt.accountItemID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
