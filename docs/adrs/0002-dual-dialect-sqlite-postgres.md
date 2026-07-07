# ADR 0002 — Dual-dialect SQLite (CE) + Postgres (prod) with enforced parity

- **Status**: Accepted
- **Date**: 2026-05-30
- **Deciders**: Denis Akpagnonite
- **Scope**: Both
- **Tags**: storage, schema, migrations, ci

## Context

Ogoune ships two runtime modes:

| | Community | Production |
|---|---|---|
| DB | SQLite (in-process) | PostgreSQL |
| Scheduler | TimingWheel | Asynq (Redis) |
| Scaling | Single binary | Stateless API + workers |

Community Edition's hard requirement is **zero external dependencies** — a sysadmin downloads one binary and runs it. SQLite is the only viable embedded store. Production Edition needs concurrent writers, replication, and operational tooling — Postgres.

Initially, migrations were authored ad-hoc per dialect with implicit "keep in sync". This produced drift: columns named differently across trees, nullability mismatches, missing tables on one side. Bugs reproduced only on one dialect. CI caught nothing because tests ran SQLite-only.

The repository follows hexagonal layering (`internal/port/`, `internal/repository/store/`), so the same Go code must work against both engines. Any schema divergence breaks that promise.

## Decision drivers

- CE single-binary promise — no Postgres dependency in CE
- Production must scale beyond single-instance — Postgres is non-negotiable for hosted/EE
- Repository code (Go) must compile and behave identically against both engines
- Drift must be caught at PR time, not in production
- Solo dev — cost of maintaining two trees must be bounded and automated

## Options considered

### Option A — Single engine, force Postgres everywhere

**Pros**
- One schema, zero drift risk
- No abstraction overhead

**Cons**
- Kills CE zero-deps promise — sysadmins must install and operate Postgres
- Container/embedded use cases (Raspberry Pi, edge boxes) become impractical
- Direct contradiction of `BUSINESS-MODEL.md` CE commitment

### Option B — Single engine, force SQLite everywhere

**Pros**
- Truly trivial deploys
- No multi-engine logic

**Cons**
- No concurrent writers, no replication — SQLite cannot back a hosted multi-tenant Cloud
- Kills EE/Cloud roadmap entirely

### Option C — Dual dialect, manual sync, hope for the best

**Pros**
- Cheap to start

**Cons**
- Drift is a question of when, not if
- Solo dev cannot reliably eyeball two trees on every PR
- Bugs discovered in production are 100x more expensive than at PR time

### Option D — Dual dialect with automated drift enforcement

**Pros**
- Keeps both runtime modes viable
- Drift detected at PR by CI
- Migration authors get fast feedback locally via `make migrations-drift-check`

**Cons**
- Two sets of migration files to write per change
- Type tokens diverge (`JSONB` vs `TEXT`, `TIMESTAMPTZ` vs `TEXT`) — drift checker must be smarter than string equality
- Some Postgres-only features (jsonb operators, advanced triggers) require workarounds in SQLite

## Decision

Ogoune supports **both SQLite and Postgres as first-class backends**, with parity enforced mechanically:

1. Migrations live in two parallel trees: `internal/database/migrations/sqlite/` and `internal/database/migrations/postgres/`.
2. One migration = two files sharing the same `NNNN_` prefix and the same intent (one per dialect).
3. **Column name + nullability MUST match** across dialects, enforced by `make migrations-drift-check`. Type tokens are intentionally **not** enforced cross-dialect (`JSONB` ↔ `TEXT`, `TIMESTAMPTZ` ↔ `TEXT`, `BIGINT` ↔ `INTEGER` are accepted equivalents).
4. JSON columns: `JSONB` on Postgres, `TEXT` on SQLite. Cross-driver Go fields use serializers, not direct typing.
5. SQLite-specific constraints documented: no `ADD COLUMN IF NOT EXISTS`, no multi-column `ALTER TABLE`, `PRAGMA`/triggers require tech-lead validation.
6. Repository tests are **dual-dialect** by default via `internal/repository/internaltest.ForEachDialect`. A repository test that runs only SQLite is the exception, not the rule.
7. CI gates: `make migrations-drift-check` before tests; `test-be-postgres` job runs the dual-dialect matrix on every PR (testcontainers).

## Consequences

### Positive
- CE keeps the zero-deps promise
- Hosted/EE keeps Postgres power
- Drift bugs caught at PR by `migrations-drift-check`, not discovered in prod
- `ForEachDialect` makes new repository tests automatically cross-dialect

### Negative
- Every schema change costs two SQL files instead of one
- Authors must remember the SQLite gotchas list — automated by drift checker but still a learning curve
- Dual-dialect test job adds ~1-2 min to CI per PR

### Neutral / to watch
- Postgres-only features (advanced jsonb, generated columns, partitioning) will create pressure — when needed, isolate behind a port interface so SQLite uses a slower fallback
- If drift checker becomes too lenient (false negatives), production bugs return — review enforcement rules during each schema audit

## Compatibility, migration & rollout

- **Existing data**: no impact — both engines already supported. This ADR formalizes and enforces what was implicit.
- **Existing migrations**: backfilled to comply with drift checker when introduced (see commit history of `feat(044)`).
- **New repositories**: must use `port` interface + dual-dialect contract test by default.
- **Doc drift**: `CLAUDE.md` "Database migrations" section, `internal/database/migrations/README.md`, `internal/repository/internaltest/README.md` all reflect this decision.
- **Rollout**: hard policy from `feat(044)` onward. PRs failing drift check are blocked at CI.

## Implementation checklist

- [x] `internal/database/migrations/sqlite/` and `postgres/` parallel trees
- [x] `make migrations-drift-check` (file pair + column name + nullability)
- [x] `internal/repository/internaltest/ForEachDialect` helper
- [x] CI job `test-be-postgres` runs dual-dialect matrix per PR (testcontainers)
- [x] `make test-be-pg` local equivalent (Docker required)
- [x] Type-mapping table in `internal/database/migrations/README.md`
- [x] `CLAUDE.md` gotchas section
- [ ] Annual audit: revisit drift checker rules for false-positive/negative balance

## References

- Specs: `specs/044-dual-dialect-test-infrastructure/`
- Commits: `0491c8f` (dual-dialect infra), `c1c0e09` (CI integration)
- Related ADRs: ADR-0003 (sqlc replaces GORM — depends on this dialect strategy)
- Internal docs: `internal/database/migrations/README.md`, `internal/repository/internaltest/README.md`
