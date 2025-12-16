package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/internal/gen"
)

func TestItemsService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListItemsOptions
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
				"items": [
					{"id": 1, "company_id": 1, "name": "商品A", "available": true, "update_date": "2024-01-15"},
					{"id": 2, "company_id": 1, "name": "商品B", "available": true, "update_date": "2024-01-16"}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with pagination",
			companyID: 1,
			opts: &ListItemsOptions{
				Limit:  int64Ptr(100),
				Offset: int64Ptr(0),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"items": [
					{"id": 3, "company_id": 1, "name": "商品C", "available": true, "update_date": "2024-01-17"}
				]
			}`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "successful list with date range filter",
			companyID: 1,
			opts: &ListItemsOptions{
				StartUpdateDate: stringPtr("2024-01-01"),
				EndUpdateDate:   stringPtr("2024-01-31"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"items": [
					{"id": 4, "company_id": 1, "name": "商品D", "available": true, "update_date": "2024-01-10"}
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
				"items": []
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
				if r.URL.Path != "/api/1/items" {
					t.Errorf("expected path /api/1/items, got %s", r.URL.Path)
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

			itemsService := accountingClient.Items()
			result, err := itemsService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.Items) != tt.wantCount {
					t.Errorf("List() got %d items, want %d", len(result.Items), tt.wantCount)
				}
				if result.Count != tt.wantCount {
					t.Errorf("List() got count %d, want %d", result.Count, tt.wantCount)
				}
			}
		})
	}
}

func TestItemsService_Get(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		itemID     int64
		mockStatus int
		mockBody   string
		wantErr    bool
		wantItemID int64
	}{
		{
			name:       "successful get",
			companyID:  1,
			itemID:     123,
			mockStatus: http.StatusOK,
			mockBody: `{
				"item": {
					"id": 123,
					"company_id": 1,
					"name": "テスト商品",
					"available": true,
					"update_date": "2024-01-15"
				}
			}`,
			wantErr:    false,
			wantItemID: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/items/123"
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

			itemsService := accountingClient.Items()
			item, err := itemsService.Get(context.Background(), tt.companyID, tt.itemID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if item == nil {
					t.Error("Get() returned nil item")
					return
				}
				if item.Item.Id != tt.wantItemID {
					t.Errorf("Get() got item ID %d, want %d", item.Item.Id, tt.wantItemID)
				}
			}
		})
	}
}

func TestItemsService_Create(t *testing.T) {
	mockStatus := http.StatusCreated
	mockBody := `{
		"item": {
			"id": 999,
			"company_id": 1,
			"name": "新規品目",
			"available": true,
			"update_date": "2024-01-25"
		}
	}`
	wantItemID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/items" {
			t.Errorf("expected path /api/1/items, got %s", r.URL.Path)
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

	params := gen.ItemParams{
		CompanyId: 1,
		Name:      "新規品目",
	}

	itemsService := accountingClient.Items()
	item, err := itemsService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}

	if item == nil {
		t.Error("Create() returned nil item")
		return
	}
	if item.Item.Id != wantItemID {
		t.Errorf("Create() got item ID %d, want %d", item.Item.Id, wantItemID)
	}
}

func TestItemsService_Update(t *testing.T) {
	itemID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"item": {
			"id": 123,
			"company_id": 1,
			"name": "更新後品目",
			"available": true,
			"update_date": "2024-01-26"
		}
	}`
	wantItemID := int64(123)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/1/items/123"
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

	params := gen.ItemParams{
		CompanyId: 1,
		Name:      "更新後品目",
	}

	itemsService := accountingClient.Items()
	item, err := itemsService.Update(context.Background(), itemID, params)

	if err != nil {
		t.Errorf("Update() unexpected error = %v", err)
		return
	}

	if item == nil {
		t.Error("Update() returned nil item")
		return
	}
	if item.Item.Id != wantItemID {
		t.Errorf("Update() got item ID %d, want %d", item.Item.Id, wantItemID)
	}
}

func TestItemsService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		itemID     int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			companyID:  1,
			itemID:     123,
			mockStatus: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "delete with 200 OK",
			companyID:  1,
			itemID:     456,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "delete not found",
			companyID:  1,
			itemID:     999,
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

			itemsService := accountingClient.Items()
			err = itemsService.Delete(context.Background(), tt.companyID, tt.itemID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestItemsService_ListIter(t *testing.T) {
	tests := []struct {
		name        string
		companyID   int64
		opts        *ListItemsOptions
		mockPages   []string
		wantErr     bool
		wantCount   int
		wantFetches int
	}{
		{
			name:      "single page iteration",
			companyID: 1,
			opts:      nil,
			mockPages: []string{
				`{
					"items": [
						{"id": 1, "company_id": 1, "name": "品目1", "available": true, "update_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "name": "品目2", "available": true, "update_date": "2024-01-16"}
					]
				}`,
			},
			wantErr:     false,
			wantCount:   2,
			wantFetches: 1,
		},
		{
			name:      "multiple page iteration",
			companyID: 1,
			opts: &ListItemsOptions{
				Limit: int64Ptr(2),
			},
			mockPages: []string{
				`{
					"items": [
						{"id": 1, "company_id": 1, "name": "品目1", "available": true, "update_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "name": "品目2", "available": true, "update_date": "2024-01-16"}
					]
				}`,
				`{
					"items": [
						{"id": 3, "company_id": 1, "name": "品目3", "available": true, "update_date": "2024-01-17"}
					]
				}`,
			},
			wantErr:     false,
			wantCount:   3,
			wantFetches: 2,
		},
		{
			name:      "empty result",
			companyID: 1,
			opts:      nil,
			mockPages: []string{
				`{
					"items": []
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
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"items": []}`))
				}
			}))
			defer server.Close()

			baseClient := client.NewClient(
				client.WithBaseURL(server.URL),
				client.WithHTTPClient(server.Client()),
			)
			accountingClient, err := NewClient(baseClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			itemsService := accountingClient.Items()
			iter := itemsService.ListIter(context.Background(), tt.companyID, tt.opts)

			var items []gen.Item
			for iter.Next() {
				items = append(items, iter.Value())
			}

			if err := iter.Err(); (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(items) != tt.wantCount {
				t.Errorf("ListIter() got %d items, want %d", len(items), tt.wantCount)
			}

			if fetchCount != tt.wantFetches {
				t.Errorf("ListIter() made %d fetches, want %d", fetchCount, tt.wantFetches)
			}
		})
	}
}
