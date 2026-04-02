# ogoune Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-30

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
- 020-heartbeat-monitoring: Added Go 1.21+, Vue 3 + TypeScrip + Chi router, GORM, Asynq/TimingWheel scheduler, existing incident and notification services
- 020-heartbeat-monitoring: Added Go 1.21+, Vue 3 + TypeScrip + Chi (HTTP router), GORM (persistence), Asynq or TimingWheel (scheduler), existing incident/notification services
- feat/019-ping-icmp-check: Added Go 1.25.1 (backend), TypeScript + Vue 3 (web UI) + `golang.org/x/net/icmp`, existing scheduler/monitoring services, existing Vue + Ant Design Vue stack


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
