# Quickstart — Maintainer Workflow

## Constructing a domain entity in any context

```go
e := domain.Resource{Name: "foo"}
e.EnsureID()                // pure: ID-only, no *gorm.DB needed
// e.ID is now a 26-char ULID, e.CreatedAt is still zero (GORM fills it on persist)
```

In an in-memory fake repository:

```go
func (f *FakeIncidentRepo) Create(i *domain.Incident) error {
    i.EnsureID()                                 // replaces i.BeforeCreate(nil)
    f.byID[i.ID] = i
    return nil
}
```

In GORM-backed code: no change. `db.Create(&entity)` continues to fire `Base.BeforeCreate(tx)` which calls `EnsureID()` internally.

## Encrypting / decrypting per concern

```go
// NotificationChannel.Config
cipher, err := crypto.EncryptChannelConfig(plaintextJSON)
plain,  err := crypto.DecryptChannelConfig(cipher)

// ResourceCredential.Password
cipher, err := crypto.EncryptCredentialPassword(rawPassword)
plain,  err := crypto.DecryptCredentialPassword(cipher)

// ResourceCredential.Options
cipher, err := crypto.EncryptCredentialOptions(jsonOptions)
plain,  err := crypto.DecryptCredentialOptions(cipher)
```

Wrappers delegate to `crypto.Encrypt`/`crypto.Decrypt`. On-disk format is unchanged — values encrypted by the generic primitive decrypt cleanly via the typed wrappers and vice-versa.

## Why both the wrappers AND the GORM hooks?

- GORM hooks stay in `internal/domain/models.go` because at least one GORM-backed repository still exists.
- Wrappers exist so the next sqlc-backed repository ticket can call a named, per-concern API without re-implementing AES/GCM.
- Both call the same underlying `crypto.Encrypt`/`Decrypt` — single source of truth, zero format drift.

## What changes for testing

Crypto tests use `crypto.SetGlobalProvider` to inject a deterministic test key. No `APP_SECRET_KEY` env var manipulation required.

```go
func TestMain(m *testing.M) {
    crypto.SetGlobalProvider(crypto.FixedKeyProvider(testKey))   // or equivalent
    os.Exit(m.Run())
}
```

## What does NOT change

- The `BeforeCreate` / `BeforeUpdate` / `AfterFind` hook signatures and methods.
- Any `gorm:"…"` struct tag on `internal/domain/models.go`.
- The on-disk encryption format.
- The application's runtime behavior, env vars, or operator-facing surfaces.
- Any code under `internal/api/`, `internal/service/`, `internal/repository/store/`, or `web/`.

Tag removal will happen repository-by-repository in later tickets (006/007/008 per `.prds/sqlc/`). Hook removal will happen in 010 once no GORM-backed repository remains.
