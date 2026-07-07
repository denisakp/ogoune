package store_test

import (
	"testing"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestAPIKeyRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedUsers(t, fx, "user-1", "user-dup", "user-2", "user-3", "user-4",
			"user-list", "user-count", "user-revoke", "user-revoke2", "user-lastused")
		repo := store.NewAPIKeyRepositorySQLC(fx.Runtime)
		runAPIKeyContract(t, repo)
	})
}
