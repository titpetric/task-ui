#!/usr/bin/env bash
# Run a job on the running task-ui over its websocket, the same way the
# run dialog does, and print the terminal output.
#
#   scripts/run.sh <project> <job> [force]
#
# Works against an instance without users, where visitors get the
# TASK_UI_ANONYMOUS role. Target another instance with TASK_UI_BASE.
set -euo pipefail

BASE="${TASK_UI_BASE:-http://localhost:3000}"

project="${1:?usage: run.sh <project> <job> [force]}"
job="${2:?usage: run.sh <project> <job> [force]}"
query="cols=120&rows=30"
if [ "${3:-}" = "force" ]; then
  query="${query}&force=1"
fi

# The socket refuses other origins, so send ours. curl prints the frames as
# they arrive and exits when the server closes the socket at the end of a run.
curl -sN --max-time 120 \
  -H "Origin: ${BASE}" \
  "${BASE/http/ws}/ws/run/${project}/${job}?${query}"
echo
