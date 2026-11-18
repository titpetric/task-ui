#!/usr/bin/env bash
# Writes the tree canonically and, without --write, fails when a file changes.
set -euo pipefail
cd "$(dirname "$0")/.."

check=1
if [ "${1:-}" = "--write" ]; then
    check=""
fi

files="$(git ls-files -co --exclude-standard | grep -E '\.(php|phpt)$' | grep -v '^ext/' | grep -v 'templates_c/')"
before="$(mktemp)"
after="$(mktemp)"
trap 'rm -f "${before}" "${after}"' EXIT

echo "${files}" | xargs sha256sum > "${before}"
phpscript fmt ./code/... ./modules/... ./stubs/... ./tools/... > /dev/null

prev="$(echo "${files}" | xargs sha256sum | sha256sum)"
for _ in $(seq 1 10); do
    echo "${files}" | xargs semgrep scan --metrics=off --quiet --config semgrep/array-syntax.yml --autofix > /dev/null 2>&1
    now="$(echo "${files}" | xargs sha256sum | sha256sum)"
    [ "${now}" = "${prev}" ] && break
    prev="${now}"
done

if [ -z "${check}" ]; then
    exit 0
fi

echo "${files}" | xargs sha256sum > "${after}"
if ! diff -q "${before}" "${after}" > /dev/null; then
    echo "formatting would rewrite:" >&2
    diff "${before}" "${after}" | grep '^[<>]' | awk '{print $3}' | sort -u >&2
    exit 1
fi
echo "every file is formatted"
