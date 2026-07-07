package store

import (
	"github.com/denisakp/ogoune/internal/port"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

// NewIncidentRepositorySQLCForTest builds an IncidentRepositorySQLC from
// pre-constructed *Queries (typically wrapped around a query-counting DBTX).
// Same pattern as NewResourceRepositorySQLCForTest. Spec 049 §FR-001 enabler.
//
// Note: IncidentRepositorySQLC does not currently carry pgPool/sqliteDB
// fields (no WithTx paths inside the impl); only the *Queries are needed.
func NewIncidentRepositorySQLCForTest(
	pgQ *pgsqlc.Queries,
	sqliteQ *sqlitesqlc.Queries,
) port.IncidentRepository {
	return &IncidentRepositorySQLC{
		pgQ:     pgQ,
		sqliteQ: sqliteQ,
	}
}
