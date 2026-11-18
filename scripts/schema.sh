#!/usr/bin/env bash
# Regenerate the data model from the migrations.
#
#   scripts/schema.sh
#
# Applies every migration to a throwaway SQLite file, lints the schema, then
# writes schema/docs, schema/schema.yml, the Go types under app/model/ and the
# front end's models.
#
# The migrations live with the code that owns each table - a module's pages, the
# key store, the Guard, the module registry - so --recursive reads the path as a
# directory name to find rather than one place, and merges every schema/ of the
# tree flat, keyed by base name. That is the same merge app.Migrations hands the
# binary, and the base names are the order a migration reading another one's
# table needs.
#
# Needs mig >= v0.6.6: go install github.com/go-bridget/mig@latest
set -euo pipefail
cd "$(dirname "$0")/.."

db="$(mktemp -u).db"
trap 'rm -f "$db"' EXIT
export MIG_DB_DSN="sqlite://file:${db}"

mig migrate --project task-ui --path schema --recursive --exclude data --apply
mig lint
mig docs --output=./schema/docs
mig docs --output=./schema --yaml --output-file schema.yml
mig gen --go.fill-json --output=./app/model

# The front end's models read the same schema.yml, so one run generates both sides.
phpscript run phpfront/app/tools/models.php
bash phpfront/app/tools/fmt.sh --write
