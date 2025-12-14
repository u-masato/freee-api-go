package gen

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_NewClient verifies that a client can be created with a valid server URL.
func TestClient_NewClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}
}

// TestClient_WithHTTPClient verifies that a custom HTTP client can be provided.
func TestClient_WithHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	customHTTPClient := &http.Client{
		Timeout: 0, // No timeout for testing
	}

	client, err := NewClient(server.URL, WithHTTPClient(customHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}
}

// TestClient_GetAccountItems verifies the GetAccountItems method can be called.
// This is a basic smoke test to ensure the generated method signatures are correct.
func TestClient_GetAccountItems(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/account_items" {
			t.Errorf("Expected path /api/1/account_items, got %s", r.URL.Path)
		}

		// Return a minimal valid response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"account_items": []}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	params := &GetAccountItemsParams{
		CompanyId: 123456,
	}

	resp, err := client.GetAccountItems(ctx, params)
	if err != nil {
		t.Fatalf("GetAccountItems() error = %v", err)
	}
	defer resp.Body.Close()

	if !requestReceived {
		t.Error("Expected request to be received by test server")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestClient_CreateAccountItem verifies the CreateAccountItem method can be called.
func TestClient_CreateAccountItem(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/1/account_items" {
			t.Errorf("Expected path /api/1/account_items, got %s", r.URL.Path)
		}

		// Verify Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Return a minimal valid response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"account_item": {"id": 1, "name": "Test Account"}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()

	// Create a minimal request body
	// Note: The exact structure depends on the generated types
	body := CreateAccountItemJSONRequestBody{
		CompanyId: 123456,
		// Other required fields would go here
	}

	resp, err := client.CreateAccountItem(ctx, body)
	if err != nil {
		t.Fatalf("CreateAccountItem() error = %v", err)
	}
	defer resp.Body.Close()

	if !requestReceived {
		t.Error("Expected request to be received by test server")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

// TestClient_ErrorResponse verifies that error responses are returned correctly.
func TestClient_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"status_code": 400,
			"errors": [{
				"type": "validation",
				"messages": ["Invalid company_id"]
			}]
		}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	params := &GetAccountItemsParams{
		CompanyId: -1, // Invalid company ID
	}

	resp, err := client.GetAccountItems(ctx, params)
	if err != nil {
		t.Fatalf("GetAccountItems() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

// TestClientWithResponses_NewClientWithResponses verifies the response client can be created.
func TestClientWithResponses_NewClientWithResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClientWithResponses(server.URL)
	if err != nil {
		t.Fatalf("NewClientWithResponses() error = %v", err)
	}

	if client == nil {
		t.Fatal("NewClientWithResponses() returned nil client")
	}
}
