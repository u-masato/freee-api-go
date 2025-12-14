package client

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestFreeeError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *FreeeError
		want string
	}{
		{
			name: "with detailed errors",
			err: &FreeeError{
				StatusCode: 400,
				Errors: []ErrorDetail{
					{
						Type:     ErrorTypeValidation,
						Messages: []string{"Invalid company_id"},
					},
				},
			},
			want: "freee API error (status 400): [validation] Invalid company_id",
		},
		{
			name: "with single message",
			err: &FreeeError{
				StatusCode: 401,
				Message:    "Unauthorized access",
			},
			want: "freee API error (status 401): Unauthorized access",
		},
		{
			name: "with messages field",
			err: &FreeeError{
				StatusCode: 403,
				Messages:   "Forbidden resource",
			},
			want: "freee API error (status 403): Forbidden resource",
		},
		{
			name: "with only status code",
			err: &FreeeError{
				StatusCode: 500,
			},
			want: "freee API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("FreeeError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFreeeError_Is(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		target     error
		want       bool
	}{
		{"400 matches ErrBadRequest", 400, ErrBadRequest, true},
		{"401 matches ErrUnauthorized", 401, ErrUnauthorized, true},
		{"403 matches ErrForbidden", 403, ErrForbidden, true},
		{"404 matches ErrNotFound", 404, ErrNotFound, true},
		{"409 matches ErrConflict", 409, ErrConflict, true},
		{"429 matches ErrTooManyRequests", 429, ErrTooManyRequests, true},
		{"500 matches ErrInternalServer", 500, ErrInternalServer, true},
		{"503 matches ErrServiceUnavailable", 503, ErrServiceUnavailable, true},
		{"400 does not match ErrUnauthorized", 400, ErrUnauthorized, false},
		{"400 does not match random error", 400, errors.New("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &FreeeError{StatusCode: tt.statusCode}
			got := err.Is(tt.target)
			if got != tt.want {
				t.Errorf("FreeeError.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFreeeError_HasValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  *FreeeError
		want bool
	}{
		{
			name: "has validation error",
			err: &FreeeError{
				Errors: []ErrorDetail{
					{Type: ErrorTypeValidation, Messages: []string{"Invalid field"}},
				},
			},
			want: true,
		},
		{
			name: "has only status error",
			err: &FreeeError{
				Errors: []ErrorDetail{
					{Type: ErrorTypeStatus, Messages: []string{"Status error"}},
				},
			},
			want: false,
		},
		{
			name: "has mixed error types",
			err: &FreeeError{
				Errors: []ErrorDetail{
					{Type: ErrorTypeStatus, Messages: []string{"Status error"}},
					{Type: ErrorTypeValidation, Messages: []string{"Validation error"}},
				},
			},
			want: true,
		},
		{
			name: "no errors",
			err:  &FreeeError{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.HasValidationError()
			if got != tt.want {
				t.Errorf("FreeeError.HasValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFreeeError_GetMessages(t *testing.T) {
	tests := []struct {
		name string
		err  *FreeeError
		want []string
	}{
		{
			name: "from Errors field",
			err: &FreeeError{
				Errors: []ErrorDetail{
					{Messages: []string{"Error 1", "Error 2"}},
					{Messages: []string{"Error 3"}},
				},
			},
			want: []string{"Error 1", "Error 2", "Error 3"},
		},
		{
			name: "from Message field",
			err: &FreeeError{
				Message: "Single error",
			},
			want: []string{"Single error"},
		},
		{
			name: "from Messages field",
			err: &FreeeError{
				Messages: "Another error",
			},
			want: []string{"Another error"},
		},
		{
			name: "from multiple sources",
			err: &FreeeError{
				Errors: []ErrorDetail{
					{Messages: []string{"Detailed error"}},
				},
				Message:  "Message error",
				Messages: "Messages error",
			},
			want: []string{"Detailed error", "Message error", "Messages error"},
		},
		{
			name: "no messages",
			err:  &FreeeError{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.GetMessages()
			if len(got) != len(tt.want) {
				t.Errorf("FreeeError.GetMessages() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FreeeError.GetMessages()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
		checkFunc  func(*testing.T, error)
	}{
		{
			name:       "valid JSON error response",
			statusCode: 400,
			body: `{
				"status_code": 400,
				"errors": [
					{
						"type": "validation",
						"messages": ["Invalid parameter"]
					}
				]
			}`,
			wantErr: true,
			checkFunc: func(t *testing.T, err error) {
				var freeeErr *FreeeError
				if !errors.As(err, &freeeErr) {
					t.Fatal("error is not FreeeError")
				}
				if freeeErr.StatusCode != 400 {
					t.Errorf("StatusCode = %d, want 400", freeeErr.StatusCode)
				}
				if !freeeErr.HasValidationError() {
					t.Error("expected validation error")
				}
			},
		},
		{
			name:       "simple message error",
			statusCode: 401,
			body:       `{"message": "Unauthorized"}`,
			wantErr:    true,
			checkFunc: func(t *testing.T, err error) {
				var freeeErr *FreeeError
				if !errors.As(err, &freeeErr) {
					t.Fatal("error is not FreeeError")
				}
				if freeeErr.Message != "Unauthorized" {
					t.Errorf("Message = %q, want %q", freeeErr.Message, "Unauthorized")
				}
			},
		},
		{
			name:       "invalid JSON",
			statusCode: 500,
			body:       "Internal Server Error",
			wantErr:    true,
			checkFunc: func(t *testing.T, err error) {
				var freeeErr *FreeeError
				if !errors.As(err, &freeeErr) {
					t.Fatal("error is not FreeeError")
				}
				if freeeErr.Message != "Internal Server Error" {
					t.Errorf("Message = %q, want %q", freeeErr.Message, "Internal Server Error")
				}
			},
		},
		{
			name:       "nil response",
			statusCode: 0,
			body:       "",
			wantErr:    true,
			checkFunc: func(t *testing.T, err error) {
				if err.Error() != "nil response" {
					t.Errorf("error = %q, want %q", err.Error(), "nil response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.name != "nil response" {
				resp = &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(tt.body)),
				}
			}

			err := ParseErrorResponse(resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseErrorResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			}
		})
	}
}

func TestIsErrorFunctions(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		checkFunc func(error) bool
		want      bool
	}{
		{"IsBadRequestError with ErrBadRequest", ErrBadRequest, IsBadRequestError, true},
		{"IsBadRequestError with FreeeError 400", &FreeeError{StatusCode: 400}, IsBadRequestError, true},
		{"IsBadRequestError with other error", errors.New("other"), IsBadRequestError, false},

		{"IsUnauthorizedError with ErrUnauthorized", ErrUnauthorized, IsUnauthorizedError, true},
		{"IsUnauthorizedError with FreeeError 401", &FreeeError{StatusCode: 401}, IsUnauthorizedError, true},
		{"IsUnauthorizedError with other error", errors.New("other"), IsUnauthorizedError, false},

		{"IsForbiddenError with ErrForbidden", ErrForbidden, IsForbiddenError, true},
		{"IsForbiddenError with FreeeError 403", &FreeeError{StatusCode: 403}, IsForbiddenError, true},

		{"IsNotFoundError with ErrNotFound", ErrNotFound, IsNotFoundError, true},
		{"IsNotFoundError with FreeeError 404", &FreeeError{StatusCode: 404}, IsNotFoundError, true},

		{"IsConflictError with ErrConflict", ErrConflict, IsConflictError, true},
		{"IsConflictError with FreeeError 409", &FreeeError{StatusCode: 409}, IsConflictError, true},

		{"IsTooManyRequestsError with ErrTooManyRequests", ErrTooManyRequests, IsTooManyRequestsError, true},
		{"IsTooManyRequestsError with FreeeError 429", &FreeeError{StatusCode: 429}, IsTooManyRequestsError, true},

		{"IsInternalServerError with ErrInternalServer", ErrInternalServer, IsInternalServerError, true},
		{"IsInternalServerError with FreeeError 500", &FreeeError{StatusCode: 500}, IsInternalServerError, true},

		{"IsServiceUnavailableError with ErrServiceUnavailable", ErrServiceUnavailable, IsServiceUnavailableError, true},
		{"IsServiceUnavailableError with FreeeError 503", &FreeeError{StatusCode: 503}, IsServiceUnavailableError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checkFunc(tt.err)
			if got != tt.want {
				t.Errorf("check function = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFreeeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"FreeeError", &FreeeError{}, true},
		{"standard error", errors.New("standard"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFreeeError(tt.err)
			if got != tt.want {
				t.Errorf("IsFreeeError() = %v, want %v", got, tt.want)
			}
		})
	}
}
