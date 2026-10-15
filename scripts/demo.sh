#!/usr/bin/env bash
# Take the screenshots the README shows.
#
#   scripts/demo.sh
#
# Builds task-ui, serves testdata/demo from throwaway instances, records a
# few real runs over the websocket with scripts/run.sh, and captures every
# view with scripts/chromium.sh in both themes. Writes .github/assets/*.png.
#
# Needs chromium, curl, atkins and task in PATH. imagemagick is optional and
# shrinks the files.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

for tool in chromium curl atkins task; do
  command -v "$tool" >/dev/null || { echo "$tool not found in PATH" >&2; exit 1; }
done

work="$(mktemp -d)"
pids=()
cleanup() {
  for pid in "${pids[@]}"; do kill "$pid" 2>/dev/null || true; done
  rm -rf "$work"
}
trap cleanup EXIT

go build -o "$work/task-ui" .

# serve starts an instance on a port, sharing the history and keys, and
# waits for it to answer. Extra arguments are KEY=value environment.
serve() {
  local port="$1"
  shift
  env TASK_UI_ASSETS="$root" \
    TASK_UI_ROOT="$root/testdata/demo" \
    TASK_UI_LISTEN="127.0.0.1:${port}" \
    TASK_UI_HISTORY_PATH="$work/history" \
    TASK_UI_SSH_DIR="$work/ssh" \
    TASK_UI_TITLE="demo" \
    "$@" "$work/task-ui" >"$work/task-ui-${port}.log" 2>&1 &
  pids+=("$!")
  for _ in $(seq 1 50); do
    curl -fs -o /dev/null "http://127.0.0.1:${port}/healthz" && return 0
    sleep 0.1
  done
  echo "task-ui on ${port} did not start" >&2
  cat "$work/task-ui-${port}.log" >&2
  exit 1
}

# An open instance: no users, visitors are admins.
admin=3101
serve "$admin" TASK_UI_ANONYMOUS=admin
export TASK_UI_BASE="http://127.0.0.1:${admin}"

# Real runs, so the history has something to show: a forced run that kills
# the one before it, a plain one, and runs in the other projects.
scripts/run.sh infra deploy >/dev/null &
first=$!
sleep 1
scripts/run.sh infra deploy force >/dev/null
scripts/run.sh infra plan >/dev/null
scripts/run.sh website build >/dev/null
scripts/run.sh backups verify >/dev/null
scripts/run.sh infra deploy >/dev/null
wait "$first"

curl -fsS -o /dev/null -H "Origin: ${TASK_UI_BASE}" \
  -d type=ed25519 -d comment=task-ui "${TASK_UI_BASE}/admin/keys/generate"

# The newest deploy run, for the dialog shots.
run="$(basename "$(ls "$work/history/infra/deploy/"*.json | sort | tail -1)" .json)"

export TASK_UI_SHOTS="$work"
shots=()
shot() {
  local name="$1"
  shift
  shots+=("$(scripts/chromium.sh "$1" "$name" "${@:2}")")
}

for theme in dark light; do
  shot "overview-${theme}" / "$theme"
  shot "project-${theme}" /p/infra "$theme"
  shot "runbook-${theme}" /p/infra/deploy "$theme"
  shot "run-history-${theme}" "/p/infra/deploy#history=${run}" "$theme"
  # Part way through the replay, so the dialog shows a run in progress.
  shot "run-replay-${theme}" "/p/infra/deploy#replay=${run}" "$theme" 1440x900 1800
  shot "keys-${theme}" /admin/keys "$theme" 1440x1000
done
shot mobile-project-dark /p/infra dark 390x844

# A read only instance over the same history.
viewer=3102
serve "$viewer" TASK_UI_ANONYMOUS=viewer
TASK_UI_BASE="http://127.0.0.1:${viewer}" shot viewer-runbook-dark /p/infra/deploy dark

# An instance with users, for the sign in screen.
login=3103
cat >"$work/task-ui.yml" <<YAML
users:
  - name: admin
    password: "$("$work/task-ui" hash demo)"
    role: admin
YAML
serve "$login" TASK_UI_CONFIG="$work/task-ui.yml"
for theme in dark light; do
  TASK_UI_BASE="http://127.0.0.1:${login}" shot "login-${theme}" /login "$theme"
done

mkdir -p .github/assets
for file in "${shots[@]}"; do
  name="$(basename "$file" .png)"
  out=".github/assets/${name#task-ui-}.png"
  if command -v magick >/dev/null 2>&1; then
    magick "$file" -strip -colors 256 -define png:compression-level=9 "$out"
  else
    cp "$file" "$out"
  fi
  echo "$out"
done
cp .github/assets/overview-dark.png .github/assets/task-ui.png
