#!/usr/bin/env bash
# check-spdx.sh — SPDX coverage guard
# Implements FR-004, FR-005, SC-002.
# Contract: specs/040-relicense-apache-ee/contracts/spdx-coverage.contract.md
#
# Rules:
#   1. Every file under internal/ee/**/*.{go,sql} MUST contain the EE SPDX line
#      in its first 20 lines.
#   2. No file outside internal/ee/** MAY contain the EE SPDX identifier anywhere.
#   3. Exclusions: internal/ee/**/testdata/**, internal/ee/**/*_generated.go,
#      internal/ee/**/mocks/**.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

EE_ID="LicenseRef-Ogoune-EE"
GO_LINE="// SPDX-License-Identifier: ${EE_ID}"
SQL_LINE="-- SPDX-License-Identifier: ${EE_ID}"

fail=0
ee_count=0

is_excluded() {
  case "$1" in
    internal/ee/*/testdata/*|internal/ee/*_generated.go|internal/ee/*/mocks/*)
      return 0
      ;;
  esac
  return 1
}

# Rule 1: MUST-be-present on internal/ee/**/*.{go,sql}
while IFS= read -r f; do
  [ -z "$f" ] && continue
  if is_excluded "$f"; then continue; fi
  ee_count=$((ee_count + 1))
  case "$f" in
    *.go)  needle="$GO_LINE"  ;;
    *.sql) needle="$SQL_LINE" ;;
    *)     continue ;;
  esac
  if ! head -n 20 "$f" | grep -qxF "$needle"; then
    printf '::error file=%s::missing SPDX header: expected %q in first 20 lines\n' "$f" "$needle"
    fail=1
  fi
done < <(find internal/ee -type f \( -name '*.go' -o -name '*.sql' \) 2>/dev/null | sort)

# Rule 2: MUST-NOT-be-present outside internal/ee/**
# Targets the SPDX HEADER form only (not prose references). The header form is
# what misleads downstream tooling (SPDX scanners, SBOM generators) about a
# file's licence. Prose mentions in user docs, this script, the EE licence text
# itself, and design artefacts are intentional and harmless.
#
# Header form regex: a line containing 'SPDX-License-Identifier:' followed by
# any whitespace, then the EE identifier as a stand-alone token.
LEAK_RE='SPDX-License-Identifier:[[:space:]]*'"$EE_ID"'([[:space:]]|$)'

leaks=$(grep -rlnE "$LEAK_RE" . 2>/dev/null \
  | grep -vE '^\./(internal/ee/|scripts/license/|\.git/|node_modules/|web/node_modules/)' \
  || true)
if [ -n "$leaks" ]; then
  while IFS= read -r f; do
    printf '::error file=%s::EE SPDX identifier present outside internal/ee/ — move the file or remove the header\n' "${f#./}"
    fail=1
  done <<< "$leaks"
fi

if [ "$fail" -eq 0 ]; then
  printf 'spdx-coverage: OK — %d files under internal/ee/ checked, all carry the header\n' "$ee_count" >&2
  exit 0
fi

printf 'spdx-coverage: FAILED\n' >&2
exit 1
