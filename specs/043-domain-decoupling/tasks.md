---

description: "Task list for 043 Domain Decoupling"
---

# Tasks: Domain Decoupling — Drop GORM Tags + Hooks

**Input**: Design documents from `/specs/043-domain-decoupling/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: REQUIRED — touches encryption surfaces (Constitution III). Round-trip + cross-version oracle is the safety net against silent ciphertext-format regression.

**Organization**: Tasks grouped by user story so each can be implemented and shipped independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Different file, no incomplete deps
- **[Story]**: Maps to spec user story (US1, US2, US3)
- Paths repo-relative under `/Users/yaovi/Projects/perso/ogoune/`

## Path Conventions

Single Go service + Vue SPA (frontend untouched). Changes scoped to `internal/domain/`, `internal/repository/fake/`, `pkg/crypto/`.

---

## Phase 1: Setup

**Purpose**: No code yet — surface inputs only.

- [X] T001 Re-grep `internal/repository/fake/` for `BeforeCreate(nil)` callers and confirm the inventory matches plan.md's expected 7 sites (`incident`, `notification`, `component`, `monitoring_activity`, `incident_event_step`, `api_key`, `resource`). If the grep returns additional files, add them to the US1 implementation tasks and update the count in the PR description
- [X] T002 Verify `pkg/crypto/crypto.go` exposes `Encrypt(string) (string, error)`, `Decrypt(string) (string, error)`, `SetGlobalProvider(KeyProvider)`. If any are missing, halt and surface the gap before US2 starts

---

## Phase 2: Foundational

**Purpose**: None — US1 and US2 are independent. Phase 2 left intentionally empty to keep the structure consistent with prior tickets.

_(no tasks)_

---

## Phase 3: User Story 1 — `Base.EnsureID()` + fakes migration (Priority: P1) 🎯 MVP

**Goal**: A pure `Base.EnsureID()` method exists, today's `Base.BeforeCreate(tx)` delegates to it, and all 7 in-memory fakes use it instead of `BeforeCreate(nil)`. GORM persistence behavior is unchanged.

**Independent Test**: Unit test constructs a `Base`-embedded entity, calls `EnsureID()`, asserts `ID` becomes a 26-char ULID; second call is a no-op; `CreatedAt`/`UpdatedAt` are untouched. Existing fake-backed tests under `internal/repository/fake/` continue to pass.

### Tests for User Story 1 ⚠️

> Write FIRST and observe FAIL on a clean `EnsureID` absence.

- [X] T003 [P] [US1] Add test file `internal/domain/base_ensure_id_test.go` with `TestEnsureID_AssignsULIDWhenEmpty`, `TestEnsureID_PreservesExistingID`, and `TestEnsureID_DoesNotTouchTimestamps`

### Implementation for User Story 1

- [X] T004 [US1] In `internal/domain/models.go`, add method `func (b *Base) EnsureID()` per `contracts/ensure-id.md`. Generate ULID via same path used today in `Base.BeforeCreate` (monotonic-entropy, `time.Now()` source). No-op when `b.ID != ""`. Do NOT touch `CreatedAt`/`UpdatedAt`
- [X] T005 [US1] In `internal/domain/models.go`, replace the body of `Base.BeforeCreate(tx *gorm.DB) (err error)` with a single call: `base.EnsureID(); return`. Signature unchanged. Outer hooks (`NotificationChannel.BeforeCreate`, `ResourceCredential.BeforeCreate`) that call `n.Base.BeforeCreate(tx)` continue to work
- [X] T006 [P] [US1] In `internal/repository/fake/incident_fake.go` line ~29, replace `incident.BeforeCreate(nil)` with `incident.EnsureID()`
- [X] T007 [P] [US1] In `internal/repository/fake/notification_fake.go` line ~36, replace `notification.BeforeCreate(nil)` with `notification.EnsureID()`
- [X] T008 [P] [US1] In `internal/repository/fake/component_fake.go` line ~26, replace `component.BeforeCreate(nil)` with `component.EnsureID()`
- [X] T009 [P] [US1] In `internal/repository/fake/monitoring_activity_fake.go` line ~37, replace `activity.BeforeCreate(nil)` with `activity.EnsureID()`
- [X] T010 [P] [US1] In `internal/repository/fake/incident_event_step_fake.go` line ~30, replace `s.BeforeCreate(nil)` with `s.EnsureID()`
- [X] T011 [P] [US1] In `internal/repository/fake/api_key_fake.go` line ~30, replace `key.BeforeCreate(nil)` with `key.EnsureID()`
- [X] T012 [P] [US1] In `internal/repository/fake/resource_fake.go` line ~33, replace `resource.BeforeCreate(nil)` with `resource.EnsureID()`
- [X] T013 [US1] Re-run `git grep -n "BeforeCreate(nil)" -- internal/repository/fake/` and assert it returns zero hits. If any remain, fix and re-assert
- [X] T014 [US1] Run `go test -race ./internal/domain/... ./internal/repository/fake/... ./internal/repository/...` and observe all pass

**Checkpoint**: US1 delivered. ID assignment is pure; fakes are GORM-handle-free.

---

## Phase 4: User Story 2 — Typed `pkg/crypto` wrappers + hook delegation (Priority: P1)

**Goal**: Six typed wrapper functions exist in `pkg/crypto/{channel,credential}.go` and the GORM hooks on `NotificationChannel` and `ResourceCredential` call them instead of `crypto.Encrypt`/`crypto.Decrypt` directly. On-disk encryption format is unchanged. Lazy plaintext-migration in `NotificationChannel.AfterFind` continues to work.

**Independent Test**: New tests in `pkg/crypto/{channel,credential}_test.go` cover (a) round-trip `Decrypt(Encrypt(p)) == p`, (b) cross-version oracle: `DecryptChannelConfig(crypto.Encrypt(p)) == p` AND `crypto.Decrypt(EncryptChannelConfig(p)) == p`, (c) the existing `internal/domain/resource_credential_hooks_test.go` continues to pass.

### Tests for User Story 2 ⚠️

- [X] T015 [P] [US2] Add `pkg/crypto/channel_test.go` covering: round-trip empty + small JSON + unicode + large blob; cross-version oracle both directions; malformed-ciphertext error path. Inject a deterministic 32-byte test key by calling `crypto.SetGlobalProvider(...)` in a `TestMain(m *testing.M)` (matches the package's existing test pattern in `crypto_test.go`); restore the previous provider in a `defer` before `m.Run()` returns
- [X] T016 [P] [US2] Add `pkg/crypto/credential_test.go` covering the same matrix for both `EncryptCredentialPassword`/`DecryptCredentialPassword` AND `EncryptCredentialOptions`/`DecryptCredentialOptions`. Distinguish password vs options round-trips so each gets independent coverage

### Implementation for User Story 2

- [X] T017 [P] [US2] Create `pkg/crypto/channel.go` with `EncryptChannelConfig(string) (string, error)` and `DecryptChannelConfig(string) (string, error)`, each a single-statement passthrough to `Encrypt`/`Decrypt` per `contracts/crypto-wrappers.md`
- [X] T018 [P] [US2] Create `pkg/crypto/credential.go` with `EncryptCredentialPassword`, `DecryptCredentialPassword`, `EncryptCredentialOptions`, `DecryptCredentialOptions`. All single-statement passthroughs to `Encrypt`/`Decrypt`
- [X] T019 [US2] In `internal/domain/models.go` `NotificationChannel.BeforeCreate(tx)`, replace `crypto.Encrypt(...)` with `crypto.EncryptChannelConfig(...)`. Body shape otherwise unchanged
- [X] T020 [US2] In `internal/domain/models.go` `NotificationChannel.BeforeUpdate(tx)`, replace `crypto.Encrypt(...)` with `crypto.EncryptChannelConfig(...)`
- [X] T021 [US2] In `internal/domain/models.go` `NotificationChannel.AfterFind(tx)`, replace `crypto.Decrypt(...)` with `crypto.DecryptChannelConfig(...)` and `crypto.Encrypt(...)` (used for legacy-plaintext re-encrypt) with `crypto.EncryptChannelConfig(...)`. Preserve the existing `tx.Save(n)` lazy-migration write — pure wrappers do not own persistence (FR-007a)
- [X] T022 [US2] In `internal/domain/models.go`, the helper `(c *ResourceCredential).encryptSecrets()` (around line 550) is the single place both `BeforeCreate` and `BeforeUpdate` route through. Replace the `crypto.Encrypt(c.Password)` call with `crypto.EncryptCredentialPassword(...)` and the Options encrypt call with `crypto.EncryptCredentialOptions(...)`. `BeforeCreate`/`BeforeUpdate` bodies themselves do not need touching
- [X] T023 [US2] _(merged into T022 — both hooks share `encryptSecrets()`; the substitution lands in one place)_
- [X] T024 [US2] In `internal/domain/models.go` `ResourceCredential.AfterFind(tx)`, replace `crypto.Decrypt` calls with `crypto.DecryptCredentialPassword` (for Password) and `crypto.DecryptCredentialOptions` (for Options). Preserve `ErrCredentialDecryption` sentinel propagation
- [X] T025 [US2] Run `go test -race ./pkg/crypto/... ./internal/domain/...` and observe all pass — including the pre-existing `resource_credential_hooks_test.go` which validates round-trip encryption through the hooks

**Checkpoint**: US2 delivered. Per-concern typed wrappers exist; hooks delegate; format byte-identical.

---

## Phase 5: User Story 3 — Doc header on `internal/domain/models.go` (Priority: P2)

**Goal**: A head-of-file comment explains progressive tag removal so future contributors don't strip tags prematurely.

**Independent Test**: Open `internal/domain/models.go`; the first comment block (above `package domain` or immediately under it) clearly states (a) GORM tags retained on purpose, (b) removed per-repository during sqlc migration, (c) pointer to `.prds/sqlc/`.

### Implementation for User Story 3

- [X] T026 [US3] In `internal/domain/models.go`, insert immediately under the `package domain` line and before the first import block a doc comment block per `research.md` R7: 4 lines explaining (a) GORM tags + hooks intentionally retained, (b) removed repository-by-repository as each migrates to sqlc, (c) see `.prds/sqlc/` (tracks 003-domain-decoupling and 006+), (d) do not submit tag-removal PRs ahead of the migration schedule

**Checkpoint**: US3 delivered.

---

## Phase N: Polish & Cross-Cutting Concerns

- [X] T027 [P] Run `go vet ./...` — clean
- [X] T028 Run `make test-be` and assert exit 0 with zero new failures (SC-004). Attach the tail to the PR description
- [X] T029 Tag-count invariant (SC-005): assert `git grep -c 'gorm:"' -- internal/domain/models.go` returns the same count on HEAD as on `042-sqlc-schema-source` (the branch base). Capture both numbers in the PR description
- [X] T030 Hook-method invariant (SC-006): run `git diff 042-sqlc-schema-source -- internal/domain/models.go | grep -E '^\-func .*\b(BeforeCreate\|BeforeUpdate\|AfterFind)\b'` and assert no matches (no hook deletions)
- [X] T031 No-new-AES-in-domain invariant (SC-007): run `git grep -E 'aes\.|cipher\.' -- internal/domain/` and assert no matches in non-test files
- [X] T032 Scope guard: run `git diff --stat 042-sqlc-schema-source -- internal/api/ internal/service/ internal/repository/store/ web/` and assert empty — this PR does not touch those layers
- [X] T033 Walk through `specs/043-domain-decoupling/quickstart.md` snippets locally; record evidence in the PR description
- [X] T034 SonarQube scan per CLAUDE.md "Code Quality"; resolve any new CRITICAL/BLOCKER under `pkg/crypto/` or `internal/domain/`
- [X] T035 Trigger CI on the branch; record the green run URL in the PR description
- [X] T036 Cross-check FRs/SCs ↔ tasks coverage map in PR description (FR-001…FR-012, SC-001…SC-007)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 Setup**: no deps — start immediately
- **Phase 2 Foundational**: empty
- **Phase 3 US1 (P1)**: needs Phase 1 — MVP candidate
- **Phase 4 US2 (P1)**: needs Phase 1; INDEPENDENT of US1 — can run in parallel by a second engineer
- **Phase 5 US3 (P2)**: needs no phase — pure doc edit, can land any time, but the PR is cleanest if it lands after US1/US2
- **Phase N Polish**: needs all targeted user stories complete

### Within Each User Story

- US1: T003 (test) before T004 (impl); fake edits T006–T012 are independent of each other and can land in parallel; T013 (grep verify) after all fake edits; T014 (test run) last
- US2: T015, T016 (tests) before implementation T017–T024; T017 and T018 (new files) independent; T019–T024 (model edits) all touch the same file (`internal/domain/models.go`) and must be serial
- US3: single task, no internal order

### Parallel Opportunities

- T006–T012 (7 fake edits) — distinct files, no shared deps
- T015 / T016 (test files) and T017 / T018 (wrapper files) — distinct files
- T027 / T029 / T030 / T031 / T032 — independent polish checks
- US1 and US2 can be developed by two engineers in parallel; they touch disjoint code paths (fakes vs hook bodies); the only overlap is `internal/domain/models.go` — US1 touches `Base` (top), US2 touches hooks (middle of file). Coordinate via small commits

---

## Parallel Example: User Story 1

```bash
# Tests first:
Task: "T003 base_ensure_id_test.go (US1)"

# Impl method:
Task: "T004 Base.EnsureID() in models.go"
Task: "T005 reshape Base.BeforeCreate to call EnsureID"

# Fake migrations in parallel (7 distinct files):
Task: "T006 incident_fake.go"
Task: "T007 notification_fake.go"
Task: "T008 component_fake.go"
Task: "T009 monitoring_activity_fake.go"
Task: "T010 incident_event_step_fake.go"
Task: "T011 api_key_fake.go"
Task: "T012 resource_fake.go"

# Verify:
Task: "T013 grep BeforeCreate(nil) returns zero"
Task: "T014 go test fakes + domain"
```

---

## Implementation Strategy

### MVP First (US1 only)

1. Phase 1 Setup
2. Phase 3 US1 — `EnsureID()` + fakes
3. STOP and VALIDATE: `TestEnsureID_*` pass; fakes test suite green; full backend tests still green

MVP delivers the smaller, lower-risk piece (no encryption touch). US2 follows once US1 is reviewed.

### Incremental Delivery

1. MVP US1 shipped → fakes no longer pass `nil` to GORM hooks
2. Add US2 (typed crypto wrappers + hook delegation) → encryption surfaces consolidated; format byte-identical
3. Add US3 (doc header) → tag-removal policy made explicit
4. Polish → merge

### Parallel Team Strategy

- Eng A: US1 (low-risk mechanical)
- Eng B (in parallel after Phase 1): US2 (encryption — highest stakes, needs the most review)
- Eng A or B picks up US3 once US1+US2 land

---

## Notes

- [P] = different files, no incomplete-task dependencies
- US2 is the highest-risk piece in this ticket — encryption format byte-equality is the invariant; the oracle test (T015/T016) protects against silent drift
- Reject scope creep: no `gorm:"…"` tag removal, no hook deletion, no field rename. T029/T030 fail the polish phase if any sneak in
- The PRD said "8 fakes"; current state has 7. T001 re-verifies and adjusts if a new fake appeared since PRD time
- Commit per task or per logical group (Conventional Commits)
- Stop at each Checkpoint and confirm the independent test before advancing
