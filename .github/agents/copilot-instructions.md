# ogoune Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-18

## Active Technologies
- Go 1.25.x (backend at repository root: `./`) and TypeScript 5 + Vue 3 (frontend in `./web`).
- Backend: Chi v5 router, GORM, Asynq, optional Redis (Asynq mode), notifier package in `pkg/notifier`.
- Frontend (`./web`): Vite, Pinia, Axios helper, Ant Design Vue.
- Monitoring supports HTTP, TCP, DNS, and optional ICMP (`ENABLE_ICMP`).
- Datastores: PostgreSQL and SQLite, with dual SQL migrations under `internal/database/migrations/{postgres,sqlite}`.
- Go 1.21+, Vue 3 + TypeScrip + Chi (HTTP router), GORM (persistence), Asynq or TimingWheel (scheduler), existing incident/notification services (020-heartbeat-monitoring)
- PostgreSQL (primary) or SQLite (Community mode); new columns on `monitors` table (020-heartbeat-monitoring)
- Go 1.21+, Vue 3 + TypeScrip + Chi router, GORM, Asynq/TimingWheel scheduler, existing incident and notification services (020-heartbeat-monitoring)
- PostgreSQL (primary) and SQLite (community mode), heartbeat fields persisted in existing resources persistence layer (020-heartbeat-monitoring)
- Go 1.25.1 + Chi v5.2.3 (router), GORM v1.31.0 (ORM), swaggo/swag (new — dev tool, not runtime), swaggo/http-swagger (new — runtime, gated by ENABLE_SWAGGER) (027-public-api-v1)
- SQLite (community) / PostgreSQL (hosted) — no schema changes (027-public-api-v1)

## Project Structure

```text
./                  # backend root (cmd, internal, pkg, configs, docs)
web/                # frontend app
internal/           # backend domain/services/api/worker/database
cmd/                # backend entrypoints
pkg/                # shared backend packages
specs/              # feature specs/tasks/verification
```

## Commands

go test -race -timeout 120s ./...
go build ./...
cd web && pnpm test
cd web && pnpm build

## Code Style

Go (backend): standard gofmt/go test conventions and existing service-repository-handler layering.
TypeScript/Vue (web frontend): existing composable/store/service separation and Ant Design Vue patterns.

## Recent Changes
- 027-public-api-v1: Added Go 1.25.1 + Chi v5.2.3 (router), GORM v1.31.0 (ORM), swaggo/swag (new — dev tool, not runtime), swaggo/http-swagger (new — runtime, gated by ENABLE_SWAGGER)
- 020-heartbeat-monitoring: Added Go 1.21+, Vue 3 + TypeScrip + Chi router, GORM, Asynq/TimingWheel scheduler, existing incident and notification services
- 020-heartbeat-monitoring: Added Go 1.21+, Vue 3 + TypeScrip + Chi (HTTP router), GORM (persistence), Asynq or TimingWheel (scheduler), existing incident/notification services


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
