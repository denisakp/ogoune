// Command migrations-drift-check enforces 1-to-1 file pairing and per-column
// name + nullability alignment between the Postgres and SQLite migration trees.
//
// See specs/042-sqlc-schema-source/contracts/drift-check-cli.md for the contract.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("migrations-drift-check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", "internal/database/migrations", "directory containing postgres/ and sqlite/ subdirs")
	verbose := fs.Bool("verbose", false, "log scanned files and counts")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	pgFiles, err := listMigrations(*root, "postgres")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	sqFiles, err := listMigrations(*root, "sqlite")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	pairs := buildPairs(pgFiles, sqFiles)
	var allMsgs []string
	allMsgs = append(allMsgs, checkPairs(pairs)...)

	// Build snapshots only from the files that have a counterpart; otherwise
	// we'd double-count drift for the same unpaired prefix.
	pgOrdered, sqOrdered := orderedPaired(pairs)

	pgSnap, err := buildSnapshot(pgOrdered)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	sqSnap, err := buildSnapshot(sqOrdered)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	if *verbose {
		fmt.Fprintf(stderr, "scanned %d postgres files, %d sqlite files; %d table(s) postgres, %d table(s) sqlite\n",
			len(pgOrdered), len(sqOrdered), len(pgSnap), len(sqSnap))
	}

	allMsgs = append(allMsgs, diffSnapshots(pgSnap, sqSnap)...)

	if len(allMsgs) == 0 {
		return 0
	}
	sort.Strings(allMsgs)
	for _, m := range allMsgs {
		fmt.Fprintln(stderr, m)
	}
	fmt.Fprintf(stderr, "migrations-drift-check: %d drift issue(s) found\n", len(allMsgs))
	return 1
}

func orderedPaired(pairs []MigrationPair) (pg, sq []*MigrationFile) {
	for _, p := range pairs {
		if p.Postgres != nil && p.SQLite != nil {
			pg = append(pg, p.Postgres)
			sq = append(sq, p.SQLite)
		}
	}
	return
}
