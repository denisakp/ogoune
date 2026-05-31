package store_test

import (
	"testing"

	"github.com/denisakp/ogoune/internal/repository/internaltest"
	"github.com/denisakp/ogoune/internal/repository/store"
)

func TestIncidentDiagnosticsRepository_SqlcContract(t *testing.T) {
	internaltest.ForEachDialect(t, func(t *testing.T, fx *internaltest.DialectFixture) {
		seedIncidents(t, fx, "inc-diag-1", "inc-diag-2", "inc-diag-3")
		repo := store.NewIncidentDiagnosticsRepositorySQLC(fx.Runtime)
		runIncidentDiagnosticsContract(t, repo)
	})
}
