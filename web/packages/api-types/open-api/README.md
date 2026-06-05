# OpenAPI Spec

The backend OpenAPI spec is placed here before running type generation:

```bash
cp ../meyebi-api/api/openapi/openapi.yaml packages/api-types/openapi/openapi.yaml
pnpm generate:types
```

The spec file (`openapi.yaml`) is **not committed** — it must be copied from the backend
repository before generating types. The generated output (`generated/schema.d.ts`) IS
committed so developers can work without the spec file locally.

If you get `Error: Cannot find module '@meyebi/api-types'`, run `pnpm generate:types`.
