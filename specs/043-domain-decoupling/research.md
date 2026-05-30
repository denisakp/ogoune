# Phase 0 — Research

## R1. Existing `pkg/crypto` is already the encryption single-source-of-truth

**Decision**: Reuse `crypto.Encrypt(plaintext string) (string, error)` and `crypto.Decrypt(ciphertext string) (string, error)` underneath the typed wrappers. No new AES/GCM code.

**Evidence**: `pkg/crypto/crypto.go` already implements AES-256-GCM with `APP_SECRET_KEY` derivation via `EnvKeyProvider`. Current hooks in `internal/domain/models.go` (`NotificationChannel.BeforeCreate`, `BeforeUpdate`, `AfterFind`; `ResourceCredential.{BeforeCreate,BeforeUpdate,AfterFind}`) already call `crypto.Encrypt`/`crypto.Decrypt` directly — there is no duplicated AES code in `internal/domain/` today.

**Rationale**: The PRD's stated goal ("no duplication") is mostly already met. The remaining work is naming + per-concern indirection so future sqlc-backed repos call `EncryptChannelConfig(…)` instead of a generic `crypto.Encrypt(…)`, making audit traceability per-concern.

## R2. `Base.BeforeCreate` semantics today — ID only, no timestamps

**Decision**: `EnsureID()` sets `ID` only. It MUST NOT touch `CreatedAt` or `UpdatedAt`.

**Evidence**: `internal/domain/models.go` L23–L31:
```go
func (base *Base) BeforeCreate(tx *gorm.DB) (err error) {
    if base.ID == "" {
        t := time.Now()
        entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
        base.ID = ulid.MustNew(ulid.Timestamp(t), entropy).String()
    }
    return
}
```
No `CreatedAt`/`UpdatedAt` writes. GORM tags (`autoCreateTime`, `autoUpdateTime`, embedded by GORM serializer) handle timestamp filling at INSERT/UPDATE.

**Rationale**: Spec FR-002 originally claimed `EnsureID` should set timestamps. That was wrong against the actual code. Spec corrected at Clarification time.

**Alternatives**: Have `EnsureID` set `CreatedAt = time.Now()` when zero (more explicit for fakes). Rejected: changes observable behavior in tests that rely on GORM's autofill; out-of-scope.

## R3. `AfterFind` legacy lazy-migration write stays in the GORM hook

**Decision** (Clarification Q2): The typed wrappers `EncryptChannelConfig`/`DecryptChannelConfig` are pure (no `tx`). The existing `AfterFind` branch that detects legacy plaintext (`Config[0] == '{'`), re-encrypts, and **persists synchronously via `tx.Save(…)`** stays inside the GORM hook.

**Rationale**:
- Pure wrappers have one job (transform bytes). Conflating that with persistence ties them to a specific data-access library.
- Lazy migration is GORM-era plumbing; when a future sqlc-backed `NotificationChannel` repository lands, it will own its own legacy-row handling decision (likely a one-off migration script per ticket 010).
- Keeps the change surface tiny in this PR: only the hook's encrypt/decrypt calls move; the write path is untouched.

**Alternatives**:
- B: Domain method `MigrateLegacyConfig(persistFn func([]byte) error)` invoked by the hook. Rejected: introduces a new domain method and a callback indirection for a path that will be deleted in 010 anyway.
- C: Stop lazy migration; write a one-off script. Rejected: out-of-scope for this ticket; risks orphaning legacy rows if the script lags.

## R4. Fake audit — 7 confirmed sites, PRD said 8

**Decision**: `grep -rn "BeforeCreate(nil)" internal/repository/fake/` returns 7 file sites today:

```
incident_fake.go:29
notification_fake.go:36
component_fake.go:26
monitoring_activity_fake.go:37
incident_event_step_fake.go:30
api_key_fake.go:30
resource_fake.go:33
```

**Rationale**: PRD said "8 fakes" — count may have predated a deletion or counted a non-`BeforeCreate` path. Implementation MUST re-run the grep at task time and migrate every match. The `git diff --stat` scope guard at polish time catches anything missed in `internal/repository/fake/`.

**Alternatives**: Trust the PRD count and hard-code "8". Rejected: PRD is older than current state; grep is authoritative.

## R5. Cross-version oracle test strategy (SC-003)

**Decision**: For each typed wrapper pair, the test does:

```go
plain := "{\"webhook_url\":\"https://example.invalid\"}"
// 1) Produce ciphertext via existing generic crypto.Encrypt (the function hooks call today).
oracle, err := crypto.Encrypt(plain)
require.NoError(t, err)
// 2) Decrypt with the new typed wrapper.
got, err := DecryptChannelConfig(oracle)
require.NoError(t, err)
require.Equal(t, plain, got)
// 3) And the reverse: typed-encrypt then generic-decrypt round-trips.
cipher, err := EncryptChannelConfig(plain)
require.NoError(t, err)
got2, err := crypto.Decrypt(cipher)
require.NoError(t, err)
require.Equal(t, plain, got2)
```

**Rationale**: This is the strongest possible local oracle without a frozen fixture from production. Because the wrappers are 1-line passthroughs, the test trivially passes today; its value is regression-prevention against any future "optimization" of the wrappers that diverges from `crypto.Encrypt`. A separate constant-fixture test using a hex-encoded captured ciphertext is also added per US2 acceptance #3.

## R6. Constant-fixture test (no live key dependency)

**Decision**: Round-trip tests use `crypto.SetGlobalProvider` to inject a stable, deterministic test key (32 bytes derived from a known constant) before each test, then restore. The captured "oracle" ciphertext is produced under the same key, so the cross-version oracle is reproducible across runs.

**Rationale**: Test must not depend on `APP_SECRET_KEY` env var being set. `pkg/crypto/crypto.go` already exposes `SetGlobalProvider(KeyProvider)`.

**Alternatives**: `os.Setenv("APP_SECRET_KEY", "…")` per test. Rejected: env mutation is racy under `-race`.

## R7. Doc header location & wording

**Decision**: Insert immediately after the `package domain` line, before the first import block:

```go
// GORM tags and GORM hooks on these models are intentionally retained.
// They are being progressively removed repository-by-repository as each
// migrates to sqlc. See .prds/sqlc/ (track 003-domain-decoupling and 006+)
// for the schedule. Do not submit PRs that rip out gorm:"…" tags ahead
// of the migration.
```

**Rationale**: First thing a reader sees. Clear contract for future contributors. Anchors against well-meaning cleanup PRs.

## R8. Hook-body shape after change

**Decision**: Each hook body becomes a thin dispatch:

```go
func (n *NotificationChannel) BeforeCreate(tx *gorm.DB) error {
    if err := n.Base.BeforeCreate(tx); err != nil {  // delegates to EnsureID via Base.BeforeCreate
        return err
    }
    if len(n.Config) == 0 { return nil }
    cipher, err := crypto.EncryptChannelConfig(string(n.Config))
    if err != nil { return err }
    n.Config = []byte(cipher)
    return nil
}
```

For `AfterFind`, the legacy-plaintext branch keeps its `tx.Save(n)` call but the encrypt/decrypt path goes through typed wrappers.

**Rationale**: Mechanical, low-risk transformation. No behavioral delta.

## R9. Scope guard (FR-009, FR-011)

**Decision**: Polish task runs:

```bash
# Tag-count invariant (SC-005)
test "$(git grep -c 'gorm:"' -- internal/domain/models.go HEAD)" = \
     "$(git grep -c 'gorm:"' -- internal/domain/models.go 041-sqlc-foundation)"

# Hook-method invariant (SC-006)
git diff 041-sqlc-foundation -- internal/domain/models.go \
  | grep -E '^\-func .*\b(BeforeCreate|BeforeUpdate|AfterFind)\b' \
  | grep -vc '^$' \
  | test "$(cat)" = "0"

# No new AES/GCM code in internal/domain/ (SC-007)
! git grep -E 'aes\.|cipher\.' internal/domain/
```

**Rationale**: Mechanized invariants prevent accidental scope creep at PR time.
