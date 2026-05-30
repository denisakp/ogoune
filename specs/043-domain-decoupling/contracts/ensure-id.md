# Contract — `Base.EnsureID()`

Package: `internal/domain`

## Signature

```go
func (b *Base) EnsureID()
```

## Behavior

1. If `b.ID == ""`, generate a fresh ULID (monotonic-entropy, time-source `time.Now()`) and assign it to `b.ID`.
2. If `b.ID != ""`, no-op.
3. MUST NOT modify `b.CreatedAt` or `b.UpdatedAt`. Timestamp filling is GORM's responsibility via struct tags (`autoCreateTime`, `autoUpdateTime`).
4. MUST NOT take any parameter; specifically no `*gorm.DB`.
5. MUST be idempotent: calling twice yields the same `ID`.
6. MUST be safe to call from any goroutine on its own `*Base` (no shared global state beyond the ULID entropy source, which is already safe).

## Compatibility

The existing `Base.BeforeCreate(tx *gorm.DB) error` MUST remain. Its body becomes:

```go
func (base *Base) BeforeCreate(tx *gorm.DB) (err error) {
    base.EnsureID()
    return
}
```

This preserves the GORM persistence path identically.

## Invariants

- `EnsureID()` callers do not need an active database transaction.
- Calling `EnsureID()` followed by GORM `Create(...)` (which fires `BeforeCreate`) is safe — second call sees a non-empty ID and is a no-op.
- ULID format: 26 chars, alphabet `0123456789ABCDEFGHJKMNPQRSTVWXYZ`.

## Out of scope

- Generating `ID` for already-loaded rows (no-op by contract).
- Setting timestamps.
- Validating ULID lexical correctness (tested elsewhere).
