# pulseguard Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-24

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
- 006-ssl-domain-expiry-alerts: Added Go 1.22 (backend), TypeScript / Vue 3 (frontend) + Chi (HTTP router), GORM (ORM), Asynq (task queue / PostgreSQL mode), TimingWheel (task scheduler / SQLite community mode), Ant Design Vue (UI components)
- 005-confirmation-window: Added Go 1.25.x (backend), TypeScript 5.x + Vue 3 (frontend) + Chi router, GORM, Asynq, in-process scheduler (timingwheel), Pinia, Axios helper
- 004-activate-dns-monitor: Added Go 1.21+ (matching backend)


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
