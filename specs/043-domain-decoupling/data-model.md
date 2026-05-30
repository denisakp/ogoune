# Data Model

This feature does not introduce any persisted entity or schema change. It adds in-process methods + helper functions.

## `Base` (extended) — `internal/domain/models.go`

Existing:

```go
type Base struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    CreatedAt time.Time `json:"created_at" gorm:"index"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

New method:

| Method | Signature | Behavior |
|--------|-----------|----------|
| `EnsureID` | `func (b *Base) EnsureID()` | If `b.ID == ""`, assigns a fresh ULID. No-op when `b.ID` already set. Does NOT touch `CreatedAt`/`UpdatedAt`. Pure: no `*gorm.DB`, no I/O. |

Modified method:

| Method | Signature | New behavior |
|--------|-----------|--------------|
| `BeforeCreate` | `func (base *Base) BeforeCreate(tx *gorm.DB) error` | Delegates: calls `EnsureID()`. Same observable behavior as today. |

## `pkg/crypto/channel.go` (new)

Pure typed wrappers around the existing generic `crypto.Encrypt` / `crypto.Decrypt`.

```go
package crypto

// EncryptChannelConfig encrypts a NotificationChannel.Config payload.
// Format is byte-identical to crypto.Encrypt — this is a typed indirection
// that lets future sqlc-backed code call a per-concern API.
func EncryptChannelConfig(plaintext string) (string, error)

// DecryptChannelConfig is the inverse of EncryptChannelConfig.
func DecryptChannelConfig(ciphertext string) (string, error)
```

Internally both call the existing `Encrypt`/`Decrypt`.

## `pkg/crypto/credential.go` (new)

Four wrappers — two per encrypted field on `ResourceCredential`.

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

All four call `Encrypt`/`Decrypt`.

## GORM Hooks (modified bodies, unchanged signatures)

| Hook | Today | After |
|------|-------|-------|
| `NotificationChannel.BeforeCreate(tx)` | calls `crypto.Encrypt` | calls `crypto.EncryptChannelConfig` |
| `NotificationChannel.BeforeUpdate(tx)` | calls `crypto.Encrypt` | calls `crypto.EncryptChannelConfig` |
| `NotificationChannel.AfterFind(tx)` | legacy branch + `crypto.Decrypt` | legacy branch (write via `tx` preserved) + `crypto.DecryptChannelConfig` / `crypto.EncryptChannelConfig` |
| `ResourceCredential.BeforeCreate(tx)` | calls `crypto.Encrypt` on Password+Options | calls `crypto.EncryptCredentialPassword`/`…Options` |
| `ResourceCredential.BeforeUpdate(tx)` | same | same |
| `ResourceCredential.AfterFind(tx)` | same | calls `crypto.DecryptCredentialPassword`/`…Options` |

**Signature contract**: signatures, count, names of hook methods are unchanged.

## Fakes (modified) — `internal/repository/fake/`

Each fake's INSERT path replaces `entity.BeforeCreate(nil)` with `entity.Base.EnsureID()` (or `entity.EnsureID()` if `Base` is embedded directly, depending on access path). Fakes do NOT carry encryption — they store plaintext domain values, mirroring today's behavior.

Confirmed sites (7):
- `incident_fake.go`
- `notification_fake.go`
- `component_fake.go`
- `monitoring_activity_fake.go`
- `incident_event_step_fake.go`
- `api_key_fake.go`
- `resource_fake.go`

Re-grep at impl time to catch any newly-added fake.

## Out of scope (no changes)

- Domain field renames, table renames, schema migrations.
- Removal of any `gorm:"…"` tag.
- Removal of any GORM hook method.
- Encryption key derivation, key rotation, on-disk format.
- Anything under `internal/api/`, `internal/service/`, `internal/repository/store/`, `web/`.
