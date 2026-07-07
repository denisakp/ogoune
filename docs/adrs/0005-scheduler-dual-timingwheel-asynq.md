# ADR 0005 — Dual scheduler: TimingWheel (CE) and Asynq (production)

- **Status**: Accepted
- **Date**: 2026-05-29
- **Deciders**: Denis Akpagnonite
- **Scope**: Both
- **Tags**: scheduler, runtime, ce-ee, async

## Context

Ogoune schedules thousands of monitor checks per minute. The scheduler is on the hot path of the product — it must be reliable, observable, and operationally appropriate for both runtime modes:

- **Community Edition** runs as a single binary with zero external dependencies (see [ADR-0002](./0002-dual-dialect-sqlite-postgres.md)). It cannot depend on Redis.
- **Production / EE** runs API and worker processes separately and needs durability across restarts, horizontal worker scaling, and operational visibility.

A single scheduler implementation cannot satisfy both modes.

## Decision drivers

- CE single-binary, zero external dependencies
- Production: durable, horizontally scalable, observable
- Same Go port interface (`port.Scheduler`) so callers stay identical
- Solo dev — bound the maintenance cost of running two implementations
- Tests must validate behavioral parity between schedulers

## Options considered

### Option A — Single scheduler everywhere (Redis-backed)

**Pros**
- One code path, zero parity surface

**Cons**
- Kills CE zero-deps promise — sysadmins must run Redis
- Embedded/edge use cases impractical

### Option B — Single scheduler everywhere (in-process timing wheel)

**Pros**
- True single binary

**Cons**
- No durability across restarts
- No horizontal worker scaling
- Production-grade observability and at-least-once semantics absent

### Option C — Two implementations behind a shared `port.Scheduler` interface

**Pros**
- Both modes get the right tool
- Callers never know which is running
- Parity testable in CI

**Cons**
- Two implementations to maintain
- Subtle semantic differences (at-most-once vs at-least-once) must be hidden behind the interface
- Tests must cover both paths

## Decision

Ogoune ships **two scheduler implementations** behind `port.Scheduler`:

- **TimingWheel** (`internal/scheduler/timingwheel.go`) — in-process hierarchical timing wheel. Default for CE. No external dependencies. State is process-local; restart loses scheduled jobs and reschedules from DB on boot.
- **Asynq** (`internal/scheduler/asynq.go`) — Redis-backed task queue. Default for production / EE. Durable, supports horizontal worker pools, has built-in observability (Asynqmon).

A factory (`internal/scheduler/scheduler_factory_test.go` pattern, runtime selection in bootstrap) selects the implementation from `SCHEDULER_MODE` env var (`timingwheel` or `asynq`).

**Parity tests** (`internal/scheduler/asynq_parity_test.go`) exercise both implementations against the same behavioral expectations: confirmation window propagation, pause/resume, reschedule, shutdown, startup recovery.

## Consequences

### Positive
- CE keeps single-binary deploys (Raspberry Pi, edge, dev laptops)
- Production gets durable scheduling and horizontal worker scale
- `port.Scheduler` interface keeps services / handlers oblivious of the runtime mode
- Parity tests catch divergence at PR time

### Negative
- Two implementations, two paths to maintain
- Behavioral edge cases (at-most-once on TimingWheel restart vs at-least-once on Asynq) must be documented and surfaced in incident postmortems
- Asynq generates more test scaffolding (Redis testcontainers)

### Neutral / to watch
- If TimingWheel restart loss becomes a CE pain point, evaluate WAL-style persistence to SQLite (not yet justified by user reports)
- If Asynq's roadmap stalls, the port interface lets us swap to River or another Redis-backed queue with bounded blast radius

## Compatibility, migration & rollout

- **API surface**: `port.Scheduler` is the contract. Adding a method requires updating both implementations and their tests in lockstep.
- **Config**: `SCHEDULER_MODE` env var (`timingwheel` default for CE, `asynq` for production)
- **Doc drift**: `CLAUDE.md` "Two runtime modes" table, `.env.example`
- **Tests**: any new scheduler behavior MUST land with parity coverage in `internal/scheduler/asynq_parity_test.go` or equivalent

## Implementation checklist

- [x] `port.Scheduler` interface in `internal/port/`
- [x] `internal/scheduler/timingwheel.go` (CE default)
- [x] `internal/scheduler/asynq.go` (production default)
- [x] Factory wiring in `internal/platform/bootstrap/`
- [x] `internal/scheduler/verify.go` compile-time interface check for both implementations
- [x] Parity tests `asynq_parity_test.go`
- [x] TimingWheel-specific tests for confirmation, pause/resume, reschedule, saturation, shutdown
- [ ] Document at-most-once vs at-least-once semantics in runbook

## References

- Code: `internal/scheduler/{timingwheel,asynq,scheduler,adapter}.go`
- Tests: `internal/scheduler/*_test.go` (parity + per-impl)
- Related: ADR-0002 (dual-dialect — same "CE zero-deps vs prod scale" tension)
- External: [Asynq docs](https://github.com/hibiken/asynq), hierarchical timing wheel paper (Varghese & Lauck, 1987)
