#!/usr/bin/env bash
# Regenerate the data model from the migrations.
#
#   scripts/schema.sh
#
# Applies schema/*.up.sql to a throwaway SQLite file, lints the schema,
# then writes schema/docs, schema/schema.yml and the Go types under model/.
# Needs mig: go install github.com/go-bridget/mig@latest
set -euo pipefail
cd "$(dirname "$0")/.."

db="$(mktemp -u).db"
trap 'rm -f "$db"' EXIT
export MIG_DB_DSN="sqlite://file:${db}"

mig migrate --project task-ui --path ./schema --apply
mig lint
mig docs --output=./schema/docs
mig docs --output=./schema --yaml --output-file schema.yml
mig gen --go.fill-json --output=./model
