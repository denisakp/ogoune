# pulseguard Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-26

## Active Technologies
- Go 1.25.x (backend), TypeScript/Vue 3 (frontend unchanged) + Chi router, GORM, Asynq (`github.com/hibiken/asynq`), Redis (Asynq mode only), existing notifier package (`pkg/notifier`) (003-in-process-scheduler)
- SQLite or PostgreSQL via existing DB runtime (`DB_DRIVER`) (003-in-process-scheduler)
- Go 1.25.x (backend), TypeScript 5.x/Vue 3 (frontend unchanged) + Chi router, GORM, Asynq (hosted path), Redis (hosted compatibility lane), notifier package (003-in-process-scheduler)
- SQLite and PostgreSQL via current DB driver abstraction (003-in-process-scheduler)
- Go 1.21+ (matching backend) (004-activate-dns-monitor)
- PostgreSQL + SQLite (existing, no changes required) (004-activate-dns-monitor)
- Go 1.25.x (backend), TypeScript 5.x + Vue 3 (frontend) + Chi router, GORM, Asynq, in-process scheduler (timingwheel), Pinia, Axios helper (005-confirmation-window)
- PostgreSQL and SQLite (dual support required) (005-confirmation-window)
- Go 1.22 (backend), TypeScript / Vue 3 (frontend) + Chi (HTTP router), GORM (ORM), Asynq (task queue / PostgreSQL mode), TimingWheel (task scheduler / SQLite community mode), Ant Design Vue (UI components) (006-ssl-domain-expiry-alerts)
- PostgreSQL (hosted/default) + SQLite (community edition) — dual migration files required for every schema change (006-ssl-domain-expiry-alerts)
- Go 1.22 (backend), TypeScript / Vue 3 (frontend) + GORM, Asynq, Chi, Ant Design Vue, Pinia (007-intelligent-alerting)
- PostgreSQL (primary) + SQLite (dev/CE default) via dual migration track (007-intelligent-alerting)
- Go 1.22 (backend), TypeScript / Vue 3.4 (frontend) + Chi v5, GORM, Asynq (backend); Vue 3, Pinia, Ant Design Vue, Vite (frontend) (008-live-monitor-refresh)
- PostgreSQL (primary/hosted) + SQLite (community/self-hosted) — raw SQL migrations (008-live-monitor-refresh)
- TypeScript 5, Vue 3.4 + Ant Design Vue 4 (`a-form-item`, `a-select`, `a-tag`), Vite 5 (009-ui-notification-cleanup)
- Go 1.25.1 (backend), TypeScript 5 + Vue 3.5 (frontend) + Chi router, GORM repositories, Asynq worker path, Ant Design Vue settings UI (010-incident-diagnostic-fixes)
- PostgreSQL/SQLite via existing GORM models (no schema migration expected) (010-incident-diagnostic-fixes)
- Go 1.25 (backend) / TypeScript 5 + Vue 3.5 (frontend) + Chi v5 (router), GORM (ORM), `crypt# Implementation Plan: API Keys Managemen (011-api-key-management)
- Go 1.23.x (backend monolith), TypeScript 5.x (frontend unaffected for this feature) + Chi router, GORM, Asynq scheduler adapter, Testify, fake repositories under `backend/internal/repository/fake` (012-fix-confirmation-window)
- PostgreSQL (primary), SQLite (community/local mode), existing resource and incident tables (012-fix-confirmation-window)

- Go 1.25.1 (backend), TypeScript/Vue 3 (frontend unaffected for this feature) + `gorm.io/gorm`, `gorm.io/driver/postgres`, `github.com/glebarez/sqlite` (new), `github.com/stretchr/testify` (002-add-db-driver-abstraction)

## Project Structure

```text
backend/
frontend/
tests/
```

## Commands

npm test && npm run lint

## Code Style

Go 1.25.1 (backend), TypeScript/Vue 3 (frontend unaffected for this feature): Follow standard conventions

## Recent Changes
- 012-fix-confirmation-window: Added Go 1.23.x (backend monolith), TypeScript 5.x (frontend unaffected for this feature) + Chi router, GORM, Asynq scheduler adapter, Testify, fake repositories under `backend/internal/repository/fake`
- 011-api-key-management: Added Go 1.25 (backend) / TypeScript 5 + Vue 3.5 (frontend) + Chi v5 (router), GORM (ORM), `crypt# Implementation Plan: API Keys Managemen
- 010-incident-diagnostic-fixes: Added Go 1.25.1 (backend), TypeScript 5 + Vue 3.5 (frontend) + Chi router, GORM repositories, Asynq worker path, Ant Design Vue settings UI


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
