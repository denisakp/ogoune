# Feature Specification: Domain Decoupling — Drop GORM Tags + Hooks

**Feature Branch**: `043-domain-decoupling`
**Created**: 2026-05-30
**Status**: Draft
**Input**: User description: "Domain decoupling: drop GORM tags + hooks — read .prds/sqlc/003-domain-decoupling.md to understand the spec"

## Clarifications

### Session 2026-05-30

- Q: Typed wrapper functions vs. generic `crypto.Encrypt`/`Decrypt`? → A: Ship typed wrappers — `EncryptChannelConfig`/`DecryptChannelConfig` in `pkg/crypto/channel.go`, plus `EncryptCredentialPassword`/`DecryptCredentialPassword`/`EncryptCredentialOptions`/`DecryptCredentialOptions` in `pkg/crypto/credential.go`. Each delegates to `crypto.Encrypt`/`crypto.Decrypt`. Reason: per-concern named API, traceable in audit, localized format-versioning later if needed.
- Spec correction (no question): Today's `Base.BeforeCreate` only sets `ID`; it does NOT touch `CreatedAt`/`UpdatedAt` (GORM tags fill them). `EnsureID()` must match that behavior — ID only.
- Q: `AfterFind` legacy-plaintext lazy migration — where does the persisting write happen post-refactor? → A: Lazy-write stays inside the GORM `AfterFind` hook (still has `tx`). Pure `DecryptChannelConfig` only decrypts. Hook detects legacy format, calls `EncryptChannelConfig`, persists via `tx`. Future sqlc-backed repo handles its own legacy path explicitly when migrated.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Maintainer can set ULIDs on domain entities without a GORM handle (Priority: P1)

A maintainer constructing a domain entity in any context — production code, in-memory fake repository, unit test, future sqlc-backed repository — calls a pure method on the entity (`EnsureID()`) and gets a deterministic ULID assignment with no GORM dependency. The existing GORM `BeforeCreate(tx *gorm.DB)` continues to work for live persistence and internally delegates to the same pure method.

**Why this priority**: Every fake and every future sqlc-backed repository needs ID assignment without a `*gorm.DB`. This is the foundational decoupling step that unblocks all subsequent migration tickets. Today, fakes pass `nil` to a GORM-shaped hook — a fragile pattern that only works because the hook never dereferences the handle.

**Independent Test**: Construct a `Base`-embedded entity in a unit test, call `EnsureID()`, assert the `ID` is now a valid 26-character ULID. Construct another entity, call `BeforeCreate(nil)`, assert the same outcome. Run the existing backend test suite and confirm no regression.

**Acceptance Scenarios**:

1. **Given** a freshly-constructed entity with an empty `ID`, **When** `EnsureID()` is called, **Then** `ID` is populated with a valid 26-character ULID and `CreatedAt`/`UpdatedAt` get set per existing semantics.
2. **Given** an entity whose `ID` is already non-empty (e.g. supplied by the caller), **When** `EnsureID()` is called, **Then** the existing `ID` is preserved.
3. **Given** the existing GORM persistence path, **When** GORM invokes `BeforeCreate(tx *gorm.DB)`, **Then** behavior is identical to before — internally delegating to `EnsureID()` and producing the same persisted row.
4. **Given** every in-memory fake repository in `internal/repository/fake/`, **When** it inserts an entity, **Then** it calls `EnsureID()` (not `BeforeCreate(nil)`).

---

### User Story 2 - Encryption logic for notification channels and resource credentials is reusable outside GORM (Priority: P1)

A maintainer needs to encrypt or decrypt a `NotificationChannel.Config` or a `ResourceCredential.{Password,Options}` payload from a context other than a GORM hook (CLI tool, future sqlc-backed repository, batch script). Two purpose-built packages — `pkg/crypto/channel.go` and `pkg/crypto/credential.go` — expose `Encrypt…` / `Decrypt…` functions that are the **single source of truth** for the encryption format. The existing GORM hooks delegate to these functions; no logic is duplicated.

**Why this priority**: Same blocker as US1 — without callable encryption helpers, any sqlc-backed repository that touches these columns has to either ship its own copy of the logic (drift risk) or invoke GORM hooks indirectly (defeats the purpose). Encryption is also the highest-stakes piece to migrate: a format regression silently corrupts stored secrets.

**Independent Test**: Encrypt a known plaintext with the new package function, then run the result through the existing GORM hook's decryption path; assert plaintext recovered. Conversely, capture ciphertext produced by the current `BeforeCreate` hook on a fixture entity, then decrypt it with the new package function; assert match. Round-trip oracle: plain → new encrypt → new decrypt = plain.

**Acceptance Scenarios**:

1. **Given** a plaintext channel config, **When** the maintainer calls `crypto.EncryptChannelConfig(plain)` → `crypto.DecryptChannelConfig(cipher)`, **Then** the original plaintext is recovered.
2. **Given** a plaintext password (resource credential), **When** the maintainer calls `crypto.EncryptCredentialPassword(plain)` → `crypto.DecryptCredentialPassword(cipher)`, **Then** the original plaintext is recovered. Same property holds for the Options field.
3. **Given** ciphertext captured from the **current** GORM `BeforeCreate` hook output (before this change lands), **When** the new package function decrypts it, **Then** the original plaintext is recovered (cross-version oracle — no format change).
4. **Given** the existing GORM hooks on `NotificationChannel` and `ResourceCredential`, **When** they run during normal persistence, **Then** they call the new package functions internally; no second copy of the encryption logic exists in `internal/domain/`.
5. **Given** a corrupted ciphertext, **When** decryption is attempted via the new package function, **Then** the same domain sentinel error (`ErrCredentialDecryption` or the `NotificationChannel` equivalent) is returned as today.

---

### User Story 3 - Domain file announces its progressive decoupling status (Priority: P2)

A future contributor opening `internal/domain/models.go` sees a header comment explaining that GORM tags remain for now and will be removed repository-by-repository as each one migrates to sqlc, with a pointer to the planning docs. They do not file PRs ripping out tags ahead of schedule.

**Why this priority**: Communication, not safety. Prevents well-intentioned but premature tag removal. Low effort.

**Independent Test**: Open `internal/domain/models.go`; the first comment block (above package or above the first type) states the policy and points at `.prds/sqlc/`.

**Acceptance Scenarios**:

1. **Given** a contributor unfamiliar with the sqlc migration plan, **When** they read the head of `internal/domain/models.go`, **Then** they understand (a) GORM tags are intentionally retained, (b) tags are removed per-repository during migration, (c) where to find the migration tickets.

---

### Edge Cases

- An entity is constructed with a pre-populated `ID` (e.g. test fixture) → `EnsureID()` preserves it; same behavior as today's `BeforeCreate`.
- A fake calls `EnsureID()` twice on the same entity → idempotent: the existing `ID` is kept.
- The new `pkg/crypto/channel.go` is called concurrently from multiple goroutines → safe (no shared state beyond the existing `APP_SECRET_KEY` derivation).
- Existing ciphertext at rest in production databases must decrypt cleanly via the new package functions — the encryption format MUST NOT change.
- A `NotificationChannel` or `ResourceCredential` with empty `Config`/`Password` → no-op, same as today's hook behavior.
- An entity already migrated to a sqlc-backed repository (none yet in this ticket) — sqlc-backed code MUST invoke the new package functions directly; it MUST NOT depend on GORM hooks.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `Base` MUST expose a public method `EnsureID()` (no receiver-internal GORM dependency) that assigns a fresh ULID to `ID` when `ID` is empty, and is a no-op when `ID` is already set.
- **FR-002**: `Base.EnsureID()` MUST NOT modify `CreatedAt` or `UpdatedAt` — current `Base.BeforeCreate` does not touch them (GORM `autoCreateTime`/`autoUpdateTime` tags fill them). `EnsureID()` is ID-only and idempotent.
- **FR-003**: The existing `Base.BeforeCreate(tx *gorm.DB) error` MUST be preserved and MUST internally delegate to `EnsureID()`. No behavioral change to the GORM persistence path.
- **FR-004**: Every in-memory fake repository under `internal/repository/fake/` MUST call `EnsureID()` instead of `BeforeCreate(nil)` on the entity it is inserting. The audit covers — at minimum — `incident_fake.go`, `notification_fake.go`, `component_fake.go`, `monitoring_activity_fake.go`, `incident_event_step_fake.go`, `api_key_fake.go`, `resource_fake.go`, and any other fake found at implementation time.
- **FR-005**: Project MUST introduce `pkg/crypto/channel.go` exposing `EncryptChannelConfig(plain string) (string, error)` and `DecryptChannelConfig(cipher string) (string, error)`. These wrap the existing `crypto.Encrypt`/`crypto.Decrypt` — no duplicated AES/GCM code, no format change.
- **FR-006**: Project MUST introduce `pkg/crypto/credential.go` exposing four typed wrappers: `EncryptCredentialPassword` / `DecryptCredentialPassword` and `EncryptCredentialOptions` / `DecryptCredentialOptions`. All four wrap `crypto.Encrypt`/`crypto.Decrypt`. Both round-trip plaintext and decrypt ciphertext produced by today's hooks unchanged.
- **FR-007**: The existing GORM hooks on `NotificationChannel` (`BeforeCreate`, `BeforeUpdate`, `AfterFind`) and `ResourceCredential` (`BeforeCreate`, `BeforeUpdate`, `AfterFind`) MUST delegate to the new typed wrappers in `pkg/crypto/{channel,credential}.go` instead of calling `crypto.Encrypt`/`crypto.Decrypt` directly. This concentrates per-concern entry points in a single audit-traceable location.
- **FR-007a**: The typed wrappers MUST be pure (no `tx`/`*gorm.DB` parameter). The existing `NotificationChannel.AfterFind` legacy-plaintext lazy-migration write (detect `Config[0]=='{'`, re-encrypt, persist via `tx.Model(n).UpdateColumn(...)`) stays inside the GORM hook where the `tx` is available. The hook calls `EncryptChannelConfig` to produce the new ciphertext then writes via `tx`. `ResourceCredential.AfterFind` has no legacy-migration branch today; its body simply switches its `crypto.Decrypt` calls to the typed wrappers. Future sqlc-backed repositories handle their own legacy path when they migrate.
- **FR-008**: The existing `ErrCredentialDecryption` sentinel (used by `ResourceCredential.AfterFind`) MUST be preserved with identical semantics — callers compare with `errors.Is` and the import path is unchanged within this PR. `NotificationChannel.AfterFind` returns the raw underlying error (no sentinel today); this behavior is preserved as-is.
- **FR-009**: Project MUST NOT remove any `gorm:"…"` struct tag from `internal/domain/models.go` in this ticket. Tags are removed per-repository in later tickets (006/007/008 per PRD).
- **FR-010**: `internal/domain/models.go` MUST gain a head-of-file comment stating that GORM tags are intentionally retained and will be removed repository-by-repository as each migrates to sqlc, with a pointer to `.prds/sqlc/`.
- **FR-011**: Project MUST NOT remove any GORM hook in this ticket. Hooks remain in place so the GORM path stays unchanged; only their *implementation* is moved/delegated.
- **FR-012**: Round-trip encryption tests MUST exist for both packages and MUST include an oracle fixture: a known plaintext + ciphertext pair captured from the **current** hook implementation (pre-PR) that the new functions decrypt successfully.

### Key Entities *(include if feature involves data)*

- **`Base`**: Embedded struct providing `ID` (ULID), `CreatedAt`, `UpdatedAt`. Gains the pure method `EnsureID()`.
- **`NotificationChannel`**: Persisted entity whose `Config` field is encrypted at rest. Encryption logic moves to `pkg/crypto/channel.go`; the entity's GORM hooks delegate.
- **`ResourceCredential`**: Persisted entity whose `Password` and `Options` fields are encrypted at rest. Encryption logic moves to `pkg/crypto/credential.go`; the entity's GORM hooks delegate.
- **Encryption format**: Byte-identical to today's output. No version bump, no header change.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of in-memory fake repositories under `internal/repository/fake/` use `EnsureID()` (zero remaining `BeforeCreate(nil)` calls), verifiable by `git grep`.
- **SC-002**: Round-trip encryption tests pass for `NotificationChannel.Config`, `ResourceCredential.Password`, and `ResourceCredential.Options` (new tests in `pkg/crypto/`).
- **SC-003**: Cross-version oracle tests pass — ciphertext produced by today's hook implementation decrypts successfully with the new package functions on the same plaintext.
- **SC-004**: Full backend test suite (`make test-be`) passes with **zero new failures** vs main (`041-sqlc-foundation` baseline already established 1100+ tests).
- **SC-005**: Number of `gorm:"…"` tags in `internal/domain/models.go` is unchanged from main (verifiable by `git grep -c 'gorm:"' internal/domain/models.go`).
- **SC-006**: Number of GORM hook methods (`BeforeCreate`, `BeforeUpdate`, `AfterFind`) on `internal/domain/models.go` types is unchanged from main; only their bodies delegate to the new packages.
- **SC-007**: No new copy of AES/GCM encryption logic exists in `internal/domain/` — `grep -r 'aes\.\|cipher\.' internal/domain/` returns no production code matches.

## Assumptions

- The current encryption format (AES-256-GCM per CLAUDE.md) is preserved byte-for-byte. No format migration in this ticket.
- `APP_SECRET_KEY` env var continues to be the encryption key source; key derivation logic moves with the encryption logic to `pkg/crypto/`.
- ULID generation uses the existing `oklog/ulid/v2` dependency.
- Fakes inventory at audit time: 7 confirmed `BeforeCreate(nil)` callers (`incident`, `notification`, `component`, `monitoring_activity`, `incident_event_step`, `api_key`, `resource`) plus any additional fake added between spec time and implementation time. PRD mentions 8 — implementation MUST grep the directory and migrate every match.
- Resource credential encryption may already live in `pkg/crypto/` from feature 026 (credential-encryption). Plan should consolidate rather than duplicate. The spec accepts either "create new file" or "extend existing file" outcomes as long as FR-005/FR-006/FR-007 hold.
- Hooks remain for the lifetime of any repository still using GORM. Full hook removal is ticket 010 per the sqlc PRD track.
- No env var or operator-visible behavior change.
- No domain field rename, no domain table rename, no schema migration.
- Independent of 042 (drift linter). Depends on 041 (foundation) being merged.
