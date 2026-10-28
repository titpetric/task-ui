#!/usr/bin/env bash
# Screenshot a page of the running task-ui front end.
#
#   scripts/chromium.sh <path> [name] [dark|light] [WxH] [budget-ms]
#
#   scripts/chromium.sh /
#   scripts/chromium.sh /p/infra project dark 1440x900
#   scripts/chromium.sh "/f/internal/infra/deploy#history=<id>" history light
#
# Writes ${TASK_UI_SHOTS:-/tmp}/task-ui-<name>.png and prints the path.
# Chromium renders the light theme by default, so pass dark explicitly when
# checking the dark palette. Always check both. Target another instance with
# TASK_UI_BASE.
#
# Every browser runs in a transient systemd scope capped at
# TASK_UI_CHROMIUM_MEM (1G) with swap off, so an over-budget shot dies alone
# and never takes the machine down with it. Without systemd the browser runs
# with its V8 heap capped and nothing else: TASK_UI_CHROMIUM_MEM has no
# effect there.
#
# At most TASK_UI_CHROMIUM_PARALLEL (2) browsers run at once, across every
# caller on the machine. A third waits for a slot, TASK_UI_CHROMIUM_WAIT
# seconds at most.
#
# Environment:
#   TASK_UI_BASE                 base URL, default http://localhost:3000
#   TASK_UI_SHOTS                output directory, default /tmp
#   TASK_UI_CHROMIUM_MEM         cgroup memory cap, default 1G (systemd only)
#   TASK_UI_CHROMIUM_PARALLEL    concurrent browsers, default 2
#   TASK_UI_CHROMIUM_WAIT        seconds to wait for a slot, default 600
#   TASK_UI_CHROMIUM_TIMEOUT     wall clock seconds per browser, default 45
#   TASK_UI_CHROMIUM_KILL_AFTER  seconds between TERM and KILL, default 5
#   TASK_UI_CHROMIUM_LOCK_DIR    slot directory, default
#                                ${XDG_RUNTIME_DIR:-$HOME/.cache}/task-ui-chromium
#   TASK_UI_CHROMIUM_PROFILE_DIR profile prefix, default
#                                ${XDG_CACHE_HOME:-$HOME/.cache}/task-ui-chromium
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

command -v chromium >/dev/null 2>&1 || { echo "chromium not found in PATH" >&2; exit 1; }

mem="${TASK_UI_CHROMIUM_MEM:-1G}"
parallel="${TASK_UI_CHROMIUM_PARALLEL:-2}"
wait_s="${TASK_UI_CHROMIUM_WAIT:-600}"
wall="${TASK_UI_CHROMIUM_TIMEOUT:-45}"
kill_after="${TASK_UI_CHROMIUM_KILL_AFTER:-5}"
cache="${XDG_CACHE_HOME:-${HOME:-/tmp}/.cache}"
lock_dir="${TASK_UI_CHROMIUM_LOCK_DIR:-${XDG_RUNTIME_DIR:-$cache}/task-ui-chromium}"
profile_base="${TASK_UI_CHROMIUM_PROFILE_DIR:-$cache/task-ui-chromium}"
case "$parallel" in ''|*[!0-9]*|0) parallel=2 ;; esac

out="${TASK_UI_SHOTS:-/tmp}/task-ui-${name}.png"
mkdir -p -- "$(dirname -- "$out")"
rm -f -- "$out"

# The lock fd and the profile are cleaned on any exit. Both are global on
# purpose: the EXIT trap runs after function scope is gone. The lock fd is
# not close-on-exec, so the browser inherits it and the slot stays held for
# as long as a browser lives.
lock_fd=""
profile=""
release() {
  if [ -n "$lock_fd" ]; then
    exec {lock_fd}>&- 2>/dev/null || true
    lock_fd=""
  fi
  if [ -n "$profile" ]; then
    rm -rf -- "$profile" 2>/dev/null || true
    profile=""
  fi
}
trap release EXIT

# One slot per browser: take the first free one, wait on slot 0 when every
# slot is busy. A fresh open per attempt, because bash re-locking its own
# descriptor always succeeds. Never unlink the slot files: a new inode is a
# different lock.
if command -v flock >/dev/null 2>&1 && mkdir -p -m 700 -- "$lock_dir" 2>/dev/null && [ -w "$lock_dir" ]; then
  deadline=$(( $(date +%s) + wait_s ))
  while :; do
    for ((i = 0; i < parallel; i++)); do
      exec {lock_fd}>"${lock_dir}/slot.${i}" 2>/dev/null || { lock_fd=""; continue; }
      if flock -n "$lock_fd"; then
        break 2
      fi
      exec {lock_fd}>&- 2>/dev/null || true
      lock_fd=""
    done
    if [ "$(date +%s)" -ge "$deadline" ]; then
      echo "no chromium slot free after ${wait_s}s (${parallel} allowed)" >&2
      exit 1
    fi
    sleep 0.2
  done
fi

# A fresh profile per shot: no state from the previous page, no lock on the
# user's own browser profile. On disk, not in /tmp, which is tmpfs. Swept
# when an earlier run was killed before its trap ran.
mkdir -p -- "$(dirname -- "$profile_base")"
find "$(dirname -- "$profile_base")" -maxdepth 1 -name "$(basename -- "$profile_base").*" -mmin +60 -exec rm -rf -- {} + 2>/dev/null || true
profile="$(mktemp -d "${profile_base}.XXXXXX")"

# The scope holds timeout and the browser only. MemoryOOMGroup kills the
# whole scope rather than one renderer, which would leave the browser
# limping to the wall clock; --collect keeps failed scopes out of
# systemctl --failed. Degrade to MemoryMax alone, then to no scope, when the
# kernel or the bus will not take the properties.
bus=()
scope=()
if [ -d /run/systemd/system ] && command -v systemd-run >/dev/null 2>&1; then
  if [ "$(id -u)" -ne 0 ]; then
    bus=(--user)
    if [ -z "${XDG_RUNTIME_DIR:-}" ] && [ -d "/run/user/$(id -u)" ]; then
      export XDG_RUNTIME_DIR="/run/user/$(id -u)"
    fi
  fi
  if systemd-run "${bus[@]}" --quiet --scope --collect \
      -p "MemoryMax=${mem}" -p MemorySwapMax=0 -p MemoryOOMGroup=yes -- true 2>/dev/null; then
    scope=(-p "MemoryMax=${mem}" -p MemorySwapMax=0 -p MemoryOOMGroup=yes)
  elif systemd-run "${bus[@]}" --quiet --scope --collect \
      -p "MemoryMax=${mem}" -- true 2>/dev/null; then
    scope=(-p "MemoryMax=${mem}")
  fi
fi

# A page that follows a topic keeps an event stream open, and virtual time
# never runs out while a request is open: the shot is taken on a timeout
# instead, the budget in real time. timeout signals the browser's process
# group, so renderers and zygotes go with it.
cmd=(
  timeout --kill-after="$kill_after" "$wall" chromium
  --headless
  --no-sandbox
  --disable-gpu
  --hide-scrollbars
  --disable-dev-shm-usage
  --blink-settings=preferredColorScheme=${scheme}
  --timeout="${budget}"
  --window-size="${size/x/,}"
  --renderer-process-limit=1
  --js-flags=--max-old-space-size=512
  --user-data-dir="$profile"
  --no-first-run
  --no-default-browser-check
  --disable-background-networking
  --disable-extensions
  --disable-sync
  --disable-breakpad
  --mute-audio
  --screenshot="$out"
  "${BASE}${path}"
)

if [ "${#scope[@]}" -gt 0 ]; then
  systemd-run "${bus[@]}" --quiet --scope --collect "${scope[@]}" -- "${cmd[@]}" 2>/dev/null || true
else
  "${cmd[@]}" 2>/dev/null || true
fi

if [ ! -s "$out" ]; then
  echo "no screenshot written for ${BASE}${path}" >&2
  exit 1
fi

echo "$out"
