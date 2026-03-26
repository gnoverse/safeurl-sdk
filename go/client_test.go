package safeurl

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestNewClientWithAPIKey_GetHealth_SendsBearerAuthorization(t *testing.T) {
	const apiKey = "sk_live_unit_test_key"
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/health" {
			t.Errorf("request path: got %q, want %q", r.URL.Path, "/health")
		}
		if r.Method != http.MethodGet {
			t.Errorf("method: got %q, want GET", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	client, err := NewClientWithAPIKey(ts.URL, apiKey, WithHTTPClient(ts.Client()))
	if err != nil {
		t.Fatalf("NewClientWithAPIKey: %v", err)
	}
	resp, err := client.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	want := "Bearer " + apiKey
	if gotAuth != want {
		t.Fatalf("Authorization header: got %q, want %q", gotAuth, want)
	}
}

func TestNewClientWithAPIKey_AcceptsBaseURLWithTrailingSlash(t *testing.T) {
	const apiKey = "sk_live_x"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("path: got %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	base := ts.URL + "/"
	client, err := NewClientWithAPIKey(base, apiKey, WithHTTPClient(ts.Client()))
	if err != nil {
		t.Fatalf("NewClientWithAPIKey: %v", err)
	}
	resp, err := client.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

func TestNewClientWithAPIKey_DeleteV1ApiKeysById_PathEncodesUUID(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	wantPath := "/v1/api-keys/" + id.String()
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodDelete {
			t.Errorf("method: got %q", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := NewClientWithAPIKey(ts.URL, "sk_live_key", WithHTTPClient(ts.Client()))
	if err != nil {
		t.Fatalf("NewClientWithAPIKey: %v", err)
	}
	resp, err := client.DeleteV1ApiKeysById(context.Background(), openapi_types.UUID(id))
	if err != nil {
		t.Fatalf("DeleteV1ApiKeysById: %v", err)
	}
	_ = resp.Body.Close()
	if gotPath != wantPath {
		t.Fatalf("path: got %q, want %q", gotPath, wantPath)
	}
}

func TestNewClientWithAPIKey_PostV1Scans_JSONBodyAndContentType(t *testing.T) {
	var gotCT string
	var gotBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/scans/" {
			t.Errorf("path: got %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: got %q", r.Method)
		}
		gotCT = r.Header.Get("Content-Type")
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"00000000-0000-0000-0000-000000000001","state":"QUEUED"}`))
	}))
	defer ts.Close()

	client, err := NewClientWithAPIKey(ts.URL, "sk_live_key", WithHTTPClient(ts.Client()))
	if err != nil {
		t.Fatalf("NewClientWithAPIKey: %v", err)
	}
	resp, err := client.PostV1Scans(context.Background(), map[string]interface{}{
		"url": "https://example.com",
	})
	if err != nil {
		t.Fatalf("PostV1Scans: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if gotCT != "application/json" {
		t.Fatalf("Content-Type: got %q, want application/json", gotCT)
	}
	if string(gotBody) != `{"url":"https://example.com"}` {
		t.Fatalf("body: got %s", gotBody)
	}
}
