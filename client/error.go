package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Common API errors.
var (
	// ErrBadRequest is returned for 400 Bad Request responses.
	ErrBadRequest = errors.New("bad request")

	// ErrUnauthorized is returned for 401 Unauthorized responses.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned for 403 Forbidden responses.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound is returned for 404 Not Found responses.
	ErrNotFound = errors.New("not found")

	// ErrConflict is returned for 409 Conflict responses.
	ErrConflict = errors.New("conflict")

	// ErrTooManyRequests is returned for 429 Too Many Requests responses.
	ErrTooManyRequests = errors.New("too many requests")

	// ErrInternalServer is returned for 500 Internal Server Error responses.
	ErrInternalServer = errors.New("internal server error")

	// ErrServiceUnavailable is returned for 503 Service Unavailable responses.
	ErrServiceUnavailable = errors.New("service unavailable")
)

// ErrorType represents the type of error in freee API responses.
type ErrorType string

const (
	// ErrorTypeStatus indicates a status-related error.
	ErrorTypeStatus ErrorType = "status"

	// ErrorTypeValidation indicates a validation error.
	ErrorTypeValidation ErrorType = "validation"

	// ErrorTypeError indicates a general error.
	ErrorTypeError ErrorType = "error"
)

// ErrorDetail represents a single error detail from the freee API.
type ErrorDetail struct {
	// Type is the error type (status, validation, error).
	Type ErrorType `json:"type"`

	// Messages contains the error messages.
	Messages []string `json:"messages"`
}

// FreeeError represents an error response from the freee API.
// It provides detailed error information including HTTP status code,
// error messages, and error types.
type FreeeError struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"status_code"`

	// Errors contains detailed error information.
	Errors []ErrorDetail `json:"errors,omitempty"`

	// Message is a single error message (used in some error responses).
	Message string `json:"message,omitempty"`

	// Messages is an alternative message field (used in some error responses).
	Messages string `json:"messages,omitempty"`

	// Response is the original HTTP response.
	Response *http.Response `json:"-"`
}

// Error implements the error interface.
func (e *FreeeError) Error() string {
	if len(e.Errors) > 0 {
		var messages []string
		for _, err := range e.Errors {
			for _, msg := range err.Messages {
				messages = append(messages, fmt.Sprintf("[%s] %s", err.Type, msg))
			}
		}
		if len(messages) > 0 {
			return fmt.Sprintf("freee API error (status %d): %s", e.StatusCode, messages[0])
		}
	}

	if e.Message != "" {
		return fmt.Sprintf("freee API error (status %d): %s", e.StatusCode, e.Message)
	}

	if e.Messages != "" {
		return fmt.Sprintf("freee API error (status %d): %s", e.StatusCode, e.Messages)
	}

	return fmt.Sprintf("freee API error (status %d)", e.StatusCode)
}

// Is allows FreeeError to be compared with standard errors.
func (e *FreeeError) Is(target error) bool {
	switch e.StatusCode {
	case http.StatusBadRequest:
		return target == ErrBadRequest
	case http.StatusUnauthorized:
		return target == ErrUnauthorized
	case http.StatusForbidden:
		return target == ErrForbidden
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusConflict:
		return target == ErrConflict
	case http.StatusTooManyRequests:
		return target == ErrTooManyRequests
	case http.StatusInternalServerError:
		return target == ErrInternalServer
	case http.StatusServiceUnavailable:
		return target == ErrServiceUnavailable
	}
	return false
}

// HasValidationError checks if the error contains validation errors.
func (e *FreeeError) HasValidationError() bool {
	for _, err := range e.Errors {
		if err.Type == ErrorTypeValidation {
			return true
		}
	}
	return false
}

// GetMessages returns all error messages as a single slice.
func (e *FreeeError) GetMessages() []string {
	var messages []string

	for _, err := range e.Errors {
		messages = append(messages, err.Messages...)
	}

	if e.Message != "" {
		messages = append(messages, e.Message)
	}

	if e.Messages != "" {
		messages = append(messages, e.Messages)
	}

	return messages
}

// ParseErrorResponse attempts to parse an HTTP error response into a FreeeError.
// If parsing fails, it returns a generic FreeeError with the status code.
func ParseErrorResponse(resp *http.Response) error {
	if resp == nil {
		return errors.New("nil response")
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &FreeeError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("failed to read error response: %v", err),
			Response:   resp,
		}
	}
	defer resp.Body.Close()

	// Try to parse as FreeeError
	var freeeErr FreeeError
	if err := json.Unmarshal(body, &freeeErr); err != nil {
		// If JSON parsing fails, use the raw body as the message
		return &FreeeError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
			Response:   resp,
		}
	}

	freeeErr.StatusCode = resp.StatusCode
	freeeErr.Response = resp
	return &freeeErr
}

// IsFreeeError checks if the error is a FreeeError.
func IsFreeeError(err error) bool {
	var freeeErr *FreeeError
	return errors.As(err, &freeeErr)
}

// IsBadRequestError checks if the error is a 400 Bad Request error.
func IsBadRequestError(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

// IsUnauthorizedError checks if the error is a 401 Unauthorized error.
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbiddenError checks if the error is a 403 Forbidden error.
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsNotFoundError checks if the error is a 404 Not Found error.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsConflictError checks if the error is a 409 Conflict error.
func IsConflictError(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsTooManyRequestsError checks if the error is a 429 Too Many Requests error.
func IsTooManyRequestsError(err error) bool {
	return errors.Is(err, ErrTooManyRequests)
}

// IsInternalServerError checks if the error is a 500 Internal Server Error.
func IsInternalServerError(err error) bool {
	return errors.Is(err, ErrInternalServer)
}

// IsServiceUnavailableError checks if the error is a 503 Service Unavailable error.
func IsServiceUnavailableError(err error) bool {
	return errors.Is(err, ErrServiceUnavailable)
}
