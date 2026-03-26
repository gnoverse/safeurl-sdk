package safeurl

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type apiKeyCreateResponse struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type apiKeyListItem struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes"`
	ExpiresAt  *string  `json:"expiresAt"`
	LastUsedAt *string  `json:"lastUsedAt"`
	CreatedAt  string   `json:"createdAt"`
	RevokedAt  *string  `json:"revokedAt"`
}

type createScanResponse struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type scanResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func TestSDKSmoke(t *testing.T) {
	baseURL := os.Getenv("SAFEURL_SDK_TEST_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	serviceSecret := os.Getenv("SAFEURL_SDK_TEST_SERVICE_SECRET")
	if serviceSecret == "" {
		serviceSecret = os.Getenv("SAFEURL_SERVICE_SECRET")
	}
	if serviceSecret == "" {
		t.Skip("SAFEURL_SDK_TEST_SERVICE_SECRET or SAFEURL_SERVICE_SECRET is required")
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serviceClient, err := NewClientWithAPIKey(baseURL, serviceSecret, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("create service client: %v", err)
	}

	healthResp, err := serviceClient.GetHealthWithResponse(ctx)
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	if healthResp.StatusCode() != http.StatusOK {
		t.Fatalf("GET /health returned status %d", healthResp.StatusCode())
	}

	keyName := "go-sdk-smoke-" + time.Now().UTC().Format("20060102150405.000000000")
	createResp, err := serviceClient.PostV1ApiKeysWithResponse(ctx, map[string]interface{}{
		"name":   keyName,
		"scopes": []string{"scan:read", "scan:write", "credits:read", "credits:write"},
	})
	if err != nil {
		t.Fatalf("POST /v1/api-keys: %v", err)
	}
	if createResp.StatusCode() != http.StatusCreated {
		t.Fatalf("POST /v1/api-keys returned status %d", createResp.StatusCode())
	}

	var created apiKeyCreateResponse
	if err := json.Unmarshal(createResp.Body, &created); err != nil {
		t.Fatalf("decode create api key response: %v", err)
	}
	if created.ID == "" {
		t.Fatal("create api key response missing id")
	}
	if !strings.HasPrefix(created.Key, "sk_live_") {
		t.Fatalf("create api key response returned unexpected key prefix: %q", created.Key)
	}
	if created.Name != keyName {
		t.Fatalf("create api key response returned name %q, want %q", created.Name, keyName)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		parsedID, parseErr := uuid.Parse(created.ID)
		if parseErr != nil {
			return
		}
		_, _ = serviceClient.DeleteV1ApiKeysByIdWithResponse(cleanupCtx, openapi_types.UUID(parsedID))
	})

	apiKeyClient, err := NewClientWithAPIKey(baseURL, created.Key, WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("create api key client: %v", err)
	}

	t.Run("key_lifecycle", func(t *testing.T) {
		apiKeysResp, err := serviceClient.GetV1ApiKeysWithResponse(ctx)
		if err != nil {
			t.Fatalf("GET /v1/api-keys: %v", err)
		}
		if apiKeysResp.StatusCode() != http.StatusOK {
			t.Fatalf("GET /v1/api-keys returned status %d", apiKeysResp.StatusCode())
		}

		apiKeyListResp, err := apiKeyClient.GetV1ApiKeysWithResponse(ctx)
		if err != nil {
			t.Fatalf("GET /v1/api-keys with api key: %v", err)
		}
		if apiKeyListResp.StatusCode() != http.StatusOK {
			t.Fatalf("GET /v1/api-keys with api key returned status %d", apiKeyListResp.StatusCode())
		}

		var keys []apiKeyListItem
		if err := json.Unmarshal(apiKeysResp.Body, &keys); err != nil {
			t.Fatalf("decode api key list response: %v", err)
		}

		found := false
		for _, key := range keys {
			if key.ID == created.ID && key.Name == keyName {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("created API key %s not found in list response", created.ID)
		}
	})

	t.Run("auth", func(t *testing.T) {
		missingAuthClient, err := NewClientWithResponses(baseURL, WithHTTPClient(httpClient))
		if err != nil {
			t.Fatalf("create unauthenticated client: %v", err)
		}
		missingAuthCreditsResp, err := missingAuthClient.GetV1CreditsWithResponse(ctx)
		if err != nil {
			t.Fatalf("GET /v1/credits without auth: %v", err)
		}
		if missingAuthCreditsResp.StatusCode() != http.StatusUnauthorized {
			t.Fatalf("GET /v1/credits without auth returned status %d", missingAuthCreditsResp.StatusCode())
		}

		invalidAuthClient, err := NewClientWithAPIKey(baseURL, "sk_live_invalid_api_key", WithHTTPClient(httpClient))
		if err != nil {
			t.Fatalf("create invalid api key client: %v", err)
		}
		invalidAuthCreditsResp, err := invalidAuthClient.GetV1CreditsWithResponse(ctx)
		if err != nil {
			t.Fatalf("GET /v1/credits with invalid api key: %v", err)
		}
		if invalidAuthCreditsResp.StatusCode() != http.StatusUnauthorized {
			t.Fatalf("GET /v1/credits with invalid api key returned status %d", invalidAuthCreditsResp.StatusCode())
		}

		// Revoke flow: create a key, verify it works, delete it, verify 401.
		revokeKeyName := "go-sdk-revoke-" + time.Now().UTC().Format("20060102150405.000000000")
		revokeCreateResp, err := serviceClient.PostV1ApiKeysWithResponse(ctx, map[string]interface{}{
			"name":   revokeKeyName,
			"scopes": []string{"scan:read", "credits:read"},
		})
		if err != nil {
			t.Fatalf("POST /v1/api-keys for revoke test: %v", err)
		}
		if revokeCreateResp.StatusCode() != http.StatusCreated {
			t.Fatalf("POST /v1/api-keys for revoke test returned status %d", revokeCreateResp.StatusCode())
		}

		var revoked apiKeyCreateResponse
		if err := json.Unmarshal(revokeCreateResp.Body, &revoked); err != nil {
			t.Fatalf("decode revoke test api key response: %v", err)
		}
		if revoked.ID == "" || revoked.Key == "" {
			t.Fatal("revoke test api key response missing id or key")
		}

		revokedClient, err := NewClientWithAPIKey(baseURL, revoked.Key, WithHTTPClient(httpClient))
		if err != nil {
			t.Fatalf("create revoked api key client: %v", err)
		}

		revokedCreditsResp, err := revokedClient.GetV1CreditsWithResponse(ctx)
		if err != nil {
			t.Fatalf("GET /v1/credits before revoke: %v", err)
		}
		if revokedCreditsResp.StatusCode() != http.StatusOK {
			t.Fatalf("GET /v1/credits before revoke returned status %d", revokedCreditsResp.StatusCode())
		}

		deleteResp, err := apiKeyClient.DeleteV1ApiKeysByIdWithResponse(ctx, openapi_types.UUID(uuid.MustParse(revoked.ID)))
		if err != nil {
			t.Fatalf("DELETE /v1/api-keys/:id: %v", err)
		}
		if deleteResp.StatusCode() != http.StatusOK {
			t.Fatalf("DELETE /v1/api-keys/:id returned status %d", deleteResp.StatusCode())
		}

		revokedCreditsAfterResp, err := revokedClient.GetV1CreditsWithResponse(ctx)
		if err != nil {
			t.Fatalf("GET /v1/credits after revoke: %v", err)
		}
		if revokedCreditsAfterResp.StatusCode() != http.StatusUnauthorized {
			t.Fatalf("GET /v1/credits after revoke returned status %d", revokedCreditsAfterResp.StatusCode())
		}
	})

	t.Run("credits", func(t *testing.T) {
		purchaseResp, err := serviceClient.PostV1CreditsPurchaseWithResponse(ctx, map[string]interface{}{
			"amount": 1,
		})
		if err != nil {
			t.Fatalf("POST /v1/credits/purchase: %v", err)
		}
		if purchaseResp.StatusCode() != http.StatusCreated {
			t.Fatalf("POST /v1/credits/purchase returned status %d", purchaseResp.StatusCode())
		}

		creditsResp, err := apiKeyClient.GetV1CreditsWithResponse(ctx)
		if err != nil {
			t.Fatalf("GET /v1/credits: %v", err)
		}
		if creditsResp.StatusCode() != http.StatusOK {
			t.Fatalf("GET /v1/credits returned status %d", creditsResp.StatusCode())
		}
	})

	t.Run("scan", func(t *testing.T) {
		scanURL := "https://example.com"
		createScanResp, err := apiKeyClient.PostV1ScansWithResponse(ctx, map[string]interface{}{
			"url": scanURL,
		})
		if err != nil {
			t.Fatalf("POST /v1/scans: %v", err)
		}
		if createScanResp.StatusCode() != http.StatusCreated {
			t.Fatalf("POST /v1/scans returned status %d", createScanResp.StatusCode())
		}

		var createdScan createScanResponse
		if err := json.Unmarshal(createScanResp.Body, &createdScan); err != nil {
			t.Fatalf("decode create scan response: %v", err)
		}
		if createdScan.ID == "" {
			t.Fatal("create scan response missing id")
		}
		if createdScan.State != "QUEUED" {
			t.Fatalf("create scan response returned state %q, want QUEUED", createdScan.State)
		}

		scanID, err := uuid.Parse(createdScan.ID)
		if err != nil {
			t.Fatalf("parse scan id: %v", err)
		}

		scanResp, err := apiKeyClient.GetV1ScansByIdWithResponse(ctx, openapi_types.UUID(scanID))
		if err != nil {
			t.Fatalf("GET /v1/scans/:id: %v", err)
		}
		if scanResp.StatusCode() != http.StatusOK {
			t.Fatalf("GET /v1/scans/:id returned status %d", scanResp.StatusCode())
		}

		var scan scanResponse
		if err := json.Unmarshal(scanResp.Body, &scan); err != nil {
			t.Fatalf("decode scan response: %v", err)
		}
		if scan.ID != createdScan.ID {
			t.Fatalf("scan response returned id %q, want %q", scan.ID, createdScan.ID)
		}
		if scan.URL != scanURL {
			t.Fatalf("scan response returned url %q, want %q", scan.URL, scanURL)
		}
		switch scan.State {
		case "QUEUED", "FETCHING", "ANALYZING", "COMPLETED", "FAILED", "TIMED_OUT":
		default:
			t.Fatalf("scan response returned unexpected state %q", scan.State)
		}
		if scan.CreatedAt == "" || scan.UpdatedAt == "" {
			t.Fatal("scan response missing timestamps")
		}
	})
}
