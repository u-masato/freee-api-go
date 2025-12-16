package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/u-masato/freee-api-go/client"
	"github.com/u-masato/freee-api-go/internal/gen"
)

func TestPartnersService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListPartnersOptions
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
				"partners": [
					{"id": 1, "company_id": 1, "name": "株式会社テスト1", "available": true, "update_date": "2024-01-15"},
					{"id": 2, "company_id": 1, "name": "株式会社テスト2", "available": true, "update_date": "2024-01-16"}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with keyword filter",
			companyID: 1,
			opts: &ListPartnersOptions{
				Keyword: stringPtr("テスト"),
				Limit:   int64Ptr(100),
				Offset:  int64Ptr(0),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"partners": [
					{"id": 3, "company_id": 1, "name": "テスト商事", "available": true, "update_date": "2024-01-17"}
				]
			}`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "successful list with date range filter",
			companyID: 1,
			opts: &ListPartnersOptions{
				StartUpdateDate: stringPtr("2024-01-01"),
				EndUpdateDate:   stringPtr("2024-01-31"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"partners": [
					{"id": 4, "company_id": 1, "name": "株式会社ABC", "available": true, "update_date": "2024-01-10"}
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
				"partners": []
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
				if r.URL.Path != "/api/1/partners" {
					t.Errorf("expected path /api/1/partners, got %s", r.URL.Path)
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

			partnersService := accountingClient.Partners()
			result, err := partnersService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.Partners.Partners) != tt.wantCount {
					t.Errorf("List() got %d partners, want %d", len(result.Partners.Partners), tt.wantCount)
				}
				if result.Count != tt.wantCount {
					t.Errorf("List() got count %d, want %d", result.Count, tt.wantCount)
				}
			}
		})
	}
}

func TestPartnersService_Get(t *testing.T) {
	tests := []struct {
		name          string
		companyID     int64
		partnerID     int64
		mockStatus    int
		mockBody      string
		wantErr       bool
		wantPartnerID int64
	}{
		{
			name:       "successful get",
			companyID:  1,
			partnerID:  123,
			mockStatus: http.StatusOK,
			mockBody: `{
				"partner": {
					"id": 123,
					"company_id": 1,
					"name": "株式会社テスト",
					"available": true,
					"update_date": "2024-01-15"
				}
			}`,
			wantErr:       false,
			wantPartnerID: 123,
		},
		{
			name:       "get with full details",
			companyID:  1,
			partnerID:  456,
			mockStatus: http.StatusOK,
			mockBody: `{
				"partner": {
					"id": 456,
					"company_id": 1,
					"name": "テスト商事株式会社",
					"long_name": "テスト商事株式会社（正式名称）",
					"name_kana": "テストショウジカブシキガイシャ",
					"code": "P001",
					"available": true,
					"email": "test@example.com",
					"phone": "03-1234-5678",
					"update_date": "2024-01-20"
				}
			}`,
			wantErr:       false,
			wantPartnerID: 456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/partners/123"
				if tt.partnerID == 456 {
					expectedPath = "/api/1/partners/456"
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

			partnersService := accountingClient.Partners()
			partner, err := partnersService.Get(context.Background(), tt.companyID, tt.partnerID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if partner == nil {
					t.Error("Get() returned nil partner")
					return
				}
				if partner.Partner.Id != tt.wantPartnerID {
					t.Errorf("Get() got partner ID %d, want %d", partner.Partner.Id, tt.wantPartnerID)
				}
			}
		})
	}
}

func TestPartnersService_Create(t *testing.T) {
	mockStatus := http.StatusCreated
	mockBody := `{
		"partner": {
			"id": 999,
			"company_id": 1,
			"name": "新規取引先株式会社",
			"available": true,
			"update_date": "2024-01-25"
		}
	}`
	wantPartnerID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/partners" {
			t.Errorf("expected path /api/1/partners, got %s", r.URL.Path)
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

	params := gen.PartnerCreateParams{
		CompanyId: 1,
		Name:      "新規取引先株式会社",
	}

	partnersService := accountingClient.Partners()
	partner, err := partnersService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}

	if partner == nil {
		t.Error("Create() returned nil partner")
		return
	}
	if partner.Partner.Id != wantPartnerID {
		t.Errorf("Create() got partner ID %d, want %d", partner.Partner.Id, wantPartnerID)
	}
}

func TestPartnersService_Update(t *testing.T) {
	partnerID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"partner": {
			"id": 123,
			"company_id": 1,
			"name": "更新後取引先株式会社",
			"available": true,
			"update_date": "2024-01-26"
		}
	}`
	wantPartnerID := int64(123)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/1/partners/123"
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

	params := gen.PartnerUpdateParams{
		CompanyId: 1,
		Name:      "更新後取引先株式会社",
	}

	partnersService := accountingClient.Partners()
	partner, err := partnersService.Update(context.Background(), partnerID, params)

	if err != nil {
		t.Errorf("Update() unexpected error = %v", err)
		return
	}

	if partner == nil {
		t.Error("Update() returned nil partner")
		return
	}
	if partner.Partner.Id != wantPartnerID {
		t.Errorf("Update() got partner ID %d, want %d", partner.Partner.Id, wantPartnerID)
	}
}

func TestPartnersService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		partnerID  int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			companyID:  1,
			partnerID:  123,
			mockStatus: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "delete with 200 OK",
			companyID:  1,
			partnerID:  456,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "delete not found",
			companyID:  1,
			partnerID:  999,
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

			partnersService := accountingClient.Partners()
			err = partnersService.Delete(context.Background(), tt.companyID, tt.partnerID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPartnersService_ListIter(t *testing.T) {
	tests := []struct {
		name        string
		companyID   int64
		opts        *ListPartnersOptions
		mockPages   []string
		wantErr     bool
		wantCount   int
		wantFetches int
	}{
		{
			name:      "single page iteration",
			companyID: 1,
			opts: &ListPartnersOptions{
				Keyword: stringPtr("テスト"),
			},
			mockPages: []string{
				`{
					"partners": [
						{"id": 1, "company_id": 1, "name": "テスト1", "available": true, "update_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "name": "テスト2", "available": true, "update_date": "2024-01-16"}
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
			opts: &ListPartnersOptions{
				Limit: int64Ptr(2),
			},
			mockPages: []string{
				// Page 1 (offset=0, limit=2)
				`{
					"partners": [
						{"id": 1, "company_id": 1, "name": "取引先1", "available": true, "update_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "name": "取引先2", "available": true, "update_date": "2024-01-16"}
					]
				}`,
				// Page 2 (offset=2, limit=2)
				`{
					"partners": [
						{"id": 3, "company_id": 1, "name": "取引先3", "available": true, "update_date": "2024-01-17"},
						{"id": 4, "company_id": 1, "name": "取引先4", "available": true, "update_date": "2024-01-18"}
					]
				}`,
				// Page 3 (offset=4, limit=2) - partial page indicates end
				`{
					"partners": [
						{"id": 5, "company_id": 1, "name": "取引先5", "available": true, "update_date": "2024-01-19"}
					]
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
					"partners": []
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
					w.Write([]byte(`{"partners": []}`))
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

			partnersService := accountingClient.Partners()

			// Create iterator
			iter := partnersService.ListIter(context.Background(), tt.companyID, tt.opts)

			// Iterate through results
			var partners []PartnerListItem
			for iter.Next() {
				partners = append(partners, iter.Value())
			}

			// Check error
			if err := iter.Err(); (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify count
			if len(partners) != tt.wantCount {
				t.Errorf("ListIter() got %d partners, want %d", len(partners), tt.wantCount)
			}

			// Verify fetch count
			if fetchCount != tt.wantFetches {
				t.Errorf("ListIter() made %d fetches, want %d", fetchCount, tt.wantFetches)
			}

			// Verify partner IDs are sequential
			for i, partner := range partners {
				expectedID := int64(i + 1)
				if partner.Id != expectedID {
					t.Errorf("partner %d: expected ID %d, got %d", i, expectedID, partner.Id)
				}
			}
		})
	}
}
