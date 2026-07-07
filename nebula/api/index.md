# API

Ogoune exposes a stable, versioned REST API under `/api/v1/`.

## OpenAPI spec

The contract is generated from Go annotations (source of truth) — see `api/openapi/v1.{yaml,json}` in the repository. At runtime the embedded spec is served at:

```
GET /api/v1/openapi.json
```

## Interactive docs

When `ENABLE_SWAGGER=true`, Swagger UI is available at:

```
/api/v1/docs/
```

## Versioning

- `/api/v1/` — stable public API, semver-protected
- `/api/` (non-versioned) — internal, may change anytime

## Full reference

The complete, always-current API reference is rendered from the OpenAPI spec:

**→ [API Reference](/api/reference)**

It is generated from `api/openapi/v1.json` via [`vitepress-openapi`](https://vitepress-openapi.vercel.app/) — regenerate the spec with `make openapi` and the docs pick it up on the next build.
