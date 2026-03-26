# SafeURL SDK Monorepo

SDKs for the [SafeURL API](https://safeurl.ai) — AI-powered URL safety screening.

## Packages

| Package | Description |
|---------|-------------|
| `go/` | Go client generated from OpenAPI via oapi-codegen |
| `ts/` | TypeScript client generated from OpenAPI via Hey API |

## Commands

```sh
just gen    # Fetch OpenAPI spec and regenerate both SDKs
just smoke  # Run Go and TypeScript smoke tests against a local API
```

Go:
```sh
cd go && go test ./...                                        # unit tests
cd go && go test -run TestSDKSmoke -count=1 ./...            # smoke only
```

TypeScript:
```sh
cd ts && bun test                                            # unit tests
cd ts && bun test tests/sdk-smoke.test.ts                   # smoke only
```

## Structure

```
safeurl-sdk/
├── go/
│   ├── client.gen.go    # Generated — do not edit by hand
│   ├── client.go        # Handwritten helpers (NewClientWithAPIKey, etc.)
│   ├── client_test.go   # Unit tests
│   └── smoke_test.go    # Integration smoke tests
├── ts/
│   ├── src/client/      # Generated — do not edit by hand
│   ├── src/index.ts     # Barrel export
│   └── tests/           # Unit and smoke tests
├── scripts/
│   ├── fetch-openapi.sh # Fetch spec + regenerate both SDKs
│   └── test-sdk-smoke.sh
├── openapi.json         # Cached OpenAPI spec (updated by just gen)
└── justfile
```

## Patterns

**Generated files:** Never edit `go/client.gen.go` or anything under `ts/src/client/` by hand. Always regenerate via `just gen`.

**TypeScript:** Use Bun exclusively — `bun test`, `bunx`, `bun install`. Never npm/yarn/pnpm/npx.

**TS imports:** Relative imports with `.js` extension in source files.

## Commits

Conventional: `type(scope): description`
Types: `feat`, `fix`, `refactor`, `test`, `chore`, `docs`, `ci`
Scopes: `go`, `ts`, `smoke`, `ci` — or omit for root-level changes

## Releases

Managed by release-please. Merging the release PR on `main` tags releases and publishes:
- `@safeurl/sdk` to npm (requires `NPM_TOKEN` repo secret)
- Go module tag on GitHub
