package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/u-masato/freee-api-go/client"
	"github.com/u-masato/freee-api-go/internal/gen"
)

func TestTagsService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListTagsOptions
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
				"tags": [
					{"id": 1, "company_id": 1, "name": "重要", "update_date": "2024-01-15"},
					{"id": 2, "company_id": 1, "name": "経費", "update_date": "2024-01-16"}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with pagination",
			companyID: 1,
			opts: &ListTagsOptions{
				Limit:  int64Ptr(100),
				Offset: int64Ptr(0),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"tags": [
					{"id": 3, "company_id": 1, "name": "売上", "update_date": "2024-01-17"}
				]
			}`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "successful list with date range filter",
			companyID: 1,
			opts: &ListTagsOptions{
				StartUpdateDate: stringPtr("2024-01-01"),
				EndUpdateDate:   stringPtr("2024-01-31"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"tags": [
					{"id": 4, "company_id": 1, "name": "仕入", "update_date": "2024-01-10"}
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
				"tags": []
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
				if r.URL.Path != "/api/1/tags" {
					t.Errorf("expected path /api/1/tags, got %s", r.URL.Path)
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

			tagsService := accountingClient.Tags()
			result, err := tagsService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.Tags) != tt.wantCount {
					t.Errorf("List() got %d tags, want %d", len(result.Tags), tt.wantCount)
				}
				if result.Count != tt.wantCount {
					t.Errorf("List() got count %d, want %d", result.Count, tt.wantCount)
				}
			}
		})
	}
}

func TestTagsService_Get(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		tagID      int64
		mockStatus int
		mockBody   string
		wantErr    bool
		wantTagID  int64
	}{
		{
			name:       "successful get",
			companyID:  1,
			tagID:      123,
			mockStatus: http.StatusOK,
			mockBody: `{
				"tag": {
					"id": 123,
					"company_id": 1,
					"name": "重要タグ",
					"update_date": "2024-01-15"
				}
			}`,
			wantErr:   false,
			wantTagID: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/tags/123"
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

			tagsService := accountingClient.Tags()
			tag, err := tagsService.Get(context.Background(), tt.companyID, tt.tagID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tag == nil {
					t.Error("Get() returned nil tag")
					return
				}
				if tag.Tag.Id != tt.wantTagID {
					t.Errorf("Get() got tag ID %d, want %d", tag.Tag.Id, tt.wantTagID)
				}
			}
		})
	}
}

func TestTagsService_Create(t *testing.T) {
	mockStatus := http.StatusCreated
	mockBody := `{
		"tag": {
			"id": 999,
			"company_id": 1,
			"name": "新規タグ",
			"update_date": "2024-01-25"
		}
	}`
	wantTagID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/tags" {
			t.Errorf("expected path /api/1/tags, got %s", r.URL.Path)
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

	params := gen.TagParams{
		CompanyId: 1,
		Name:      "新規タグ",
	}

	tagsService := accountingClient.Tags()
	tag, err := tagsService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}

	if tag == nil {
		t.Error("Create() returned nil tag")
		return
	}
	if tag.Tag.Id != wantTagID {
		t.Errorf("Create() got tag ID %d, want %d", tag.Tag.Id, wantTagID)
	}
}

func TestTagsService_Update(t *testing.T) {
	tagID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"tag": {
			"id": 123,
			"company_id": 1,
			"name": "更新後タグ",
			"update_date": "2024-01-26"
		}
	}`
	wantTagID := int64(123)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/1/tags/123"
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

	params := gen.TagParams{
		CompanyId: 1,
		Name:      "更新後タグ",
	}

	tagsService := accountingClient.Tags()
	tag, err := tagsService.Update(context.Background(), tagID, params)

	if err != nil {
		t.Errorf("Update() unexpected error = %v", err)
		return
	}

	if tag == nil {
		t.Error("Update() returned nil tag")
		return
	}
	if tag.Tag.Id != wantTagID {
		t.Errorf("Update() got tag ID %d, want %d", tag.Tag.Id, wantTagID)
	}
}

func TestTagsService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		tagID      int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			companyID:  1,
			tagID:      123,
			mockStatus: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "delete with 200 OK",
			companyID:  1,
			tagID:      456,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "delete not found",
			companyID:  1,
			tagID:      999,
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

			tagsService := accountingClient.Tags()
			err = tagsService.Delete(context.Background(), tt.companyID, tt.tagID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagsService_ListIter(t *testing.T) {
	tests := []struct {
		name        string
		companyID   int64
		opts        *ListTagsOptions
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
					"tags": [
						{"id": 1, "company_id": 1, "name": "タグ1", "update_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "name": "タグ2", "update_date": "2024-01-16"}
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
			opts: &ListTagsOptions{
				Limit: int64Ptr(2),
			},
			mockPages: []string{
				`{
					"tags": [
						{"id": 1, "company_id": 1, "name": "タグ1", "update_date": "2024-01-15"},
						{"id": 2, "company_id": 1, "name": "タグ2", "update_date": "2024-01-16"}
					]
				}`,
				`{
					"tags": [
						{"id": 3, "company_id": 1, "name": "タグ3", "update_date": "2024-01-17"}
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
					"tags": []
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
					w.Write([]byte(`{"tags": []}`))
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

			tagsService := accountingClient.Tags()
			iter := tagsService.ListIter(context.Background(), tt.companyID, tt.opts)

			var tags []gen.Tag
			for iter.Next() {
				tags = append(tags, iter.Value())
			}

			if err := iter.Err(); (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tags) != tt.wantCount {
				t.Errorf("ListIter() got %d tags, want %d", len(tags), tt.wantCount)
			}

			if fetchCount != tt.wantFetches {
				t.Errorf("ListIter() made %d fetches, want %d", fetchCount, tt.wantFetches)
			}
		})
	}
}
