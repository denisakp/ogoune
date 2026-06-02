# ADR NNNN — <Title>

- **Status**: Proposed | Accepted | Superseded by ADR-XXXX | Deprecated
- **Date**: YYYY-MM-DD
- **Deciders**: <names>
- **Scope**: CE only | EE only | Both
- **Tags**: <storage, scheduler, license, frontend, agent, security, ...>
- **Supersedes**: ADR-XXXX (optional)

## Context

<!--
  Factual description of the situation in Ogoune today.
  What constraints apply (Go 1.25, dual-dialect SQLite+Postgres, CE single-binary
  vs EE API+worker+Redis, open-core Apache 2.0 + LicenseRef-Ogoune-EE,
  zero-telemetry CE promise, slice-based solo dev delivery)? What triggered this
  decision? Stay descriptive — no opinion in this section.
-->

## Decision drivers

- Driver 1 (e.g. "must work in CE single-binary with zero external deps")
- Driver 2 (e.g. "must not break dual-dialect SQLite/Postgres parity")
- Driver 3 (e.g. "must hold zero-telemetry CE commitment")

## Options considered

### Option A — short label

Description.

**Pros**
-

**Cons**
-

### Option B — short label

Description.

**Pros**
-

**Cons**
-

### Option C — (if applicable)

...

## Decision

<!--
  The choice, stated affirmatively in the present tense.
  "Ogoune uses sqlc for type-safe queries across SQLite and Postgres."
  NOT "We will use" or "It was decided to use".
-->

## Consequences

### Positive
-

### Negative
-

### Neutral / to watch
-

## Compatibility, migration & rollout

<!--
  Mandatory. Address whichever apply:
  - Dual-dialect impact: does this affect SQLite ↔ Postgres parity (schema, queries, drift checks)?
  - In-flight migration impact: does this collide with GORM→sqlc migration in progress?
  - CE ↔ EE impact: does this cross the open-core boundary (internal/ee/, license-gated)?
  - Spec drift: does any specs/NNN-name/plan.md need updating?
  - Doc drift: CLAUDE.md / AGENTS.md / README.md / BUSINESS-MODEL.md updates?
  - User-visible: API change, CLI/env change, DB migration, CHANGELOG.md entry needed?
  - Rollout: feature-flagged? gradual slice? hard cutover?
  If none apply, write "No compatibility impact."
-->

## Implementation checklist

- [ ] Concrete step 1 (reference real paths, e.g. `internal/repository/sqlc/queries/`)
- [ ] Concrete step 2 (tests, e.g. `internal/repository/internaltest/ForEachDialect`)
- [ ] Concrete step 3 (docs, e.g. `CLAUDE.md`, `docs/adrs/README.md` index, CHANGELOG)

## References

- Specs: `specs/NNN-name/plan.md`
- Related ADRs: ADR-XXXX
- External: RFCs, upstream docs, CVEs, business docs (`BUSINESS-MODEL.md`, `.private/STRATEGY.md`)
