# Generate Go and TypeScript clients from the OpenAPI spec
gen:
    sh scripts/fetch-openapi.sh

# Run Go and TypeScript SDK smoke tests against a local SafeURL API
smoke:
    sh scripts/test-sdk-smoke.sh
