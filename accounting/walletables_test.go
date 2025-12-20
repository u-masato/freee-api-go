package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/u-masato/freee-api-go/client"
	"github.com/u-masato/freee-api-go/internal/gen"
)

func TestWalletablesService_List(t *testing.T) {
	mockStatus := http.StatusOK
	mockBody := `{
		"walletables": [
			{"id": 1, "name": "Main Wallet", "type": "wallet"},
			{"id": 2, "name": "Bank Account", "type": "bank_account"}
		],
		"meta": {
			"up_to_date": true
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/walletables" {
			t.Errorf("expected path /api/1/walletables, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("company_id") != "1" {
			t.Errorf("expected company_id=1, got %s", query.Get("company_id"))
		}
		if query.Get("type") != "wallet" {
			t.Errorf("expected type=wallet, got %s", query.Get("type"))
		}
		if query.Get("with_balance") != "true" {
			t.Errorf("expected with_balance=true, got %s", query.Get("with_balance"))
		}
		if query.Get("with_sync_status") != "true" {
			t.Errorf("expected with_sync_status=true, got %s", query.Get("with_sync_status"))
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

	opts := &ListWalletablesOptions{
		Type:           stringPtr("wallet"),
		WithBalance:    boolPtr(true),
		WithSyncStatus: boolPtr(true),
	}

	walletablesService := accountingClient.Walletables()
	result, err := walletablesService.List(context.Background(), 1, opts)

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
	if result.UpToDate == nil || !*result.UpToDate {
		t.Error("List() expected up_to_date=true")
	}
}

func TestWalletablesService_Get(t *testing.T) {
	companyID := int64(1)
	walletableID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"walletable": {
			"id": 123,
			"name": "Main Wallet",
			"type": "wallet"
		},
		"meta": {
			"up_to_date": true
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		expectedPath := "/api/1/walletables/wallet/123"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("company_id") != "1" {
			t.Errorf("expected company_id=1, got %s", query.Get("company_id"))
		}
		if query.Get("with_last_synced_at") != "true" {
			t.Errorf("expected with_last_synced_at=true, got %s", query.Get("with_last_synced_at"))
		}
		if query.Get("with_sync_status") != "true" {
			t.Errorf("expected with_sync_status=true, got %s", query.Get("with_sync_status"))
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

	opts := &GetWalletableOptions{
		WithLastSyncedAt: boolPtr(true),
		WithSyncStatus:   boolPtr(true),
	}

	walletablesService := accountingClient.Walletables()
	result, err := walletablesService.Get(context.Background(), companyID, "wallet", walletableID, opts)

	if err != nil {
		t.Errorf("Get() unexpected error = %v", err)
		return
	}
	if result == nil {
		t.Error("Get() returned nil result")
		return
	}
	if result.Walletable.Id != walletableID {
		t.Errorf("Get() got walletable ID %d, want %d", result.Walletable.Id, walletableID)
	}
	if result.UpToDate == nil || !*result.UpToDate {
		t.Error("Get() expected up_to_date=true")
	}
}

func TestWalletablesService_Create(t *testing.T) {
	mockStatus := http.StatusCreated
	mockBody := `{
		"walletable": {
			"id": 999,
			"bank_id": 1,
			"name": "New Wallet",
			"type": "wallet"
		}
	}`
	wantID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/walletables" {
			t.Errorf("expected path /api/1/walletables, got %s", r.URL.Path)
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

	params := gen.WalletableCreateParams{
		CompanyId: 1,
		Name:      "New Wallet",
		Type:      "wallet",
	}

	walletablesService := accountingClient.Walletables()
	walletable, err := walletablesService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}
	if walletable == nil {
		t.Error("Create() returned nil walletable")
		return
	}
	if walletable.Walletable.Id != wantID {
		t.Errorf("Create() got walletable ID %d, want %d", walletable.Walletable.Id, wantID)
	}
}

func TestWalletablesService_Update(t *testing.T) {
	walletableID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"walletable": {
			"id": 123,
			"name": "Updated Wallet",
			"type": "wallet"
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/1/walletables/wallet/123"
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

	params := gen.WalletableUpdateParams{
		CompanyId: 1,
		Name:      "Updated Wallet",
	}

	walletablesService := accountingClient.Walletables()
	walletable, err := walletablesService.Update(context.Background(), "wallet", walletableID, params)

	if err != nil {
		t.Errorf("Update() unexpected error = %v", err)
		return
	}
	if walletable == nil {
		t.Error("Update() returned nil walletable")
		return
	}
	if walletable.Id != walletableID {
		t.Errorf("Update() got walletable ID %d, want %d", walletable.Id, walletableID)
	}
}

func TestWalletablesService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		companyID    int64
		walletableID int64
		mockStatus   int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			companyID:    1,
			walletableID: 123,
			mockStatus:   http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "delete with 200 OK",
			companyID:    1,
			walletableID: 456,
			mockStatus:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "delete not found",
			companyID:    1,
			walletableID: 999,
			mockStatus:   http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE request, got %s", r.Method)
				}
				expectedPath := "/api/1/walletables/wallet/" + strconv.FormatInt(tt.walletableID, 10)
				if r.URL.Path != expectedPath {
					t.Errorf("expected path /api/1/walletables/wallet/%d, got %s", tt.walletableID, r.URL.Path)
				}

				query := r.URL.Query()
				if query.Get("company_id") != "1" {
					t.Errorf("expected company_id=1, got %s", query.Get("company_id"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
			}))
			defer server.Close()

			baseClient := client.NewClient(client.WithBaseURL(server.URL))
			accountingClient, err := NewClient(baseClient)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			walletablesService := accountingClient.Walletables()
			err = walletablesService.Delete(context.Background(), tt.companyID, "wallet", tt.walletableID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
