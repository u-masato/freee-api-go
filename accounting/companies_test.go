package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/u-masato/freee-api-go/client"
)

func TestCompaniesService_List(t *testing.T) {
	mockStatus := http.StatusOK
	mockBody := `{
		"companies": [
			{
				"id": 1,
				"company_number": "1234567890",
				"display_name": "Company A",
				"name": "Company A",
				"role": "admin"
			},
			{
				"id": 2,
				"company_number": "0987654321",
				"display_name": "Company B",
				"name": "Company B",
				"role": "read_only"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/companies" {
			t.Errorf("expected path /api/1/companies, got %s", r.URL.Path)
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

	companiesService := accountingClient.Companies()
	result, err := companiesService.List(context.Background())

	if err != nil {
		t.Errorf("List() unexpected error = %v", err)
		return
	}
	if result == nil {
		t.Error("List() returned nil result")
		return
	}
	if result.Count != 2 {
		t.Errorf("List() got count %d, want %d", result.Count, 2)
	}
}

func TestCompaniesService_Get(t *testing.T) {
	companyID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"company": {
			"id": 123,
			"company_number": "1234567890",
			"display_name": "Test Company",
			"corporate_number": "1234567890123",
			"amount_fraction": 0,
			"fiscal_years": []
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		expectedPath := "/api/1/companies/123"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("details") != "true" {
			t.Errorf("expected details=true, got %s", query.Get("details"))
		}
		if query.Get("walletables") != "true" {
			t.Errorf("expected walletables=true, got %s", query.Get("walletables"))
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

	opts := &GetCompanyOptions{
		Details:     boolPtr(true),
		Walletables: boolPtr(true),
	}

	companiesService := accountingClient.Companies()
	company, err := companiesService.Get(context.Background(), companyID, opts)

	if err != nil {
		t.Errorf("Get() unexpected error = %v", err)
		return
	}
	if company == nil {
		t.Error("Get() returned nil company")
		return
	}
	if company.Company.Id != companyID {
		t.Errorf("Get() got company ID %d, want %d", company.Company.Id, companyID)
	}
}
