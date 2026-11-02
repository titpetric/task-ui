#!/usr/bin/env bash
# Compare a page of the Go front end against the same page of the one being
# ported, and report what differs.
#
#   scripts/compare.sh /notes
#   scripts/compare.sh /runs /runs/01hz /schedules
#
# Several paths report one line each and the diffs below them, and the exit
# code is the count that differ, so a round's verification gates on it.
#
# Fetches <base><path> and <base>/development<path> from the running instance,
# normalises both and diffs them. Exits 1 when they differ, so a round's
# verification gates on it.
#
# Normalising is what makes the diff readable, and it is deliberately narrow:
# the mount prefix comes off every link, runs of whitespace between tags
# collapse, the asset version is dropped, and an identifier that changes per
# run is masked. Everything else is a real difference.
#
# Works against an instance signed in with TASK_UI_AUTOLOGIN, since curl has
# no session. Target another with TASK_UI_BASE.
set -euo pipefail

BASE="${TASK_UI_BASE:-http://localhost:3000}"
PREFIX="${TASK_UI_PHP_BASE:-/development}"

if [ $# -eq 0 ]; then
  echo "usage: compare.sh <path> [path...]" >&2
  exit 2
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# fetch writes a page and reports its status.
fetch() {
  curl -sS -o "$2" -w '%{http_code}' "${BASE}$1"
}

# normalise reads a page and writes it in the shape the diff compares: the
# prefix off the links, one tag per line, the per-run values masked.
normalise() {
  sed -e "s#\"${PREFIX}/#\"/#g" -e "s#\"${PREFIX}\"#\"/\"#g" "$1" |
    sed -e 's/?v=[0-9a-f]\{12\}//g' |
    sed -e 's/\b[0-9a-z]\{26\}\b/ID/g' |
    sed -e 's/>/>\n/g' |
    sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' |
    grep -v '^$'
}

# compare reports one path and returns whether it differs.
compare() {
  local path="$1"
  local go_status php_status lines

  go_status="$(fetch "$path" "$work/go.html")"
  php_status="$(fetch "${PREFIX}${path}" "$work/php.html")"

  if [ "$go_status" != "$php_status" ]; then
    echo "${path}: go answered ${go_status}, php answered ${php_status}" >&2
    return 1
  fi

  normalise "$work/go.html" >"$work/go.txt"
  normalise "$work/php.html" >"$work/php.txt"

  if diff -u "$work/go.txt" "$work/php.txt" >"$work/diff"; then
    echo "${path}: the same, ${go_status}"
    return 0
  fi

  lines="$(grep -c '^[-+]' "$work/diff" || true)"
  echo "${path}: ${lines} lines differ, ${go_status}" >&2
  if command -v colordiff >/dev/null; then
    colordiff <"$work/diff" >&2
  else
    cat "$work/diff" >&2
  fi
  return 1
}

differ=0
for path in "$@"; do
  compare "$path" || differ=$((differ + 1))
done

if [ "$differ" -gt 0 ]; then
  echo "${differ} of $# differ" >&2
fi
exit "$differ"
