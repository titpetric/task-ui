#!/usr/bin/env bash
# Start a job on the running task-ui through its API, wait for it to end
# and print the last frame of its terminal.
#
#   scripts/run.sh <folder> <job> [force [reason]]
#
# Works against an instance signed in with TASK_UI_AUTOLOGIN, since curl
# has no session. Target another instance with TASK_UI_BASE. The run ID
# and the final state go to stderr.
set -euo pipefail

BASE="${TASK_UI_BASE:-http://localhost:3000}"

folder="${1:?usage: run.sh <folder> <job> [force]}"
job="${2:?usage: run.sh <folder> <job> [force]}"
force=false
if [ "${3:-}" = "force" ]; then
  force=true
fi
reason="${4:-}"

# The API refuses other origins on writes, so send ours.
api() {
  curl -sS -H "Origin: ${BASE}" -H "Content-Type: application/json" "$@"
}

field() {
  grep -o "\"$1\":\"[^\"]*\"" | head -1 | cut -d'"' -f4
}

started="$(api -X POST -d "{\"folder\":\"${folder}\",\"job\":\"${job}\",\"force\":${force},\"reason\":\"${reason}\",\"cols\":120,\"rows\":30}" "${BASE}/api/runs")"
id="$(printf '%s' "$started" | field id)"
if [ -z "$id" ]; then
  echo "$started" >&2
  exit 1
fi

state=running
for _ in $(seq 1 1200); do
  state="$(api "${BASE}/api/runs/${id}" | field state)"
  if [ "$state" != "running" ]; then
    break
  fi
  sleep 0.2
done

api "${BASE}/api/runs/${id}/frame.txt"
echo
echo "${id} ${state}" >&2
