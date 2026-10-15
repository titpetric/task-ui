#!/usr/bin/env bash
# Screenshot a page of the running task-ui front end.
#
#   scripts/chromium.sh <path> [name] [dark|light] [WxH] [budget-ms]
#
#   scripts/chromium.sh /
#   scripts/chromium.sh /p/infra project dark 1440x900
#   scripts/chromium.sh "/p/infra/deploy#history=<id>" history light
#
# Writes /tmp/task-ui-<name>.png and prints the path. Chromium renders the
# light theme by default, so pass dark explicitly when checking the dark
# palette. Always check both. Target another instance with TASK_UI_BASE.
set -euo pipefail

BASE="${TASK_UI_BASE:-http://localhost:3000}"

path="${1:-/}"
name="${2:-shot}"
theme="${3:-dark}"
size="${4:-1440x900}"
budget="${5:-3000}"

case "$theme" in
  dark) scheme=0 ;;
  light) scheme=1 ;;
  *) echo "theme must be dark or light" >&2; exit 2 ;;
esac

out="${TASK_UI_SHOTS:-/tmp}/task-ui-${name}.png"
rm -f "$out"

# Pages poll /api/runs, which keeps the virtual time budget from settling on
# its own; the budget bounds it.
timeout 45 chromium \
  --headless \
  --no-sandbox \
  --disable-gpu \
  --hide-scrollbars \
  --blink-settings=preferredColorScheme=${scheme} \
  --virtual-time-budget="${budget}" \
  --window-size="${size/x/,}" \
  --screenshot="$out" \
  "${BASE}${path}" 2>/dev/null || true

if [ ! -s "$out" ]; then
  echo "no screenshot written for ${BASE}${path}" >&2
  exit 1
fi

echo "$out"
