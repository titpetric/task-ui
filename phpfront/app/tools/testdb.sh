#!/usr/bin/env bash
# Build the fixture database from every migration, fresh on each run.
set -euo pipefail
root="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
cd "$root"
db="${root}/phpfront/app/tests/task-ui.db"
rm -f "${db}" "${db}-shm" "${db}-wal"
export MIG_DB_DSN="sqlite://file:${db}"
mig migrate --project task-ui --path schema --recursive --exclude data --apply
