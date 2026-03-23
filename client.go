package safeurl

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config oapi-config.yaml openapi.json

import (
	"context"
	"net/http"
	"strings"
)

// DefaultBaseURL is the default SafeURL API base URL.
const DefaultBaseURL = "https://api.safeurl.ai"

// NewClientWithAPIKey returns a client that sends the given API key as Bearer token on every request.
// server is the base URL (e.g. DefaultBaseURL); it will get a trailing slash if missing.
// Use the returned *ClientWithResponses for typed methods (PostV1ScansWithResponse, GetV1ScansByIdWithResponse, etc.).
func NewClientWithAPIKey(server, apiKey string, opts ...ClientOption) (*ClientWithResponses, error) {
	server = strings.TrimSuffix(server, "/")
	opts = append([]ClientOption{
		WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+apiKey)
			return nil
		}),
	}, opts...)
	return NewClientWithResponses(server, opts...)
}
