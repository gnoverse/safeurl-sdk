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
npm install @gnoverse/safeurl-ts
```

## Development

These SDKs are generated from the SafeURL OpenAPI specification.
The fetched spec is stored once at the repository root as `openapi.json`.

### Regenerating All SDKs

From the root of this repository:

```bash
sh scripts/fetch-openapi.sh
```

## License

Apache-2.0
