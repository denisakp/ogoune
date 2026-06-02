# Changelog

All notable changes to this project will be documented in this file. Format
follows [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/).

## [2.0.0-rc.1] - 2026-06-XX

Release-candidate. **GORM removed; sqlc is the only ORM.** Closes the 9-month
sqlc chantier (specs 041–052). See
[ADR 0001](docs/adr/0001-migrate-gorm-to-sqlc.md) for the decision record and
lessons learned.

A 7-day community-feedback window opens at RC tag. The final `v2.0.0` follows
only if no GORM-removal regression is reported during the window. Report
issues on the GitHub Discussion thread linked in the release notes.

### Changed

- Migrated all repositories from GORM to sqlc-generated query bindings. Every
  `internal/repository/store/*_sqlc.go` wrapper is now the sole implementation
  of its `port.*Repository` interface.
- Postgres now driven directly by `pgx/v5` + `pgxpool` (no more
  `database/sql` wrapping). SQLite driven by `modernc.org/sqlite` via
  `database/sql` (no more `glebarez/sqlite` GORM dialector).
- Migrations are now applied by a thin `database/sql` runner over the
  existing SQL files (`internal/database/migrations/{postgres,sqlite}/`).
  `gorm.AutoMigrate` is gone. Startup fails fast on any apply error and wraps
  the failing file + dialect in the returned error.
- `Runtime.GormDB()` accessor removed. Only `Runtime.PgxPool()` (Postgres)
  and `Runtime.SQLiteDB()` (SQLite) remain.

### Removed

- **GORM dependency** (`gorm.io/gorm`, `gorm.io/driver/postgres`,
  `gorm.io/driver/sqlite`) and `github.com/glebarez/sqlite` from `go.mod`.
- **~115 `gorm:"..."` struct tags** from `internal/domain/models.go`.
- **7 GORM lifecycle hooks**: `Base.BeforeCreate`,
  `NotificationChannel.{BeforeCreate,BeforeUpdate,AfterFind}`,
  `ResourceCredential.{BeforeCreate,BeforeUpdate,AfterFind}`. The pure logic
  they wrapped (ID generation, encryption, decryption) now lives in the sqlc
  Create/Update wrappers explicitly. `EnsureID()` is the remaining plain
  method on `Base`.
- **15 legacy `*_repository.go` GORM impl files** under
  `internal/repository/store/`. Their `*_sqlc.go` siblings are now the sole
  implementations.
- **Paired-bench harness** from spec 049
  (`internal/repository/store/*_bench_test.go`, the `RunPairedBench` /
  `SeedPairedBenchFixture` helpers, the `make test-be-bench` target).
  Round-trip-bound tests are retained — they validate sqlc in isolation.
- **3 CI flag-matrix lanes** (`test-be-sqlc-tags`, `test-be-sqlc-wave1`,
  `test-be-sqlc-wave3`) from `.github/workflows/{ci,test}.yml`. The
  concurrent-update soak step moves into the main `test` / `backend-tests`
  lane.

### Deprecated

- **Legacy `SQLC_*` environment flags** (16 vars: `SQLC_TAGS`,
  `SQLC_API_KEY`, `SQLC_USER`, `SQLC_NOTIFICATION_CHANNEL`,
  `SQLC_EXPIRY_NOTIFICATION_LOG`, `SQLC_STATUSPAGE_SETTINGS`,
  `SQLC_INCIDENT_DIAGNOSTICS`, `SQLC_RESOURCE_CREDENTIAL`,
  `SQLC_COMPONENT`, `SQLC_MAINTENANCE`, `SQLC_NOTIFICATION`,
  `SQLC_MONITORING_ACTIVITY`, `SQLC_INCIDENT_EVENT_STEP`, `SQLC_RESOURCE`,
  `SQLC_INCIDENT`). These are no longer read by the binary. They are silently
  ignored — leaving them in your `.env` is safe and has no effect. You may
  delete them at your convenience. The regression test
  `TestLegacyFlagsSilentlyIgnored` guards this behaviour going forward.

### Performance

- `GET /api/v1/monitors` p95 latency unchanged within ±10 % vs the v1.x
  baseline captured immediately before the GORM-removal commit (numbers in
  `specs/052-decommission-gorm/retro.md`).
- `GET /api/v1/incidents` p95 latency unchanged within ±10 %.
- Cold-boot p95 unchanged within ±10 %.

### Documentation

- **New ADR**: `docs/adr/0001-migrate-gorm-to-sqlc.md` — the formal decision
  record covering context, alternatives considered (GORM kept, ent, raw
  `database/sql`, SQLBoiler, hybrid), and 8 lessons learned compiled from
  prior spec retros.
- **New contributor guide**: `internal/repository/sqlc/README.md` — 9-step
  walkthrough for adding a new repository, sqlc-only.
- **New patterns catalogue**: `internal/repository/sqlc/PATTERNS.md` —
  canonical implementations of recurring shapes (M2M, controlled-N+1
  preload, dynamic filters via `dynquery`, transactions, encrypted columns,
  dialect-specific aggregates).
- **Updated `CLAUDE.md`**: "Adding a new repository" section rewritten as
  10-step sqlc-only recipe; "Domain models" + "Repositories" + "Database
  migrations" sections updated.

### Breaking changes

- None to the public HTTP API. `/api/v1/*` is byte-identical.
- None to the database schema. v2.0.0-rc.1 is forward-compatible with v1.x
  (no migration delta). Rollback is a redeploy; no DB surgery needed.
- Internal: any external Go consumer importing
  `internal/repository/store.*RepositoryImpl` will fail to compile. Go's
  `internal/` package convention forbids this so n/a in practice.

### Upgrade notes

1. Pull the v2.0.0-rc.1 image / tag.
2. Boot as usual. If your `.env` still has `SQLC_*=...` lines, no change
   needed — they're silently ignored.
3. Watch the monitor + incident list endpoints over the first 24-48h.
   Report any GORM-removal regression on the linked GitHub Discussion
   within the 7-day RC window.

---

## [Unreleased]

### Added

- **Keyword / content check monitor** — new monitor type (`keyword`) that performs an HTTP GET and verifies the response body contains or does not contain a literal string. Detects content failures that return HTTP 200 with degraded content.
- **`contains` / `not_contains` modes** — case-sensitive keyword matching. `contains` raises an incident when the keyword is absent; `not_contains` raises one when the keyword is present.
- **512 KB body cap** — the strategy reads at most 512 KB of the response body; content beyond this limit is silently discarded and `body_truncated` is set to `true`.
- **Keyword failure diagnostics** — `IncidentDiagnostics` extended with `keyword`, `keyword_mode`, and `keyword_found` fields, populated for keyword monitor incidents only.
- **Enriched notifications** — alert emails and webhook payloads include keyword, match mode, and human-readable cause message for keyword monitor failures (FR-013).
- **Incident detail view** — keyword diagnostics panel in `IncidentView.vue` showing keyword, mode, match result, body excerpt, body size, and truncation flag.
- **DB migration `0011_keyword_fields`** — additive nullable columns on `resources` and `incident_diagnostics` for both SQLite and PostgreSQL.
