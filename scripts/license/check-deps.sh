#!/usr/bin/env bash
# check-deps.sh — Runtime-dependency license guard
# Implements FR-008, SC-003.
# Contract: specs/040-relicense-apache-ee/contracts/deps-license.contract.md
#
# Scope:
#   - Go: ./cmd/..., ./internal/... (excluding ./internal/ee/...), ./pkg/...
#   - Web: web/package.json `dependencies` only (excluding `devDependencies`)
#
# Denied license families: GPL-*, LGPL-*, AGPL-*, SSPL-*, BUSL-*.
# Unknown license = denied.
#
# Artefacts:
#   dist/license-report-go.csv
#   dist/license-report-web.json

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

SCRIPT_DIR="$REPO_ROOT/scripts/license"
DENIED_FILE="$SCRIPT_DIR/denied-deps-licenses.txt"

mkdir -p dist

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'check-deps: missing tool %s. Install: %s\n' "$1" "$2" >&2
    exit 2
  fi
}

require go "https://go.dev/dl/"
require jq "brew install jq  # or: apt install jq"
require pnpm "npm install -g pnpm"

if ! command -v go-licenses >/dev/null 2>&1; then
  if [ -x "${GOPATH:-$HOME/go}/bin/go-licenses" ]; then
    export PATH="${GOPATH:-$HOME/go}/bin:$PATH"
  else
    printf 'check-deps: missing tool go-licenses. Install: go install github.com/google/go-licenses@v1.6.0\n' >&2
    exit 2
  fi
fi

# Build the denied-regex from the configuration file (skip comments + blanks).
denied_regex="$(grep -vE '^\s*(#|$)' "$DENIED_FILE" | paste -sd'|' -)"
if [ -z "$denied_regex" ]; then
  printf 'check-deps: denied-deps-licenses.txt is empty — refusing to run\n' >&2
  exit 2
fi

fail=0

# --- Go audit -----------------------------------------------------------------
# `go-licenses report` lists every transitive dep with its detected license.
# We exclude internal/ee paths from the failure scan (EE scope governed separately)
# but they still appear in the archived report for traceability.

MODULE_PATH="$(awk '/^module /{print $2}' go.mod)"
EE_PREFIX="${MODULE_PATH}/internal/ee/"

# Known-good packages whose license go-licenses cannot detect from the upstream
# module (no LICENSE file in distributed module; license is identified in the
# project's own repository or documentation). Each entry is the package path
# prefix and the known SPDX licence (documented for traceability).
#
#   modernc.org/mathutil      — BSD-3-Clause (https://gitlab.com/cznic/mathutil)
#
# These prefixes are passed to `--ignore` so go-licenses does not fail the
# build; they are listed manually in dist/license-report-go.csv below.
GO_IGNORE=(
  "modernc.org/mathutil"
)

ignore_args=()
for p in "${GO_IGNORE[@]}"; do
  ignore_args+=("--ignore=$p")
done

printf 'check-deps: running go-licenses csv on Go runtime scope...\n' >&2
if ! go-licenses csv \
       "${ignore_args[@]}" \
       ./cmd/... ./internal/... ./pkg/... \
       > dist/license-report-go.csv 2> dist/license-report-go.err; then
  cat dist/license-report-go.err >&2
  printf '::error::go-licenses csv failed — fix above errors and retry\n'
  fail=1
fi
rm -f dist/license-report-go.err

# Append known-good packages to the report for completeness (manual entries).
{
  printf 'modernc.org/mathutil,https://gitlab.com/cznic/mathutil,BSD-3-Clause\n'
} >> dist/license-report-go.csv

if [ -s dist/license-report-go.csv ]; then
  # CSV columns: <module>,<repo>,<license name>
  while IFS=, read -r dep _repo license; do
    [ -z "$dep" ] && continue
    case "$dep" in
      "$EE_PREFIX"*) continue ;;
      "$MODULE_PATH"*) continue ;;  # our own module
    esac
    if printf '%s' "$license" | grep -qE "^(${denied_regex})"; then
      printf '::error::Go runtime dependency %s has disallowed license %s\n' "$dep" "$license"
      fail=1
    fi
    if [ "$license" = "Unknown" ] || [ -z "$license" ]; then
      printf '::error::Go runtime dependency %s has unknown license — clarify upstream or add to GO_IGNORE with documented licence\n' "$dep"
      fail=1
    fi
  done < dist/license-report-go.csv
fi

# --- Web audit ----------------------------------------------------------------
printf 'check-deps: running pnpm licenses ls --prod on web/...\n' >&2
if ! pnpm -C web licenses ls --prod --json > dist/license-report-web.json 2> dist/license-report-web.err; then
  cat dist/license-report-web.err >&2
  printf '::error::pnpm licenses ls failed — fix above errors and retry\n'
  fail=1
fi
rm -f dist/license-report-web.err

if [ -s dist/license-report-web.json ]; then
  # pnpm output is a map of license -> [{name,version,...}]
  # Flatten + filter for denied regex.
  bad=$(jq -r --arg re "^(${denied_regex})" '
    to_entries
    | map(
        .key as $lic
        | .value[]?
        | select($lic | test($re))
        | "\(.name)@\(.version): \($lic)"
      )
    | .[]?
  ' dist/license-report-web.json 2>/dev/null || true)
  if [ -n "$bad" ]; then
    while IFS= read -r line; do
      printf '::error::Web runtime dependency %s — disallowed license\n' "$line"
      fail=1
    done <<< "$bad"
  fi

  # Unknown licenses on the web side
  unknown=$(jq -r '
    to_entries
    | map(
        select(.key == "" or .key == null or (.key | ascii_downcase) == "unknown")
        | .value[]?
        | "\(.name)@\(.version)"
      )
    | .[]?
  ' dist/license-report-web.json 2>/dev/null || true)
  if [ -n "$unknown" ]; then
    while IFS= read -r line; do
      printf '::error::Web runtime dependency %s has unknown license — clarify upstream or replace\n' "$line"
      fail=1
    done <<< "$unknown"
  fi
fi

# --- Summary ------------------------------------------------------------------
go_count=$(wc -l < dist/license-report-go.csv 2>/dev/null | tr -d ' ' || echo 0)
web_count=$(jq -r '[.[] | length] | add // 0' dist/license-report-web.json 2>/dev/null || echo 0)

if [ "$fail" -eq 0 ]; then
  printf 'deps-license: OK — Go: %s modules audited, Web: %s packages audited\n' "$go_count" "$web_count" >&2
  exit 0
fi

printf 'deps-license: FAILED\n' >&2
exit 1
