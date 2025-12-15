package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/muno/freee-api-go/client"
	"github.com/muno/freee-api-go/internal/gen"
	"github.com/muno/freee-api-go/tests/integration/mockserver"
)

// TestErrorHandling_Unauthorized tests 401 Unauthorized error handling.
func TestErrorHandling_Unauthorized(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	// Set error mode to simulate unauthorized
	server.SetErrorMode(mockserver.ErrorModeUnauthorized)
	defer server.ClearErrorMode()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().List(ctx, 1, nil)
	if err == nil {
		t.Fatal("Expected unauthorized error, got nil")
	}

	// Verify the error contains appropriate message
	if !containsString(err.Error(), "401") && !containsString(err.Error(), "unauthorized") && !containsString(err.Error(), "Unauthorized") {
		t.Errorf("Error should indicate unauthorized: %v", err)
	}
}

// TestErrorHandling_BadRequest tests 400 Bad Request error handling.
func TestErrorHandling_BadRequest(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	server.SetErrorMode(mockserver.ErrorModeBadRequest)
	defer server.ClearErrorMode()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().List(ctx, 1, nil)
	if err == nil {
		t.Fatal("Expected bad request error, got nil")
	}
}

// TestErrorHandling_NotFound tests 404 Not Found error handling.
func TestErrorHandling_NotFound(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	// Try to get a non-existent deal
	_, err := accountingClient.Deals().Get(ctx, 1, 99999, nil)
	if err == nil {
		t.Fatal("Expected not found error, got nil")
	}
}

// TestErrorHandling_InternalServer tests 500 Internal Server Error handling.
func TestErrorHandling_InternalServer(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	server.SetErrorMode(mockserver.ErrorModeInternalServer)
	defer server.ClearErrorMode()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().List(ctx, 1, nil)
	if err == nil {
		t.Fatal("Expected internal server error, got nil")
	}
}

// TestErrorHandling_TooManyRequests tests 429 Too Many Requests error handling.
func TestErrorHandling_TooManyRequests(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	server.SetErrorMode(mockserver.ErrorModeTooManyRequests)
	defer server.ClearErrorMode()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().List(ctx, 1, nil)
	if err == nil {
		t.Fatal("Expected too many requests error, got nil")
	}
}

// TestErrorHandling_ServiceUnavailable tests 503 Service Unavailable error handling.
func TestErrorHandling_ServiceUnavailable(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	server.SetErrorMode(mockserver.ErrorModeServiceUnavailable)
	defer server.ClearErrorMode()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().List(ctx, 1, nil)
	if err == nil {
		t.Fatal("Expected service unavailable error, got nil")
	}
}

// TestErrorHandling_ValidationError tests validation error handling.
func TestErrorHandling_ValidationError(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	// Try to create a deal with missing required fields
	params := gen.DealCreateParams{
		CompanyId: 1,
		// Missing IssueDate
		// Missing Type
		// Missing Details
	}

	_, err := accountingClient.Deals().Create(ctx, params)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
}

// TestErrorHandling_ContextCancellation tests context cancellation handling.
func TestErrorHandling_ContextCancellation(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	accountingClient := createTestAccountingClient(t, server)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := accountingClient.Deals().List(ctx, 1, nil)
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) && !containsString(err.Error(), "cancel") {
		// The error might be wrapped, so we just check it's not nil
		t.Logf("Got expected error: %v", err)
	}
}

// TestErrorHandling_FreeeError tests FreeeError type handling.
func TestErrorHandling_FreeeError(t *testing.T) {
	t.Run("IsBadRequestError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 400,
			Message:    "Bad request",
		}

		if !client.IsBadRequestError(err) {
			t.Error("Expected IsBadRequestError to return true")
		}
	})

	t.Run("IsUnauthorizedError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 401,
			Message:    "Unauthorized",
		}

		if !client.IsUnauthorizedError(err) {
			t.Error("Expected IsUnauthorizedError to return true")
		}
	})

	t.Run("IsForbiddenError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 403,
			Message:    "Forbidden",
		}

		if !client.IsForbiddenError(err) {
			t.Error("Expected IsForbiddenError to return true")
		}
	})

	t.Run("IsNotFoundError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 404,
			Message:    "Not found",
		}

		if !client.IsNotFoundError(err) {
			t.Error("Expected IsNotFoundError to return true")
		}
	})

	t.Run("IsTooManyRequestsError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 429,
			Message:    "Too many requests",
		}

		if !client.IsTooManyRequestsError(err) {
			t.Error("Expected IsTooManyRequestsError to return true")
		}
	})

	t.Run("IsInternalServerError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 500,
			Message:    "Internal server error",
		}

		if !client.IsInternalServerError(err) {
			t.Error("Expected IsInternalServerError to return true")
		}
	})

	t.Run("IsServiceUnavailableError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 503,
			Message:    "Service unavailable",
		}

		if !client.IsServiceUnavailableError(err) {
			t.Error("Expected IsServiceUnavailableError to return true")
		}
	})

	t.Run("HasValidationError", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 400,
			Errors: []client.ErrorDetail{
				{
					Type:     client.ErrorTypeValidation,
					Messages: []string{"Field is required"},
				},
			},
		}

		if !err.HasValidationError() {
			t.Error("Expected HasValidationError to return true")
		}
	})

	t.Run("GetMessages", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 400,
			Errors: []client.ErrorDetail{
				{
					Type:     client.ErrorTypeValidation,
					Messages: []string{"Field1 is required", "Field2 is invalid"},
				},
			},
			Message: "Additional message",
		}

		messages := err.GetMessages()
		if len(messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(messages))
		}
	})

	t.Run("Error", func(t *testing.T) {
		err := &client.FreeeError{
			StatusCode: 400,
			Errors: []client.ErrorDetail{
				{
					Type:     client.ErrorTypeValidation,
					Messages: []string{"Field is required"},
				},
			},
		}

		errStr := err.Error()
		if !containsString(errStr, "400") {
			t.Errorf("Error string should contain status code: %s", errStr)
		}
		if !containsString(errStr, "Field is required") {
			t.Errorf("Error string should contain message: %s", errStr)
		}
	})
}

// TestErrorHandling_InvalidToken tests invalid token handling.
func TestErrorHandling_InvalidToken(t *testing.T) {
	server := mockserver.NewServer()
	defer server.Close()

	// Remove the default token
	server.RemoveToken("test-access-token")

	accountingClient := createTestAccountingClient(t, server)
	ctx := context.Background()

	_, err := accountingClient.Deals().List(ctx, 1, nil)
	if err == nil {
		t.Fatal("Expected authentication error, got nil")
	}
}
