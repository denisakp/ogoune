# Pulseguard Backend Architecture

Version: October 2025

This document describes the Go backend that powers Pulseguard. It focuses solely on the JSON API, the background worker, persistence, scheduling, real‑time updates, and notifications. It intentionally excludes any frontend implementation details.

## Table of Contents

1. High‑Level Overview
2. Technology Stack (Backend)
3. Core Architecture (API + Worker)
4. Scheduling Engine (Ticker)
5. Real‑time System (WebSockets)
6. Backend Directory Structure
7. Database Models (Overview)
8. Incident Management Logic
9. Notification System
10. Deployment Notes
11. Performance & Scalability
12. Security

---

## 1) High‑Level Overview

The backend is a headless service that exposes a REST/JSON API and executes health checks asynchronously. Responsibilities:

- Expose endpoints to create/read/update/delete monitoring resources, tags, integrations, and incidents.
- Enqueue and process check jobs for HTTP/TCP targets.
- Persist activities and incidents to PostgreSQL.
- Dispatch notifications via pluggable providers.
- Broadcast activity/incident events to connected clients over WebSockets.

Principles:

- Clean layering (domain, repository, api, worker). Domain has no infra dependencies.
- API handlers must never perform network checks inline; they only enqueue work.
- PostgreSQL is the source of truth; Redis is used for job transport.

---

## 2) Technology Stack (Backend)

- Go 1.25+
- Chi (HTTP router)
- GORM (ORM for PostgreSQL)
- PostgreSQL (primary datastore)
- Redis + Asynq (job queue and worker)
- nhooyr.io/websocket (WebSockets)

---

## 3) Core Architecture (API + Worker)

Single deployable that starts two long‑running components:

- API Server

  - Validates input, applies business rules via services, persists via repositories.
  - Enqueues jobs (e.g., monitoring:check) into Redis (Asynq) for the worker.
  - Manages WebSocket connections (upgrade + message routing to clients).

- Background Worker
  - Consumes jobs from Redis (Asynq server) and executes checks using domain strategies.
  - Writes MonitoringActivity rows, updates Resource state and Incident lifecycle.
  - Emits events to the in‑process broadcast channel consumed by the WebSocket hub.

Flow (simplified):

1. Client calls API (create/update resource) → DB write.
2. API enqueues job → Redis.
3. Worker consumes job → executes check (HTTP/TCP) → persists results.
4. Incident service creates/resolves incidents as needed.
5. Notifications sent; activity broadcasted to WebSocket hub for real‑time UI updates.

---

## 4) Scheduling Engine (Ticker)

The backend uses a ticker‑driven scheduler (internal/scheduler/ticker.go) instead of asynq.Scheduler.

Rationale:

- Robustness and simplicity: a single loop is easier to reason about and operate.
- Database as source of truth: reads active resources and their intervals directly from PostgreSQL.
- Dynamic reconfiguration: interval or pause changes are honored on the next tick; no out‑of‑band schedule writes.

Design:

- A `time.Ticker` wakes up at a small cadence (e.g., 5s).
- On each tick: query active resources; for each resource whose interval elapsed since LastChecked, enqueue a monitoring job and update LastChecked.
- Back‑pressure is handled by Asynq concurrency and retries.

---

## 5) Real‑time System (WebSockets)

Pulseguard pushes live updates to clients using WebSockets built with `nhooyr.io/websocket`.

Components:

- WebSocket Hub: manages client connections; broadcasts JSON messages to all or selected subscribers.
- In‑process Events Channel (e.g., `events.ActivityBroadcast`): a buffered channel decoupling producers (worker) from consumers (hub).

Data Flow:

1. Worker finishes a check → creates a MonitoringActivity and possibly an Incident.
2. Worker publishes an event to `ActivityBroadcast` with the minimal JSON payload (resource id, status, response time, incident id, etc.).
3. The Hub goroutine listens on the channel, marshals to JSON, and writes to all connected WebSocket clients.

Notes:

- Messages are compact JSON objects; version the payload if fields evolve.
- The hub enforces write deadlines and cleans up dead connections.

---

## 6) Backend Directory Structure

High‑level overview of `backend/internal` packages:

- api: routing (Chi), HTTP handlers, request/response mapping.
- config: environment loading and configuration.
- domain: pure business types and logic (validation, check strategies, incident rules).
- repository: interfaces + PostgreSQL implementations (GORM) for resources, incidents, activities, integrations, tags.
- service: application services orchestrating domain + repos.
- worker: Asynq server and task handlers (monitoring jobs, notification dispatch).
- (scheduler): ticker loop responsible for enqueuing due checks.
- (websocket): hub managing WebSocket clients and broadcasting.
- pkg/notifier: providers (SMTP, Slack, Discord, Google Chat), plus templates for emails.

---

## 7) Database Models (Overview)

Core entities (see `internal/domain/models.go` for exact fields):

- Resource: target to monitor (type: http|tcp, target, interval, timeout, status, failure count, timestamps).
- MonitoringActivity: per‑check log (success, message, response time, raw data optional).
- Incident: downtime record (reason/cause, started_at, resolved_at NULL while active).
- Integration: notifier configuration (type + config JSON, subscribed event types).
- Tag: labels for organizing resources (many‑to‑many).

Indexes focus on frequent filters: resource status/active; incident resolved_at/resource_id; activity resource_id/created_at.

---

## 8) Incident Management Logic

- Stateful rule: create an Incident after N consecutive failures (default 3). Reset failure count on success.
- Idempotency: ensure a resource has at most one active incident; if already active, append steps/details but don’t duplicate.
- Resolution: on first successful check after downtime, set `resolved_at` and emit an "up" event.

---

## 9) Notification System

Two layers:

1. System SMTP (optional): sends to a default admin recipient if configured (environment variables). Acts as safety net.
2. User Integrations: per‑integration type (smtp, slack, discord, googlechat, webhook), filtered by `event_types` (e.g., ["down","up"]). Matching integrations are executed in parallel.

Providers live in `pkg/notifier`. Email templates under `pkg/notifier/templates/`.

---

## 10) Deployment Notes

- Single process binary (API + worker + scheduler + websocket hub).
- Requires PostgreSQL and Redis. Configuration via environment variables (DATABASE*URL, REDIS_URL, PORT, SMTP*\* …).
- Horizontal scale by running multiple instances; Asynq coordinates job distribution.

---

## 11) Performance & Scalability

- Concurrency controlled by Asynq; tune worker concurrency per instance.
- DB pool suggested defaults: MaxOpen=25, MaxIdle=5, 30m lifetime.
- Consider caching hot read paths in Redis (future) and partitioning large activity tables.

---

## 12) Security

- Secrets via environment variables (consider external secret managers later).
- Authentication/authorization TBD (API keys/JWT roadmap).
- TLS termination recommended at ingress/reverse proxy.
