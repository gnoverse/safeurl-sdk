package safeurl

import (
	"time"
)

// ScanState represents the state of a scan.
type ScanState string

// Scan states returned by the API.
const (
	ScanStateQueued    ScanState = "QUEUED"
	ScanStateFetching  ScanState = "FETCHING"
	ScanStateAnalyzing ScanState = "ANALYZING"
	ScanStateCompleted ScanState = "COMPLETED"
	ScanStateFailed    ScanState = "FAILED"
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

// ScanResult contains the analysis result from a completed scan.
type ScanResult struct {
	RiskScore   float64  `json:"riskScore"`
	Categories  []string `json:"categories,omitempty"`
	Reasoning   string   `json:"reasoning,omitempty"`
	ContentType string   `json:"contentType,omitempty"`
}

// ScanResponse represents a scan result from the API.
type ScanResponse struct {
	ID           string      `json:"id"`
	URL          string      `json:"url"`
	State        ScanState   `json:"state"`
	Deduplicated bool        `json:"deduplicated,omitempty"`
	Result       *ScanResult `json:"result,omitempty"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	ExpiresAt    *time.Time  `json:"expiresAt,omitempty"`
}

// IsComplete returns true if the scan has finished processing.
func (s *ScanResponse) IsComplete() bool {
	return s.State.IsTerminal()
}

// GetVerdict returns the safety verdict based on the risk score.
// Risk score thresholds: 0-30 safe, 31-60 suspect, 61-100 malicious.
func (s *ScanResponse) GetVerdict() Verdict {
	if s.Result == nil {
		return VerdictUnknown
	}
	riskScore := s.Result.RiskScore
	switch {
	case riskScore <= 30:
		return VerdictSafe
	case riskScore <= 60:
		return VerdictSuspect
	default:
		return VerdictMalicious
	}
}

// BatchScanResponse represents the response from a batch scan request.
type BatchScanResponse struct {
	BatchID string         `json:"batchId"`
	Jobs    []ScanResponse `json:"jobs"`
}

// GetScanResponse represents the response when fetching a scan by ID.
type GetScanResponse = ScanResponse

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}
