#!/usr/bin/env sh
set -eu

SPEC_URL="${SPEC_URL:-http://localhost:8081/openapi/json}"
REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SPEC_JSON="${REPO_DIR}/openapi.json"

echo "Fetching OpenAPI spec from ${SPEC_URL} ..."
if ! curl -sf -o "${SPEC_JSON}" "${SPEC_URL}"; then
  echo "Failed to fetch spec. Is the API running at ${SPEC_URL}?" >&2
  exit 1
fi

echo "Generating Go client..."
cd "${REPO_DIR}/go"
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
  -config oapi-config.yaml \
  ../openapi.json

echo "Generating TypeScript client..."
cd "${REPO_DIR}/ts"
bun x openapi-ts --input ../openapi.json --output ./src/client --client @hey-api/client-fetch

echo "Done. All SDKs updated."
