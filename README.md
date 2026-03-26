# SafeURL SDKs

Official SDKs for the [SafeURL API](https://safeurl.ai), an AI-powered URL safety screening service.

This repository contains:

- [**Go SDK**](./go) — Native Go client generated from OpenAPI.
- [**TypeScript SDK**](./ts) — Type-safe TypeScript client generated from OpenAPI using Hey API.

## Repository Structure

```
safeurl-sdk/
├── go/    # Go SDK source and generation scripts
└── ts/    # TypeScript SDK source and generation scripts
```

## Installation

### Go

```bash
go get github.com/gnoverse/safeurl-sdk/go
```

### TypeScript

```bash
npm install @safeurl/sdk
```

## Releases

This repo uses [release-please](https://github.com/googleapis/release-please) with [`release-please-config.json`](./release-please-config.json). Merge the release pull request on `main` to tag releases and publish changelogs for the TypeScript (`ts/`) and Go (`go/`) packages.

Publishing `@safeurl/sdk` to npm from CI requires a repository secret **`NPM_TOKEN`** (automation token with publish access to the `@safeurl` scope). Without it, the release still completes on GitHub; the npm publish step fails until the secret is configured.

## Development

These SDKs are generated from the SafeURL OpenAPI specification.
The fetched spec is stored once at the repository root as `openapi.json`.

### Regenerating All SDKs

From the root of this repository:

```bash
sh scripts/fetch-openapi.sh
```

### Running SDK Smoke Tests

Run both SDK smoke tests against a local SafeURL API:

```bash
sh scripts/test-sdk-smoke.sh
```

The script expects the local SafeURL stack to be running and reads the shared
secret from `~/dev/safeurl/.env` by default, or from `SAFEURL_ENV_FILE` /
`SAFEURL_REPO_DIR` if you need to point at a different checkout.
You can still override the secret with `SAFEURL_SDK_TEST_SERVICE_SECRET` or
`SAFEURL_SERVICE_SECRET`. `SAFEURL_SDK_TEST_BASE_URL` defaults to
`http://localhost:8081`.

## License

Apache-2.0
