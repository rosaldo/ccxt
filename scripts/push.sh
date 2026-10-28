#!/usr/bin/env bash
#
# push.sh — publishes the fork: runs the gate, then pushes the branch and the tags.
# Run by the repository owner only. Nothing is built or released: the Go module is
# consumed straight from git via its vX.Y.Z tags.
#
# Usage: scripts/push.sh [--force]
#   --force   force-push the branch and tags (needed once: this history replaced the
#             old fork history on origin, which shares no commit with it)
#
# Order — every check runs before anything leaves this machine:
#   0. checks: on a branch, clean tree, tag v$VERSION exists
#   1. gate (scripts/commit.sh --gate)
#   2. git push of the branch, then of the v* tags
#
# Exit codes: 0 pushed · 1 refused or failed (the message says which step)
set -euo pipefail

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
log() { printf '\n==> %s\n' "$*"; }
ok()  { printf '    ok  %s\n' "$*"; }
die() { printf '\nerror: %s\n' "$*" >&2; exit 1; }
g()   { git -C "$ROOT" "$@"; }

FORCE=()
case "${1:-}" in
  "") ;;
  --force) FORCE=(--force) ;;
  -h | --help) sed -n '3,17p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
  *) die "unknown option '$1' (usage: scripts/push.sh [--force])" ;;
esac

# --- 0. checks ---------------------------------------------------------------
BRANCH="$(g symbolic-ref -q --short HEAD)" || die "HEAD is detached — check out the branch to publish"
[[ -z "$(g status --porcelain)" ]] || die "uncommitted changes — close the version with scripts/commit.sh first"
ok "clean tree on $BRANCH"
VERSION="$(cat "$ROOT/VERSION" 2>/dev/null)" && [[ -n "$VERSION" ]] || die "VERSION missing at $ROOT"
g rev-parse -q --verify "refs/tags/v$VERSION" > /dev/null || die "tag v$VERSION does not exist — commit.sh creates it with the release commit"
ok "tag v$VERSION exists"

# --- 1. gate -------------------------------------------------------------------
log "gate"
"$ROOT/scripts/commit.sh" --gate || die "gate failed — nothing was pushed"

# --- 2. git --------------------------------------------------------------------
log "git push origin $BRANCH ${FORCE[*]}"
g push "${FORCE[@]}" -u origin "$BRANCH"
log "git push v* tags ${FORCE[*]}"
mapfile -t TAGS < <(g tag -l 'v*')
g push "${FORCE[@]}" origin "${TAGS[@]/#/refs/tags/}"
log "pushed $BRANCH and ${#TAGS[@]} tags (v$VERSION)"
