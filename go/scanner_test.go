package safeurl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewScanner(t *testing.T) {
	scanner, err := NewScanner("test-api-key")
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}
	if scanner == nil {
		t.Fatal("NewScanner() returned nil")
	}
	if scanner.client == nil {
		t.Fatal("Scanner.client is nil")
	}
}

func TestNewScannerWithOptions(t *testing.T) {
	scanner, err := NewScanner("test-api-key",
		WithPollInterval(100*time.Millisecond),
		WithMaxWait(5*time.Second),
	)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}

	if scanner.config.PollInterval != 100*time.Millisecond {
		t.Errorf("PollInterval = %v, want %v", scanner.config.PollInterval, 100*time.Millisecond)
	}
	if scanner.config.MaxWait != 5*time.Second {
		t.Errorf("MaxWait = %v, want %v", scanner.config.MaxWait, 5*time.Second)
	}
}

func TestScanURL_InvalidURL(t *testing.T) {
	scanner, _ := NewScanner("test-api-key")
	testCases := []string{
		"ht!tp://",
		"example.com",
		"ftp://example.com", // Scheme exists but ParseRequestURI might be strict
		"",
	}

	for _, tc := range testCases {
		_, err := scanner.ScanURL(context.Background(), tc)
		if err != ErrInvalidURL {
			t.Errorf("ScanURL(%q) error = %v, want %v", tc, err, ErrInvalidURL)
		}
	}
}

func TestScanBatch_TooLarge(t *testing.T) {
	scanner, _ := NewScanner("test-api-key")
	// Create unique URLs to exceed batch limit after deduplication
	urls := make([]string, BatchScanMaxURLs+1)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://example.com/%d", i)
	}

	_, err := scanner.ScanBatch(context.Background(), urls)
	if err != ErrBatchTooLarge {
		t.Errorf("ScanBatch() error = %v, want %v", err, ErrBatchTooLarge)
	}
}

func TestScanURL_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/scans/":
			if r.Method == "POST" {
				resp := ScanResponse{
					ID:      "test-scan-id",
					URL:     "https://example.com",
					State:   ScanStateCompleted,
					Verdict: VerdictSafe,
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(resp)
				return
			}
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	scanner, err := NewScannerWithBaseURL(server.URL, "test-api-key")
	if err != nil {
		t.Fatalf("NewScannerWithBaseURL() error = %v", err)
	}

	result, err := scanner.ScanURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("ScanURL() error = %v", err)
	}

	if result.URL != "https://example.com" {
		t.Errorf("result.URL = %q, want %q", result.URL, "https://example.com")
	}
	if result.State != ScanStateCompleted {
		t.Errorf("result.State = %v, want %v", result.State, ScanStateCompleted)
	}
	if result.Verdict != VerdictSafe {
		t.Errorf("result.Verdict = %v, want %v", result.Verdict, VerdictSafe)
	}
}

func TestScanURL_WithPolling(t *testing.T) {
	pollCount := 0
	scanID := "550e8400-e29b-41d4-a716-446655440000" // Valid UUID
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/v1/scans/" && r.Method == "POST":
			resp := ScanResponse{
				ID:    scanID,
				URL:   "https://example.com",
				State: ScanStatePending,
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)

		case r.URL.Path == "/v1/scans/"+scanID && r.Method == "GET":
			pollCount++
			var resp ScanResponse
			if pollCount >= 2 {
				resp = ScanResponse{
					ID:      scanID,
					URL:     "https://example.com",
					State:   ScanStateCompleted,
					Verdict: VerdictSafe,
				}
			} else {
				resp = ScanResponse{
					ID:    scanID,
					URL:   "https://example.com",
					State: ScanStateProcessing,
				}
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	scanner, _ := NewScannerWithBaseURL(server.URL, "test-api-key",
		WithPollInterval(10*time.Millisecond),
	)

	result, err := scanner.ScanURL(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("ScanURL() error = %v", err)
	}

	if pollCount < 2 {
		t.Errorf("expected at least 2 poll requests, got %d", pollCount)
	}
	if result.State != ScanStateCompleted {
		t.Errorf("result.State = %v, want %v", result.State, ScanStateCompleted)
	}
}

func TestScanURLs_Empty(t *testing.T) {
	scanner, _ := NewScanner("test-api-key")
	results, err := scanner.ScanURLs(context.Background(), []string{})
	if err != nil {
		t.Fatalf("ScanURLs([]) error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("ScanURLs([]) returned %d results, want 0", len(results))
	}
}

func TestScanBatch_Deduplication(t *testing.T) {
	var receivedURLs []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/v1/scans/batch" && r.Method == "POST" {
			var req BatchScanRequest
			json.NewDecoder(r.Body).Decode(&req)
			receivedURLs = req.URLs

			// Return completed scans for all URLs
			resp := BatchScanResponse{
				BatchID: "test-batch-id",
				Jobs:    make([]ScanResponse, len(req.URLs)),
			}
			for i, u := range req.URLs {
				resp.Jobs[i] = ScanResponse{
					ID:      fmt.Sprintf("scan-%d", i),
					URL:     u,
					State:   ScanStateCompleted,
					Verdict: VerdictSafe,
				}
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	scanner, _ := NewScannerWithBaseURL(server.URL, "test-api-key")

	// Submit duplicate URLs
	urls := []string{
		"https://example.com/a",
		"https://example.com/b",
		"https://example.com/a", // duplicate
		"https://example.com/c",
		"https://example.com/b", // duplicate
		"https://example.com/a", // duplicate
	}

	results, err := scanner.ScanBatch(context.Background(), urls)
	if err != nil {
		t.Fatalf("ScanBatch() error = %v", err)
	}

	// Should only send 3 unique URLs to the API
	if len(receivedURLs) != 3 {
		t.Errorf("expected 3 unique URLs sent to API, got %d: %v", len(receivedURLs), receivedURLs)
	}

	// Results should contain all 3 unique URLs
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Verify all unique URLs have results
	for _, u := range []string{"https://example.com/a", "https://example.com/b", "https://example.com/c"} {
		if _, ok := results[u]; !ok {
			t.Errorf("missing result for URL: %s", u)
		}
	}
}

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
	}{
		{
			name:       "success",
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "created",
			statusCode: 201,
			wantErr:    false,
		},
		{
			name:       "not found",
			statusCode: 404,
			body:       `{"error": "not found"}`,
			wantErr:    true,
		},
		{
			name:       "server error",
			statusCode: 500,
			body:       `{"message": "internal error"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			err := checkResponse(resp, []byte(tt.body))
			if (err != nil) != tt.wantErr {
				t.Errorf("checkResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultScannerConfig(t *testing.T) {
	cfg := DefaultScannerConfig()

	if cfg.PollInterval != 500*time.Millisecond {
		t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, 500*time.Millisecond)
	}
	if cfg.MaxWait != 30*time.Second {
		t.Errorf("MaxWait = %v, want %v", cfg.MaxWait, 30*time.Second)
	}
}
