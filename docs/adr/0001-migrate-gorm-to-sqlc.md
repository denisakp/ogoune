# 0001. Migrate from GORM to sqlc

- **Status**: Accepted
- **Date**: 2026-06-02
- **Authors**: Denis AKPAGNONITE (@from-togo)
- **Supersedes**: none (first ADR)
- **Superseded by**: none

## Context

Ogoune started with GORM v2 as the persistence layer. GORM gave the project rapid CRUD scaffolding, automatic schema generation via struct tags + `AutoMigrate`, and lifecycle hooks (`BeforeCreate`, `AfterFind`, etc.) that bundled cross-cutting concerns (ID generation, encryption, decryption) onto the domain models themselves. For the first ~40 specs that was the right trade-off.

Three pressures accumulated:

1. **Dual-dialect drift**. Ogoune ships in two modes: Community (AGPL, SQLite, in-process) and Enterprise (PostgreSQL, hosted-future). GORM hides the dialect at the call site, but the abstractions leak — `JSONB` vs `TEXT` for JSON columns, `TIMESTAMPTZ` vs `TEXT` for timestamps, `pgvector`-shaped columns that have no SQLite equivalent. Bugs landed where the GORM tags were correct on one driver and silently wrong on the other. Discovered painfully in resource credential storage and notification channel config encryption.

2. **Opaque preloads + N+1 surprises**. GORM's `Preload("Tags")` does a separate SELECT per relation; this is fine until you hit the API list path under load. Spec 049 introduced a paired-bench harness and immediately measured a **4.08× regression** of `sqlc Resource.List` vs `GORM Resource.List` on Postgres. Spec 050 audited the call site and discovered three of the four preloads were dead code — no consumer read them. Deleting the dead preloads closed the gap to **1.64×**. That kind of audit is intractable against an ORM that hides its preloads behind a single `.Find()` call; sqlc forces you to write the queries you actually run.

3. **Hooks doing real work, invisibly**. `NotificationChannel.AfterFind` decrypted the encrypted `Config` column on every read. Subtle: the in-memory shape and the persisted shape differ; if a caller forgets the hook is the boundary, they get ciphertext when they expected plaintext or vice versa. We had several incidents where the hook was the right answer for the wrong reason. sqlc forces the encrypt/decrypt boundary to be explicit (`encryptChannelConfig` in `Create`, `decryptChannelConfig` after `Scan`).

The migration ran across 11 specs over ~9 months:

| Spec | What |
|---|---|
| 041 | sqlc foundation + dual-dialect generator config |
| 042 | Schema-as-source-of-truth — `migrations-drift-check` |
| 043 | Domain decoupling — remove leaky GORM types from service/handler layers |
| 044 | Test infra — `internaltest.ForEachDialect` for dual-dialect contract tests |
| 045 | Pilot — `Tags` repository, behind `SQLC_TAGS` flag |
| 046 | Wave 1 — 7 simple CRUD repos behind per-repo flags |
| 047 | Wave 2 — 5 mid-complexity repos (M2M, ClaimPending, aggregations) |
| 048 | Wave 3 — `Resource` + `Incident` (preloads, controlled N+1) |
| 049 | Test infra v2 — paired-bench harness, soak tests, CI flag matrix |
| 050 | Resource.List perf fix — drop dead preloads |
| 051 | Dynamic filters (squirrel + dynquery package) |
| **052** | **GORM decommission (this ADR)** |

Each spec landed independently behind a flag so production traffic could move repository-by-repository. By spec 052, every sqlc impl had been validated for months against its GORM sibling via the CI flag-matrix lanes. The dual-impl carried a permanent maintenance tax — every schema change touched two implementations, the CI matrix had 4 lanes, the paired-bench harness reported a comparison nobody acted on anymore.

This ADR records the decision to physically remove GORM.

## Decision

We migrate every repository from GORM to sqlc, then physically delete the GORM code in spec 052. Post-migration, the project uses:

- **sqlc** for all generated typed queries (under `internal/repository/sqlc/{pg,sqlite}/`).
- **`pgx/v5`** + **`pgxpool`** as the Postgres driver — production handle is `pgxpool.Pool` for sqlc-generated queries; an `stdlib.OpenDBFromPool` wrapper exposes a `*sql.DB` for migrations and the startup schema validator.
- **`modernc.org/sqlite`** (pure-Go) as the SQLite driver, via `database/sql`.
- **`Masterminds/squirrel`** for dynamic-filter query construction (spec 051 — the `dynquery` package).
- Hand-written wrappers in `internal/repository/store/*_sqlc.go` that satisfy the `port.*Repository` interfaces, with shared helpers for controlled-N+1 preloads.

Schema authority moves to the raw SQL migration files under `internal/database/migrations/{postgres,sqlite}/`. `gorm.AutoMigrate` is gone; a thin `database/sql` runner applies the files. `make migrations-drift-check` enforces that file pairs + column names + nullability stay consistent across dialects.

## Consequences

### Positive

- **One impl, one path**. Schema changes touch one file pair (`pg.sql` + `sqlite.sql` queries) instead of two implementations + tag drift.
- **Compile-time type safety**. `pgtype.Timestamptz` vs `time.Time` mismatches caught at build, not at runtime when a JSON serialization layer disagrees.
- **Hot paths are competitive**. Post-spec-050 + spec-052, `Resource.List` p95 is 1.64× sqlc vs GORM on PG (down from initial 4.08×). `Incident.GetIncidentStats` is **0.58×** — sqlc faster than GORM, via a CTE one-pass that replaced two correlated sub-queries.
- **Explicit boundaries**. Encryption, ID generation, timestamp assignment are all explicit in the sqlc Create/Update wrappers. No magic hooks.
- **Boot path simplified**. No `SQLC_*` flag matrix; one less branching point at startup.
- **Operator config simpler**. 16 deprecated env flags removed from `.env.example`.

### Negative

- **Wrapper boilerplate is more verbose than GORM's reflection-based `Create(&x)`**. Each repository method is ~20-40 LoC instead of ~5. Mitigated by `internal/repository/sqlc/PATTERNS.md` documenting recurring shapes.
- **Dual-dialect divergence lives in queries, not in struct tags**. Slightly more cognitive load per query. Mitigated by `make migrations-drift-check` enforcing column-name + nullability parity.
- **Loss of GORM's "magic" generic helpers** (Scopes, Hooks). Accepted — the magic was the source of the bugs we wanted out.
- **modernc.org/sqlite quirks**. Compared to glebarez/sqlite (GORM's previous dialector), modernc binds `time.Time` parameters in Go's default `String()` format rather than RFC 3339. Required a `substr(col, 1, 19)` workaround in one query (`FindMissedHeartbeatsSQLite`) — see commit message of fix(052) on branch `052-decommission-gorm` for the analysis.

### Operational

- `v2.0.0-rc.1` published with a 7-day community-feedback window before the final `v2.0.0` tag (spec 052 Q1 clarification — there is no SaaS hosted by us today, so "canari" means RC + observation, not blue/green).
- Legacy `SQLC_*` env flags are silently ignored — no operator action required to upgrade (spec 052 FR-004, regression test `TestLegacyFlagsSilentlyIgnored`).
- Rollback to v1.x is a redeploy — schema is forward-compatible, no migration delta.

## Alternatives Considered

### 1. Keep GORM, fix incrementally

**Rejected.** The dual-dialect bugs continued accumulating, and the perf regressions were silently hidden behind GORM's preload semantics. The original problem doesn't go away by waiting; you just pay the dual-impl tax longer.

### 2. ent (`entgo.io/ent`)

**Rejected.** Trades GORM's runtime reflection for ent's code generation + graph traversal model. Solves type safety but introduces a heavier abstraction (entity graph, schema-as-code) — opposite direction from "explicit SQL, thin wrappers". Also weaker dual-dialect support at the time of decision.

### 3. Raw `database/sql` + hand-written mappers

**Rejected.** Solves all the same problems sqlc does, plus more flexibility, but the hand-mapping boilerplate is 3× what sqlc generates. sqlc gives us the type-safety win at a fraction of the maintenance cost.

### 4. SQLBoiler

**Rejected.** Schema-first like sqlc, but the generated API leans on its own query DSL rather than letting us write SQL directly. We wanted SQL to remain readable and reviewable as-is.

### 5. Hybrid (GORM for writes, sqlc for hot reads)

**Rejected.** All the worst of both. Twice the maintenance, doubled cognitive load on contributors, doubled CI test surface. We considered this seriously in spec 047 retro — voted down unanimously.

## Lessons Learned

(Compiled from prior spec retros 045-051 + spec 052's experience.)

1. **Invest in measurement tooling early.** The 4.08× sqlc-vs-GORM regression on `Resource.List` would not have been caught without spec 049's paired-bench harness measuring both impls in the same process on the same fixture. The harness is now retired (no comparant) but it earned its keep many times.

2. **Audit consumers before optimising query shape.** The 4.08× → 1.64× perf gain in spec 050 came not from any new pattern, but from **deleting preloads that no caller read**. Always grep callers of the hot method before architecting an optim. The bench was measuring fictional work that GORM happened to be good at.

3. **Controlled N+1 is fine when bounded.** 1 main + N preloads is acceptable as long as N is bounded by the number of associations (constant), not by row count. Spec 049's round-trip-bound tests are the regression guard, and they're retained post-decom.

4. **Dual-dialect type tokens cannot be enforced cross-dialect.** `TIMESTAMPTZ` (PG) vs `TEXT` (SQLite) is intentional and necessary. Only column NAMES and NULLABILITY can be enforced — captured in `make migrations-drift-check`.

5. **Hooks doing real work hide their work.** Every `BeforeCreate` / `AfterFind` we deleted in spec 052 was carrying load-bearing logic (ID gen, encryption). Migrating that logic to explicit wrapper calls forced us to audit it (research §3 grep audit before the deletion). That audit caught zero issues — but it could have caught one, and we'd never have known without the explicit migration.

6. **Speckit traceability paid for itself.** 11 specs over 9 months stayed coherent because every change had a spec + plan + tasks + retro. Recommend keeping this for future major chantiers.

7. **Decommission deserves its own spec.** The deletion is half the chantier's value — it's where the maintenance tax stops. Don't fold it into the last migration spec; give it its own scope, ADR, and RC window. Spec 052 was 62 tasks, ~6500 lines deleted, ~900 added; if we'd tried to slide it under spec 051 it would have collapsed scope.

8. **modernc.org/sqlite is not glebarez/sqlite.** Time-binding differs. Test the SQLite path explicitly; don't assume "all SQLite drivers behave the same."

## References

- `specs/041-sqlc-foundation/` through `specs/051-dynamic-filters/` — the full chantier
- `specs/052-decommission-gorm/retro.md` — this chantier's retro, with bench numbers
- `specs/052-decommission-gorm/spec.md` — this decommission's spec
- `internal/repository/sqlc/README.md` — contributor onboarding (post-decom)
- `internal/repository/sqlc/PATTERNS.md` — canonical implementations of recurring patterns
- `.prds/sqlc/010-decommission.md` — the original PRD that triggered this spec
