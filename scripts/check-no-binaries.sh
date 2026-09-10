#!/usr/bin/env bash
#
# Fail if any tracked file is a compiled binary.
#
# A .gitignore entry only stops the artifacts someone thought to name. This
# checks what is actually in the index, so an artifact with an unforeseen name
# is caught the first time rather than after it has been in history for weeks.
#
# It happened: `go build ./cmd/agent` writes ./agent into the working directory,
# and a 9 MB macOS binary plus an 8 MB Windows one reached the repository that
# way. Nothing built from this repo belongs in git.
set -euo pipefail

cd "$(dirname "$0")/.."

# Types that are legitimately binary and legitimately tracked: images, fonts,
# archives, and anything git itself stores as LFS pointers.
allowed_mime='^(image/|font/|application/font|application/vnd\.ms-fontobject|application/zip|application/gzip|application/x-tar|application/pdf)'

offenders=""
while IFS= read -r -d '' f; do
	mime="$(file --mime-type -b -- "$f" 2>/dev/null || echo unknown)"
	case "$mime" in
		text/* | inode/* | application/json | application/xml | application/javascript | application/x-empty)
			continue
			;;
	esac
	if [[ "$mime" =~ $allowed_mime ]]; then
		continue
	fi
	offenders+="  $f  ($mime)"$'\n'
done < <(git ls-files -z)

if [[ -n "$offenders" ]]; then
	echo "=== Compiled binaries are tracked in git ==="
	printf '%s' "$offenders"
	cat <<'MSG'

Nothing built from this repository belongs in version control.

If this is a build artifact:
    git rm --cached <file> && rm <file>
and add it to .gitignore. A bare `go build ./cmd/<name>` writes ./<name> into
the working directory — use `go build -o dist/<name> ./cmd/<name>` instead.

If it is a genuine asset (an image, a font), add its type to allowed_mime in
scripts/check-no-binaries.sh, with a note saying why.
MSG
	exit 1
fi

echo "no compiled binaries tracked"
