# Pulseguard Backend Architecture

This document describes the technical architecture of the Pulseguard backend. It covers the API surface, background processing, persistence, scheduling, incidents, notifications, configuration, and operational guidance, with pointers to the main packages and files.

Scope: Backend only (Go). No frontend details are included.

---

## 1) High‑Level Overview

The backend is a single Go binary that, on startup, initializes:
- PostgreSQL (via GORM) and auto-migrates core models.
- Redis-backed Asynq components:
  - Periodic scheduler: registers recurring checks per resource.
  - Worker server: processes monitoring jobs from Redis.
- HTTP JSON API (Chi router) with CORS enabled.

Core responsibilities:
- CRUD for monitored resources, tags, and integrations.
- Periodic health checks (HTTP, TCP).
- Persisted monitoring activities.
- Incident lifecycle (create on persistent failures, resolve on recovery).
- Notifications via user-configured channels (SMTP/Slack/Webhook).
- Aggregated stats and status endpoints for dashboards/status pages.

Source of truth: PostgreSQL.
Transport for background work: Redis + Asynq.

---

## 2) Technology Stack

- Language/runtime: Go 1.25+
- HTTP: Chi router
- Persistence: GORM ORM + PostgreSQL
- Background jobs: Redis + Asynq (periodic scheduler + worker server)
- Notifications: user-configured channels (SMTP, Slack, Webhook)
- Configuration: Environment variables (dotenv supported in development)

Key modules and files:
- Entry point: `backend/cmd/api/main.go`
- Config: `backend/internal/config/config.go`
- Database: `backend/internal/repository/postgres/database`
- API router: `backend/internal/api/router.go`
- Worker: `backend/internal/worker`
- Monitoring pipeline: `backend/internal/monitoring`
- Domain models and strategies: `backend/internal/domain`, `backend/internal/monitoring/strategy`
- Notifiers: `backend/pkg/notifier`

---

## 3) Process Topology

Single process with three long‑running components:
- HTTP API server
- Asynq periodic scheduler
- Asynq worker server

Startup sequence (simplified):
1. Load config and init DB (auto-migrations).
2. Init Redis/Asynq client, inspector, scheduler.
3. Bootstrap: list active resources and schedule each with Asynq periodic tasks.
4. Start Asynq worker (consumes monitoring jobs).
5. Start HTTP server.

Note: The current binary starts all three components by default in every instance. See Scaling considerations in section 12.

---

## 4) Directory Structure (Backend)

- `internal/api`
  - `router.go`: route wiring, CORS, content-type.
  - `handler/*`: HTTP handlers for resources, activities, tags, integrations, incidents, status page, stats, notifications.
- `internal/config`
  - `config.go`: env loading, dotenv support.
- `internal/domain`
  - `models.go`: core entities (Resource, Incident, Event Steps, Integration, NotificationEvent, MonitoringActivity, Tags).
  - `check.go`: check strategies interface and executor contract.
- `internal/monitoring`
  - `scheduler_service.go`: Asynq periodic scheduling per resource.
  - `incident_service.go`: incident creation/resolution and notifications.
  - `strategy/*`: concrete check strategies (HTTP, TCP; DNS file exists but core strategies are HTTP/TCP).
- `internal/repository`
  - `interfaces.go`: repository interfaces.
  - `postgres/*`: GORM implementations and DB setup (auto-migrate).
- `internal/service`
  - `resource_service.go`, `statuspage_service.go`, `stats_service.go`, etc.: orchestration logic used by handlers.
- `internal/worker`
  - `processor.go`: Asynq server and mux registration.
  - `handler_monitoring.go`: monitoring job handler, resource status transitions, activity persistence, incident handoff.
- `pkg/notifier`
  - `smtp.go` (with HTML templates under `templates/`), `slack.go`, `webhook.go`, `factory.go`.

---

## 5) Data Model (Summary)

- Resource
  - Fields: `id`, `name`, `type` (http|tcp), `target`, `interval` (seconds), `timeout` (seconds), `status` (up/down/error/warning/pending/unknown/paused), `is_active`, `failure_count`, `last_checked`, `created_at`, `updated_at`.
  - Relations: many‑to‑many `tags`, one‑to‑many `incidents`, one‑to‑many `monitoring_activities`.
- MonitoringActivity
  - Per-check log: `resource_id`, `success` (bool), `response_time` (ms), `message`, `response_data`, timestamps.
- Incident
  - Downtime record: `resource_id`, `cause` (structured string), `started_at`, `resolved_at` (NULL while active), `details` (bytes), event steps.
- IncidentEventStep
  - Timeline events for each incident: `detected`, `resolved`, `resource_down_alert`, `resource_up_alert`, etc.
- Integration
  - Notification integration config (JSON) + subscribed event types (JSON).
- NotificationEvent
  - Audit of notification attempts (type up/down/expiry).
- Tags
  - Labels for resources (many‑to‑many).

GORM auto-migrates all above on startup.

---

## 6) API Surface (JSON only)

- Health
  - `GET /health` → "OK" (text)
- Status Page
  - `GET /status` → global status + 90‑day per-resource summary (see `docs/STATUS_ENDPOINT.md`)
  - `GET /status/{resourceId}` → detailed single-resource view (see `docs/resource-detail-status-example.json`)
- Resources
  - `GET /resources`
  - `POST /resources`
  - `GET /resources/{id}`
  - `PATCH /resources/{id}`
  - `DELETE /resources/{id}`
  - `POST /resources/{id}/pause`
  - `POST /resources/{id}/resume`
  - `POST /resources/{resourceID}/tags`
  - `DELETE /resources/{resourceID}/tags/{tagID}`
  - `GET /resources/{resourceId}/uptime-stats` (hourly aggregation)
- Monitoring Activities
  - `GET /monitoring-activities` (supports `?resource_id=...`)
- Incidents
  - `GET /incidents` (supports `?unresolved=true`, `?limit`, `?offset`)
  - `GET /incidents/{id}`
  - `GET /incidents/{id}/event-steps`
- Integrations
  - `GET /integrations`
  - `POST /integrations`
  - `PATCH /integrations/{id}`
- Notifications
  - `POST /notifications/test`
- Stats
  - `GET /stats/summary?range=2h|24h|7d|30d` (see `docs/STATS_API.md`)

CORS: permissive wildcard by default (adjust for production).

Authentication/authorization: not implemented at present.

---

## 7) Monitoring Pipeline

1. Scheduling
   - On boot, active resources are read from PostgreSQL and scheduled as recurring Asynq tasks using the `@every <interval>s` cronspec.
   - Each resource has a unique periodic entry id `monitor:<resourceID>`. Updating a resource reschedules by unregistering and re-registering its periodic task.
   - Inactive resources are not scheduled.

2. Execution
   - Asynq scheduler enqueues "monitoring:check" tasks into the "monitoring" queue at the requested interval.
   - The Asynq worker (concurrency default: 10) consumes tasks via `handler_monitoring.go`.

3. Strategy
   - `CheckExecutor` dispatches by `Resource.Type`:
     - HTTP: HEAD request with timeout; maps common error substrings to structured causes (timeout/refused/dns/ssl).
     - TCP: dial target with timeout and infer availability from connection result.

4. Persistence
   - A `MonitoringActivity` row is created per check with success flag and response metrics.

5. State transitions
   - Previous resource status is read; `failure_count` increments on consecutive downs, resets on up.
   - After each check, resource `status` and `last_checked` are updated.

6. Incidents
   - On exactly the 3rd consecutive failure: create incident if none is active.
   - On recovery (first `up` after a `down`): resolve the most recent active incident.

7. Notifications
   - Channel-based: notification channels (SMTP/Slack/Webhook/SMS) are stored in the database and dispatched via the notifier factory.
   - Testing: `/notification-channels/{id}/test` for saved channels, `/notification-channels/test-config` to validate before saving.

8. Status and Stats
   - `/status`: pre-aggregated 90‑day uptime and per‑day status array per resource (includes "no_data" days before resource creation). See `docs/STATUS_ENDPOINT.md`.
   - `/stats/summary`: global uptime %, incident counts, affected monitors across time windows. See `docs/STATS_API.md`.

---

## 8) Scheduling Details (Asynq)

- Component: `internal/monitoring/scheduler_service.go`
- API:
  - `Schedule(ctx, *domain.Resource)`:
    - Unregister previous periodic entry `monitor:<id>` (ignore if not present).
    - If resource is active, register `asynq.NewTask("monitoring:check", payload)` with `@every <interval>s` in the "monitoring" queue.
  - `Unschedule(ctx, resourceID)`:
    - Unregister entry `monitor:<id>`.

Operational notes:
- Rescheduling is idempotent: safe on updates.
- The scheduler runs in‑process and requires Redis connectivity.
- At present, every running instance also starts a scheduler; see Scaling (section 12).

---

## 9) Worker Server

- Component: `internal/worker/processor.go`
- Server config:
  - Concurrency: 10
  - Queues: `monitoring: 10`
- Mux:
  - `"monitoring:check"` → `MonitoringTaskHandler.ProcessTask`
- Handler steps:
  - Parse payload, re-fetch resource from DB (honors current `is_active`, interval changes, etc.).
  - Execute check via `CheckExecutor`.
  - Persist activity.
  - Transition resource status and `failure_count`.
  - Create or resolve incident when appropriate.
  - Trigger notifications via configured channels (SMTP/Slack/Webhook) filtered by event type.

Error handling and retries:
- Handler returns errors for fatal problems (e.g., resource not found). Asynq will retry per its defaults.
- Functional errors (e.g., failing to persist an activity) are logged; incident logic continues to avoid dropping state transitions.

---

## 10) Persistence

- Component: `internal/repository/postgres`
- Initialization: `database.Init(ctx, dsn)` opens the GORM connection, sets pool limits (MaxOpen=25, MaxIdle=5, lifetime=30m), and auto-migrates all registered models.
- Repositories expose CRUD and aggregated queries:
  - Activities: global uptime %, per-resource hourly stats, recent response times.
  - Incidents: unresolved filter, per-resource, incident stats (count and affected monitors).
  - Resources, Tags, Integrations, Incident Event Steps, Notification Events.

PostgreSQL is the sole source of truth; Redis is used only for job scheduling/transport.

---

## 11) Notifications

  - Channel-based delivery: SMTP/Slack/Webhook/SMS channels are created and stored via `/notification-channels` APIs.
  - Tests: `/notification-channels/{id}/test` for saved channels; `/notification-channels/test-config` to validate before saving.

- Integrations (user-level):
  - Providers: Slack, Webhook (and SMTP as a channel).
  - Selection: by `Integration.Config.type` and `Integration.EventTypes` (contains the current event type).
  - Dispatch: parallelized per incident event via notifier factory.
  - Audit: each attempt creates a `NotificationEvent` entry.

---

## 12) Operations & Scaling

- Single process model:
  - API server, Asynq periodic scheduler, and worker server all run in the same binary.

- Horizontal scaling:
  - Worker: multiple instances safely share the Redis queue; this scales check throughput linearly.
  - Scheduler: current code starts a scheduler in every instance. Running multiple schedulers may register the same periodic entries concurrently. To avoid double-enqueueing checks, prefer running a single instance of the binary in “scheduler-enabled” mode. If you deploy multiple instances today, be aware of possible duplicated scheduling until a toggle/leadership mechanism is introduced.

- Database:
  - Tune pool sizes based on workload. Auto-migration runs at startup; ensure least-privilege and pre-migrate in production if needed.

- Redis:
  - Ensure proper sizing and persistence policy for Asynq keys. Network latency to Redis directly affects scheduling jitter.

- CORS/security:
  - CORS is wide open by default; restrict origins for production.
  - TLS termination should be handled by an ingress/reverse proxy.
  - Authentication/authorization is not implemented; gate API access at the edge if required.

- Logging:
  - Structured logs to stdout; GORM warns slow queries.

- Backups:
  - RPO depends on PostgreSQL backup cadence; Redis is ephemeral for job transport.

---

## 13) Configuration

Environment variables (examples in parentheses indicate typical local defaults):
- Server:
  - `PORT` (e.g., `8080`)
  - `APP_ENV` (e.g., `development`)
- Database:
  - `DATABASE_URL` (e.g., `postgres://user:password@localhost:5432/pulseguard?sslmode=disable`)
- Redis:
  - `REDIS_URL` (e.g., `localhost:6379`)

Dotenv:
- In development, `.env` is loaded if present.

Required:
- `DATABASE_URL` must be set; the process will exit if missing.

---

## 14) Local Development

Prerequisites:
- PostgreSQL accessible via `DATABASE_URL`.
- Redis accessible via `REDIS_URL`.

Typical flow:
- Create and export environment variables (or use a `.env` file).
- Run the backend:
  - Change directory to `backend` and execute `go run ./cmd/api`.
  - On startup, the app will:
    - Connect to PostgreSQL and migrate.
    - Connect to Redis, start the Asynq scheduler and worker.
    - Bootstrap scheduling for active resources.
    - Start the JSON API server at `http://localhost:<PORT>`.

Useful endpoints:
- Health: `/health`
- Status page data: `/status`
- Stats: `/stats/summary`
- Resources CRUD: `/resources`
- Incidents: `/incidents`

---

## 15) Incident Management Logic (Details)

- Creation:
  - Increment `failure_count` on each consecutive `down`.
  - On exactly the 3rd consecutive failure: create an incident if none is active.
  - Add an `IncidentEventStep` with `detected`.
  - Send notifications through configured channels subscribed to `down`.

- Resolution:
  - When a `down` resource transitions to `up`, resolve the most recent active incident (set `resolved_at`).
  - Add `IncidentEventStep` with `resolved`.
  - Send notifications for `up`.

- Idempotency:
  - Only one active incident per resource at a time; checks guard against duplicates.

- Cause classification:
  - `Incident.Cause` derived from `CheckResult` (timeout/refused/dns/invalid_status_code/ssl_certificate_error/etc) for consistent reporting.

---

## 16) Status and Statistics

- Status endpoint (`/status`):
  - Returns `global_status`, generation timestamp, and per-resource block with:
    - `current_status` simplified (up/down/degraded/… mapping).
    - `uptime_percentage_last_90_days`.
    - `daily_status_last_90_days` (exactly 90 entries; `no_data` for days before creation).
  - See `docs/STATUS_ENDPOINT.md`, `docs/STATUS_ENDPOINT_CHANGELOG.md`, and `docs/NO_DATA_STATUS_FIX.md`.

- Stats endpoint (`/stats/summary`):
  - Aggregations across all resources over `2h`, `24h`, `7d`, `30d`:
    - Overall uptime %, total incidents, affected monitors, placeholder “without incidents duration”.
  - See `docs/STATS_API.md`.

---

## 17) Future Enhancements (Non‑exhaustive)

- Scheduler leadership/toggle to safely run multiple API instances without duplicate scheduling.
- WebSocket or Server‑Sent Events for real‑time UI updates.
- Caching for heavy status aggregations (e.g., Redis TTL).
- Authentication/authorization (API keys/JWT).
- Metrics and tracing for checks, DB queries, and queue latencies.
- More check strategies (DNS, TLS/SSL expiry, WHOIS/domain expiry fully integrated).
- Fine‑grained notification throttling and templating across providers.

---

## 18) Quick Cross‑Reference

- Entry and wiring: `backend/cmd/api/main.go`
- Router and routes: `backend/internal/api/router.go`
- Scheduling: `backend/internal/monitoring/scheduler_service.go`
- Worker server: `backend/internal/worker/processor.go`
- Monitoring handler: `backend/internal/worker/handler_monitoring.go`
- Incident service: `backend/internal/monitoring/incident_service.go`
- Check executor & strategies: `backend/internal/domain/check.go`, `backend/internal/monitoring/strategy`
- Stats service: `backend/internal/service/stats_service.go`
- Status page service: `backend/internal/service/statuspage_service.go`
- GORM DB setup: `backend/internal/repository/postgres/database`
- Notifiers: `backend/pkg/notifier`

---

This document reflects the current implementation in the repository and is intended to help Go developers navigate and extend the backend effectively.