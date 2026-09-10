package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/postgres/*.sql migrations/sqlite/*.sql
var embeddedMigrations embed.FS

type migrationFS = fs.FS

type migrationFile struct {
	Version string
	Name    string
	Path    string
	SQL     string
}

// runMigrations applies pending SQL migration files in lexicographic order
// against the provided *sql.DB. Constitution IV gate: fails fast on any apply
// error; the returned error wraps the failing file path and dialect.
func runMigrations(ctx context.Context, db *sql.DB, driver Driver, migrationFS migrationFS) error {
	migrations, err := loadMigrations(migrationFS, driver)
	if err != nil {
		return fmt.Errorf("db init: failed to load migrations: %w", err)
	}
	if len(migrations) == 0 {
		return fmt.Errorf("db init: no migrations found for driver %s", driver)
	}

	bootstrap := migrations[0]
	if err := executeStatements(ctx, db, splitSQLStatements(bootstrap.SQL)); err != nil {
		return fmt.Errorf("db init: failed to bootstrap schema_migrations table (%s on %s): %w", bootstrap.Path, driver, err)
	}
	if err := recordMigration(ctx, db, driver, bootstrap); err != nil {
		return fmt.Errorf("db init: failed to record bootstrap migration (%s on %s): %w", bootstrap.Path, driver, err)
	}

	appliedVersions, err := appliedMigrationVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("db init: failed to query applied migrations: %w", err)
	}

	for _, migration := range migrations[1:] {
		if _, ok := appliedVersions[migration.Version]; ok {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("db init: failed to start migration transaction for %s on %s: %w", migration.Path, driver, err)
		}

		if err := executeStatementsTx(ctx, tx, splitSQLStatements(migration.SQL)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db init: migrate %s on %s: %w", migration.Path, driver, err)
		}
		if err := recordMigrationTx(ctx, tx, driver, migration); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("db init: failed to record migration %s on %s: %w", migration.Path, driver, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("db init: failed to commit migration %s on %s: %w", migration.Path, driver, err)
		}
	}

	return nil
}

func loadMigrations(migrationFS migrationFS, driver Driver) ([]migrationFile, error) {
	dir := path.Join("migrations", string(driver))
	entries, err := fs.ReadDir(migrationFS, dir)
	if err != nil {
		return nil, err
	}

	files := make([]migrationFile, 0, len(entries))
	seen := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		// Undo scripts are not migrations. This loader used to pick them up with
		// everything else and execute them FORWARD, which is only survivable
		// because ".down.sql" sorts before ".up.sql" and the statements are all
		// IF EXISTS: the drop ran against a table that did not exist yet, then the
		// up created it. Nothing guarantees that order -- versions are compared,
		// and sort.Slice makes no promise about ties -- so the day it flipped, a
		// fresh install would have created a table and immediately dropped it.
		//
		// It also made schema_migrations lie: 19 of 34 rows on a fresh database
		// recorded the name of a down file as the applied migration.
		//
		// There is no down path in this migrator. These files exist as
		// documentation of the inverse and are kept paired by
		// migrations-drift-check; they are never run.
		if strings.HasSuffix(entry.Name(), ".down.sql") {
			continue
		}

		fullPath := path.Join(dir, entry.Name())
		contents, err := fs.ReadFile(migrationFS, fullPath)
		if err != nil {
			return nil, err
		}

		version, name := parseMigrationFileName(entry.Name())
		// Applied state is keyed on the version alone, so two migrations sharing
		// one prefix are indistinguishable: on a database that already recorded
		// that version, the second would be skipped forever -- silently, and only
		// on installs that upgrade rather than start fresh. Refusing to start is
		// the only honest answer, and it is an answer the author gets immediately
		// rather than an operator gets in production.
		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf(
				"db init: migrations %q and %q share version %q; every migration needs its own number",
				other, entry.Name(), version)
		}
		seen[version] = entry.Name()

		files = append(files, migrationFile{
			Version: version,
			Name:    name,
			Path:    fullPath,
			SQL:     string(contents),
		})
	}

	// Versions are unique by the check above, so this ordering is total and the
	// comparison never has to break a tie.
	sort.Slice(files, func(i, j int) bool {
		return files[i].Version < files[j].Version
	})

	return files, nil
}

func parseMigrationFileName(name string) (string, string) {
	trimmed := strings.TrimSuffix(name, ".sql")
	parts := strings.SplitN(trimmed, "_", 2)
	if len(parts) == 1 {
		return trimmed, trimmed
	}
	return parts[0], parts[1]
}

func appliedMigrationVersions(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = struct{}{}
	}

	return versions, rows.Err()
}

func recordMigration(ctx context.Context, db *sql.DB, driver Driver, migration migrationFile) error {
	query := recordQuery(driver)
	_, err := db.ExecContext(ctx, query, migration.Version, migration.Name, time.Now().UTC())
	return err
}

func recordMigrationTx(ctx context.Context, tx *sql.Tx, driver Driver, migration migrationFile) error {
	query := recordQuery(driver)
	_, err := tx.ExecContext(ctx, query, migration.Version, migration.Name, time.Now().UTC())
	return err
}

func recordQuery(driver Driver) string {
	switch driver {
	case DriverPostgres:
		return "INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3) ON CONFLICT (version) DO NOTHING"
	case DriverSQLite:
		return "INSERT OR IGNORE INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)"
	}
	return "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)"
}

func executeStatements(ctx context.Context, db *sql.DB, statements []string) error {
	for _, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func executeStatementsTx(ctx context.Context, tx *sql.Tx, statements []string) error {
	for _, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func splitSQLStatements(sql string) []string {
	var (
		statements     []string
		builder        strings.Builder
		inSingleQuote  bool
		inLineComment  bool
		inBlockComment bool
	)

	runes := []rune(sql)
	for index := 0; index < len(runes); index++ {
		current := runes[index]
		var next rune
		if index+1 < len(runes) {
			next = runes[index+1]
		}

		if inLineComment {
			builder.WriteRune(current)
			if current == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			builder.WriteRune(current)
			if current == '*' && next == '/' {
				builder.WriteRune(next)
				index++
				inBlockComment = false
			}
			continue
		}

		if !inSingleQuote && current == '-' && next == '-' {
			builder.WriteRune(current)
			builder.WriteRune(next)
			index++
			inLineComment = true
			continue
		}
		if !inSingleQuote && current == '/' && next == '*' {
			builder.WriteRune(current)
			builder.WriteRune(next)
			index++
			inBlockComment = true
			continue
		}

		if current == '\'' {
			inSingleQuote = !inSingleQuote
			builder.WriteRune(current)
			continue
		}

		if current == ';' && !inSingleQuote {
			statement := strings.TrimSpace(builder.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			builder.Reset()
			continue
		}

		builder.WriteRune(current)
	}

	if trailing := strings.TrimSpace(builder.String()); trailing != "" {
		statements = append(statements, trailing)
	}

	return statements
}
