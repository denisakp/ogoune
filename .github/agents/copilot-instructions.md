# pulseguard Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-23

## Active Technologies
- Go 1.25.x (backend), TypeScript/Vue 3 (frontend unchanged) + Chi router, GORM, Asynq (`github.com/hibiken/asynq`), Redis (Asynq mode only), existing notifier package (`pkg/notifier`) (003-in-process-scheduler)
- SQLite or PostgreSQL via existing DB runtime (`DB_DRIVER`) (003-in-process-scheduler)
- Go 1.25.x (backend), TypeScript 5.x/Vue 3 (frontend unchanged) + Chi router, GORM, Asynq (hosted path), Redis (hosted compatibility lane), notifier package (003-in-process-scheduler)
- SQLite and PostgreSQL via current DB driver abstraction (003-in-process-scheduler)

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
- 003-in-process-scheduler: Added Go 1.25.x (backend), TypeScript 5.x/Vue 3 (frontend unchanged) + Chi router, GORM, Asynq (hosted path), Redis (hosted compatibility lane), notifier package
- 003-in-process-scheduler: Added Go 1.25.x (backend), TypeScript/Vue 3 (frontend unchanged) + Chi router, GORM, Asynq (`github.com/hibiken/asynq`), Redis (Asynq mode only), existing notifier package (`pkg/notifier`)

- 002-add-db-driver-abstraction: Added Go 1.25.1 (backend), TypeScript/Vue 3 (frontend unaffected for this feature) + `gorm.io/gorm`, `gorm.io/driver/postgres`, `github.com/glebarez/sqlite` (new), `github.com/stretchr/testify`

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
