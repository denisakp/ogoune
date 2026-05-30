# Implementation Plan: Domain Decoupling — Drop GORM Tags + Hooks

**Branch**: `043-domain-decoupling` | **Date**: 2026-05-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/043-domain-decoupling/spec.md`

## Summary

Decouple `internal/domain/` from GORM at two surface points without changing any persisted behavior:

1. **ID assignment**: add a pure `Base.EnsureID()` method that today's `Base.BeforeCreate(tx)` delegates to; migrate all 7 in-memory fakes from `BeforeCreate(nil)` → `EnsureID()`.
2. **Encryption per concern**: add typed wrappers in `pkg/crypto/channel.go` (`EncryptChannelConfig`/`DecryptChannelConfig`) and `pkg/crypto/credential.go` (`EncryptCredentialPassword`/`DecryptCredentialPassword` + `…Options` pair). All wrap the existing `crypto.Encrypt`/`crypto.Decrypt` — no AES/GCM duplication, no on-disk format change. GORM hooks on `NotificationChannel` and `ResourceCredential` are reshaped to call the typed wrappers; the `AfterFind` legacy-plaintext lazy-migration write stays inside the GORM hook (where `tx` is available).
3. **Doc header**: head-of-file comment on `internal/domain/models.go` explaining the progressive tag-removal plan + pointer to `.prds/sqlc/`.

GORM tags and GORM hooks are NOT removed in this ticket (deferred to 006/007/008 + 010 per PRD). No domain schema change. No env var change.

## Technical Context

**Language/Version**: Go 1.25.1.
**Primary Dependencies**: Existing `pkg/crypto` (AES-256-GCM via `APP_SECRET_KEY`); `oklog/ulid/v2`; existing GORM hooks. No new module deps.
**Storage**: Encryption format MUST be byte-identical (FR-005, FR-006, SC-003).
**Testing**: Unit tests under `pkg/crypto/` for round-trip + oracle. Fake tests stay green. Full `make test-be` must pass with zero regression (SC-004).
**Target Platform**: unchanged.
**Project Type**: Single Go service + Vue SPA (frontend untouched).
**Performance Goals**: No measurable perf delta — wrappers are 1-line passthroughs.
**Constraints**:
- Byte-identical ciphertext.
- Zero GORM hook removal (FR-011), zero tag removal (FR-009).
- `EnsureID()` is pure and ID-only (FR-002).
- Typed wrappers are pure (FR-007a); lazy-migration write stays in `AfterFind`.
- Audit at impl time grep'd **7** fakes (`incident`, `notification`, `component`, `monitoring_activity`, `incident_event_step`, `api_key`, `resource`). PRD said 8 — re-grep before sign-off.
**Scale/Scope**:
- ~80 LOC new wrappers + `EnsureID()`.
- ~50 LOC fake edits.
- ~150 LOC tests.
- ~10 lines doc header.

## Constitution Check

| Principle | Verdict | Notes |
|-----------|---------|-------|
| I. Layered Boundary Integrity | PASS | Typed wrappers live in `pkg/crypto/`. Hooks remain in `internal/domain/`. Fakes still implement port interfaces. |
| II. Community Simplicity, Hosted Continuity | PASS | Pure additive; both runtime modes unchanged. No env var, no new dep, no operator-visible change. |
| III. Automated Verification for Runtime Changes | PASS | Persistence-adjacent → tests mandatory. SC-003 cross-version oracle = safety net against silent ciphertext-format regression. |
| IV. Migration and Startup Safety | PASS | No schema migration. Startup unchanged. Hooks still fire identically. |
| V. Spec-to-Execution Traceability | PASS | spec → clarify → plan → tasks chain in place. Fake audit + doc header explicitly tracked. |

No violations.

## Project Structure

### Documentation (this feature)

```text
specs/043-domain-decoupling/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── ensure-id.md
│   └── crypto-wrappers.md
└── checklists/requirements.md
```

### Source Code (repository root)

```text
ogoune/
├── internal/domain/
│   ├── models.go                                    # MODIFIED: header, EnsureID, hook bodies
│   └── resource_credential_hooks_test.go            # MAYBE: unchanged assertions
├── internal/repository/fake/
│   ├── incident_fake.go                             # MODIFIED: BeforeCreate(nil) → EnsureID()
│   ├── notification_fake.go                         # MODIFIED
│   ├── component_fake.go                            # MODIFIED
│   ├── monitoring_activity_fake.go                  # MODIFIED
│   ├── incident_event_step_fake.go                  # MODIFIED
│   ├── api_key_fake.go                              # MODIFIED
│   └── resource_fake.go                             # MODIFIED
├── pkg/crypto/
│   ├── crypto.go                                    # UNCHANGED
│   ├── channel.go                                   # NEW
│   ├── channel_test.go                              # NEW
│   ├── credential.go                                # NEW
│   └── credential_test.go                           # NEW
```

**Structure Decision**: Touches only `internal/domain/`, `internal/repository/fake/`, `pkg/crypto/`. Nothing under `internal/api/`, `internal/service/`, `internal/repository/store/`, or `web/`.

## Phase 0 — Research

See [`research.md`](./research.md). Resolved:

1. Existing `pkg/crypto` already centralizes AES/GCM — wrappers are 1-line passthroughs, zero format risk.
2. `Base.BeforeCreate` today is ID-only; `EnsureID()` matches.
3. `AfterFind` legacy lazy-migration write stays in GORM hook (Clarification Q2).
4. Fake audit: 7 sites confirmed; re-grep at impl time.
5. Cross-version oracle test = encrypt via existing `crypto.Encrypt`, decrypt via new wrapper, plus reverse.
6. Tests inject deterministic key via existing `crypto.SetGlobalProvider`.
7. Doc header wording locked.

## Phase 1 — Design & Contracts

- [`data-model.md`](./data-model.md) — `Base.EnsureID()` + typed-wrapper signatures
- [`contracts/ensure-id.md`](./contracts/ensure-id.md) — pure-ID semantics
- [`contracts/crypto-wrappers.md`](./contracts/crypto-wrappers.md) — wrapper signatures + delegation contract
- [`quickstart.md`](./quickstart.md) — maintainer workflow

Post-design Constitution re-check: all 5 PASS.

## Complexity Tracking

No violations.
