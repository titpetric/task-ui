#!/usr/bin/env bash
#
# test-coverage runs every suite in turn and reports what each one costs and
# what each one covers, then the coverage of all of them together.
#
# The question it answers is which suite is worth its time. A suite that takes
# a minute and moves the total by a point is a minute spent on a point; one
# that takes three seconds and moves it by forty is the suite to keep. Neither
# number means anything on its own, so they are printed side by side.
#
# Every suite writes Go coverage into a directory of its own, in the binary
# format `go tool covdata` reads, and every suite is built with
# -coverpkg=./... so the denominators match and the profiles merge. The
# screenshot suites are not Go tests at all: their server is built with
# `go build -cover` and writes its counters when it exits, which is what
# makes a browser driving a real instance count towards the total.
#
#   scripts/test-coverage.sh                 # every suite
#   scripts/test-coverage.sh unit storage    # the suites named
#   scripts/test-coverage.sh --list          # the names
#
# Serial on purpose, one suite at a time and -p=1 inside each: the durations
# are the point, and two suites sharing four cores measure neither.
#
# It leaves cover/merged.txt, the merged profile in the text format
# `go tool cover` and `uncover` read:
#
#   go tool cover -func=cover/merged.txt | tail -1
#   uncover cover/merged.txt frontend
#
# and cover/logs/<suite>.log, each suite's whole output. A failure prints
# the last thirty lines and the path to the rest, because a suite that
# compares a hundred pages names the one that differed where it reached
# it and not at the end.
set -u -o pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

procs="${TASK_UI_TEST_PROCS:-2}"
mem="${TASK_UI_TEST_MEM:-2GiB}"
out="${TASK_UI_COVER_OUT:-cover}"

# The suites, in the order they run: cheapest first, so a failure in the
# fast gate is reported before a minute of screenshots.
suites=(unit storage php shots inventory)

usage() {
	printf 'suites: %s\n' "${suites[*]}"
}

if [ "${1:-}" = "--list" ]; then
	usage
	exit 0
fi

wanted=("$@")
if [ "${#wanted[@]}" -eq 0 ]; then
	wanted=("${suites[@]}")
fi
for name in "${wanted[@]}"; do
	case " ${suites[*]} " in
	*" ${name} "*) ;;
	*)
		echo "test-coverage: no suite ${name}" >&2
		usage >&2
		exit 2
		;;
	esac
done

mkdir -p "${out}"
# Absolute, because a test binary runs with its own package directory as
# the working directory and would resolve a relative one against that.
data="${root}/${out}/data"
rm -rf "${data}"
mkdir -p "${data}"
# Beside the counters and kept, not a temporary directory: a suite that
# reads a hundred pages names the one that differed somewhere above the
# last thirty lines, and a log removed on exit turns that name into a
# code hunt. The directory is emptied at the start of a run instead, so
# what is in it is this run's.
logs="${root}/${out}/logs"
rm -rf "${logs}"
mkdir -p "${logs}"

status=0
rows=()
covered=()

# run executes one suite, times it, and records its row. A suite's Go
# coverage lands in ${data}/<name>; a suite with none records a dash.
run() {
	local name="$1"
	shift
	local dir="${data}/${name}"
	mkdir -p "${dir}"
	printf '\n== %s\n' "${name}" >&2
	# EPOCHREALTIME rather than date: this machine's date takes %N but
	# not the width that would truncate it, so %s%3N is nanoseconds.
	local start="${EPOCHREALTIME}"
	local code=0
	if ! "$@" >"${logs}/${name}.log" 2>&1; then
		code=1
		status=1
	fi
	local seconds
	seconds="$(awk -v a="${start}" -v b="${EPOCHREALTIME}" \
		'BEGIN { d = b - a; printf (d < 10 ? "%.2fs" : "%.1fs"), d }')"
	if [ "${code}" -ne 0 ]; then
		tail -30 "${logs}/${name}.log" >&2
		printf '\nthe whole log is %s/logs/%s.log\n' "${out}" "${name}" >&2
	fi
	local percent
	percent="$(percent "${dir}")"
	if [ "${percent}" != "-" ]; then
		covered+=("${dir}")
	fi
	local verdict=ok
	[ "${code}" -ne 0 ] && verdict=FAILED
	rows+=("${name}|${seconds}|${percent}|${verdict}")
	printf '%s: %s, %s, %s\n' "${name}" "${seconds}" "${percent}" "${verdict}" >&2
}

# percent is the share of the module's statements a directory of counters
# covers, or a dash where it holds none.
percent() {
	local dir="$1"
	# covdata reads a directory with no counters as an error, which is
	# what a suite that ran no Go code leaves.
	# The last line is "total <tabs> (statements) <tabs> 40.6%", so the
	# share is the last field, by any whitespace.
	go tool covdata func -i="${dir}" 2>/dev/null | awk 'END { if ($NF ~ /%$/) print $NF; else print "-" }'
}

# The unit suite: every package, no build tag, the fast gate.
suite_unit() {
	env GOFLAGS=-p=1 "GOMAXPROCS=${procs}" "GOMEMLIMIT=${mem}" \
		go test -count=1 -cover -coverpkg=./... ./... \
		-args -test.gocoverdir="${data}/unit"
}

# The integration suite: the storage package, which is the whole of it.
# docs/testing.md says why nothing else has one.
suite_storage() {
	env GOFLAGS=-p=1 "GOMAXPROCS=${procs}" "GOMEMLIMIT=${mem}" \
		go test -count=1 -tags integration -cover -coverpkg=./... ./storage/ \
		-args -test.gocoverdir="${data}/storage"
}

# The PHP fixtures, which phpscript runs and Go coverage cannot see: the row
# carries its duration and no percent. Two trees hold them, the module's and
# the front end being ported, and each roots its fixtures at its own
# phpscript.yml, so each runs from its own directory.
suite_php() {
	command -v phpscript >/dev/null || {
		echo "phpscript not in PATH"
		return 1
	}
	(cd phpscripts/modules/titpetric/refinement && phpscript test ./tests/...) || return 1
	(cd phpfront/app && phpscript test ./tests/...)
}

# The README's screenshots: four instances over one database, real runs,
# every link crawled, and chromium over every view in both themes. The
# server is instrumented, so this is the only suite that measures the
# pages as a browser reaches them.
suite_shots() {
	command -v chromium >/dev/null || {
		echo "chromium not in PATH"
		return 1
	}
	# inventory-shots runs at the end of demo.sh and keeps its own
	# counters, so the two suites are reported apart.
	env TASK_UI_COVER_DIR="${data}/shots" TASK_UI_SHOT_SKIP_INVENTORY=1 \
		scripts/demo.sh
}

# The Inventory page's shots, from a fake hub and hosts of their own.
suite_inventory() {
	command -v chromium >/dev/null || {
		echo "chromium not in PATH"
		return 1
	}
	env TASK_UI_COVER_DIR="${data}/inventory" scripts/inventory-shots.sh
}

for name in "${wanted[@]}"; do
	run "${name}" "suite_${name}"
done

# The total is every suite's counters read together, which is not the sum
# of the columns: two suites covering the same line cover one line.
merged="-"
if [ "${#covered[@]}" -gt 0 ]; then
	joined="$(
		IFS=,
		echo "${covered[*]}"
	)"
	go tool covdata textfmt -i="${joined}" -o "${out}/merged.txt" 2>/dev/null
	merged="$(percent "${joined}")"
fi

printf '\n| suite | duration | coverage |\n|-------|----------|----------|\n'
for row in "${rows[@]}"; do
	IFS='|' read -r name seconds pct verdict <<<"${row}"
	[ "${verdict}" = FAILED ] && name="${name} (FAILED)"
	printf '| %s | %s | %s |\n' "${name}" "${seconds}" "${pct}"
done
printf '| **all merged** | | **%s** |\n' "${merged}"

printf '\nthe merged profile is %s/merged.txt\n' "${out}"
if [ "${status}" -ne 0 ]; then
	echo
	echo "test-coverage: a suite failed above" >&2
fi
exit "${status}"
