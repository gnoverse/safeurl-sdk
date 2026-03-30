package safeurl

import (
	"errors"
	"fmt"
)

// Common errors returned by the SDK.
var (
	// ErrUnauthorized is returned when the API key is invalid or missing.
	ErrUnauthorized = errors.New("safeurl: unauthorized - invalid or missing API key")

	// ErrNotFound is returned when a scan ID is not found.
	ErrNotFound = errors.New("safeurl: scan not found")

	// ErrRateLimited is returned when the API rate limit is exceeded.
	ErrRateLimited = errors.New("safeurl: rate limit exceeded")

	// ErrTimeout is returned when a scan times out waiting for completion.
	ErrTimeout = errors.New("safeurl: scan timed out waiting for completion")

	// ErrBatchTooLarge is returned when a batch scan request exceeds the maximum URL count.
	ErrBatchTooLarge = errors.New("safeurl: batch size exceeds maximum of 50 URLs")

	// ErrInvalidURL is returned when a URL is invalid or empty.
	ErrInvalidURL = errors.New("safeurl: invalid or empty URL")

	// ErrScanFailed is returned when a scan fails to complete.
	ErrScanFailed = errors.New("safeurl: scan failed")
)

// APIError represents an error response from the SafeURL API.
type APIError struct {
	StatusCode int
	Message    string
	Code       string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("safeurl: API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("safeurl: API error %d: %s", e.StatusCode, e.Message)
}

// IsUnauthorized returns true if this is an unauthorized error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == 401
}

// IsNotFound returns true if this is a not found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsRateLimited returns true if this is a rate limit error.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}

// IsServerError returns true if this is a server error (5xx).
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// Is implements errors.Is matching for APIError.
// This allows using errors.Is(err, ErrUnauthorized) etc. with APIError instances.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrUnauthorized:
		return e.StatusCode == 401
	case ErrNotFound:
		return e.StatusCode == 404
	case ErrRateLimited:
		return e.StatusCode == 429
	default:
		return false
	}
}
