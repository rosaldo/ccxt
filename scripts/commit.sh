#!/usr/bin/env bash
#
# commit.sh — the only way to commit on the fork's branch: runs the gate, bumps
# VERSION by the commit type, commits with the version in the subject, regenerates
# CHANGELOG.md into that same commit, and tags it. It never pushes.
#
# Usage: scripts/commit.sh <type> "<message>"     (stage your files first)
#        scripts/commit.sh --gate                   only run the gate (push.sh uses it)
#   fix | security | debug       patch   subject: fix(vX.Y.Z): message
#   docs | chore | test          patch   subject: <type>(vX.Y.Z): message
#   build | perf                 patch
#   feat | refactor              minor
#   major                        major
#
# Tag: vX.Y.Z — the module github.com/rosaldo/ccxt/v5 lives at the repository root,
# so each tag name equals the version (go.dev/ref/mod, "Mapping versions to commits");
# the major is v5, so VERSION must stay at 5.x.y unless the module path changes too.
#
# Exit codes: 0 committed and tagged · 1 gate or git failure · 2 usage
# Emergency only: CCXT_SKIP_GATE=1 scripts/commit.sh ...
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
VERSION_FILE="$REPO_DIR/VERSION"
TAG_PREFIX="v"
# shellcheck source=lib-version.sh
. "$REPO_DIR/scripts/lib-version.sh"

fail() { echo "error: $*" >&2; exit 1; }
usage() {
  echo "Usage: $0 <fix|security|debug|docs|chore|test|build|perf|feat|refactor|major> \"<message>\"  |  $0 --gate" >&2
  exit 2
}

# The gate runs BEFORE the bump: a broken commit never becomes a version, and a
# version never becomes a tag. Every test here runs offline (no exchange calls).
gate() {
  if [[ "${CCXT_SKIP_GATE:-}" == 1 ]]; then
    echo "warning: gate skipped (CCXT_SKIP_GATE=1)" >&2
    return
  fi
  command -v go > /dev/null || fail "go is not installed"

  echo "-> gate: bash -n scripts/*.sh"
  local sh
  for sh in "$REPO_DIR"/scripts/*.sh; do bash -n "$sh" || fail "syntax error in $sh"; done

  echo "-> gate: go build + go vet + go test (module root)"
  (cd "$REPO_DIR" && go build ./... && go vet ./... && go test -count=1 ./...) \
    || fail "go build, go vet or go test failed"

  # upstream's offline base suite (REST + WS helpers); the exchange and request
  # suites need upstream ts/ fixtures or the network and stay out
  echo "-> gate: base tests (tests --baseTests, REST + WS)"
  (cd "$REPO_DIR/tests" && go run . --baseTests > /dev/null && go run . --baseTests --ws > /dev/null) \
    || fail "tests: base tests failed"

  echo "-> gate: go build (cli)"
  (cd "$REPO_DIR/cli" && go build -o /dev/null .) || fail "cli: go build failed"

  echo "gate ok"
}

release() {
  local type="$1" msg="$2" current new subject
  [[ -f "$VERSION_FILE" ]] || fail "VERSION not found at $VERSION_FILE"
  current="$(cat "$VERSION_FILE")"
  new="$(bump_version "$type" "$current")" || fail "VERSION holds '$current', expected X.Y.Z"
  git -C "$REPO_DIR" rev-parse -q --verify "refs/tags/$TAG_PREFIX$new" > /dev/null \
    && fail "tag $TAG_PREFIX$new already exists — VERSION ($current) is behind the tags; set it to the latest tag's version"

  echo "$new" > "$VERSION_FILE"
  subject="$(subject_prefix "$type")(v${new}): ${msg}"
  git -C "$REPO_DIR" add "$VERSION_FILE"
  git -C "$REPO_DIR" commit -q -m "$subject"

  # CHANGELOG goes into the SAME commit, before the tag, so the tag points at a
  # commit that already contains its own entry.
  "$REPO_DIR/scripts/gen-changelog.sh" > /dev/null
  git -C "$REPO_DIR" add CHANGELOG.md
  git -C "$REPO_DIR" commit -q --amend --no-edit

  git -C "$REPO_DIR" tag -a "$TAG_PREFIX$new" HEAD -m "Release $TAG_PREFIX$new"
  echo "Committed: $subject"
  echo "Tagged:    $TAG_PREFIX$new (local; nothing pushed)"
}

if [[ "${1:-}" == --gate && $# -eq 1 ]]; then gate; exit 0; fi
[[ $# -eq 2 && -n "$2" ]] || usage
case "$1" in
  fix | security | debug | docs | chore | test | build | perf | feat | refactor | major) ;;
  *) echo "error: unknown type '$1'" >&2; usage ;;
esac

gate
release "$1" "$2"
