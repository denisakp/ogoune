# Copilot Instructions for Ogoune

Purpose: help AI coding agents be productive immediately in this monorepo. Keep changes aligned with the existing layering, data flow, and tooling.

## Big picture

- Monorepo with two apps:
  - `backend/` Go API + background worker + scheduler (Asynq/Redis) + Postgres via GORM.
  - `frontend/` Vue 3 + TS (Vite), Axios service layer, Pinia stores/composables, Ant Design Vue UI.
- Runtime flow:
  1. API writes/reads via repositories; scheduling is delegated to a Scheduler service (no checks in handlers).
  2. Scheduler enqueues monitoring tasks in Redis (Asynq).
  3. Worker executes checks using strategies, persists activities, drives incident lifecycle, sends notifications.
  4. Frontend consumes JSON endpoints, never hits DB directly.

## How to run (local)

- Backend (requires Docker for Postgres + Redis):
  - From repo root: `make docker-up` then `make run` (runs API+worker). Health: `GET /health`.
  - Env vars: see `backend/.env.example` (PORT, DATABASE*URL, REDIS_URL*\*).
  - Tests: `make test` (alias for `go test -v ./...`). Build: `make build`.
- Frontend:
  - `cd frontend && pnpm install && pnpm dev` with `VITE_API_BASE_URL=http://localhost:8080`.

## Backend architecture and conventions

- Entry point: `backend/cmd/api/main.go`
  - Bootstraps DB, repositories, Asynq (client/inspector/scheduler), worker, HTTP server.
  - Registers monitoring strategies: HTTP/TCP in `internal/monitoring/strategy/*` via a map in `main.go`.
- HTTP layer:
  - Router: `internal/api/router.go` (Chi). JSON-only; CORS enabled; sets `Content-Type: application/json`.
  - Handlers in `internal/api/handler/*` call services; handlers do not perform DB queries or checks directly.
- Services layer (`internal/service/*`): orchestrates repositories + scheduler, applies domain validation.
  - Example: `ResourceService` schedules/unschedules via `repository.Scheduler`; uses `ErrValidationFailed`, `ErrResourceNotFound` (see `internal/service/errors.go`).
- Repositories (`internal/repository/*`): interfaces in `interfaces.go`, Postgres impls under `postgres/`.
- Monitoring runtime (`internal/worker/*`, `internal/monitoring/*`):
  - Worker `Processor` consumes `monitoring:check` tasks.
  - `MonitoringTaskHandler` executes checks via `domain.CheckExecutor` and updates resource status.
  - Incident rules: incident is created on the 3rd consecutive DOWN; resolving triggers when UP after DOWN.
  - Notifications (two layers): default SMTP (if enabled) + user integrations (Slack/Webhooks) via `pkg/notifier/*` and `NotifierFactory`.

## Frontend architecture and conventions

- Do not call Axios from components.
  - HTTP in `src/services/*` using `src/libs/axios.helper.ts` (base URL from `VITE_API_BASE_URL`).
  - State via Pinia stores and thin composables (e.g., `src/composables/useResources.ts` wraps `src/stores/resourceStore.ts`).
- Routing: `src/router/index.ts` (Monitors, Incidents, Activities, Settings routes).
- Types: centralised in `src/types/index.ts`; keep service return types and component props aligned.

## When adding features

- New API endpoint:
  1. Define service method in `internal/service/*` using repository interfaces.
  2. Add handler in `internal/api/handler/*` and route in `internal/api/router.go`.
  3. If scheduling/monitoring-related, go through `repository.Scheduler` instead of running checks.
- New monitor type:
  - Implement `domain.CheckStrategy` in `internal/monitoring/strategy/`, register in `main.go` strategies map.
- Frontend page/feature:
  - Add service in `src/services/`, types in `src/types/`, composable/store in `src/composables/` or `src/stores/`, route in `src/router/index.ts`, and a view in `src/views/`.

## Gotchas and project-specific patterns

- IDs are ULIDs (set in `domain.Base.BeforeCreate`). GORM models live in `internal/domain/*`.
- Repository errors: map `repository.ErrNotFound` to service-level errors used by handlers.
- All API responses are JSON; keep CORS headers intact; avoid blocking on scheduler failures (log, return domain error like `ErrSchedulerSync`).
- Incident event steps (`detected`, `resource_down_alert`, `resolved`, `resource_up_alert`) are persisted; don’t assume steps are always present.
- Frontend uses Ant Design Vue; components should stay presentational and delegate logic to services/stores.

## Key files to reference

- Backend: `cmd/api/main.go`, `internal/api/router.go`, `internal/service/resource_service.go`, `internal/worker/handler_monitoring.go`, `internal/monitoring/incident_service.go`, `internal/repository/interfaces.go`, `internal/domain/models.go`.
- Frontend: `src/libs/axios.helper.ts`, `src/services/resourceService.ts`, `src/composables/useResources.ts`, `src/router/index.ts`, `src/types/index.ts`, `src/App.vue`.
