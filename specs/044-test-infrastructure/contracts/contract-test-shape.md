# Contract — Repository Contract Test Shape (post-refactor)

Each of the 6 contract tests under `internal/repository/store/` adopts the same shape after this ticket.

## Template

```go
package store

import (
    "testing"

    "github.com/denisakp/ogoune/internal/port"
    "github.com/denisakp/ogoune/internal/repository/internaltest"
    "gorm.io/gorm"
)

// Default factory wires the GORM-backed repository.
// Future sqlc-backed implementations add additional factory test functions
// that call runTagsContract with their own factory.
func TestTagsRepository_Contract(t *testing.T) {
    internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
        repo := NewTagsRepository(fx.Runtime.GormDB())
        runTagsContract(t, repo)
    })
}

// runTagsContract is the actual contract — preserved verbatim from the
// current fake-targeting test body, with one substitution: every reference
// to the fake's internal map / direct-poke is removed or moved to a
// sibling tags_repository_fake_test.go file.
func runTagsContract(t *testing.T, repo port.TagsRepository) {
    t.Run("Create", func(t *testing.T) { /* … */ })
    t.Run("GetByID", func(t *testing.T) { /* … */ })
    // … all existing per-method sub-tests preserved …
}
```

## Migration recipe (per contract file)

1. Rename existing `TestXxxRepository_Contract(t *testing.T)` body into `runXxxContract(t, repo)`.
2. New `TestXxxRepository_Contract(t *testing.T)` becomes the dual-dialect wrapper above.
3. Inside `runXxxContract`: substitute every `repo.<method>` to use the parameter, not a captured local fake.
4. Identify assertions that target the fake's internal state (e.g. `len(fake.byID) == N`). Move them to `internal/repository/store/<entity>_repository_fake_test.go` and keep the contract test free of fake-specific knowledge.
5. Drop the `fake.New<Entity>Fake()` import; add `internaltest` + `port`.

## What MUST be preserved

- Every assertion that describes an observable behavior of the repository (created row appears in `List`, `GetByID` returns the row, errors map correctly, etc.).
- Sub-test names (`"Create"`, `"GetByID"`, …) so failing-test selectors in CI logs remain stable.
- The package — files stay in `internal/repository/store/`.

## What MUST change

- The implementation under test: `fake.New…Fake()` → `NewXxxRepository(fx.Runtime.GormDB())`.
- The wrapping: bare `t.Run` → `internaltest.ForEachDialect` outside, per-method sub-tests inside.

## Sqlc extension point (future)

Once a sqlc-backed `tagsSqlcRepository` exists (ticket 005+):

```go
func TestTagsRepository_SqlcContract(t *testing.T) {
    internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
        // pgxpool / *sql.DB depending on dialect
        repo := newSqlcTagsRepository(fx.Runtime)
        runTagsContract(t, repo)
    })
}
```

The contract body is reused unchanged — exactly the property this refactor exists to enable.
