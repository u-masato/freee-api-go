package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muno/freee-api-go/client"
)

func TestJournalsService_Download(t *testing.T) {
	tests := []struct {
		name         string
		companyID    int64
		downloadType string
		opts         *DownloadJournalsOptions
		mockStatus   int
		mockBody     string
		wantErr      bool
		wantID       int64
	}{
		{
			name:         "successful download with no options",
			companyID:    1,
			downloadType: "csv",
			opts:         nil,
			mockStatus:   http.StatusAccepted,
			mockBody: `{
				"journals": {
					"id": 123,
					"company_id": 1,
					"download_type": "csv",
					"start_date": "2024-01-01",
					"end_date": "2024-01-31"
				}
			}`,
			wantErr: false,
			wantID:  123,
		},
		{
			name:         "successful download with options",
			companyID:    1,
			downloadType: "csv",
			opts: &DownloadJournalsOptions{
				StartDate: stringPtr("2024-01-01"),
				EndDate:   stringPtr("2024-01-31"),
				Encoding:  stringPtr("utf-8"),
			},
			mockStatus: http.StatusAccepted,
			mockBody: `{
				"journals": {
					"id": 456,
					"company_id": 1,
					"download_type": "csv",
					"start_date": "2024-01-01",
					"end_date": "2024-01-31",
					"encoding": "utf-8"
				}
			}`,
			wantErr: false,
			wantID:  456,
		},
		{
			name:         "download with visible tags and IDs",
			companyID:    1,
			downloadType: "generic",
			opts: &DownloadJournalsOptions{
				StartDate:   stringPtr("2024-01-01"),
				EndDate:     stringPtr("2024-01-31"),
				VisibleTags: &[]string{"partner", "item", "tag"},
				VisibleIds:  &[]string{"deal_id", "manual_journal_id"},
			},
			mockStatus: http.StatusAccepted,
			mockBody: `{
				"journals": {
					"id": 789,
					"company_id": 1,
					"download_type": "generic",
					"start_date": "2024-01-01",
					"end_date": "2024-01-31",
					"visible_tags": ["partner", "item", "tag"],
					"visible_ids": ["deal_id", "manual_journal_id"]
				}
			}`,
			wantErr: false,
			wantID:  789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/1/journals" {
					t.Errorf("expected path /api/1/journals, got %s", r.URL.Path)
				}

				// Verify query parameters
				query := r.URL.Query()
				if query.Get("company_id") != "1" {
					t.Errorf("expected company_id=1, got %s", query.Get("company_id"))
				}
				if query.Get("download_type") != tt.downloadType {
					t.Errorf("expected download_type=%s, got %s", tt.downloadType, query.Get("download_type"))
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

			journalsService := accountingClient.Journals()
			result, err := journalsService.Download(context.Background(), tt.companyID, tt.downloadType, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("Download() returned nil result")
					return
				}
				if result.Journals.Journals.Id != tt.wantID {
					t.Errorf("Download() got ID %d, want %d", result.Journals.Journals.Id, tt.wantID)
				}
			}
		})
	}
}

func TestJournalsService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListManualJournalsOptions
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
				"manual_journals": [
					{
						"id": 1,
						"company_id": 1,
						"issue_date": "2024-01-15",
						"txn_number": "001"
					},
					{
						"id": 2,
						"company_id": 1,
						"issue_date": "2024-01-16",
						"txn_number": "002"
					}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with date filter",
			companyID: 1,
			opts: &ListManualJournalsOptions{
				StartIssueDate: stringPtr("2024-01-01"),
				EndIssueDate:   stringPtr("2024-01-31"),
				Limit:          int64Ptr(50),
				Offset:         int64Ptr(0),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"manual_journals": [
					{
						"id": 3,
						"company_id": 1,
						"issue_date": "2024-01-20",
						"txn_number": "003"
					}
				]
			}`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:      "successful list with amount filter",
			companyID: 1,
			opts: &ListManualJournalsOptions{
				MinAmount: int64Ptr(1000),
				MaxAmount: int64Ptr(10000),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"manual_journals": [
					{
						"id": 4,
						"company_id": 1,
						"issue_date": "2024-01-25",
						"txn_number": "004"
					}
				]
				
			}`,
			wantErr:   false,
			wantCount: 1,
			
		},
		{
			name:      "successful list with account and partner filters",
			companyID: 1,
			opts: &ListManualJournalsOptions{
				AccountItemId: int64Ptr(12345),
				PartnerId:     int64Ptr(67890),
				PartnerCode:   stringPtr("PARTNER001"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"manual_journals": [
					{
						"id": 5,
						"company_id": 1,
						"issue_date": "2024-01-28",
						"txn_number": "005"
					}
				]
				
			}`,
			wantErr:   false,
			wantCount: 1,
			
		},
		{
			name:      "successful list with entry side filter",
			companyID: 1,
			opts: &ListManualJournalsOptions{
				EntrySide: stringPtr("debit"),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"manual_journals": [
					{
						"id": 6,
						"company_id": 1,
						"issue_date": "2024-01-30",
						"txn_number": "006"
					}
				]
				
			}`,
			wantErr:   false,
			wantCount: 1,
			
		},
		{
			name:      "successful list with segment filters",
			companyID: 1,
			opts: &ListManualJournalsOptions{
				Segment1TagId: int64Ptr(111),
				Segment2TagId: int64Ptr(222),
				Segment3TagId: int64Ptr(333),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"manual_journals": [
					{
						"id": 7,
						"company_id": 1,
						"issue_date": "2024-01-31",
						"txn_number": "007"
					}
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
				"manual_journals": []
				
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
				if r.URL.Path != "/api/1/manual_journals" {
					t.Errorf("expected path /api/1/manual_journals, got %s", r.URL.Path)
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

			journalsService := accountingClient.Journals()
			result, err := journalsService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.ManualJournals) != tt.wantCount {
					t.Errorf("List() got %d manual journals, want %d", len(result.ManualJournals), tt.wantCount)
				}
			}
		})
	}
}
