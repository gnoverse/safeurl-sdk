#!/usr/bin/env sh
set -eu

SPEC_URL="${SPEC_URL:-http://localhost:8081/openapi/json}"
SDK_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SPEC_JSON="${SDK_DIR}/openapi.json"

echo "Fetching OpenAPI spec from ${SPEC_URL} ..."
if ! curl -sf -o "${SPEC_JSON}" "${SPEC_URL}"; then
  echo "Failed to fetch spec. Is the API running at ${SPEC_URL}?" >&2
  exit 1
fi

echo "Generating TypeScript client..."
cd "${SDK_DIR}"
bun x openapi-ts --input ./openapi.json --output ./src/client --client @hey-api/client-fetch

echo "Done. src/client updated."
