#!/usr/bin/env bash
# check-docs.sh — Documentation AGPL-drift guard
# Implements FR-007, SC-004.
# Contract: specs/040-relicense-apache-ee/contracts/docs-drift.contract.md
#
# Disallowed patterns (case-insensitive) on README.md, CONTRIBUTING.md,
# roadmap.md, and web/src/**/*.{vue,ts}. Lines containing any allowlist
# phrase from docs-allowlist.txt are exempt (historical / explanatory mentions).

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

ALLOWLIST_FILE="$REPO_ROOT/scripts/license/docs-allowlist.txt"

# Build the allowlist regex from the configuration file (skip comments/blanks).
allowlist_regex="$(grep -vE '^\s*(#|$)' "$ALLOWLIST_FILE" | sed 's/[][\/.*^$]/\\&/g' | paste -sd'|' -)"

# Patterns considered current-tense AGPL claims on the core scope.
# (Anchored with -E -i grep below.)
patterns=(
  'license-AGPL'
  'under AGPL'
  'under the GNU Affero'
  'AGPL v3[[:space:]]*\|'
  'AGPL v3 — see'
  'licence AGPL'
  'licence AGPL v3'
)

# File set in scope.
scan_targets=(
  README.md
  CONTRIBUTING.md
  roadmap.md
)

# Web file set — Vue + TS only.
while IFS= read -r f; do
  [ -n "$f" ] && scan_targets+=("$f")
done < <(find web/src -type f \( -name '*.vue' -o -name '*.ts' \) 2>/dev/null | sort)

fail=0
scanned=0

for f in "${scan_targets[@]}"; do
  [ -f "$f" ] || continue
  scanned=$((scanned + 1))

  # Build a single regex from the disallowed pattern set.
  joined="$(IFS='|'; echo "${patterns[*]}")"

  # Web files (.vue/.ts) — any literal "AGPL" is disallowed unless allowlisted.
  case "$f" in
    web/src/*)
      pattern_re="AGPL|${joined}"
      ;;
    *)
      pattern_re="${joined}"
      ;;
  esac

  while IFS= read -r hit; do
    [ -z "$hit" ] && continue
    line="${hit#*:}"
    line="${line#*:}"
    # Skip lines matching the allowlist.
    if [ -n "$allowlist_regex" ] && printf '%s' "$line" | grep -qiE "$allowlist_regex"; then
      continue
    fi
    lineno="${hit#*:}"
    lineno="${lineno%%:*}"
    printf '::error file=%s,line=%s::current-tense AGPL claim — rewrite or add an allowlist phrase if historical\n' "$f" "$lineno"
    fail=1
  done < <(grep -nE -i "$pattern_re" "$f" 2>/dev/null || true)
done

if [ "$fail" -eq 0 ]; then
  printf 'docs-drift: OK — %d files scanned, 0 disallowed patterns\n' "$scanned" >&2
  exit 0
fi

printf 'docs-drift: FAILED\n' >&2
exit 1
