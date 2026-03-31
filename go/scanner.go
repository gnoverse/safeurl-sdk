package safeurl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ScannerConfig configures the Scanner behavior.
type ScannerConfig struct {
	// PollInterval is how often to poll for scan completion (default: 500ms).
	PollInterval time.Duration
	// MaxWait is the maximum time to wait for a scan to complete (default: 30s).
	MaxWait time.Duration
}

// DefaultScannerConfig returns sensible defaults for the scanner.
func DefaultScannerConfig() ScannerConfig {
	return ScannerConfig{
		PollInterval: 500 * time.Millisecond,
		MaxWait:      30 * time.Second,
	}
}

// Scanner provides high-level methods for scanning URLs.
// It wraps the low-level ClientWithResponses and adds polling, error handling,
// and typed responses.
type Scanner struct {
	client *ClientWithResponses
	config ScannerConfig
}

// NewScanner creates a new Scanner with the given API key and optional configuration.
func NewScanner(apiKey string, opts ...ScannerOption) (*Scanner, error) {
	return NewScannerWithBaseURL(DefaultBaseURL, apiKey, opts...)
}

// NewScannerWithBaseURL creates a new Scanner with a custom base URL.
func NewScannerWithBaseURL(baseURL, apiKey string, opts ...ScannerOption) (*Scanner, error) {
	client, err := NewClientWithAPIKey(baseURL, apiKey)
	if err != nil {
		return nil, err
	}

	s := &Scanner{
		client: client,
		config: DefaultScannerConfig(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// ScannerOption configures a Scanner.
type ScannerOption func(*Scanner)

// WithPollInterval sets the polling interval for waiting on scan completion.
func WithPollInterval(d time.Duration) ScannerOption {
	return func(s *Scanner) {
		s.config.PollInterval = d
	}
}

// WithMaxWait sets the maximum time to wait for scan completion.
func WithMaxWait(d time.Duration) ScannerOption {
	return func(s *Scanner) {
		s.config.MaxWait = d
	}
}

// WithScannerConfig sets the full scanner configuration.
func WithScannerConfig(cfg ScannerConfig) ScannerOption {
	return func(s *Scanner) {
		s.config = cfg
	}
}

// ScanURL scans a single URL and waits for the result.
// It returns the completed scan result or an error if the scan fails or times out.
func (s *Scanner) ScanURL(ctx context.Context, rawURL string) (*ScanResponse, error) {
	if rawURL == "" {
		return nil, ErrInvalidURL
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, ErrInvalidURL
	}

	// Submit the scan
	req := ScanRequest{URL: rawURL}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("safeurl: failed to marshal request: %w", err)
	}

	resp, err := s.client.PostV1ScansWithBodyWithResponse(ctx, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("safeurl: failed to submit scan: %w", err)
	}

	if err := checkResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}

	var scan ScanResponse
	if err := json.Unmarshal(resp.Body, &scan); err != nil {
		return nil, fmt.Errorf("safeurl: failed to parse response: %w", err)
	}

	// If already complete, return immediately
	if scan.IsComplete() {
		if scan.State == ScanStateFailed {
			return &scan, ErrScanFailed
		}
		return &scan, nil
	}

	// Poll for completion
	return s.waitForScan(ctx, scan.ID)
}

// ScanURLs scans multiple URLs and waits for all results.
// Duplicate URLs are automatically deduplicated before scanning.
// URLs are automatically chunked if they exceed the batch limit (50).
// Returns a map of URL to scan result.
func (s *Scanner) ScanURLs(ctx context.Context, urls []string) (map[string]*ScanResponse, error) {
	if len(urls) == 0 {
		return make(map[string]*ScanResponse), nil
	}

	// Deduplicate URLs upfront to avoid splitting duplicates across chunks
	seen := make(map[string]bool)
	uniqueURLs := make([]string, 0, len(urls))
	for _, u := range urls {
		if !seen[u] {
			seen[u] = true
			uniqueURLs = append(uniqueURLs, u)
		}
	}

	results := make(map[string]*ScanResponse)

	// Process in chunks
	for i := 0; i < len(uniqueURLs); i += BatchScanMaxURLs {
		end := i + BatchScanMaxURLs
		if end > len(uniqueURLs) {
			end = len(uniqueURLs)
		}
		chunk := uniqueURLs[i:end]

		chunkResults, err := s.ScanBatch(ctx, chunk)
		if err != nil {
			return results, err
		}

		for url, result := range chunkResults {
			results[url] = result
		}
	}

	return results, nil
}

// ScanBatch scans a single batch of URLs (max 50) and waits for completion.
// Duplicate URLs are automatically deduplicated before sending to the API.
// Returns ErrBatchTooLarge if the number of unique URLs exceeds BatchScanMaxURLs.
func (s *Scanner) ScanBatch(ctx context.Context, urls []string) (map[string]*ScanResponse, error) {
	// Deduplicate URLs
	seen := make(map[string]bool)
	uniqueURLs := make([]string, 0, len(urls))
	for _, u := range urls {
		if !seen[u] {
			seen[u] = true
			uniqueURLs = append(uniqueURLs, u)
		}
	}

	if len(uniqueURLs) == 0 {
		return make(map[string]*ScanResponse), nil
	}

	if len(uniqueURLs) > BatchScanMaxURLs {
		return nil, ErrBatchTooLarge
	}

	req := BatchScanRequest{URLs: uniqueURLs}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("safeurl: failed to marshal request: %w", err)
	}

	resp, err := s.client.PostV1ScansBatchWithBodyWithResponse(ctx, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("safeurl: failed to submit batch scan: %w", err)
	}

	if err := checkResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}

	var batchResp BatchScanResponse
	if err := json.Unmarshal(resp.Body, &batchResp); err != nil {
		return nil, fmt.Errorf("safeurl: failed to parse response: %w", err)
	}

	results := make(map[string]*ScanResponse)
	pending := make(map[string]string) // scanID -> url

	for i := range batchResp.Jobs {
		scan := &batchResp.Jobs[i]
		if scan.IsComplete() {
			results[scan.URL] = scan
		} else {
			pending[scan.ID] = scan.URL
		}
	}

	// Poll for pending scans
	var firstErr error
	for scanID, url := range pending {
		scan, err := s.waitForScan(ctx, scanID)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			// Continue processing other URLs, but track the error
			continue
		}
		results[url] = scan
	}

	return results, firstErr
}

func (s *Scanner) waitForScan(ctx context.Context, scanID string) (*ScanResponse, error) {
	deadline := time.Now().Add(s.config.MaxWait)
	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, ErrTimeout
			}

			scan, err := s.GetScan(ctx, scanID)
			if err != nil {
				return nil, err
			}

			if scan.IsComplete() {
				if scan.State == ScanStateFailed {
					return scan, ErrScanFailed
				}
				return scan, nil
			}
		}
	}
}

// GetScan retrieves a scan by ID.
func (s *Scanner) GetScan(ctx context.Context, scanID string) (*ScanResponse, error) {
	parsedID, err := uuid.Parse(scanID)
	if err != nil {
		return nil, fmt.Errorf("safeurl: invalid scan ID: %w", err)
	}

	resp, err := s.client.GetV1ScansByIdWithResponse(ctx, openapi_types.UUID(parsedID))
	if err != nil {
		return nil, fmt.Errorf("safeurl: failed to get scan: %w", err)
	}

	if err := checkResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}

	var scan ScanResponse
	if err := json.Unmarshal(resp.Body, &scan); err != nil {
		return nil, fmt.Errorf("safeurl: failed to parse response: %w", err)
	}

	return &scan, nil
}

// checkResponse checks the HTTP response for errors and returns an appropriate error.
func checkResponse(resp *http.Response, body []byte) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    http.StatusText(resp.StatusCode),
	}

	// Try to parse error response
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errResp.Message != "" {
			apiErr.Message = errResp.Message
		} else if errResp.Error != "" {
			apiErr.Message = errResp.Error
		}
		apiErr.Code = errResp.Code
	}

	return apiErr
}

// Client returns the underlying ClientWithResponses for advanced use cases.
func (s *Scanner) Client() *ClientWithResponses {
	return s.client
}

// QuickScan is a convenience function that creates a Scanner, scans a URL, and returns the result.
// For scanning multiple URLs or repeated scans, create a Scanner instance instead.
func QuickScan(ctx context.Context, apiKey, url string) (*ScanResponse, error) {
	scanner, err := NewScanner(apiKey)
	if err != nil {
		return nil, err
	}
	return scanner.ScanURL(ctx, url)
}
