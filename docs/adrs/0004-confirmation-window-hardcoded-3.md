# ADR 0004 — Hardcoded N=3 confirmation window before alerting

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: Denis Akpagnonite
- **Scope**: Both
- **Tags**: monitoring, incident, false-positive

## Context

Single-shot uptime monitors (Pingdom-style) page on every transient blip. Real production failures look identical to a momentary network hiccup at the check origin: one HTTP 503, one DNS timeout, one TCP RST. Paging on those produces alert fatigue, which destroys trust in the system.

Ogoune's core product differentiator is the **confirmation window**: a resource must fail N consecutive checks before an incident is created and notifications dispatched. This is encoded in `internal/monitoring/incident_service.go:55` — `CreateIncident` only fires when `r.FailureCount` reaches 3.

The question is: should N be hardcoded, configurable per resource, configurable globally, or configurable per check type?

## Decision drivers

- Confirmation-before-alerting is the product's promise — must be unambiguous
- Solo dev — fewer knobs = fewer support cases and fewer footguns
- A user who needs N=1 (paging immediately) is misaligned with the product
- A user who needs N=10 likely has a different problem (flap detection, not confirmation)
- Empirical: 3 consecutive failures at typical 60s intervals = ~3 min detection lag, which is the industry sweet spot

## Options considered

### Option A — Hardcoded N=3

**Pros**
- Unambiguous product behavior
- No misconfiguration possible
- Reduces support surface

**Cons**
- Users with strong opinions cannot tune it
- One-size-fits-all is technically suboptimal

### Option B — Configurable globally via env var (`CONFIRMATION_THRESHOLD=N`)

**Pros**
- Single tuning point
- Easy to roll back

**Cons**
- Tempts users to set N=1, killing the product's promise on their instance
- Documentation burden ("why doesn't Ogoune page faster?")

### Option C — Per-resource configurable

**Pros**
- Maximum flexibility
- Critical resources can have N=5, dev resources N=2

**Cons**
- UI complexity
- Most users will leave it at default; the rest will misconfigure
- Adds a column to a hot table (`resources`)

## Decision

Ogoune **hardcodes N=3** in `internal/monitoring/incident_service.go`. The value is not user-configurable in CE or EE.

Adjacent settings (interval, retries, timeout, **flap threshold**) remain configurable per resource — those govern check frequency and post-incident behavior, not the confirmation logic itself.

## Consequences

### Positive
- Product promise is mechanical, not negotiable
- Zero support burden on "tuning confirmation"
- Alert quality is consistent across the user base

### Negative
- Users with edge cases (very flaky public APIs they want flagged immediately, or very stable internal services they want 5+ confirmations on) have no escape hatch
- A future enterprise customer may demand per-resource override — at that point, evaluate as EE feature

### Neutral / to watch
- If the support channel surfaces repeated "N=3 is wrong for us" complaints from credible users, revisit as Option C (per-resource, default 3) behind a feature gate
- Document N=3 prominently in `README.md` and the in-product onboarding so users self-select

## Compatibility, migration & rollout

- **API/DB**: no change — `FailureCount` already exists on `resources`
- **CE ↔ EE**: identical behavior in both editions
- **Doc**: `CLAUDE.md` line 5 already states "N consecutive failures required" — keep `N=3` out of the public README intentionally so the value can be revisited later without breaking a documented contract
- **Rollout**: no migration; this ADR formalizes the existing behavior

## Implementation checklist

- [x] Hardcoded `3` in `internal/monitoring/incident_service.go`
- [x] `FailureCount int` field on `Resource` in `internal/domain/models.go`
- [x] Test coverage in `internal/monitoring/incident_service_test.go`
- [ ] Onboarding copy explains confirmation window so users do not expect single-shot paging

## References

- Code: `internal/monitoring/incident_service.go:19,55`
- Tests: `internal/monitoring/incident_service_test.go` (FailureCount transitions)
- Related: `Resource.FlapThreshold` (governs post-resolution flap suppression, distinct from confirmation)
