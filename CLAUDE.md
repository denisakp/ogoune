# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Ogoune

Uptime monitoring app that confirms failures before alerting (N consecutive failures required). Open-core model: Community Edition (AGPL v3, SQLite + TimingWheel) and Enterprise Edition (Postgres + Redis/Asynq).

## Commands

```bash
# Build
make build              # frontend + backend
make build-be           # Go binary → dist/ogoune
make build-fe           # Vue SPA → web/dist

# Test
make test               # both
make test-be            # go test -race ./...
make test-fe            # cd web && pnpm test
go test -race ./internal/scheduler/...  # single package

# Lint
make lint               # go vet + pnpm lint

# Run locally (community mode, zero deps)
cp .env.example .env    # set APP_SECRET_KEY (openssl rand -hex 32)
DB_DRIVER=sqlite SQLITE_PATH=./ogoune.db SCHEDULER_MODE=timingwheel go run ./cmd/api

# Run locally (full stack)
docker compose up -d    # Postgres + Redis
go run ./cmd/api

# Frontend dev
cd web && pnpm install && pnpm dev  # http://localhost:5173, needs VITE_API_BASE_URL

# Swagger
make swag               # regenerate docs/ from annotations, commit the result
                        # UI at /api/v1/docs/* (ENABLE_SWAGGER=true)

# Docker
make docker             # builds ogoune:test image
```

## Architecture

### Runtime flow

```
HTTP Router (Chi) → Handlers → Services → Repositories → DB (SQLite or Postgres)
                                  ↓
                            Scheduler (TimingWheel or Asynq)
                                  ↓
                            Worker pool → Check Strategies → Incident lifecycle → Notifications
```

Handlers never run checks or query DB directly. Scheduling goes through the Scheduler service. Workers execute checks via strategies, persist results, manage incidents, and dispatch notifications.

### Two runtime modes

| | Community | Production |
|---|---|---|
| DB | SQLite (in-process) | PostgreSQL |
| Scheduler | TimingWheel (in-process) | Asynq (Redis) |
| Scaling | Single binary | Stateless API + external workers |

### Key layers

- **Entry point**: `cmd/api/main.go` — bootstraps everything, wires dependencies
- **HTTP**: `internal/api/router.go` (Chi router), handlers in `internal/api/handler/`
- **Domain models**: `internal/domain/models.go` — source of truth, IDs are ULIDs (set in `BeforeCreate` GORM hook)
- **Services**: `internal/service/` — business logic, domain-level errors (not HTTP errors)
- **Repositories**: `internal/repository/interfaces.go` (contracts), implementations in `internal/repository/store/database/`
- **Scheduler**: `internal/scheduler/` — `Scheduler` interface with TimingWheel and Asynq implementations
- **Workers**: `internal/worker/` — `handler_monitoring.go` (check execution + incident triggering), `handler_expiry.go`, `handler_notification.go`
- **Check strategies**: `internal/monitoring/strategy/` — HTTP, TCP, DNS, ICMP, Keyword, Protocol
- **Incident logic**: `internal/monitoring/incident_service.go` — confirmation window, flap detection, alert grouping
- **Notifications**: `pkg/notifier/` — SMTP, Slack, Discord, Google Chat, Teams, webhooks
- **Encryption**: `pkg/crypto/` — AES-256-GCM for notification channel credentials
- **Migrations**: `internal/database/migrations/sqlite/` and `postgres/` — dual trees, keep in sync

### Frontend (web/)

Vue 3 Composition API + TypeScript + Pinia + Ant Design Vue.

- API calls only through `web/src/services/*.ts` (Axios via `web/src/libs/axios.helper.ts`)
- State: Pinia stores + composables (e.g., `web/src/composables/useResources.ts`)
- Types centralized in `web/src/types/index.ts`
- No Options API, no raw fetch/axios in components

### API versioning

- `/api/v1/` — stable public API, semver-protected
- `/api/` (non-versioned) — internal, may change anytime

## Patterns to follow

### Adding a new monitor type

1. Implement `CheckStrategy` in `internal/monitoring/strategy/yourtype.go`
2. Add `ResourceYourType` constant to `internal/domain/models.go`
3. Register in the strategies map in `cmd/api/main.go`

### Adding a new notification channel

1. Implement `Notifier` in `pkg/notifier/yournotifier.go`
2. Add constant to `internal/domain/models.go`
3. Add dispatch case in `internal/monitoring/incident_service.go`
4. Add config validation in `internal/service/notification_service.go`

### Adding a new v1 API endpoint

1. Handler in `internal/api/handler/v1/` (interface + impl pattern)
2. DTOs in `internal/dto/v1/`
3. Route in `internal/api/router.go` (v1 sub-group)
4. Swaggo annotations, then `make swag` and commit `docs/`
5. Test scope enforcement (read-only API key → 403 on writes)

### Database migrations

- Add to both `sqlite/` and `postgres/` trees
- SQLite: no `ADD COLUMN IF NOT EXISTS`, no multi-column `ALTER TABLE`
- Use `GORM serializer:json` instead of `type:jsonb` for cross-driver JSON fields
- Naming: `XXXX_description.up.sql` / `.down.sql`

### Testing

- DB tests use `setupTestDB(t)` helper (SQLite in-memory)
- Table-driven tests for multi-case scenarios
- `go test -race ./...` must pass
- Frontend: Vitest + jsdom, `*.spec.ts` files

## Code Quality — SonarQube (MANDATORY)

Before completing any task or commit, you MUST:

1. **Ensure coverage reports are generated** (both stacks):
   ```bash
   make test-be          # generates coverage/unit.out (Go)
   make test-fe          # generates web/coverage/lcov.info (Vue/TS)
   ```

2. **Run the SonarQube scanner**:
   ```bash
   sonar-scanner -Dsonar.token=<YOUR_TOKEN>
   ```
   (Set `SONAR_TOKEN` env var or pass via `-D` flag. Config is in `sonar-project.properties`.)

3. **Verify quality gate** via SonarQube dashboard (local: http://localhost:9009):
   - Project: `ogoune`
   - Check for **CRITICAL** and **BLOCKER** issues in **both**:
     - Go backend (`cmd/`, `internal/`, `pkg/`)
     - Vue frontend (`web/src/`)

4. **If issues are detected**:
   - **Go issues**: fix in `cmd/`, `internal/`, `pkg/`, re-run `make test-be && sonar-scanner`
   - **TS/Vue issues**: fix in `web/src/`, re-run `make test-fe && sonar-scanner`
   - **Maximum 3 cycles total** — if still failing after 3 attempts, stop and report remaining issues
   
5. **Never merge with Quality Gate = FAILED**
   - This is in addition to `make lint` and `make test` (which also remain mandatory)

### Why monorepo matters
- Two coverage reports are required (Go + TS/lcov)
- Each stack may have different critical issues
- Fixing one stack's issue doesn't block the other

## Gotchas

- `APP_SECRET_KEY` env var is mandatory — app refuses to start without it
- Repository errors: map `repository.ErrNotFound` to service-level errors, handlers map those to HTTP status
- Incident event steps (`detected`, `resource_down_alert`, `resolved`, `resource_up_alert`) may not all be present
- Never block on scheduler failures — log and return `ErrSchedulerSync`
- Commits use Conventional Commits format (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`)

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
# currentDate
@specs/034-frontend-quality-improvements/plan.md
<!-- SPECKIT END -->
