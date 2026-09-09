package strategy

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

// MySQL health collection (spec 088). The mirror of the PostgreSQL side: same
// four signals, same names, same units, so a consumer never branches on engine.
//
// The statements are stateless, which matters here: database/sql pools, and even
// with MaxOpenConns(1) there is no guarantee of the same server-side session
// across calls. Nothing below sets a session variable or depends on one.

// mysqlMinSupportedVersion is the oldest server this enrichment reads. Below it
// the fields are omitted and unsupported_version is set. 8.0 is the branch still
// supported upstream, and it is where SHOW REPLICA STATUS was introduced -- older
// servers would need a second query shape for no good reason.
const mysqlMinSupportedVersion = 80000

// collectMySQLHealth gathers what the server says about itself. Returns nil when
// nothing at all could be read.
func collectMySQLHealth(ctx context.Context, db *sql.DB, budget time.Duration) *domain.DatabaseHealth {
	if budget <= 0 {
		return nil
	}
	start := time.Now()

	hctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	h := &domain.DatabaseHealth{}

	var versionStr string
	if err := db.QueryRowContext(hctx, "SELECT @@version").Scan(&versionStr); err != nil {
		return nil
	}
	if mysqlVersionNum(versionStr) < mysqlMinSupportedVersion {
		h.UnsupportedVersion = true
		h.CollectionDuration = time.Since(start)
		return h
	}

	// Saturation. Threads_connected and max_connections need no privilege beyond
	// connecting, mirroring the Postgres side. Both or neither.
	var active, max int64
	errA := db.QueryRowContext(hctx,
		"SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME = 'Threads_connected'").Scan(&active)
	errB := db.QueryRowContext(hctx, "SELECT @@max_connections").Scan(&max)
	if errA == nil && errB == nil {
		h.ConnectionsActive = &active
		h.ConnectionsMax = &max
	}

	collectMySQLPrivilegedHealth(hctx, db, h)

	h.CollectionDuration = time.Since(start)
	if h.ConnectionsActive == nil && h.LongestQuerySeconds == nil && h.ReplicationLagSeconds == nil && !h.UnsupportedVersion {
		return nil
	}
	return h
}

// collectMySQLPrivilegedHealth adds the two gated signals.
//
// The same trap as PostgreSQL, arrived at differently: without PROCESS, a role
// sees only ITS OWN threads in the process list. A maximum over what is visible
// therefore returns this monitor's own connection age and presents it as the
// database's longest running query -- plausible, wrong, undetectable. So the
// privilege is established first and the field is omitted otherwise (FR-013).
func collectMySQLPrivilegedHealth(ctx context.Context, db *sql.DB, h *domain.DatabaseHealth) {
	var canSeeAllThreads bool
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) > 0 FROM information_schema.user_privileges
		WHERE GRANTEE = CONCAT("'", SUBSTRING_INDEX(CURRENT_USER(), '@', 1), "'@'",
		                       SUBSTRING_INDEX(CURRENT_USER(), '@', -1), "'")
		  AND PRIVILEGE_TYPE IN ('PROCESS', 'SUPER')`).Scan(&canSeeAllThreads); err != nil {
		h.PrivilegeLimited = true
		return
	}
	if !canSeeAllThreads {
		h.PrivilegeLimited = true
		return
	}

	// Longest running statement: its AGE only, never the statement itself
	// (FR-007). PROCESSLIST_INFO carries the SQL text and is deliberately not
	// selected.
	var longest sql.NullFloat64
	if err := db.QueryRowContext(ctx, `
		SELECT MAX(PROCESSLIST_TIME) FROM performance_schema.threads
		WHERE PROCESSLIST_COMMAND = 'Query' AND PROCESSLIST_ID <> CONNECTION_ID()`).
		Scan(&longest); err == nil && longest.Valid {
		v := longest.Float64
		h.LongestQuerySeconds = &v
	}

	collectMySQLReplicationLag(ctx, db, h)
}

// collectMySQLReplicationLag reads the replica's own lag. As on PostgreSQL the
// two sides of replication answer different questions: a primary knows about its
// connected replicas, a replica knows how far behind it is. This reads the replica
// side, which is the one an operator monitors.
//
// Absent when this server is not a replica -- never 0, which would claim it is
// perfectly in sync (FR-015).
func collectMySQLReplicationLag(ctx context.Context, db *sql.DB, h *domain.DatabaseHealth) {
	var lag sql.NullFloat64
	err := db.QueryRowContext(ctx, `
		SELECT TIMESTAMPDIFF(SECOND, LAST_APPLIED_TRANSACTION_ORIGINAL_COMMIT_TIMESTAMP, NOW())
		FROM performance_schema.replication_applier_status_by_worker
		ORDER BY 1 DESC LIMIT 1`).Scan(&lag)
	if err != nil || !lag.Valid {
		// No replication configured, or the table is not readable. Either way the
		// honest answer is absence.
		return
	}
	v := lag.Float64
	h.ReplicationLagSeconds = &v
}

// mysqlVersionNum turns "8.0.36-log" into 80036 for comparison, mirroring the
// server_version_num PostgreSQL exposes directly. Returns 0 when the string
// cannot be parsed, which reads as "below the floor" and skips the enrichment --
// the safe direction.
func mysqlVersionNum(v string) int {
	// Strip any suffix: "8.0.36-0ubuntu0.22.04.1" -> "8.0.36"
	if i := strings.IndexAny(v, "-+ "); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return 0
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	patch := 0
	if len(parts) > 2 {
		patch, _ = strconv.Atoi(parts[2])
	}
	return major*10000 + minor*100 + patch
}
