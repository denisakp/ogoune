# ADR 0009 — AES-256-GCM for notification credential encryption at rest

- **Status**: Accepted
- **Date**: 2026-05-30
- **Deciders**: Denis Akpagnonite
- **Scope**: Both
- **Tags**: crypto, security, secrets, notifications

## Context

Notification channels (`NotificationChannel.Config`) store user-supplied secrets: Slack webhook URLs, Discord tokens, SMTP passwords, Teams webhook tokens, generic webhook bearer tokens. These secrets:

- Are user-owned and must round-trip exactly (decrypt must return what the user typed)
- Live in the database (SQLite file or Postgres rows) which is part of routine backups
- Are referenced at notification dispatch time by workers, potentially many times per minute under incident bursts

Storing them plaintext means a database leak directly compromises the user's third-party accounts. The CE single-binary deployment model rules out external secret managers (Vault, AWS KMS) — the system must encrypt locally with a key the operator controls.

`pkg/crypto/crypto.go` implements the encryption boundary. `pkg/crypto/channel.go` provides a typed wrapper (`EncryptChannelConfig` / `DecryptChannelConfig`) used by the notification layer.

## Decision drivers

- Must work with a single operator-managed key — CE has no external KMS
- Must provide both confidentiality and integrity (a tampered ciphertext must not silently decrypt to attacker-controlled plaintext)
- Must be standard-library-only on the Go side — no third-party crypto dependency to audit
- Must produce ciphertexts safe to store in a TEXT column on both SQLite and Postgres ([ADR-0002](./0002-dual-dialect-sqlite-postgres.md))
- Must fail closed when the key is missing — refuse to start rather than silently store plaintext
- Solo-dev cryptography rule: pick the boring authenticated mode, do not invent

## Options considered

### Option A — Plaintext storage

**Pros**: trivial.
**Cons**: a single DB leak compromises every user's third-party tokens. Unacceptable.

### Option B — AES-256-CBC + HMAC-SHA256 (Encrypt-then-MAC)

**Pros**: well-understood, both primitives in stdlib.
**Cons**: two-key construction adds complexity; padding oracle classes of bugs in custom CBC code; GCM gives identical guarantees with less code.

### Option C — AES-256-GCM (authenticated encryption)

**Pros**: AEAD in one primitive, in `crypto/cipher` stdlib, single key, FIPS-validatable. Industry standard for data-at-rest at this scale.
**Cons**: nonce reuse is catastrophic — must generate unique nonces (random 12-byte from `crypto/rand` is safe up to ~2^32 messages per key).

### Option D — ChaCha20-Poly1305

**Pros**: faster than AES on platforms lacking AES-NI, also AEAD.
**Cons**: not in stdlib's main `crypto/cipher` — requires `golang.org/x/crypto`; on modern x86/ARM with AES-NI, AES-GCM is hardware-accelerated and outperforms; smaller ecosystem expectation in audit contexts.

### Option E — age (modern envelope encryption)

**Pros**: high-level, opinionated, hard to misuse.
**Cons**: tailored for file-at-rest and recipient-based encryption; overkill for short-string column encryption; introduces a dependency.

## Decision

Ogoune uses **AES-256-GCM** for all credential encryption at rest:

- Key: 32 bytes (256-bit), supplied by the operator via `APP_SECRET_KEY` as a 64-character hex string. The app refuses to start if the env var is missing, empty, or malformed (`pkg/crypto/crypto.go` line 27-37).
- Nonce: 12 bytes (`gcm.NonceSize()`), freshly generated per encryption from `crypto/rand`.
- Ciphertext layout: `base64(nonce || ciphertext_with_auth_tag)`.
- Empty plaintext is preserved as empty string (no encryption), so unset configs do not produce ciphertext.
- Stored in TEXT columns on both dialects ([ADR-0002](./0002-dual-dialect-sqlite-postgres.md)).
- Typed wrappers (`EncryptChannelConfig`, `DecryptChannelConfig`) provide concern-specific entry points without changing the on-disk format — purely an audit and refactor lever.

No key rotation mechanism ships at decision time: operators rotate by re-encrypting all rows after generating a new key, manually. A future ADR can introduce a versioned envelope (`vN.nonce.ciphertext`) if rotation becomes a concrete requirement.

## Consequences

### Positive
- Confidentiality + integrity in one primitive
- Stdlib only (`crypto/aes`, `crypto/cipher`, `crypto/rand`) — no third-party crypto to audit
- Hardware-accelerated on every modern deploy target
- Fail-closed startup: missing key = refuse to start, never silent plaintext
- Identical scheme in CE and EE — no boundary surprise at upgrade

### Negative
- Key rotation is manual today — operator burden
- Nonce-reuse risk is real if `crypto/rand` is misused or seeded poorly (the current implementation reads directly from `crypto/rand.Reader`, which is correct)
- No version byte in the ciphertext envelope — future algorithm migration requires re-encryption, not in-place upgrade

### Neutral / to watch
- If credential volume grows large (e.g. 100k+ channels), birthday-bound on random 12-byte nonces (~2^32 safe messages per key) becomes worth tracking — still ~7 orders of magnitude from current scale
- If FIPS validation becomes a customer requirement, AES-GCM is already on the validated-algorithm list

## Compatibility, migration & rollout

- **DB schema**: existing TEXT columns for `notification_channels.config` and similar; ciphertext is base64 ASCII, safe across dialects
- **Boot semantics**: missing `APP_SECRET_KEY` aborts startup with a clear error — documented in `.env.example` and `CLAUDE.md` "Gotchas"
- **Key rotation**: not automated. Operators run a re-encrypt script (one-shot Go program reading old key from env, new key from second env, re-writing rows). Document as runbook when first requested.
- **CE ↔ EE**: identical scheme; EE Cloud may later introduce per-tenant keys behind a separate ADR
- **Doc drift**: `CLAUDE.md` "Encryption" line, `.env.example` `APP_SECRET_KEY` example with `openssl rand -hex 32` guidance

## Implementation checklist

- [x] `pkg/crypto/crypto.go` — `Encrypt`, `Decrypt`, `GetEncryptionKey`
- [x] `pkg/crypto/channel.go` — typed wrappers for notification channel configs
- [x] `pkg/crypto/credential.go` — typed wrappers for other credentials
- [x] Tests `pkg/crypto/crypto_test.go`, `channel_test.go`, `credential_test.go` — round-trip, oracle-against-generic, error cases
- [x] Startup refuses missing/malformed `APP_SECRET_KEY`
- [x] `.env.example` and `CLAUDE.md` mention `openssl rand -hex 32`
- [ ] Runbook: key rotation procedure (when first requested)
- [ ] If versioned envelope is needed, introduce `v1:` prefix as a successor ADR — must keep old-format decryption path until full re-encryption

## References

- Code: `pkg/crypto/crypto.go` (primitive), `pkg/crypto/channel.go` (typed wrapper)
- Tests: `pkg/crypto/{crypto,channel,credential}_test.go`
- Related ADRs: ADR-0002 (dual-dialect — ciphertext stored as TEXT), ADR-0007 (zero telemetry — keys stay local, no escrow)
- External: [NIST SP 800-38D (GCM)](https://csrc.nist.gov/publications/detail/sp/800-38d/final), Go `crypto/cipher` docs
