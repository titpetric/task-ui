#!/usr/bin/env bash
#
# test-integration runs the integration suite one package at a time, under a
# memory cap, splitting the packages that hold more tests than the cap fits.
#
# `go test -tags integration ./...` is what this replaces. -p=1 builds and runs
# the packages in order, but a cap binds a process and not a run: one test
# binary holding more than the cap takes the whole suite with it, and the OOM
# names no package. One `go test` per package keeps a failure local and says
# which package it is.
#
# A package over the batch size is split by test name. The names come from
# `go test -list`, not from a list kept here, so a test added to a split
# package runs without this file being edited.
#
#   scripts/test-integration.sh                  # every package
#   scripts/test-integration.sh ./storage/...    # one package or pattern
#   TASK_UI_TEST_BATCH=8 TASK_UI_TEST_MEM=4GiB scripts/test-integration.sh
set -u -o pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

pattern="${1:-./...}"
mem="${TASK_UI_TEST_MEM:-2GiB}"
procs="${TASK_UI_TEST_PROCS:-2}"
batch="${TASK_UI_TEST_BATCH:-20}"

# Packages that hold at least one test under the tag. go list reads the build
# tag, so a package with no tagged file is left out instead of being built for
# nothing.
mapfile -t packages < <(
	go list -tags integration -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' "${pattern}" | sort -u
)

if [ "${#packages[@]}" -eq 0 ]; then
	echo "test-integration: no package under ${pattern} holds a test" >&2
	exit 2
fi

status=0

# one runs a single go test and reports what it covered. An empty filter runs
# the whole package.
one() {
	local package="$1" filter="$2" label="$3"
	printf '\n== %s\n' "${label}"
	local args=(test -tags integration -count=1 -timeout 15m)
	if [ -n "${filter}" ]; then
		args+=(-run "${filter}")
	fi
	args+=("${package}")
	if ! env GOFLAGS=-p=1 "GOMAXPROCS=${procs}" "GOMEMLIMIT=${mem}" go "${args[@]}"; then
		status=1
	fi
}

for package in "${packages[@]}"; do
	mapfile -t names < <(
		env GOFLAGS=-p=1 "GOMEMLIMIT=${mem}" go test -tags integration -list '.*' "${package}" 2>/dev/null \
			| grep -E '^(Test|Fuzz|Example)'
	)
	if [ "${#names[@]}" -le "${batch}" ]; then
		one "${package}" "" "${package} (${#names[@]} tests)"
		continue
	fi

	# RE2 has no negative lookahead, so a batch names the tests it runs:
	# ^(TestOne|TestTwo)$ and nothing else. Every name is in exactly one
	# batch, so nothing is run twice and nothing is skipped.
	total="${#names[@]}"
	for ((i = 0; i < total; i += batch)); do
		filter="$(printf '%s|' "${names[@]:i:batch}")"
		one "${package}" "^(${filter%|})\$" "${package} $((i + 1))-$((i + batch > total ? total : i + batch)) of ${total}"
	done
done

if [ "${status}" -ne 0 ]; then
	echo >&2
	echo "test-integration: a package failed above" >&2
fi
exit "${status}"
