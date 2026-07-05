# Bulk resource import (YAML manifest)

Declare monitored resources in a YAML manifest and import them in bulk (spec 078).

- **Schema**: [`manifest.schema.json`](./manifest.schema.json) — validate your manifest against this before importing.
- **Example**: [`example-manifest.yaml`](./example-manifest.yaml) — one row per resource type.

## Endpoints (v1)

- `POST /api/v1/monitors/import` — bulk import. `?dryRun=true` validates only (writes nothing). `?duplicatePolicy=skip|error` (default `skip`). Body: raw `text/yaml` or a multipart file field `manifest`. Write scope required.
- `GET /api/v1/monitors/export` — export all current resources as a round-trippable manifest. Read scope.

## Rules

- Tags and components are referenced **by name** and auto-created if missing.
- Notification channels are referenced **by name** and **must pre-exist** (they hold secrets and are never created from a manifest). A missing channel is a row error.
- Parsing is **strict**: an unknown field is a row error (catches typos).
- Duplicate detection is exact, case-sensitive, and global. Under `skip` the row is skipped; under `error` the whole (all-or-nothing) import is rejected.
- Import is **all-or-nothing**: if any row is invalid, nothing is written — fix and re-import. Always dry-run first.
- Maximum **500** resources per manifest.

See `specs/078-bulk-resource-import/quickstart.md` for curl examples.
