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

## Development

These SDKs are generated from the SafeURL OpenAPI specification.

### Regenerating Go SDK

```bash
cd go
sh scripts/fetch-openapi.sh
```

### Regenerating TypeScript SDK

```bash
cd ts
sh scripts/fetch-openapi.sh
```

## License

Apache-2.0
