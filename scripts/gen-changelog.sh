#!/usr/bin/env bash
#
# gen-changelog.sh — writes CHANGELOG.md from git. Do not edit CHANGELOG.md by hand.
#
# An entry is a commit whose subject carries a version, "type(vX.Y.Z): message" —
# every commit made by commit.sh (and the initial commit). Newest first.
# commit.sh runs this BEFORE it creates the go/vX.Y.Z tag, so the entry comes from
# the subject, not from the tag.
set -euo pipefail

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
OUT="$ROOT/CHANGELOG.md"

{
  echo "# Changelog"
  echo
  echo "> Generated from git by \`scripts/gen-changelog.sh\` — do not edit by hand."
  echo
  git -C "$ROOT" log --pretty=$'%cs\t%s' | awk -F'\t' '
    match($2, /^(feat|fix|refactor|docs|chore|test|major|build|perf)\(v[0-9]+\.[0-9]+\.[0-9]+\): /) {
      subj = $2
      type = substr(subj, 1, index(subj, "(") - 1)
      ver = substr(subj, index(subj, "(") + 1); ver = substr(ver, 1, index(ver, ")") - 1)
      printf "- **%s** (%s) · `%s` — %s\n", ver, $1, type, substr(subj, RLENGTH + 1)
    }'
} > "$OUT"

echo "CHANGELOG.md generated: $(grep -c '^- ' "$OUT") entries."
