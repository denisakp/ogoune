# Copilot Project Instructions (Pulseguard)

Purpose: Enable AI coding agents to contribute productively and consistently to the Pulseguard Go codebase.
Keep responses practical, incremental, and aligned with the asynchronous API+Worker architecture.

## Core Architecture (Must Respect)

- Two deployable binaries: `cmd/api` (HTTP + templates) and `cmd/worker` (job processor).
- All monitor execution happens in workers. API must enqueue jobs (never perform checks inline).
- Layers (clean boundaries):
  1. `internal/domain` – pure business types & logic (no HTTP, DB, Redis).
  2. `internal/repository` – persistence + queue abstractions (Postgres, Redis impls in subfolders).
  3. `internal/api` – HTTP routing, handlers, request/response mapping, template rendering.
  4. `internal/worker` – job polling + dispatch to domain logic.
  5. `pkg/notifier` – outward notification clients (Slack, Email, etc.) suitable for reuse.
- `web/template` (Go html/templates + partials) and `web/static` (CSS/JS) – prefer HTMX partial updates.

## When Adding Code

- Put new business rules in `internal/domain/*` with small, testable functions.
- Define interfaces for persistence in `internal/repository` (e.g., `MonitorStore`, `CheckResultStore`). Implement Postgres in `repository/postgres`, queue/cache in `repository/redis`.
- API layer: accept/validate input -> call domain -> enqueue (Redis) or query repos -> render template or JSON.
- Worker: consume queue message -> invoke domain check executor -> persist results -> trigger notifier(s).
- Never import `internal/api` from domain or repositories; enforce one-way dependency (domain is ignorant of infra).

## Conventions

- Go version: from `go.mod` (currently 1.25.x). Use stdlib first; only add deps with clear justification.
- File/package naming: lowercase, short, domain-oriented (`sslcheck`, `uptime`, `scheduler`). Avoid generic `utils`.
- Errors: return rich `error` values; wrap with `fmt.Errorf("context: %w", err)` at boundaries.
- Configuration: centralize future env parsing in `internal/config/config.go` (expand instead of scattering `os.Getenv`).
- Jobs: model as small structs (e.g., `CheckJob{MonitorID, Type, ScheduledAt}`) serialized for Redis (JSON unless a better encoding is added). Keep stable for backward compatibility.
- Templates: prefer small partials that map to route handlers (e.g., `/monitors/list` -> `monitors/list.html`). HTMX responses should return only the fragment required.
- **Database**: Use `internal/repository/postgres/database` package for all DB access. Call `database.Init()` once at startup, use `database.Instance()` in repositories. GORM auto-migration handles schema changes. Pool settings: MaxOpen=25, MaxIdle=5, 30m lifetime.

## Testing (Establish Early)

- Unit tests for domain packages (no external calls). Table-driven style.
- Repository tests may use a temporary Postgres (later: use docker or testcontainers) – if not available yet, scaffold interfaces + in‑memory fake to unblock logic tests.
- Worker logic: test job -> side-effects using mock interfaces (define minimal interfaces to allow this).

## Adding a New Monitor Type (Pattern)

1. Add core type & validation in `internal/domain/monitor` (e.g., `type SSLMonitor struct {...}`).
2. Extend a generic `Monitor` model or registry map (`Type -> executor`).
3. Implement check executor in `internal/domain/check` (pure function returning result struct).
4. Repository additions (schema, migrations later) + method on `MonitorStore`.
5. API: create form handler -> validate -> store -> enqueue initial check.
6. Worker: dispatch on job type -> call executor -> persist -> call notifiers.

## Notifications

- Place provider-specific code in `pkg/notifier/<provider>.go` returning a simple interface `Send(ctx, message)`.
- Keep formatting decisions near the notifier (not spread through worker logic).

## Performance & Scalability Hooks

- Ensure any blocking I/O (network checks) runs in goroutines batched by worker concurrency limit (config-driven later).
- Cache recent check results in Redis (read path in API should first attempt cache once that layer exists).

## Safe Changes / Guardrails

- Do not collapse directory layering even though current files are skeletal—future growth depends on this separation.
- Do not add runtime dependencies in domain (no SQL, Redis, HTTP clients there).
- Keep new public types in `pkg/` minimal and stability-minded (prefer internal until stable).

## Minimal TODO Seeds (Acceptable to Implement Incrementally)

- Populate `internal/config/config.go` with struct + loader stub.
- Implement router scaffolding in `internal/api/router.go` with Chi (defer advanced middleware until needed).
- Introduce job struct & interface definitions in repositories to unblock worker implementation.

## Example: Enqueue Flow (Target Pattern)

(API handler) parse form -> build domain command `CreateMonitorCmd` -> repo.Save(...) -> queue.Enqueue(CheckJob{MonitorID,...}). Return 202 + HTMX swap.
(Worker) queue.Pop() -> resolve executor -> result := executor.Run(ctx, monitor) -> repo.RecordResult(result) -> notifier.Send(...)

## What NOT To Do

- Don’t perform monitor checks directly in API handlers.
- Don’t let infra types (Redis/Postgres clients) leak into domain code.
- Don’t create circular imports (respect direction: api -> domain & repos; worker -> domain & repos; domain -> stdlib only).

## Agent Response Expectations

- Before coding: restate intent + placement (package & file) succinctly.
- Prefer small, incremental PR-sized edits (one concern per change).
- Provide follow-up suggestions after implementing a request (tests, docs, edge handling).

Feedback Welcome: If any ambiguity (e.g., naming a new interface) surface 1–2 options, pick a default, proceed unless explicitly blocked.
