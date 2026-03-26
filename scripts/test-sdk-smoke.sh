#!/usr/bin/env sh
set -eu

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SAFEURL_REPO_DIR="${SAFEURL_REPO_DIR:-$HOME/dev/safeurl}"
SAFEURL_ENV_FILE="${SAFEURL_ENV_FILE:-${SAFEURL_REPO_DIR}/.env}"
BASE_URL="${SAFEURL_SDK_TEST_BASE_URL:-http://localhost:8081}"

if [ ! -f "${SAFEURL_ENV_FILE}" ]; then
  echo "Missing SafeURL env file: ${SAFEURL_ENV_FILE}" >&2
  echo "Set SAFEURL_REPO_DIR if your safeurl checkout lives elsewhere." >&2
  exit 1
fi

set -a
. "${SAFEURL_ENV_FILE}"
set +a

SERVICE_SECRET="${SAFEURL_SDK_TEST_SERVICE_SECRET:-${SAFEURL_SERVICE_SECRET:-}}"

if [ -z "${SERVICE_SECRET}" ]; then
  echo "SAFEURL_SDK_TEST_SERVICE_SECRET or SAFEURL_SERVICE_SECRET is required." >&2
  echo "Loaded env file: ${SAFEURL_ENV_FILE}" >&2
  echo "Expected the local SafeURL API to be running, usually at ${BASE_URL}." >&2
  exit 1
fi

export SAFEURL_SDK_TEST_BASE_URL="${BASE_URL}"
export SAFEURL_SDK_TEST_SERVICE_SECRET="${SERVICE_SECRET}"

echo "Loaded SafeURL env from ${SAFEURL_ENV_FILE}"
echo "Running Go SDK smoke test against ${SAFEURL_SDK_TEST_BASE_URL} ..."
(
  cd "${REPO_DIR}/go"
  GOCACHE="${GOCACHE:-/tmp/go-build}" go test ./...
)

echo "Running TypeScript SDK smoke test against ${SAFEURL_SDK_TEST_BASE_URL} ..."
(
  cd "${REPO_DIR}/ts"
  bun test tests/sdk-smoke.test.ts
)

echo "SDK smoke tests passed."
