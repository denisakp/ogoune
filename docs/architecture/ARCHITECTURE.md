# Ogoune Backend Architecture

Concise technical reference for the Ogoune backend. This document unifies architecture, runtime flow, APIs, and operations into one source.

## 1) System Overview

Ogoune is a single Go service that:

- Monitors HTTP and TCP resources
- Persists check activity and incident lifecycle in SQL
- Schedules and executes checks via background workers
- Sends notifications through channel-based providers
- Exposes JSON APIs for UI and integrations

Runtime components:

- HTTP API server (Chi router)
- Scheduler (Asynq or in-process timingwheel depending on mode)
- Worker processors (monitoring checks and expiry checks)
- SQL database (PostgreSQL or SQLite)
- Redis (required for Asynq lane)

## 2) Core Stack

- Go
- Chi (HTTP routing)
- GORM + versioned SQL migrations
- PostgreSQL or SQLite runtime
- Asynq + Redis for hosted/compatibility lane
- Timingwheel scheduler for SQLite community lane
- Notifier providers: SMTP, Slack, Webhook, SMS (via channel config)

## 3) Project Layout (Key Paths)

- `cmd/api/main.go`: thin entry point (~26 lines), delegates to bootstrap package
- `internal/platform/bootstrap/`: application composition root (config, DB, scheduler, worker, HTTP)
- `internal/api/router.go`: routes and middleware wiring
- `internal/api/handler/*`: transport handlers (JSON)
- `internal/service/*`: business orchestration layer
- `internal/repository/interfaces.go`: repository contracts
- `internal/repository/store/*`: SQL-backed repository implementations
- `internal/monitoring/*`: scheduling and incident orchestration
- `internal/monitoring/strategy/*`: check strategies (HTTP/TCP)
- `internal/worker/*`: task processors (`monitoring:check`, `expiry:check`)
- `internal/database/*`: driver initialization and migrations
- `internal/domain/*`: entities and domain rules
- `pkg/notifier/*`: channel notifier implementations and factory

## 4) Runtime Modes

Ogoune supports two execution lanes:

1. Community lane

- Typical setup: SQLite + timingwheel scheduler
- Redis not required

2. Hosted compatibility lane

- Typical setup: PostgreSQL + Asynq scheduler/worker
- Redis required

Selection is controlled by `DB_DRIVER` and optional `SCHEDULER_MODE`.

## 5) Monitoring Flow

1. Resource scheduling

- Active resources are scheduled according to `interval`
- Schedule/unschedule operations are driven through scheduler abstraction

2. Check execution

- Worker consumes check tasks
- Handler reloads resource state and executes strategy by type

3. Activity persistence

- Each run writes a monitoring activity record (success, message, response time, metadata)

4. Resource state transition

- Failure increments `failure_count`
- Success resets `failure_count`
- `status` and `last_checked` are updated each cycle

5. Incident lifecycle

- Incident opens when failures cross confirmation threshold
- Incident resolves when service recovers after a confirmed outage
- Event steps are persisted (`detected`, `resolved`, alert events)

6. Notification dispatch

- Delivery uses configured notification channels (resource/component/global resolution)
- Delivery attempts are audited through notification events
- Startup recovery retries recent pending notification events and expires stale ones

## 6) Confirmation Logic

- `confirmation_checks` controls how many consecutive failures are required before incident creation
- While below threshold, resource is in confirmation phase (no incident yet)
- `confirmation_interval` can temporarily increase check cadence for faster confirmation
- Recovery before threshold is treated as false positive and does not create incidents

## 7) Expiry Monitoring

A daily `expiry:check` task evaluates SSL/domain expiry for active HTTP resources:

- Applies global thresholds (configurable) or per-resource overrides
- Deduplicates alerts to avoid repeated notifications at the same threshold
- Uses the same channel-based notification system as incidents

## 8) API Surface (High Level)

Primary endpoints:

- Health: `GET /health`
- Resources: CRUD + pause/resume + tags + uptime stats
- Activities: `GET /api/monitoring-activities`
- Incidents: list/details/event steps
- Status page data: global and per-resource status
- Stats summary: aggregated uptime/incidents over a time range
- Notification channels: create/test/validate configuration

Static serving rules:

- Unmatched `/api/*` stays API-first and returns 404
- `/status` and `/status/*` prefer `status.html`, with fallback behavior when absent

## 9) Data Model (Essential)

- `Resource`: monitor target/config/state (`status`, `failure_count`, timing fields)
- `MonitoringActivity`: per-check result record
- `Incident`: outage window (`started_at`, `resolved_at`, cause/details)
- `IncidentEventStep`: incident timeline markers
- `NotificationEvent`: notification attempt lifecycle (`pending`, `sent`, `failed`, `expired`)
- `Tag`: resource classification

IDs are ULIDs. Schema is controlled by migrations in `internal/database/migrations`.

## 10) Configuration Essentials

Key environment variables:

- `PORT`
- `APP_ENV`
- `DB_DRIVER` (`postgres` or `sqlite`)
- `DATABASE_URL` (required for postgres mode)
- `SQLITE_PATH` (sqlite mode)
- `REDIS_URL` (required for Asynq lane)
- `SCHEDULER_MODE` (`timingwheel` or `asynq`)
- `CONFIRMATION_CHECKS`
- `CONFIRMATION_INTERVAL`
- `STATIC_DIR`

Notification channels are database-configured (not environment-default SMTP).

## 11) Operations and Scaling

- API layer is stateless and horizontally scalable
- Worker throughput scales horizontally with shared queue
- Run one scheduler authority in hosted environments to avoid duplicate periodic registration
- Use standard SQL backup strategy (database is source of truth)
- Redis is transport/scheduling infrastructure, not authoritative data storage

## 12) Extension Points

Common ways to extend:

- New monitor type: implement a new check strategy and register it at bootstrap
- New API endpoint: add service method, handler, and route wiring
- New notifier: add provider in `pkg/notifier` and hook factory resolution
- New background task: add worker handler and scheduler registration

## 13) Local Run (Minimal)

Community lane:

- `DB_DRIVER=sqlite SQLITE_PATH=./ogoune.db SCHEDULER_MODE=timingwheel go run ./cmd/api`

Hosted lane:

- Start PostgreSQL + Redis
- Configure `DB_DRIVER=postgres`, `DATABASE_URL`, `REDIS_URL`
- `go run ./cmd/api`

---

For deeper service behavior, inspect:

- `internal/platform/bootstrap/` (application wiring)
- `internal/worker/handler_monitoring.go`
- `internal/monitoring/incident_service.go`
- `internal/service/statuspage_service.go`
- `internal/service/stats_service.go`
