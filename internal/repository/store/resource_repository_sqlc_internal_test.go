package store

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denisakp/ogoune/internal/port"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

// NewResourceRepositorySQLCForTest builds a ResourceRepositorySQLC from
// pre-constructed *Queries (typically wrapped around a query-counting DBTX).
// Lives in *_internal_test.go so it's only reachable from tests but can still
// be called from store_test package via the public symbol.
//
// Production code MUST continue to use NewResourceRepositorySQLC(rt).
// Spec 049 §FR-001 enabler — feeds the round-trip-bound test.
func NewResourceRepositorySQLCForTest(
	pgQ *pgsqlc.Queries,
	sqliteQ *sqlitesqlc.Queries,
	pgPool *pgxpool.Pool,
	sqliteDB *sql.DB,
) port.ResourceRepository {
	return &ResourceRepositorySQLC{
		pgQ:      pgQ,
		sqliteQ:  sqliteQ,
		pgPool:   pgPool,
		sqliteDB: sqliteDB,
	}
}
