#!/usr/bin/env sh
# Fetches the OpenAPI spec (JSON) from the running SafeURL API (Scalar at http://localhost:8081/openapi).
# Saves as openapi.json and runs oapi-codegen. Ensure the API is running.
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SDK_DIR="$(dirname "$SCRIPT_DIR")"
SPEC_URL="${SPEC_URL:-http://localhost:8081/openapi/json}"
SPEC_JSON="${SDK_DIR}/openapi.json"

echo "Fetching OpenAPI spec from ${SPEC_URL} ..."
if ! curl -sf -o "${SPEC_JSON}" "${SPEC_URL}"; then
  echo "Failed to fetch spec. Is the API running at ${SPEC_URL}?" >&2
  exit 1
fi

echo "Generating Go client..."
cd "${SDK_DIR}"
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
  -config oapi-config.yaml \
  openapi.json

echo "Done. client.gen.go updated."
