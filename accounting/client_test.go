package accounting

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/muno/freee-api-go/client"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *client.Client
		wantErr bool
	}{
		{
			name: "valid client with defaults",
			setup: func() *client.Client {
				return client.NewClient()
			},
			wantErr: false,
		},
		{
			name: "valid client with custom HTTP client",
			setup: func() *client.Client {
				return client.NewClient(
					client.WithHTTPClient(&http.Client{}),
				)
			},
			wantErr: false,
		},
		{
			name: "valid client with custom base URL",
			setup: func() *client.Client {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				t.Cleanup(server.Close)
				return client.NewClient(
					client.WithBaseURL(server.URL),
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseClient := tt.setup()
			accountingClient, err := NewClient(baseClient)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if accountingClient == nil {
					t.Error("NewClient() returned nil client")
					return
				}
				if accountingClient.client == nil {
					t.Error("NewClient() client field is nil")
				}
				if accountingClient.genClient == nil {
					t.Error("NewClient() genClient field is nil")
				}
			}
		})
	}
}

func TestClient_Deals(t *testing.T) {
	baseClient := client.NewClient()
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// First call should initialize the service
	deals1 := accountingClient.Deals()
	if deals1 == nil {
		t.Error("Deals() returned nil")
		return
	}

	// Second call should return the same instance (lazy initialization)
	deals2 := accountingClient.Deals()
	if deals1 != deals2 {
		t.Error("Deals() returned different instances, expected same instance")
	}

	// Verify the service has required fields
	if deals1.client == nil {
		t.Error("DealsService.client is nil")
	}
	if deals1.genClient == nil {
		t.Error("DealsService.genClient is nil")
	}
}

func TestClient_Journals(t *testing.T) {
	baseClient := client.NewClient()
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// First call should initialize the service
	journals1 := accountingClient.Journals()
	if journals1 == nil {
		t.Error("Journals() returned nil")
		return
	}

	// Second call should return the same instance (lazy initialization)
	journals2 := accountingClient.Journals()
	if journals1 != journals2 {
		t.Error("Journals() returned different instances, expected same instance")
	}

	// Verify the service has required fields
	if journals1.client == nil {
		t.Error("JournalsService.client is nil")
	}
	if journals1.genClient == nil {
		t.Error("JournalsService.genClient is nil")
	}
}

func TestClient_WalletTxns(t *testing.T) {
	baseClient := client.NewClient()
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// First call should initialize the service
	walletTxns1 := accountingClient.WalletTxns()
	if walletTxns1 == nil {
		t.Error("WalletTxns() returned nil")
		return
	}

	// Second call should return the same instance (lazy initialization)
	walletTxns2 := accountingClient.WalletTxns()
	if walletTxns1 != walletTxns2 {
		t.Error("WalletTxns() returned different instances, expected same instance")
	}

	// Verify the service has required fields
	if walletTxns1.client == nil {
		t.Error("WalletTxnService.client is nil")
	}
	if walletTxns1.genClient == nil {
		t.Error("WalletTxnService.genClient is nil")
	}
}

func TestClient_Transfers(t *testing.T) {
	baseClient := client.NewClient()
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// First call should initialize the service
	transfers1 := accountingClient.Transfers()
	if transfers1 == nil {
		t.Error("Transfers() returned nil")
		return
	}

	// Second call should return the same instance (lazy initialization)
	transfers2 := accountingClient.Transfers()
	if transfers1 != transfers2 {
		t.Error("Transfers() returned different instances, expected same instance")
	}

	// Verify the service has required fields
	if transfers1.client == nil {
		t.Error("TransfersService.client is nil")
	}
	if transfers1.genClient == nil {
		t.Error("TransfersService.genClient is nil")
	}
}

func TestClient_BaseClient(t *testing.T) {
	baseClient := client.NewClient()
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	retrievedClient := accountingClient.BaseClient()
	if retrievedClient == nil {
		t.Error("BaseClient() returned nil")
		return
	}
	if retrievedClient != baseClient {
		t.Error("BaseClient() returned different client than originally provided")
	}
}

func TestClient_GenClient(t *testing.T) {
	baseClient := client.NewClient()
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	genClient := accountingClient.GenClient()
	if genClient == nil {
		t.Error("GenClient() returned nil")
		return
	}
	if genClient != accountingClient.genClient {
		t.Error("GenClient() returned different client than internal genClient")
	}
}

func TestClient_AllServices(t *testing.T) {
	// Test that all services can be accessed together
	baseClient := client.NewClient()
	accountingClient, err := NewClient(baseClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	deals := accountingClient.Deals()
	journals := accountingClient.Journals()
	walletTxns := accountingClient.WalletTxns()
	transfers := accountingClient.Transfers()

	if deals == nil {
		t.Error("Deals service is nil")
	}
	if journals == nil {
		t.Error("Journals service is nil")
	}
	if walletTxns == nil {
		t.Error("WalletTxns service is nil")
	}
	if transfers == nil {
		t.Error("Transfers service is nil")
	}

	// Verify all services share the same underlying clients
	if deals.client != journals.client {
		t.Error("Services do not share the same base client")
	}
	if deals.genClient != journals.genClient {
		t.Error("Services do not share the same generated client")
	}
}
