# Pulseguard Backend – Unified Overview

This document provides a consolidated, developer‑oriented overview of the Pulseguard backend. It explains the core architecture, how checks are scheduled and processed, what the main APIs do (with emphasis on Status and Stats), how data flows through the system, and how to configure and operate the service.

Contents
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

The system is a single Go binary that initializes the database, the Redis/Asynq scheduler and worker, and the HTTP server.

---

## 2) Stack and layout

- Language/runtime: Go 1.25+
- HTTP router: Chi
- ORM/DB: GORM + PostgreSQL
- Background processing: Redis + Asynq (periodic scheduler + worker server)
- Notifications: SMTP (HTML templates) + Slack, Webhook
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
- internal/worker/handler_monitoring.go: “monitoring:check” task handler
- internal/repository/postgres/*: GORM repository implementations + DB setup
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

See router for the definitive list and mapping.

---

## 4) Data model and persistence

GORM entities (see internal/domain/models.go):
- Resource: id, name, type (http|tcp), target, interval(s), timeout(s), status, is_active, failure_count, last_checked, created_at, updated_at; relations to tags, incidents, activities
- MonitoringActivity: resource_id, success(bool), message, response_time(ms), response_data, created_at
- Incident: resource_id, cause, started_at, resolved_at(NULL while active), details(bytes), event steps
- IncidentEventStep: incident_id, step (detected, resolved, resource_down_alert, resource_up_alert, …), message
- Integration: config(JSON), event_types(JSON), is_active, name
- NotificationEvent: incident_id, type (up|down|expiry), timestamps
- Tags: many‑to‑many with resources

DB initialization:
- On startup, connection pool is configured (MaxOpen=25, MaxIdle=5, Lifetime=30m)
- Auto‑migrations run for all models

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
   - Layer 1 (system SMTP, optional): if SMTP env vars are fully set, send default “down” and “up” emails and log `NotificationEvent`.
   - Layer 2 (integrations): fetch active integrations, filter by `event_types` (e.g., ["down","up"]), then dispatch using the notifier factory (Slack, Discord, Google Chat). Each attempt logs a `NotificationEvent`.

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
  - DATABASE_URL (postgres://user:password@localhost:5432/pulseguard?sslmode=disable)
- Redis:
  - REDIS_URL (localhost:6379)
- SMTP (optional; all required to enable system SMTP notifications):
  - SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD, SMTP_SENDER, DEFAULT_RECIPIENT_EMAIL

Validation:
- DATABASE_URL is required; the process exits if missing
- SMTP notifications are enabled only if all SMTP variables are non‑empty

---

## 9) Notifications

Two‑layer fan‑out:
- System SMTP (optional):
  - Sends default admin notifications for DOWN and UP
  - Uses embedded HTML templates for emails (subject and content tailored per event)
  - A test endpoint is available: POST /notifications/test
- User integrations (Slack, Discord, Google Chat):
  - Each `Integration` carries a JSON config (`config`) that includes `"type": "slack"|"webhook"` and any provider settings
  - Subscriptions are filtered via `event_types` JSON array
  - A simple factory resolves notifier implementations by type
  - Every attempt is audited by a `NotificationEvent`

---

## 10) Local development and operations

Prerequisites:
- A running PostgreSQL instance reachable via DATABASE_URL
- A running Redis server reachable via REDIS_URL

Run:
- From backend/, `go run ./cmd/api`
- On startup the app:
  - Connects to PostgreSQL and auto‑migrates
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
- PostgreSQL is the source of truth; ensure regular backups
- Redis is used for job transport and scheduling; it does not store authoritative data

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