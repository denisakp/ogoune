package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// MigrationFile represents one .sql migration file located under a dialect directory.
type MigrationFile struct {
	Prefix  string // four-digit numeric prefix, e.g. "0013"
	Name    string // slug after the prefix
	Path    string
	Dialect string // "postgres" or "sqlite"
}

// MigrationPair pairs the postgres + sqlite file sharing a numeric prefix.
type MigrationPair struct {
	Prefix   string
	Postgres *MigrationFile
	SQLite   *MigrationFile
}

var prefixRe = regexp.MustCompile(`^(\d{4})_([A-Za-z0-9_]+)\.sql$`)

// listMigrations walks one dialect directory and returns its migration files keyed by prefix.
func listMigrations(root, dialect string) (map[string]*MigrationFile, error) {
	dir := filepath.Join(root, dialect)
	out := make(map[string]*MigrationFile)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		m := prefixRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		out[m[1]] = &MigrationFile{
			Prefix:  m[1],
			Name:    m[2],
			Path:    filepath.Join(dir, e.Name()),
			Dialect: dialect,
		}
	}
	return out, nil
}

// buildPairs combines per-dialect listings into prefix-keyed pairs, sorted by prefix.
func buildPairs(pg, sqlite map[string]*MigrationFile) []MigrationPair {
	prefixes := make(map[string]struct{}, len(pg)+len(sqlite))
	for k := range pg {
		prefixes[k] = struct{}{}
	}
	for k := range sqlite {
		prefixes[k] = struct{}{}
	}
	sorted := make([]string, 0, len(prefixes))
	for k := range prefixes {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	pairs := make([]MigrationPair, 0, len(sorted))
	for _, p := range sorted {
		pairs = append(pairs, MigrationPair{Prefix: p, Postgres: pg[p], SQLite: sqlite[p]})
	}
	return pairs
}

// checkPairs returns a list of human-readable drift messages for unpaired prefixes.
func checkPairs(pairs []MigrationPair) []string {
	var msgs []string
	for _, p := range pairs {
		switch {
		case p.Postgres == nil:
			msgs = append(msgs, fmt.Sprintf("missing pair for prefix %s: postgres=(missing), sqlite=%s", p.Prefix, p.SQLite.Path))
		case p.SQLite == nil:
			msgs = append(msgs, fmt.Sprintf("missing pair for prefix %s: postgres=%s, sqlite=(missing)", p.Prefix, p.Postgres.Path))
		}
	}
	return msgs
}
