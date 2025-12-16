package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/u-masato/freee-api-go/client"
	"github.com/u-masato/freee-api-go/internal/gen"
)

func TestSectionsService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListSectionsOptions
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
				"sections": [
					{"id": 1, "company_id": 1, "name": "営業部", "available": true},
					{"id": 2, "company_id": 1, "name": "開発部", "available": true}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with date range filter",
			companyID: 1,
			opts: &ListSectionsOptions{
				StartUpdateDate: stringPtr("2024-01-01"),
				EndUpdateDate:   stringPtr("2024-01-31"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"sections": [
					{"id": 3, "company_id": 1, "name": "経理部", "available": true}
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
				"sections": []
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
				if r.URL.Path != "/api/1/sections" {
					t.Errorf("expected path /api/1/sections, got %s", r.URL.Path)
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

			sectionsService := accountingClient.Sections()
			result, err := sectionsService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.Sections) != tt.wantCount {
					t.Errorf("List() got %d sections, want %d", len(result.Sections), tt.wantCount)
				}
				if result.Count != tt.wantCount {
					t.Errorf("List() got count %d, want %d", result.Count, tt.wantCount)
				}
			}
		})
	}
}

func TestSectionsService_Get(t *testing.T) {
	tests := []struct {
		name          string
		companyID     int64
		sectionID     int64
		mockStatus    int
		mockBody      string
		wantErr       bool
		wantSectionID int64
	}{
		{
			name:       "successful get",
			companyID:  1,
			sectionID:  123,
			mockStatus: http.StatusOK,
			mockBody: `{
				"section": {
					"id": 123,
					"company_id": 1,
					"name": "営業部",
					"available": true
				}
			}`,
			wantErr:       false,
			wantSectionID: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/sections/123"
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

			sectionsService := accountingClient.Sections()
			section, err := sectionsService.Get(context.Background(), tt.companyID, tt.sectionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if section == nil {
					t.Error("Get() returned nil section")
					return
				}
				if section.Section.Id != tt.wantSectionID {
					t.Errorf("Get() got section ID %d, want %d", section.Section.Id, tt.wantSectionID)
				}
			}
		})
	}
}

func TestSectionsService_Create(t *testing.T) {
	mockStatus := http.StatusCreated
	mockBody := `{
		"section": {
			"id": 999,
			"company_id": 1,
			"name": "新規部門",
			"available": true
		}
	}`
	wantSectionID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/sections" {
			t.Errorf("expected path /api/1/sections, got %s", r.URL.Path)
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

	params := gen.SectionParams{
		CompanyId: 1,
		Name:      "新規部門",
	}

	sectionsService := accountingClient.Sections()
	section, err := sectionsService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}

	if section == nil {
		t.Error("Create() returned nil section")
		return
	}
	if section.Section.Id != wantSectionID {
		t.Errorf("Create() got section ID %d, want %d", section.Section.Id, wantSectionID)
	}
}

func TestSectionsService_Update(t *testing.T) {
	sectionID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"section": {
			"id": 123,
			"company_id": 1,
			"name": "更新後部門",
			"available": true
		}
	}`
	wantSectionID := int64(123)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/1/sections/123"
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

	params := gen.SectionParams{
		CompanyId: 1,
		Name:      "更新後部門",
	}

	sectionsService := accountingClient.Sections()
	section, err := sectionsService.Update(context.Background(), sectionID, params)

	if err != nil {
		t.Errorf("Update() unexpected error = %v", err)
		return
	}

	if section == nil {
		t.Error("Update() returned nil section")
		return
	}
	if section.Section.Id != wantSectionID {
		t.Errorf("Update() got section ID %d, want %d", section.Section.Id, wantSectionID)
	}
}

func TestSectionsService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		sectionID  int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			companyID:  1,
			sectionID:  123,
			mockStatus: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "delete with 200 OK",
			companyID:  1,
			sectionID:  456,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "delete not found",
			companyID:  1,
			sectionID:  999,
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

			sectionsService := accountingClient.Sections()
			err = sectionsService.Delete(context.Background(), tt.companyID, tt.sectionID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
