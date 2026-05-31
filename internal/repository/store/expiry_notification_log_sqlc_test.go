package store_test

import (
	"testing"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestExpiryNotificationLogRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedResources(t, fx, "res-enl-1", "res-enl-2", "res-enl-old")
		repo := store.NewExpiryNotificationLogRepositorySQLC(fx.Runtime)
		runExpiryNotificationLogContract(t, repo)
	})
}
