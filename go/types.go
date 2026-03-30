package safeurl

import (
	"time"
)

// ScanState represents the state of a scan.
type ScanState string

// Scan states returned by the API.
const (
	ScanStatePending    ScanState = "pending"
	ScanStateProcessing ScanState = "processing"
	ScanStateCompleted  ScanState = "completed"
	ScanStateFailed     ScanState = "failed"
)

// IsTerminal returns true if the scan state is terminal (completed or failed).
func (s ScanState) IsTerminal() bool {
	return s == ScanStateCompleted || s == ScanStateFailed
}

// Verdict represents the safety verdict for a URL.
type Verdict string

// Verdicts returned by the API after a scan completes.
const (
	VerdictSafe      Verdict = "safe"
	VerdictMalicious Verdict = "malicious"
	VerdictSuspect   Verdict = "suspect"
	VerdictUnknown   Verdict = "unknown"
)

// IsSafe returns true if the verdict indicates the URL is safe.
func (v Verdict) IsSafe() bool {
	return v == VerdictSafe
}

// IsUnsafe returns true if the verdict indicates the URL is potentially dangerous.
func (v Verdict) IsUnsafe() bool {
	return v == VerdictMalicious || v == VerdictSuspect
}

// ScanRequest represents the request body for creating a scan.
type ScanRequest struct {
	URL string `json:"url"`
}

// BatchScanRequest represents the request body for creating a batch scan.
type BatchScanRequest struct {
	URLs []string `json:"urls"`
}

// ScanResponse represents a scan result from the API.
type ScanResponse struct {
	ID        string     `json:"id"`
	URL       string     `json:"url"`
	State     ScanState  `json:"state"`
	Verdict   Verdict    `json:"verdict,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// IsComplete returns true if the scan has finished processing.
func (s *ScanResponse) IsComplete() bool {
	return s.State.IsTerminal()
}

// BatchScanResponse represents the response from a batch scan request.
type BatchScanResponse struct {
	BatchID string         `json:"batchId"`
	Scans   []ScanResponse `json:"scans"`
}

// GetScanResponse represents the response when fetching a scan by ID.
type GetScanResponse = ScanResponse

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}
