package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/internal/gen"
)

func TestTransfersService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListTransfersOptions
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
				"transfers": [
					{
						"id": 1,
						"company_id": 1,
						"amount": 10000,
						"date": "2024-01-15",
						"from_walletable_id": 100,
						"from_walletable_type": "bank_account",
						"to_walletable_id": 200,
						"to_walletable_type": "wallet"
					},
					{
						"id": 2,
						"company_id": 1,
						"amount": 20000,
						"date": "2024-01-16",
						"from_walletable_id": 100,
						"from_walletable_type": "bank_account",
						"to_walletable_id": 300,
						"to_walletable_type": "credit_card"
					}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with options",
			companyID: 1,
			opts: &ListTransfersOptions{
				StartDate: stringPtr("2024-01-01"),
				EndDate:   stringPtr("2024-01-31"),
				Limit:     int64Ptr(10),
				Offset:    int64Ptr(0),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"transfers": [
					{
						"id": 3,
						"company_id": 1,
						"amount": 5000,
						"date": "2024-01-17",
						"from_walletable_id": 100,
						"from_walletable_type": "bank_account",
						"to_walletable_id": 200,
						"to_walletable_type": "wallet"
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
				"transfers": []
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
				if r.URL.Path != "/api/1/transfers" {
					t.Errorf("expected path /api/1/transfers, got %s", r.URL.Path)
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

			transfersService := accountingClient.Transfers()
			result, err := transfersService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.Transfers) != tt.wantCount {
					t.Errorf("List() got %d transfers, want %d", len(result.Transfers), tt.wantCount)
				}
			}
		})
	}
}

func TestTransfersService_Get(t *testing.T) {
	tests := []struct {
		name          string
		companyID     int64
		transferID    int64
		mockStatus    int
		mockBody      string
		wantErr       bool
		wantTransferID int64
	}{
		{
			name:       "successful get",
			companyID:  1,
			transferID: 123,
			mockStatus: http.StatusOK,
			mockBody: `{
				"transfer": {
					"id": 123,
					"company_id": 1,
					"amount": 10000,
					"date": "2024-01-15",
					"from_walletable_id": 100,
					"from_walletable_type": "bank_account",
					"to_walletable_id": 200,
					"to_walletable_type": "wallet"
				}
			}`,
			wantErr:        false,
			wantTransferID: 123,
		},
		{
			name:       "successful get another transfer",
			companyID:  1,
			transferID: 456,
			mockStatus: http.StatusOK,
			mockBody: `{
				"transfer": {
					"id": 456,
					"company_id": 1,
					"amount": 20000,
					"date": "2024-01-16",
					"from_walletable_id": 100,
					"from_walletable_type": "bank_account",
					"to_walletable_id": 300,
					"to_walletable_type": "credit_card",
					"description": "Transfer to credit card"
				}
			}`,
			wantErr:        false,
			wantTransferID: 456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/transfers/123"
				if tt.transferID == 456 {
					expectedPath = "/api/1/transfers/456"
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

			transfersService := accountingClient.Transfers()
			transfer, err := transfersService.Get(context.Background(), tt.companyID, tt.transferID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if transfer == nil {
					t.Error("Get() returned nil transfer")
					return
				}
				if transfer.Transfer.Id != tt.wantTransferID {
					t.Errorf("Get() got transfer ID %d, want %d", transfer.Transfer.Id, tt.wantTransferID)
				}
			}
		})
	}
}

func TestTransfersService_Create(t *testing.T) {
	mockStatus := http.StatusCreated
	mockBody := `{
		"transfer": {
			"id": 999,
			"company_id": 1,
			"amount": 10000,
			"date": "2024-01-15",
			"from_walletable_id": 100,
			"from_walletable_type": "bank_account",
			"to_walletable_id": 200,
			"to_walletable_type": "wallet"
		}
	}`
	wantTransferID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/transfers" {
			t.Errorf("expected path /api/1/transfers, got %s", r.URL.Path)
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

	params := gen.TransferParams{
		CompanyId:          1,
		Date:               "2024-01-15",
		Amount:             10000,
		FromWalletableId:   100,
		FromWalletableType: "bank_account",
		ToWalletableId:     200,
		ToWalletableType:   "wallet",
	}

	transfersService := accountingClient.Transfers()
	transfer, err := transfersService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}

	if transfer == nil {
		t.Error("Create() returned nil transfer")
		return
	}
	if transfer.Transfer.Id != wantTransferID {
		t.Errorf("Create() got transfer ID %d, want %d", transfer.Transfer.Id, wantTransferID)
	}
}

func TestTransfersService_Update(t *testing.T) {
	transferID := int64(123)
	mockStatus := http.StatusOK
	mockBody := `{
		"transfer": {
			"id": 123,
			"company_id": 1,
			"amount": 15000,
			"date": "2024-01-20",
			"from_walletable_id": 100,
			"from_walletable_type": "bank_account",
			"to_walletable_id": 200,
			"to_walletable_type": "wallet",
			"description": "Updated transfer"
		}
	}`
	wantTransferID := int64(123)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
		}
		expectedPath := "/api/1/transfers/123"
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

	desc := "Updated transfer"
	params := gen.TransferParams{
		CompanyId:          1,
		Date:               "2024-01-20",
		Amount:             15000,
		FromWalletableId:   100,
		FromWalletableType: "bank_account",
		ToWalletableId:     200,
		ToWalletableType:   "wallet",
		Description:        &desc,
	}

	transfersService := accountingClient.Transfers()
	transfer, err := transfersService.Update(context.Background(), transferID, params)

	if err != nil {
		t.Errorf("Update() unexpected error = %v", err)
		return
	}

	if transfer == nil {
		t.Error("Update() returned nil transfer")
		return
	}
	if transfer.Transfer.Id != wantTransferID {
		t.Errorf("Update() got transfer ID %d, want %d", transfer.Transfer.Id, wantTransferID)
	}
}

func TestTransfersService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		transferID int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			companyID:  1,
			transferID: 123,
			mockStatus: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "delete with 200 OK",
			companyID:  1,
			transferID: 456,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "delete not found",
			companyID:  1,
			transferID: 999,
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

			transfersService := accountingClient.Transfers()
			err = transfersService.Delete(context.Background(), tt.companyID, tt.transferID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransfersService_List_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListTransfersOptions
		mockStatus int
		mockBody   string
		wantErr    bool
	}{
		{
			name:       "server error",
			companyID:  1,
			opts:       nil,
			mockStatus: http.StatusInternalServerError,
			mockBody:   `{"errors": [{"messages": ["Internal server error"]}]}`,
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			companyID:  1,
			opts:       nil,
			mockStatus: http.StatusUnauthorized,
			mockBody:   `{"errors": [{"messages": ["Invalid access token"]}]}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			transfersService := accountingClient.Transfers()
			_, err = transfersService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransfersService_Get_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		transferID int64
		mockStatus int
		mockBody   string
		wantErr    bool
	}{
		{
			name:       "not found",
			companyID:  1,
			transferID: 999,
			mockStatus: http.StatusNotFound,
			mockBody:   `{"errors": [{"messages": ["Resource not found"]}]}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			companyID:  1,
			transferID: 123,
			mockStatus: http.StatusInternalServerError,
			mockBody:   `{"errors": [{"messages": ["Internal server error"]}]}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			transfersService := accountingClient.Transfers()
			_, err = transfersService.Get(context.Background(), tt.companyID, tt.transferID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransfersService_Create_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   string
		wantErr    bool
	}{
		{
			name:       "bad request",
			mockStatus: http.StatusBadRequest,
			mockBody:   `{"errors": [{"messages": ["date is required"]}]}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			mockStatus: http.StatusInternalServerError,
			mockBody:   `{"errors": [{"messages": ["Internal server error"]}]}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			params := gen.TransferParams{
				CompanyId:          1,
				Date:               "2024-01-15",
				Amount:             10000,
				FromWalletableId:   100,
				FromWalletableType: "bank_account",
				ToWalletableId:     200,
				ToWalletableType:   "wallet",
			}

			transfersService := accountingClient.Transfers()
			_, err = transfersService.Create(context.Background(), params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransfersService_Update_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		transferID int64
		mockStatus int
		mockBody   string
		wantErr    bool
	}{
		{
			name:       "not found",
			transferID: 999,
			mockStatus: http.StatusNotFound,
			mockBody:   `{"errors": [{"messages": ["Resource not found"]}]}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			transferID: 123,
			mockStatus: http.StatusInternalServerError,
			mockBody:   `{"errors": [{"messages": ["Internal server error"]}]}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			params := gen.TransferParams{
				CompanyId:          1,
				Date:               "2024-01-20",
				Amount:             15000,
				FromWalletableId:   100,
				FromWalletableType: "bank_account",
				ToWalletableId:     200,
				ToWalletableType:   "wallet",
			}

			transfersService := accountingClient.Transfers()
			_, err = transfersService.Update(context.Background(), tt.transferID, params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransfersService_ListIter_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		companyID int64
		opts      *ListTransfersOptions
		mockPages []struct {
			status int
			body   string
		}
		wantErr   bool
		wantCount int
	}{
		{
			name:      "error on first fetch",
			companyID: 1,
			opts:      nil,
			mockPages: []struct {
				status int
				body   string
			}{
				{http.StatusInternalServerError, `{"errors": [{"messages": ["Server error"]}]}`},
			},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetchCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if fetchCount < len(tt.mockPages) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockPages[fetchCount].status)
					w.Write([]byte(tt.mockPages[fetchCount].body))
					fetchCount++
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"transfers": []}`))
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

			transfersService := accountingClient.Transfers()
			iter := transfersService.ListIter(context.Background(), tt.companyID, tt.opts)

			var transfers []gen.Transfer
			for iter.Next() {
				transfers = append(transfers, iter.Value())
			}

			err = iter.Err()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(transfers) != tt.wantCount {
				t.Errorf("ListIter() got %d transfers, want %d", len(transfers), tt.wantCount)
			}
		})
	}
}

func TestTransfersService_ListIter(t *testing.T) {
	tests := []struct {
		name        string
		companyID   int64
		opts        *ListTransfersOptions
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
					"transfers": [
						{"id": 1, "company_id": 1, "amount": 10000, "date": "2024-01-15"},
						{"id": 2, "company_id": 1, "amount": 20000, "date": "2024-01-16"}
					]
				}`,
				`{"transfers": []}`,
			},
			wantErr:     false,
			wantCount:   2,
			wantFetches: 2,
		},
		{
			name:      "multiple page iteration",
			companyID: 1,
			opts: &ListTransfersOptions{
				Limit: int64Ptr(2),
			},
			mockPages: []string{
				`{
					"transfers": [
						{"id": 1, "company_id": 1, "amount": 10000, "date": "2024-01-15"},
						{"id": 2, "company_id": 1, "amount": 20000, "date": "2024-01-16"}
					]
				}`,
				`{
					"transfers": [
						{"id": 3, "company_id": 1, "amount": 30000, "date": "2024-01-17"}
					]
				}`,
				`{"transfers": []}`,
			},
			wantErr:     false,
			wantCount:   3,
			wantFetches: 3,
		},
		{
			name:      "empty result",
			companyID: 1,
			opts:      nil,
			mockPages: []string{
				`{"transfers": []}`,
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
					w.Write([]byte(`{"transfers": []}`))
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

			transfersService := accountingClient.Transfers()

			iter := transfersService.ListIter(context.Background(), tt.companyID, tt.opts)

			var transfers []gen.Transfer
			for iter.Next() {
				transfers = append(transfers, iter.Value())
			}

			if err := iter.Err(); (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(transfers) != tt.wantCount {
				t.Errorf("ListIter() got %d transfers, want %d", len(transfers), tt.wantCount)
			}

			if fetchCount != tt.wantFetches {
				t.Errorf("ListIter() made %d fetches, want %d", fetchCount, tt.wantFetches)
			}
		})
	}
}
