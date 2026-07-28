#!/usr/bin/env bash
# worktree.sh — create/list/remove git worktrees for the mandatory
# worktree + PR workflow (see AGENTS.md / CONTRIBUTING.md).
#
# All development happens on a branch in an isolated worktree; `main` is never
# committed to directly. Worktrees live under ../.worktrees/<branch> so they sit
# beside the repo, not inside it.
#
# Usage:
#   scripts/worktree.sh new  <type>/<scope>/<name>   # e.g. feat/studio/inputs-batch
#   scripts/worktree.sh new  sprint/<id>-<scope>     # e.g. sprint/S2-workspace-crud
#   scripts/worktree.sh list
#   scripts/worktree.sh rm   <branch>
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
WT_DIR="$(cd "${REPO_ROOT}/.." && pwd)/.worktrees"
DEFAULT_BASE="main"

usage() { sed -n '2,16p' "$0"; exit "${1:-0}"; }

valid_branch() {
  case "$1" in
    feat/*/*|fix/*/*|chore/*/*|sprint/*) return 0 ;;
    *) return 1 ;;
  esac
}

cmd_new() {
  local branch="${1:-}"
  [ -n "$branch" ] || { echo "error: branch name required" >&2; usage 1; }
  if ! valid_branch "$branch"; then
    echo "error: branch must match feat/<scope>/<name>, fix/<scope>/<name>," >&2
    echo "       chore/<scope>/<name>, or sprint/<id>-<scope>" >&2
    exit 1
  fi
  local path="${WT_DIR}/${branch}"
  git -C "$REPO_ROOT" fetch --quiet origin "$DEFAULT_BASE" || true
  mkdir -p "$(dirname "$path")"
  git -C "$REPO_ROOT" worktree add -b "$branch" "$path" "origin/${DEFAULT_BASE}" 2>/dev/null \
    || git -C "$REPO_ROOT" worktree add -b "$branch" "$path" "$DEFAULT_BASE"
  echo "worktree ready: $path"
  echo "  cd '$path' && task up && task shell"
}

cmd_list() { git -C "$REPO_ROOT" worktree list; }

cmd_rm() {
  local branch="${1:-}"
  [ -n "$branch" ] || { echo "error: branch required" >&2; exit 1; }
  git -C "$REPO_ROOT" worktree remove "${WT_DIR}/${branch}"
  echo "removed worktree ${WT_DIR}/${branch} (branch '$branch' kept)"
}

case "${1:-}" in
  new)  shift; cmd_new "$@" ;;
  list) cmd_list ;;
  rm)   shift; cmd_rm "$@" ;;
  -h|--help|"") usage 0 ;;
  *) echo "unknown command: $1" >&2; usage 1 ;;
esac
