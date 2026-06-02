# ADR 0008 — ULIDs as primary IDs across all domain entities

- **Status**: Accepted
- **Date**: 2026-05-30
- **Deciders**: Denis Akpagnonite
- **Scope**: Both
- **Tags**: schema, identifiers, domain

## Context

Every domain entity in `internal/domain/models.go` (`Resource`, `Incident`, `NotificationChannel`, `Tag`, etc.) needs a globally unique primary identifier. The ID is exposed in URLs, logs, exports, and the v1 API contract — changing it later is expensive.

Original implementation used a GORM `BeforeCreate` hook to assign the ID at insert time. The migration from GORM to sqlc ([ADR-0003](./0003-sqlc-replaces-gorm.md)) removed that hook layer; assignment is now done explicitly via `Base.EnsureID()` (`internal/domain/models.go:28`) called from service wrappers (see commit `cad6fdb`).

The identifier choice is independent of the assignment mechanism. This ADR records why **ULID** rather than UUIDv4, UUIDv7, auto-increment, Snowflake, or KSUID.

## Decision drivers

- Must be globally unique across CE single-binary and EE multi-instance deploys
- Must be safe to expose in URLs and logs (no security info leak)
- Must sort roughly by creation time for human debugging (`ORDER BY id` ≈ time order)
- Must be compact textually — used in JSON payloads, log lines, monitoring tags
- Must work cross-dialect on SQLite TEXT and Postgres TEXT/varchar (see [ADR-0002](./0002-dual-dialect-sqlite-postgres.md))
- Must be assigned client-side (Go-side) before INSERT — no `RETURNING id` ping-pong, no DB-side default function

## Options considered

### Option A — Auto-increment integers

**Pros**: smallest storage, fastest index.
**Cons**: leaks total count, predictable URLs, breaks across-instance uniqueness in EE multi-writer scenarios, fragile under merge/import.

### Option B — UUIDv4

**Pros**: ubiquitous, well supported.
**Cons**: random — no time order, harmful for index locality (random insert position causes B-tree page splits), 36-char canonical form is bulky.

### Option C — UUIDv7

**Pros**: time-ordered prefix, UUID compatibility for libraries expecting UUID.
**Cons**: spec finalised recently — ecosystem support uneven at decision time; still 36-char canonical with hyphens; minor gain over UUIDv4 for our scale.

### Option D — Snowflake / TSID

**Pros**: 64-bit, time-ordered, compact.
**Cons**: requires worker ID coordination — fights the CE single-binary deployment model.

### Option E — ULID

**Pros**: time-ordered prefix (48-bit millisecond timestamp + 80-bit randomness), 26-char Crockford base32 (URL-safe, case-insensitive, no hyphens), lexicographically sortable as plain text, no central coordinator, mature Go library (`github.com/oklog/ulid/v2`).
**Cons**: not a UUID — some external systems require UUID format; monotonic generator needs care under concurrent inserts (handled by `ulid.Monotonic`).

## Decision

Ogoune uses **ULID v2** (`github.com/oklog/ulid/v2`) as the primary identifier for every domain entity.

- Storage: stringified Crockford base32 (26 chars) in a TEXT column on both SQLite and Postgres
- Generation: client-side via `Base.EnsureID()` in `internal/domain/models.go`, using `ulid.Monotonic` entropy keyed on the current time
- Assignment timing: before persistence, in service-layer wrappers around repository writes (previously in a GORM `BeforeCreate` hook — equivalent semantics, explicit call site after [ADR-0003](./0003-sqlc-replaces-gorm.md))
- Idempotency: `EnsureID()` is a no-op if `b.ID` is already set, so callers can pre-assign for tests or imports

## Consequences

### Positive
- Time-ordered IDs make `ORDER BY id` a free approximation of `ORDER BY created_at` for debugging
- Index locality stays reasonable — new inserts append at the end of the B-tree
- URL-safe and log-safe out of the box
- Identical column type (`TEXT`) across SQLite and Postgres — zero dialect divergence
- No DB-side function dependency — `id` column has no default, application owns the value

### Negative
- ULID is not a UUID — external systems demanding UUID format need a converter (rare in Ogoune's domain)
- 26-char text is larger than a 16-byte UUID binary representation; storage cost accepted for human readability

### Neutral / to watch
- If a future use case demands UUID compatibility (e.g. SCIM integration), `Base.EnsureID()` is the single rewrite point
- Monotonic generator behavior under high concurrency is library-dependent — covered by `internal/domain/base_ensure_id_test.go`

## Compatibility, migration & rollout

- **Schema**: `id TEXT PRIMARY KEY` on every entity table, both dialects. No migration needed; this has been the contract since the project's start.
- **GORM → sqlc**: ID assignment moved from `BeforeCreate` hook (GORM-coupled) to explicit `EnsureID()` calls in service wrappers (post-`feat(043)`). Behaviorally identical; tests in `internal/domain/base_ensure_id_test.go` lock the contract.
- **API contract**: ULIDs already exposed publicly in v1 API responses — no client change
- **Doc**: `CLAUDE.md` "Domain models" section notes "IDs are ULIDs"

## Implementation checklist

- [x] `Base.EnsureID()` in `internal/domain/models.go:28`
- [x] `ulid.Monotonic` entropy for safe concurrent generation
- [x] Tests `internal/domain/base_ensure_id_test.go` (assigns when empty, preserves when set)
- [x] Service wrappers call `EnsureID()` before repository writes (post-`feat(043)`)
- [x] `CLAUDE.md` documents the convention
- [ ] If/when UUID interop is required, add a typed converter in `internal/domain/` rather than changing the ID storage

## References

- Code: `internal/domain/models.go:28` (`Base.EnsureID`)
- Tests: `internal/domain/base_ensure_id_test.go`
- Commit: `cad6fdb` (`feat(043): domain decoupling — Base.EnsureID()`)
- Related ADRs: ADR-0003 (sqlc migration that motivated extracting the ID hook)
- External: [ULID spec](https://github.com/ulid/spec), [`oklog/ulid/v2`](https://github.com/oklog/ulid)
