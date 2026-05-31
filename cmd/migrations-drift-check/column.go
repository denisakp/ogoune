package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ColumnDef describes one column extracted from a CREATE TABLE or ALTER TABLE ADD COLUMN.
type ColumnDef struct {
	Table      string
	Column     string
	NotNull    bool
	SourceFile string
	Line       int
}

// SchemaSnapshot maps table → column → ColumnDef.
type SchemaSnapshot map[string]map[string]ColumnDef

var (
	createTableRe   = regexp.MustCompile(`(?is)^\s*CREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+([A-Za-z_][A-Za-z0-9_]*)\s*\((.*)\)\s*$`)
	alterTableHdrRe = regexp.MustCompile(`(?is)^\s*ALTER\s+TABLE\s+([A-Za-z_][A-Za-z0-9_]*)\s+(.*)$`)
	addColumnRe     = regexp.MustCompile(`(?is)^\s*ADD\s+COLUMN(?:\s+IF\s+NOT\s+EXISTS)?\s+([A-Za-z_][A-Za-z0-9_]*)\b(.*)$`)
	columnLeadRe    = regexp.MustCompile(`(?s)^\s*([A-Za-z_][A-Za-z0-9_]*)\b(.*)$`)
	skipKwRe        = regexp.MustCompile(`(?i)^\s*(CONSTRAINT|PRIMARY\s+KEY|FOREIGN\s+KEY|UNIQUE|CHECK|INDEX)\b`)
)

// buildSnapshot scans the given migration files in order and returns a snapshot.
// Files MUST be passed in numerical (prefix) order so later ALTERs override earlier defs.
func buildSnapshot(files []*MigrationFile) (SchemaSnapshot, error) {
	snap := make(SchemaSnapshot)
	for _, f := range files {
		if f == nil {
			continue
		}
		if err := scanFile(f.Path, snap); err != nil {
			return nil, err
		}
	}
	return snap, nil
}

func scanFile(path string, snap SchemaSnapshot) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	stripped, lineMap := stripCommentsTrackLines(string(data))
	statements := splitStatements(stripped)
	for _, stmt := range statements {
		text := strings.TrimSpace(stmt.text)
		if text == "" {
			continue
		}
		line := lineMap[stmt.offset]
		if m := createTableRe.FindStringSubmatch(text); m != nil {
			table := strings.ToLower(m[1])
			body := m[2]
			for _, part := range splitTopLevel(body, ',') {
				seg := strings.TrimSpace(part)
				if seg == "" || skipKwRe.MatchString(seg) {
					continue
				}
				cm := columnLeadRe.FindStringSubmatch(seg)
				if cm == nil {
					continue
				}
				col := strings.ToLower(cm[1])
				rest := cm[2]
				notNull := isNotNull(rest) || strings.Contains(strings.ToUpper(rest), "PRIMARY KEY")
				addColumn(snap, ColumnDef{Table: table, Column: col, NotNull: notNull, SourceFile: path, Line: line})
			}
			continue
		}
		if m := alterTableHdrRe.FindStringSubmatch(text); m != nil {
			table := strings.ToLower(m[1])
			tail := m[2]
			// An ALTER TABLE may contain multiple comma-separated actions:
			//   ALTER TABLE x ADD COLUMN a INT, ADD COLUMN b TEXT, DROP COLUMN c
			// We split at top-level commas and pick up each ADD COLUMN clause.
			for _, action := range splitTopLevel(tail, ',') {
				am := addColumnRe.FindStringSubmatch(action)
				if am == nil {
					continue
				}
				col := strings.ToLower(am[1])
				rest := am[2]
				addColumn(snap, ColumnDef{Table: table, Column: col, NotNull: isNotNull(rest), SourceFile: path, Line: line})
			}
		}
	}
	return nil
}

func addColumn(snap SchemaSnapshot, c ColumnDef) {
	cols, ok := snap[c.Table]
	if !ok {
		cols = make(map[string]ColumnDef)
		snap[c.Table] = cols
	}
	cols[c.Column] = c
}

func isNotNull(rest string) bool {
	return strings.Contains(strings.ToUpper(rest), "NOT NULL")
}

// stmt is a single SQL statement with its byte offset into the comment-stripped source.
type stmt struct {
	text   string
	offset int
}

// splitStatements splits SQL source into statements terminated by `;` at paren depth 0.
// String literals are NOT specially handled — our migrations don't put semicolons in strings.
func splitStatements(src string) []stmt {
	var out []stmt
	depth := 0
	start := 0
	for i, r := range src {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ';':
			if depth == 0 {
				out = append(out, stmt{text: src[start:i], offset: start})
				start = i + 1
			}
		}
	}
	if start < len(src) {
		out = append(out, stmt{text: src[start:], offset: start})
	}
	return out
}

// splitTopLevel splits s by sep at paren depth 0.
func splitTopLevel(s string, sep rune) []string {
	var out []string
	depth := 0
	start := 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case sep:
			if depth == 0 {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// stripCommentsTrackLines removes `--` line comments and returns the cleaned
// source plus a byte-offset → original-line-number map.
func stripCommentsTrackLines(src string) (string, map[int]int) {
	var b strings.Builder
	b.Grow(len(src))
	lineMap := make(map[int]int)
	line := 1
	inComment := false
	for i := 0; i < len(src); i++ {
		c := src[i]
		if !inComment && c == '-' && i+1 < len(src) && src[i+1] == '-' {
			inComment = true
			i++
			continue
		}
		if inComment {
			if c == '\n' {
				inComment = false
				lineMap[b.Len()] = line
				b.WriteByte('\n')
				line++
			}
			continue
		}
		lineMap[b.Len()] = line
		b.WriteByte(c)
		if c == '\n' {
			line++
		}
	}
	return b.String(), lineMap
}

// diffSnapshots compares two snapshots and emits drift messages.
// Type tokens are NOT compared (Clarification Q2 / Research R5).
func diffSnapshots(pg, sqlite SchemaSnapshot) []string {
	var msgs []string

	tables := make(map[string]struct{})
	for t := range pg {
		tables[t] = struct{}{}
	}
	for t := range sqlite {
		tables[t] = struct{}{}
	}

	for table := range tables {
		pgCols := pg[table]
		sqCols := sqlite[table]
		cols := make(map[string]struct{})
		for c := range pgCols {
			cols[c] = struct{}{}
		}
		for c := range sqCols {
			cols[c] = struct{}{}
		}
		for col := range cols {
			pCol, pOk := pgCols[col]
			sCol, sOk := sqCols[col]
			switch {
			case pOk && !sOk:
				msgs = append(msgs, fmt.Sprintf("column-name drift: table=%s, column=%s present in postgres (%s:%d) but missing in sqlite",
					table, col, pCol.SourceFile, pCol.Line))
			case !pOk && sOk:
				msgs = append(msgs, fmt.Sprintf("column-name drift: table=%s, column=%s present in sqlite (%s:%d) but missing in postgres",
					table, col, sCol.SourceFile, sCol.Line))
			case pOk && sOk && pCol.NotNull != sCol.NotNull:
				msgs = append(msgs, fmt.Sprintf("nullability drift: table=%s, column=%s, postgres=%s (%s:%d), sqlite=%s (%s:%d)",
					table, col,
					nullLabel(pCol.NotNull), pCol.SourceFile, pCol.Line,
					nullLabel(sCol.NotNull), sCol.SourceFile, sCol.Line))
			}
		}
	}
	return msgs
}

func nullLabel(notNull bool) string {
	if notNull {
		return "NOT_NULL"
	}
	return "nullable"
}
