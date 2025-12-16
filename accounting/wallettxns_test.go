package accounting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/u-masato/freee-api-go/client"
	"github.com/u-masato/freee-api-go/internal/gen"
)

func TestWalletTxnService_List(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListWalletTxnsOptions
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
				"wallet_txns": [
					{"id": 1, "company_id": 1, "amount": 10000, "date": "2024-01-15", "description": "Payment 1", "entry_side": "income"},
					{"id": 2, "company_id": 1, "amount": 20000, "date": "2024-01-16", "description": "Payment 2", "entry_side": "expense"}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:      "successful list with options",
			companyID: 1,
			opts: &ListWalletTxnsOptions{
				WalletableType: stringPtr("bank_account"),
				WalletableId:   int64Ptr(12345),
				EntrySide:      stringPtr("income"),
				Limit:          int64Ptr(10),
				Offset:         int64Ptr(0),
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"wallet_txns": [
					{"id": 3, "company_id": 1, "amount": 5000, "date": "2024-01-17", "description": "Income", "entry_side": "income"}
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
				"wallet_txns": []
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
				if r.URL.Path != "/api/1/wallet_txns" {
					t.Errorf("expected path /api/1/wallet_txns, got %s", r.URL.Path)
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

			walletTxnService := accountingClient.WalletTxns()
			result, err := walletTxnService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("List() returned nil result")
					return
				}
				if len(result.WalletTxns) != tt.wantCount {
					t.Errorf("List() got %d wallet txns, want %d", len(result.WalletTxns), tt.wantCount)
				}
			}
		})
	}
}

func TestWalletTxnService_Get(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		txnID      int64
		mockStatus int
		mockBody   string
		wantErr    bool
		wantTxnID  int64
	}{
		{
			name:       "successful get",
			companyID:  1,
			txnID:      123,
			mockStatus: http.StatusOK,
			mockBody: `{
				"wallet_txn": {
					"id": 123,
					"company_id": 1,
					"amount": 10000,
					"date": "2024-01-15",
					"description": "Payment",
					"entry_side": "income",
					"walletable_id": 12345,
					"walletable_type": "bank_account"
				}
			}`,
			wantErr:   false,
			wantTxnID: 123,
		},
		{
			name:       "successful get with different txn",
			companyID:  1,
			txnID:      456,
			mockStatus: http.StatusOK,
			mockBody: `{
				"wallet_txn": {
					"id": 456,
					"company_id": 1,
					"amount": 20000,
					"date": "2024-01-16",
					"description": "Another payment",
					"entry_side": "expense",
					"walletable_id": 67890,
					"walletable_type": "credit_card"
				}
			}`,
			wantErr:   false,
			wantTxnID: 456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				expectedPath := "/api/1/wallet_txns/123"
				if tt.txnID == 456 {
					expectedPath = "/api/1/wallet_txns/456"
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

			walletTxnService := accountingClient.WalletTxns()
			txn, err := walletTxnService.Get(context.Background(), tt.companyID, tt.txnID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if txn == nil {
					t.Error("Get() returned nil txn")
					return
				}
				if txn.WalletTxn.Id != tt.wantTxnID {
					t.Errorf("Get() got txn ID %d, want %d", txn.WalletTxn.Id, tt.wantTxnID)
				}
			}
		})
	}
}

func TestWalletTxnService_Create(t *testing.T) {
	mockStatus := http.StatusCreated
	mockBody := `{
		"wallet_txn": {
			"id": 999,
			"company_id": 1,
			"amount": 10000,
			"date": "2024-01-15",
			"description": "New transaction",
			"entry_side": "income",
			"walletable_id": 12345,
			"walletable_type": "bank_account"
		}
	}`
	wantTxnID := int64(999)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/wallet_txns" {
			t.Errorf("expected path /api/1/wallet_txns, got %s", r.URL.Path)
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

	params := gen.WalletTxnParams{
		CompanyId:      1,
		WalletableId:   12345,
		WalletableType: "bank_account",
		Date:           "2024-01-15",
		Amount:         10000,
		EntrySide:      "income",
	}

	walletTxnService := accountingClient.WalletTxns()
	txn, err := walletTxnService.Create(context.Background(), params)

	if err != nil {
		t.Errorf("Create() unexpected error = %v", err)
		return
	}

	if txn == nil {
		t.Error("Create() returned nil txn")
		return
	}
	if txn.WalletTxn.Id != wantTxnID {
		t.Errorf("Create() got txn ID %d, want %d", txn.WalletTxn.Id, wantTxnID)
	}
}

func TestWalletTxnService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		txnID      int64
		mockStatus int
		wantErr    bool
	}{
		{
			name:       "successful delete",
			companyID:  1,
			txnID:      123,
			mockStatus: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "delete with 200 OK",
			companyID:  1,
			txnID:      456,
			mockStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "delete not found",
			companyID:  1,
			txnID:      999,
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

			walletTxnService := accountingClient.WalletTxns()
			err = walletTxnService.Delete(context.Background(), tt.companyID, tt.txnID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWalletTxnService_List_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		opts       *ListWalletTxnsOptions
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

			walletTxnService := accountingClient.WalletTxns()
			_, err = walletTxnService.List(context.Background(), tt.companyID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWalletTxnService_Get_ErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		companyID  int64
		txnID      int64
		mockStatus int
		mockBody   string
		wantErr    bool
	}{
		{
			name:       "not found",
			companyID:  1,
			txnID:      999,
			mockStatus: http.StatusNotFound,
			mockBody:   `{"errors": [{"messages": ["Resource not found"]}]}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			companyID:  1,
			txnID:      123,
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

			walletTxnService := accountingClient.WalletTxns()
			_, err = walletTxnService.Get(context.Background(), tt.companyID, tt.txnID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWalletTxnService_Create_ErrorCases(t *testing.T) {
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

			params := gen.WalletTxnParams{
				CompanyId:      1,
				WalletableId:   12345,
				WalletableType: "bank_account",
				Date:           "2024-01-15",
				Amount:         10000,
				EntrySide:      "income",
			}

			walletTxnService := accountingClient.WalletTxns()
			_, err = walletTxnService.Create(context.Background(), params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWalletTxnService_ListIter_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		companyID int64
		opts      *ListWalletTxnsOptions
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
					w.Write([]byte(`{"wallet_txns": []}`))
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

			walletTxnService := accountingClient.WalletTxns()
			iter := walletTxnService.ListIter(context.Background(), tt.companyID, tt.opts)

			var txns []gen.WalletTxn
			for iter.Next() {
				txns = append(txns, iter.Value())
			}

			err = iter.Err()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(txns) != tt.wantCount {
				t.Errorf("ListIter() got %d transactions, want %d", len(txns), tt.wantCount)
			}
		})
	}
}

func TestWalletTxnService_ListIter(t *testing.T) {
	tests := []struct {
		name        string
		companyID   int64
		opts        *ListWalletTxnsOptions
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
					"wallet_txns": [
						{"id": 1, "company_id": 1, "amount": 10000, "date": "2024-01-15"},
						{"id": 2, "company_id": 1, "amount": 20000, "date": "2024-01-16"}
					]
				}`,
				`{"wallet_txns": []}`,
			},
			wantErr:     false,
			wantCount:   2,
			wantFetches: 2,
		},
		{
			name:      "multiple page iteration",
			companyID: 1,
			opts: &ListWalletTxnsOptions{
				Limit: int64Ptr(2),
			},
			mockPages: []string{
				`{
					"wallet_txns": [
						{"id": 1, "company_id": 1, "amount": 10000, "date": "2024-01-15"},
						{"id": 2, "company_id": 1, "amount": 20000, "date": "2024-01-16"}
					]
				}`,
				`{
					"wallet_txns": [
						{"id": 3, "company_id": 1, "amount": 30000, "date": "2024-01-17"}
					]
				}`,
				`{"wallet_txns": []}`,
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
				`{"wallet_txns": []}`,
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
					w.Write([]byte(`{"wallet_txns": []}`))
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

			walletTxnService := accountingClient.WalletTxns()

			iter := walletTxnService.ListIter(context.Background(), tt.companyID, tt.opts)

			var txns []gen.WalletTxn
			for iter.Next() {
				txns = append(txns, iter.Value())
			}

			if err := iter.Err(); (err != nil) != tt.wantErr {
				t.Errorf("ListIter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(txns) != tt.wantCount {
				t.Errorf("ListIter() got %d transactions, want %d", len(txns), tt.wantCount)
			}

			if fetchCount != tt.wantFetches {
				t.Errorf("ListIter() made %d fetches, want %d", fetchCount, tt.wantFetches)
			}
		})
	}
}
