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
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		fullPath := path.Join(dir, entry.Name())
		contents, err := fs.ReadFile(migrationFS, fullPath)
		if err != nil {
			return nil, err
		}

		version, name := parseMigrationFileName(entry.Name())
		files = append(files, migrationFile{
			Version: version,
			Name:    name,
			Path:    fullPath,
			SQL:     string(contents),
		})
	}

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
