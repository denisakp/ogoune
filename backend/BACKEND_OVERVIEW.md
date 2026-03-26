# Pulseguard Backend – Unified Overview

This document provides a consolidated, developer‑oriented overview of the Pulseguard backend. It explains the core architecture, how checks are scheduled and processed, what the main APIs do (with emphasis on Status and Stats), how data flows through the system, and how to configure and operate the service.

Contents:

- What this backend does
- Stack and layout
- Endpoints (high level)
- Data model and persistence
- Monitoring pipeline (schedule → execute → persist → incidents → notifications)
- Status endpoint (90‑day model, “no_data”)
- Stats endpoint (global uptime, incidents, affected monitors)
- Configuration and environment variables
- Notifications (SMTP and integrations)
- Local development and operations
- Performance, scaling, and security notes
- Pointers to code and docs

---

## 1) What this backend does

Pulseguard monitors resources (HTTP/TCP), stores check results, detects incidents, and emits notifications. It exposes a pure JSON API to:

- Manage resources, tags, and integrations
- Inspect incidents and activities
- Provide status page data with pre‑aggregated 90‑day history
- Provide aggregated statistics across all monitors

The system is a single Go binary that initializes the configured database runtime, the Redis/Asynq scheduler and worker, and the HTTP server.

---

## 2) Stack and layout

- Language/runtime: Go 1.25+
- HTTP router: Chi
- ORM/DB: GORM + a driver-aware PostgreSQL/SQLite runtime
- Background processing: Redis + Asynq (periodic scheduler + worker server)
- Notifications: user-configured channels (SMTP with templates, Slack, Webhook)
- Config: environment variables (dotenv supported in development)

Key packages and files (relative to backend/):

- cmd/api/main.go: application entrypoint (wires DB, Asynq scheduler/worker, and HTTP server)
- internal/api/router.go: all JSON routes and CORS
- internal/api/handler/*: HTTP handlers (resources, incidents, status page, stats, etc.)
- internal/config/config.go: env loading, dotenv, SMTP enablement flag
- internal/domain/models.go: core entities and enums
- internal/domain/check.go: check strategies interface + executor
- internal/monitoring/scheduler_service.go: Asynq periodic scheduling per resource
- internal/monitoring/incident_service.go: incident lifecycle + notifications fan‑out
- internal/monitoring/strategy/{http,tcp}.go: concrete monitoring strategies
- internal/worker/processor.go: Asynq server and task mux
- internal/worker/handler_monitoring.go: “monitoring:check” task handler- internal/worker/handler_expiry.go: "expiry:check" daily background task handler
- internal/service/expiry_notification_service.go: threshold evaluation, dedup, and notification dispatch for SSL/domain expiry- internal/database/*: driver selection, openers, startup migration runner, authoritative SQL migrations
- internal/repository/postgres/*: GORM repository implementations + legacy DB wrapper
- pkg/notifier/*: SMTP + Slack/Webhook providers and factory

---

## 3) Endpoints (high level)

- Health: GET /health
- Status Page:
  - GET /status → global status + each resource’s 90‑day daily status and uptime percentage
  - GET /status/{resourceId} → single resource detail (90‑day series, recent events, response time summary)
- Resources (Monitors):
  - GET /resources, POST /resources
  - GET /resources/{id}, PATCH /resources/{id}, DELETE /resources/{id}
  - POST /resources/{id}/pause, POST /resources/{id}/resume
  - POST /resources/{resourceID}/tags, DELETE /resources/{resourceID}/tags/{tagID}
  - GET /resources/{resourceId}/uptime-stats → hourly uptime for last 24h
- Monitoring Activities: GET /monitoring-activities (supports ?resource_id=...)
- Incidents:
  - GET /incidents (supports ?unresolved=true, ?limit, ?offset)
  - GET /incidents/{id}
  - GET /incidents/{id}/event-steps
- Integrations:
  - GET /integrations, POST /integrations, PATCH /integrations/{id}
- Notifications:
  - POST /notifications/test (SMTP test)
- Stats:
  - GET /stats/summary?range=2h|24h|7d|30d
- API Keys (JWT-only management routes):
  - POST /account/api-keys
  - GET /account/api-keys
  - DELETE /account/api-keys/{id}

API key auth on regular API routes accepts:

- `X-API-Key: pk_live_*`
- `Authorization: Bearer pk_live_*`

Scope behavior:

- `read` keys can access read endpoints
- `read_write` keys can access mutating endpoints guarded by `RequireReadWrite`

See router for the definitive list and mapping.

---

## 4) Data model and persistence

GORM entities (see internal/domain/models.go):

- Resource: id, name, type (http|tcp), target, interval(s), timeout(s), status, is_active, failure_count, last_checked, created_at, updated_at; relations to tags, incidents, activities
- MonitoringActivity: resource_id, success(bool), message, response_time(ms), response_data, created_at
- Incident: resource_id, cause, started_at, resolved_at(NULL while active), details(bytes), event steps
- IncidentEventStep: incident_id, step (detected, resolved, resource_down_alert, resource_up_alert, …), message
- Integration: config(JSON), event_types(JSON), is_active, name
- NotificationEvent: incident_id, type (up|down|expiry), status (pending|sent|failed|expired), claim_owner, claimed_at, processed_at, last_error, timestamps
- Tags: many‑to‑many with resources

DB initialization:

- On startup, PulseGuard resolves `DB_DRIVER`, opens PostgreSQL or SQLite, configures pool policy, and applies pending versioned SQL migrations before serving requests
- SQLite mode enables WAL mode and a busy timeout for embedded community deployments

---

## 5) Monitoring pipeline

End‑to‑end flow:

1) Scheduling
   - On boot, active resources are fetched and each is registered as an Asynq periodic task.
   - A unique entry id `monitor:<resourceID>` is used. Re‑scheduling unregisters then registers the new spec.
   - Cronspec uses `@every <interval>s` to enqueue “monitoring:check” tasks at the resource’s interval.

2) Execution
   - Asynq worker (default concurrency 10) consumes “monitoring:check”.
   - The handler re‑loads the resource from the DB (respects is_active/updated interval) and runs a check via the `CheckExecutor`.

3) Strategies
   - HTTP: HEAD request using resource timeout; success for 2xx/3xx; error mapping for common failure classes (timeout/refused/dns/ssl/invalid status).
   - TCP: attempts to connect within timeout; success on connection, failure otherwise.

4) Persistence
   - Each check writes a `MonitoringActivity` with success flag, response time, and optional response metadata.

5) Resource state transitions
   - On down: increment `failure_count`, update `status`, update `last_checked`.
   - On up: reset `failure_count` to 0, set `status=up`, update `last_checked`.

6) Incidents
   - Create on the 3rd consecutive failure (guards against duplicates if one is already active).
   - Resolve on the first up transition after a down (resolves the most recent active incident).
   - Event steps are stored for `detected` and `resolved`, and for notification attempts.

7) Notifications
   - Channel-based: notification channels (SMTP/Slack/Webhook/SMS) are stored in the database and dispatched via the notifier factory.
  - Resolution order is resource channels, then component channels, then default-enabled global channels.
  - If no channel is resolved, incident creation continues and a warning log is emitted with remediation guidance.
  - Notification events are persisted as `pending` before dispatch and terminalized as `sent` or `failed` after immediate delivery attempts.
  - At startup, a single recovery pass retries recent pending down/up events, expires stale events, and skips already-claimed rows to prevent duplicate dispatch in multi-instance boot scenarios.
   - Testing: `/notification-channels/{id}/test` for saved channels; `/notification-channels/test-config` to validate before saving.

9) Incident diagnostics and API readability
  - Incident and monitoring activity APIs return byte-backed diagnostic fields as readable text.
  - HTTP incident diagnostics persist response headers when available.
  - Persisted request headers remove the Authorization key entirely before storage.

8) SSL/Domain Expiry Alerts (daily background task)
   - A separate `expiry:check` task runs once per day (via Asynq `@daily` scheduler or a 24-hour TimingWheel ticker in community/SQLite deployments).
   - `ExpiryTaskHandler` (internal/worker/handler_expiry.go) iterates all active HTTP resources, enriches SSL/WHOIS metadata, detects certificate/domain renewals (resets dedup logs), and delegates to `ExpiryNotificationService`.
   - `ExpiryNotificationService` (internal/service/expiry_notification_service.go) evaluates configurable threshold sets (global default: `EXPIRY_ALERT_THRESHOLDS=30,14,7,1`; per-resource override via `expiry_alert_thresholds` field), deduplicates via `expiry_notification_logs`, and dispatches alerts through the same SMTP/Webhook channel routing used for incidents.
   - Resource API (`PATCH /resources/{id}`) accepts `expiry_alert_thresholds` (comma-separated integers 1–365) to override global defaults per monitor.

---

## 6) Status endpoint (90‑day data)

- GET /status returns:
  - global_status: all_systems_operational | some_systems_down
  - generated_at: timestamp
  - resources[]:
    - id, name
    - current_status (simplified mapping from internal status)
    - uptime_percentage_last_90_days (float 0..100)
    - daily_status_last_90_days: exactly 90 entries, oldest → newest

Daily status values:

- "up", "degraded", "down", "no_data"
- “no_data” is returned for days before the resource was created (fixed in v2.0.1)
- Major incident coverage and success rate thresholds influence degraded/down

Important notes:

- The 90‑day array is always length 90 for UI consistency
- Uptime percentage excludes “no_data” days to avoid penalizing new resources

See:

- docs/STATUS_ENDPOINT.md
- docs/NO_DATA_STATUS_FIX.md
- docs/STATUS_ENDPOINT_CHANGELOG.md
- docs/status-endpoint-example.json
- docs/resource-detail-status-example.json

---

## 7) Stats endpoint (global aggregates)

- GET /stats/summary?range=2h|24h|7d|30d returns:
  - range: the requested range string
  - overall_uptime: average uptime across all resources
  - incidents: total count of incidents starting in the window
  - affected_monitors: distinct resources with incidents
  - without_incidents_duration: currently "0m" (placeholder; planned enhancement)

Implementation details:
- Uptime computed from `monitoring_activities` success ratios in the time window
- Incident counts and affected monitors computed from `incidents.started_at` within the window

See:
- docs/STATS_API.md

---

## 8) Configuration and environment variables

The app reads configuration from environment variables and also supports loading a local .env file in development.

Required (typical local defaults in parentheses):
- Server:

  - PORT (8080)
  - APP_ENV (development)
- Database:
  - DB_DRIVER (postgres or sqlite)
  - DATABASE_URL (postgres://user:password@localhost:5432/pulseguard?sslmode=disable) when DB_DRIVER=postgres
  - SQLITE_PATH (pulseguard.db) when DB_DRIVER=sqlite
  - DB_LOG_LEVEL (silent, error, warn, info)
- Redis:
  - REDIS_URL (localhost:6379)

Notifications are now configured via notification channels (stored in the database). There is no default SMTP configuration from environment variables; add SMTP/Slack/Webhook channels via the UI or `/notification-channels` APIs.

Validation:

- DATABASE_URL is required only for PostgreSQL mode; startup fails fast on missing DSN or failed SQL migrations

---

## 9) Notifications

Channel-based delivery:

- Channels (SMTP/Slack/Webhook/SMS) are created via `/notification-channels` and stored in the database.
- Test a saved channel with `POST /notification-channels/{id}/test`; validate before saving with `POST /notification-channels/test-config`.
- Each attempt is audited by a `NotificationEvent`.

---

## 10) Local development and operations

Prerequisites:

- A running Redis server reachable via REDIS_URL
- Either a running PostgreSQL instance reachable via DATABASE_URL or a writable SQLite path when DB_DRIVER=sqlite
- A running Redis server reachable via REDIS_URL

Run:

- From backend/, `go run ./cmd/api`
- On startup the app:
  - Connects to the configured PostgreSQL or SQLite runtime and applies versioned SQL migrations
  - Connects to Redis and starts:
    - Asynq periodic scheduler
    - Asynq worker server (queue: "monitoring", concurrency: 10)
  - Bootstraps scheduling for all active resources
  - Starts HTTP server at http://localhost:${PORT}

Observability:

- Logs to stdout; GORM logs slow queries (threshold 200ms)
- Consider reverse proxy for TLS termination and request logs

Scaling:

- Worker throughput scales horizontally: multiple instances can consume the "monitoring" queue concurrently
- Scheduler runs in every instance by default (current implementation). To avoid duplicate periodic registrations, prefer running a single instance as the scheduler until a leadership/toggle is introduced.

Backups and durability:

- PostgreSQL or SQLite is the source of truth depending on DB_DRIVER; ensure regular backups of the active runtime
- Redis is used for job transport and scheduling; it does not store authoritative data

Troubleshooting confirmation enforcement:

- Below-threshold failures should not create incidents or down-alert steps. If alerts appear too early, verify `confirmation_checks` on the resource and inspect `failure_count` progression in the DB.
- If `failure_count` is already above threshold but no unresolved incident exists, the next DOWN cycle should create an incident immediately. Check worker logs for `[INCIDENT_RECONCILE]` entries.
- When persistence of failure progression fails, incident creation is intentionally skipped for that cycle and retried on the next check. Check worker logs for `failed to persist failure progression` warnings.

---

## 11) Performance, scaling, and security notes

Performance/scaling:

- Asynq worker concurrency defaults to 10; tune per instance
- DB pool defaults are conservative; adjust for production
- Heavier aggregations (status, 90‑day windows) can be cached via reverse proxy or future Redis TTL caches

Security:

- CORS is permissive by default; restrict origins in production
- TLS should be terminated by an ingress or reverse proxy
- Authentication/authorization is not implemented; deploy behind a protected network or add edge auth

---

## 12) Pointers to code and docs

Code:

- Entry and wiring: cmd/api/main.go
- Router and routes: internal/api/router.go
- DB setup and auto‑migrations: internal/repository/postgres/database
- Repositories: internal/repository/postgres/*
- Monitoring scheduler: internal/monitoring/scheduler_service.go
- Worker and task handler: internal/worker/processor.go, internal/worker/handler_monitoring.go
- Incident lifecycle and notifications: internal/monitoring/incident_service.go
- Check executor and strategies: internal/domain/check.go, internal/monitoring/strategy/*
- Stats service: internal/service/stats_service.go
- Status page service: internal/service/statuspage_service.go
- Notifiers: pkg/notifier/* (+ templates/)

Docs:

- docs/STATUS_ENDPOINT.md (full response shape and semantics)
- docs/NO_DATA_STATUS_FIX.md (90‑day "no_data" handling)
- docs/STATUS_ENDPOINT_CHANGELOG.md (breaking changes and v2.0.1 fix)
- docs/STATS_API.md (summary stats endpoint)

---

With this overview, a Go engineer should be able to navigate the codebase, run the backend locally, understand how data is processed end‑to‑end, and extend the system (new strategies, endpoints, or notifiers) with confidence.