# ogoune Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-12

## Active Technologies

- Go 1.25.1 + GORM v1.31.0 (model hooks), `pkg/crypto` (existing AES-256-GCM — extends with `KeyProvider`) (026-credential-encryption)
- SQLite (community/dev) via `glebarez/sqlite`; PostgreSQL (hosted/production) via `gorm.io/driver/postgres` — encryption storage-agnostic (026-credential-encryption)

- Go 1.25.1 (backend), Vue 3 + TypeScript (frontend) + GORM, glebarez/sqlite, gorm.io/driver/postgres, asynq (task queue), testify (tests), Vite/Vitest (frontend) (023-keyword-monitor)

## Project Structure

```text
src/
tests/
```

## Commands

npm test && npm run lint

## Code Style

Go 1.25.1 (backend), Vue 3 + TypeScript (frontend): Follow standard conventions

## Recent Changes
- 026-credential-encryption: Added Go 1.25.1 + GORM v1.31.0 (model hooks), `pkg/crypto` (existing AES-256-GCM — extends with `KeyProvider`)


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->