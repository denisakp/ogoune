# ogoune Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-12

## Active Technologies
- Go 1.25.1 + Chi v5 (router), GORM (ORM), prometheus/client_golang (new — not yet in go.mod), testify v1.8.1 (tests) (025-prometheus-metrics-endpoint)
- SQLite (community/dev) and PostgreSQL (hosted/production) — both supported equally (025-prometheus-metrics-endpoint)

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
- 025-prometheus-metrics-endpoint: Added Go 1.25.1 + Chi v5 (router), GORM (ORM), prometheus/client_golang (new — not yet in go.mod), testify v1.8.1 (tests)

- 023-keyword-monitor: Added Go 1.25.1 (backend), Vue 3 + TypeScript (frontend) + GORM, glebarez/sqlite, gorm.io/driver/postgres, asynq (task queue), testify (tests), Vite/Vitest (frontend)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
