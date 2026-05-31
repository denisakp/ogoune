package store_test

import (
	"testing"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestStatusPageSettingsRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		repo := store.NewStatusPageSettingsRepositorySQLC(fx.Runtime)
		runStatusPageSettingsContract(t, repo)
	})
}
