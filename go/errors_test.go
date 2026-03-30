package safeurl

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "with code",
			err:      &APIError{StatusCode: 400, Message: "invalid request", Code: "INVALID_REQUEST"},
			expected: "safeurl: API error 400 (INVALID_REQUEST): invalid request",
		},
		{
			name:     "without code",
			err:      &APIError{StatusCode: 500, Message: "internal server error"},
			expected: "safeurl: API error 500: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAPIError_Is(t *testing.T) {
	tests := []struct {
		name       string
		err        *APIError
		unauthorized bool
		notFound   bool
		rateLimited bool
		serverError bool
	}{
		{
			name:         "unauthorized",
			err:          &APIError{StatusCode: 401},
			unauthorized: true,
		},
		{
			name:     "not found",
			err:      &APIError{StatusCode: 404},
			notFound: true,
		},
		{
			name:        "rate limited",
			err:         &APIError{StatusCode: 429},
			rateLimited: true,
		},
		{
			name:        "server error 500",
			err:         &APIError{StatusCode: 500},
			serverError: true,
		},
		{
			name:        "server error 503",
			err:         &APIError{StatusCode: 503},
			serverError: true,
		},
		{
			name: "client error",
			err:  &APIError{StatusCode: 400},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsUnauthorized(); got != tt.unauthorized {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.unauthorized)
			}
			if got := tt.err.IsNotFound(); got != tt.notFound {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.notFound)
			}
			if got := tt.err.IsRateLimited(); got != tt.rateLimited {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.rateLimited)
			}
			if got := tt.err.IsServerError(); got != tt.serverError {
				t.Errorf("IsServerError() = %v, want %v", got, tt.serverError)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	// Ensure sentinel errors are distinct
	errs := []error{
		ErrUnauthorized,
		ErrNotFound,
		ErrRateLimited,
		ErrTimeout,
		ErrBatchTooLarge,
		ErrInvalidURL,
		ErrScanFailed,
	}

	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("errors.Is(%v, %v) should be false", err1, err2)
			}
		}
	}
}

func TestAPIError_IsMethod(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		target   error
		expected bool
	}{
		{
			name:     "401 matches ErrUnauthorized",
			err:      &APIError{StatusCode: 401, Message: "unauthorized"},
			target:   ErrUnauthorized,
			expected: true,
		},
		{
			name:     "404 matches ErrNotFound",
			err:      &APIError{StatusCode: 404, Message: "not found"},
			target:   ErrNotFound,
			expected: true,
		},
		{
			name:     "429 matches ErrRateLimited",
			err:      &APIError{StatusCode: 429, Message: "rate limited"},
			target:   ErrRateLimited,
			expected: true,
		},
		{
			name:     "401 does not match ErrNotFound",
			err:      &APIError{StatusCode: 401, Message: "unauthorized"},
			target:   ErrNotFound,
			expected: false,
		},
		{
			name:     "500 does not match any sentinel",
			err:      &APIError{StatusCode: 500, Message: "server error"},
			target:   ErrUnauthorized,
			expected: false,
		},
		{
			name:     "does not match ErrTimeout",
			err:      &APIError{StatusCode: 408, Message: "timeout"},
			target:   ErrTimeout,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.expected {
				t.Errorf("errors.Is(APIError{%d}, %v) = %v, want %v",
					tt.err.StatusCode, tt.target, got, tt.expected)
			}
		})
	}
}
