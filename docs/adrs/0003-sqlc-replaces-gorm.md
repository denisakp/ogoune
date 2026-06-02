# ADR 0003 — sqlc replaces GORM for all repositories

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: Denis Akpagnonite
- **Scope**: Both
- **Tags**: storage, repositories, type-safety, migration

## Context

Ogoune started with **GORM** as the persistence layer for all `internal/repository/store/` implementations. GORM provided fast initial productivity: struct tags, automatic CRUD, hooks for `BeforeCreate` ULID assignment, and a unified Go API across SQLite and Postgres.

As the codebase matured to ~20 repositories with dual-dialect support (see [ADR-0002](./0002-dual-dialect-sqlite-postgres.md)), GORM's limitations compounded:

- **Runtime SQL generation** — queries are strings built by reflection at runtime. Typos, wrong column names, and joins against nonexistent tables fail at request time, not at build time.
- **Implicit cross-dialect behavior** — GORM emits subtly different SQL on SQLite vs Postgres for the same Go code, hiding bugs that only manifest in one engine.
- **Performance opacity** — N+1 queries, missing indexes, and inefficient generated SQL are invisible at code review.
- **Bench/profiling difficulty** — abstract layers make `EXPLAIN ANALYZE` hard to map back to call sites.
- **Schema drift risk** — GORM auto-migration was disabled long ago, but struct tags still claim to "know" the schema, which drifts from the migration source of truth.

A type-safe alternative that compiles SQL at build time would catch these classes of bugs before commit.

## Decision drivers

- Catch bad queries at `go build`, not at request time
- Make every query reviewable as plain SQL — no hidden behavior in a query builder
- Preserve dual-dialect SQLite + Postgres support ([ADR-0002](./0002-dual-dialect-sqlite-postgres.md))
- Allow incremental migration — repositories shipped one at a time, no big-bang
- Keep contract tests as the source of truth for behavior parity
- Solo dev — choose the boring, well-supported option

## Options considered

### Option A — Stay on GORM

**Pros**
- Zero migration cost
- Familiar to most Go developers

**Cons**
- All the runtime-error and dual-dialect-drift problems above persist forever
- Performance debt accumulates silently

### Option B — Hand-rolled `database/sql` with sqlx helpers

**Pros**
- Maximum control
- No code generation step
- Cross-dialect by SQL discipline

**Cons**
- Boilerplate for every Scan/Args mapping
- No compile-time verification of SQL syntax or column matches
- Tedious refactors when schema changes

### Option C — sqlc (query → typed Go code generation)

**Pros**
- SQL is the source — reviewable, copy-pasteable into `psql`
- Compile-time errors for wrong columns, types, or missing tables (parse-time, before `go build`)
- First-class support for SQLite and Postgres with separate query dirs
- Generated code is mechanical — no magic at runtime
- Trivial to benchmark (`EXPLAIN ANALYZE` the literal SQL)
- Generator pin via `make sqlc-generate`, drift caught by `make sqlc-check` in CI

**Cons**
- Two query directories per repository (`sqlite/`, `postgres/`) doubling SQL surface
- Generator version pinning matters (Go version compat — see `feat(044)` downgrade history)
- More files to commit per change
- Learning curve for the team (irrelevant solo dev, will matter post-hire)

### Option D — ent

**Pros**
- Graph-style API, type-safe queries
- Schema migrations included

**Cons**
- Still an ORM — same opacity class as GORM, just with codegen
- Less idiomatic SQL output
- Smaller community than sqlc

## Decision

Ogoune **migrates all repositories from GORM to sqlc**, progressively.

- SQL queries live in `internal/repository/sqlc/queries/{sqlite,postgres}/`.
- Generated Go code lives in `internal/repository/sqlc/{sqlite,pg}/` and is **committed**.
- `make sqlc-generate` regenerates after any `.sql` edit; `make sqlc-check` fails CI on drift.
- Repository implementations under `internal/repository/store/` move from GORM to sqlc one at a time, gated behind a per-repository feature flag (`SQLC_<NAME>`) until contract tests prove parity.
- Contract tests (`internal/repository/internaltest.ForEachDialect`) are the source of truth for behavior parity between GORM and sqlc implementations during the transition.
- Once a repository's sqlc implementation is green on both dialects across CE and EE paths, the GORM implementation and its flag are removed.

The migration ships in **waves** (pilot → wave 1 → wave 2 → wave 3 → decommission), each shipping under its own `specs/0NN-*` plan.

## Consequences

### Positive
- Query bugs caught at build time, not in production
- SQL is reviewable as SQL — no translation layer for an oncall to debug
- Dual-dialect divergence becomes explicit (two .sql files) instead of implicit (GORM dialect logic)
- Performance work becomes tractable — bench files target generated functions directly
- Schema source of truth tightens: migrations + sqlc queries, no GORM struct tags claiming to know schema

### Negative
- ~2x SQL files per repository (sqlite + postgres dirs)
- Migration period (waves) doubles repository implementations temporarily
- sqlc generator version pin is a CI fragility surface (e.g. Go 1.25 ↔ sqlc v1.30 compat)
- BeforeCreate ULID hook pattern from GORM must be replaced with explicit assignment at call sites or in a domain `Base.EnsureID()` helper (see `feat(043)`)

### Neutral / to watch
- If sqlc's roadmap stalls on a feature we need (e.g. specific Postgres operators), evaluate `pgx` direct calls behind the port interface as a per-query escape hatch
- Some advanced patterns (dynamic WHERE building, optional filters) sit awkwardly in sqlc — accept a small amount of hand-rolled SQL on those rare endpoints

## Compatibility, migration & rollout

- **Wire/API**: no impact. Internal refactor only.
- **DB schema**: no impact. Schemas (migrations) stay the same; only the Go-side query layer changes.
- **Feature flags**: each wave introduces `SQLC_<REPO>` boolean flags so production rollout is per-repository. Flags removed in the decommission spec once a wave is stable.
- **Tests**: contract tests are run against both GORM and sqlc implementations during the overlap window. Removal of the GORM implementation removes the flag and the GORM test wiring.
- **CI**: `make sqlc-check` blocks PR if generated code drifts from queries. `make ci-local` mirrors this lane.
- **Doc drift**: `CLAUDE.md` "sqlc" section, `internal/repository/sqlc/PATTERNS.md`, `specs/0NN-*/plan.md` per wave.
- **Rollout**: incremental over branches `041` (foundation) → `045` (pilot, tags) → `046` (wave 1) → `047` (wave 2) → `048` (wave 3) → `052` (decommission GORM).

## Implementation checklist

- [x] `feat(041)` — sqlc foundation, drift linter, generate/check Makefile targets
- [x] `feat(042)` — sqlc schema source aligned with migrations
- [x] `feat(043)` — domain decoupling (`Base.EnsureID()`) so ULIDs do not depend on a GORM hook
- [x] `feat(044)` — dual-dialect test infrastructure + CI `test-be-postgres` job
- [x] `feat(045)` — pilot: `TagsRepository` behind `SQLC_TAGS`
- [x] `feat(046)` — wave 1: 7 repositories
- [ ] `feat(047)` — wave 2: mid-complexity repositories (in progress)
- [ ] `feat(048)` — wave 3: `ResourceRepository`, `IncidentRepository`
- [ ] `feat(049)` — sqlc test infra + benches
- [ ] `feat(052)` — decommission GORM, remove flags and dependency

## References

- Specs: `specs/041-sqlc-foundation/`, `specs/042-sqlc-schema-source/`, `specs/043-domain-decoupling/`, `specs/045-pilot-tags-sqlc/`, `specs/046-wave1-sqlc-crud/`, `specs/047-wave2-sqlc-mid/`, `specs/052-decommission-gorm/`
- Commits: `eb4a2e3` (foundation), `0f7e3d0` (schema source), `cad6fdb` (domain decoupling), `aa92e0b` (pilot), `548800d` (wave 1)
- Patterns: `internal/repository/sqlc/PATTERNS.md`
- Related ADRs: ADR-0002 (dual-dialect — sqlc inherits and reinforces the parity model)
- External: [sqlc.dev](https://sqlc.dev), upstream issue tracker for SQLite/Postgres feature coverage
