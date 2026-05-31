# Contract — `pkg/crypto` Typed Wrappers

Package: `pkg/crypto`

## Files & exports

### `pkg/crypto/channel.go`

```go
package crypto

// EncryptChannelConfig encrypts a NotificationChannel.Config payload.
// Format: byte-identical to crypto.Encrypt — this is per-concern typed
// indirection over the generic primitive, allowing future sqlc-backed
// callers to depend on a named API per encrypted field.
func EncryptChannelConfig(plaintext string) (string, error)

// DecryptChannelConfig is the inverse of EncryptChannelConfig. Returns the
// same error semantics as crypto.Decrypt — callers comparing via errors.Is
// against any package-level sentinel see no change.
func DecryptChannelConfig(ciphertext string) (string, error)
```

### `pkg/crypto/credential.go`

```go
package crypto

// EncryptCredentialPassword encrypts a ResourceCredential.Password payload.
func EncryptCredentialPassword(plaintext string) (string, error)

// DecryptCredentialPassword is the inverse.
func DecryptCredentialPassword(ciphertext string) (string, error)

// EncryptCredentialOptions encrypts a ResourceCredential.Options payload.
func EncryptCredentialOptions(plaintext string) (string, error)

// DecryptCredentialOptions is the inverse.
func DecryptCredentialOptions(ciphertext string) (string, error)
```

## Delegation contract

Every wrapper MUST be a single-statement passthrough:

```go
func EncryptChannelConfig(p string) (string, error) { return Encrypt(p) }
func DecryptChannelConfig(c string) (string, error) { return Decrypt(c) }
```

No additional pre/post processing. The contract forbids transformations that would diverge the wrappers from the generic primitive's on-disk format. Format divergence becomes a deliberate, reviewed change at the wrapper site rather than implicit at the call site.

## Purity contract

- No `*gorm.DB` parameter. No I/O beyond what `Encrypt`/`Decrypt` already do. No callback parameter.
- Lazy-migration writes (currently in `NotificationChannel.AfterFind`) stay in the GORM hook with its `tx`. The wrapper produces the ciphertext; the hook persists it.

## Error semantics

- Encryption failures propagate the underlying `Encrypt` error verbatim.
- Decryption failures propagate the underlying `Decrypt` error verbatim.
- Existing sentinels (e.g. `domain.ErrCredentialDecryption`) remain where they are. Their definitions and import paths are unchanged in this ticket.

## Tests

`pkg/crypto/channel_test.go` and `pkg/crypto/credential_test.go` MUST cover:

1. Round-trip: `Decrypt(Encrypt(p)) == p` for representative payloads (empty, small JSON, large JSON, unicode).
2. Cross-version oracle:
   - `DecryptXxx(crypto.Encrypt(p)) == p` — new wrapper decrypts ciphertext made by the generic primitive.
   - `crypto.Decrypt(EncryptXxx(p)) == p` — generic primitive decrypts ciphertext made by the new wrapper.
3. Error path: malformed/short ciphertext produces a non-nil error.

Tests inject a deterministic key via `crypto.SetGlobalProvider` so they do not depend on `APP_SECRET_KEY`.

## Out of scope

- Key rotation, key derivation changes.
- Envelope versioning or header bytes.
- Streaming variants.
- Async helpers.
