# SafeURL Go SDK

Go client for the [SafeURL API](https://api.safeurl.ai), for URL safety screening. Designed for server-side use (e.g. [gno.land](https://gno.land) and other Go backends).

## Installation

```bash
go get github.com/gnoverse/safeurl-sdk/go
```

**Using from a local checkout:** In your Go project's `go.mod`:

```go
replace github.com/gnoverse/safeurl-sdk/go => /path/to/safeurl-sdk/go
```

## Quick start

The SDK provides two levels of abstraction:

### High-level Scanner (Recommended)

The `Scanner` provides convenience methods with automatic polling and typed responses:

```go
package main

import (
	"context"
	"fmt"
	"log"

	safeurl "github.com/gnoverse/safeurl-sdk/go"
)

func main() {
	// Create a scanner with your API key
	scanner, err := safeurl.NewScanner("sk_live_your_api_key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Scan a single URL - waits for completion automatically
	result, err := scanner.ScanURL(ctx, "https://example.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("URL: %s\n", result.URL)
	fmt.Printf("Verdict: %s (safe: %v)\n", result.Verdict, result.Verdict.IsSafe())
	fmt.Printf("State: %s\n", result.State)

	// Scan multiple URLs - automatically chunks if > 50 URLs
	results, err := scanner.ScanURLs(ctx, []string{
		"https://example.com",
		"https://google.com",
		"https://suspicious-site.xyz",
	})
	if err != nil {
		log.Fatal(err)
	}

	for url, scan := range results {
		fmt.Printf("%s -> %s\n", url, scan.Verdict)
	}

	// Or use ScanBatch for a single batch (max 50, returns ErrBatchTooLarge if exceeded)
	batch, err := scanner.ScanBatch(ctx, []string{"https://a.com", "https://b.com"})
}
```

### One-liner with QuickScan

For simple use cases:

```go
result, err := safeurl.QuickScan(ctx, "sk_live_your_api_key", "https://example.com")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Verdict: %s\n", result.Verdict)
```

### Low-level Client

For advanced use cases, use the generated client directly:

```go
import (
	"bytes"
	"encoding/json"
)

client, err := safeurl.NewClientWithAPIKey(safeurl.DefaultBaseURL, "sk_live_your_api_key")
if err != nil {
	log.Fatal(err)
}

// The generated client uses PostV1ScansWithBodyWithResponse for custom JSON
body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
resp, err := client.PostV1ScansWithBodyWithResponse(ctx, "application/json", bytes.NewReader(body))
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Status: %d, Body: %s\n", resp.StatusCode(), string(resp.Body))
// Handle polling manually by calling GetV1ScansByIdWithResponse...
```

## Scanner Configuration

```go
scanner, err := safeurl.NewScanner(
	"sk_live_...",
	safeurl.WithPollInterval(200 * time.Millisecond), // How often to poll for completion
	safeurl.WithMaxWait(60 * time.Second),            // Maximum time to wait for scan
)
```

Or with a custom base URL:

```go
scanner, err := safeurl.NewScannerWithBaseURL(
	"https://custom-api.example.com",
	"sk_live_...",
)
```

## Types

### Verdict

The safety verdict for a scanned URL:

```go
safeurl.VerdictSafe       // URL is safe
safeurl.VerdictMalicious  // URL is malicious
safeurl.VerdictSuspect    // URL is suspicious
safeurl.VerdictUnknown    // Could not determine safety

// Helper methods
verdict.IsSafe()   // true if VerdictSafe
verdict.IsUnsafe() // true if VerdictMalicious or VerdictSuspect
```

### ScanState

The processing state of a scan:

```go
safeurl.ScanStatePending     // Scan is queued
safeurl.ScanStateProcessing  // Scan is in progress
safeurl.ScanStateCompleted   // Scan finished successfully
safeurl.ScanStateFailed      // Scan failed

// Helper method
state.IsTerminal() // true if Completed or Failed
```

### ScanResponse

The result returned from scans:

```go
type ScanResponse struct {
	ID        string
	URL       string
	State     ScanState
	Verdict   Verdict
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt *time.Time
}

// Helper method
scan.IsComplete() // true if State is terminal
```

## Error Handling

The SDK provides typed errors for common scenarios:

```go
result, err := scanner.ScanURL(ctx, url)
if err != nil {
	switch {
	case errors.Is(err, safeurl.ErrUnauthorized):
		// Invalid API key
	case errors.Is(err, safeurl.ErrTimeout):
		// Scan took too long
	case errors.Is(err, safeurl.ErrInvalidURL):
		// URL was empty or invalid
	case errors.Is(err, safeurl.ErrScanFailed):
		// Scan failed on the server
	default:
		// Check for API errors
		var apiErr *safeurl.APIError
		if errors.As(err, &apiErr) {
			if apiErr.IsRateLimited() {
				// Handle rate limiting
			}
		}
	}
}
```

### Sentinel Errors

- `ErrUnauthorized` - Invalid or missing API key
- `ErrNotFound` - Scan ID not found
- `ErrRateLimited` - API rate limit exceeded
- `ErrTimeout` - Scan timed out waiting for completion
- `ErrBatchTooLarge` - Batch exceeds 50 URLs
- `ErrInvalidURL` - URL is empty or invalid
- `ErrScanFailed` - Scan failed to complete

### APIError

For HTTP errors from the API:

```go
apiErr.StatusCode     // HTTP status code
apiErr.Message        // Error message
apiErr.Code           // Error code (if provided)

apiErr.IsUnauthorized() // 401
apiErr.IsNotFound()     // 404
apiErr.IsRateLimited()  // 429
apiErr.IsServerError()  // 5xx
```

## Constants

- `safeurl.BatchScanMaxURLs` — maximum URLs per batch (50)
- `safeurl.DefaultBaseURL` — default API base URL

## Low-level API Coverage

| Method                                | Endpoint                       | Description          |
| ------------------------------------- | ------------------------------ | -------------------- |
| `PostV1ScansWithResponse`             | `POST /v1/scans`               | Create a single scan |
| `PostV1ScansBatchWithResponse`        | `POST /v1/scans/batch`         | Create batch scan    |
| `GetV1ScansWithResponse`              | `GET /v1/scans`                | List scans           |
| `GetV1ScansByIdWithResponse`          | `GET /v1/scans/{id}`           | Get scan by ID       |
| `GetV1ScansByIdAnalyticsWithResponse` | `GET /v1/scans/{id}/analytics` | Scan analytics       |
| `GetV1CreditsWithResponse`            | `GET /v1/credits`              | Credit balance       |
| `PostV1CreditsPurchaseWithResponse`   | `POST /v1/credits/purchase`    | Purchase credits     |
| `GetHealthWithResponse`               | `GET /health`                  | Health check         |

## Regenerating the Client

The low-level client is generated from OpenAPI using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

**From running API:**

```bash
sh scripts/fetch-openapi.sh
```

**With existing `openapi.json`:**

```bash
go generate ./...
```

## License

Same as the SafeURL project.
