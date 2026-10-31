#!/usr/bin/env bash
#
# test runs the packages a change touched, and holds every test to a
# budget.
#
# Two ideas, both borrowed from how a linter is run here. Scope: a change
# to one package is a reason to test that package and the packages that
# import it, not the ninety that cannot have noticed. Budget: a test over
# a second is a bug in the test, and the run says which one rather than
# leaving it to be noticed a year later.
#
#   scripts/test.sh                  # the packages changed against main
#   scripts/test.sh --all            # every package
#   scripts/test.sh --base HEAD~3    # changed against another ref
#   scripts/test.sh ./runner/...     # a pattern, scoped by hand
#
# Flags:
#   --all              every package, not the changed ones
#   --base REF         what "changed" is measured against, main by default
#   --budget DURATION  the longest a test may take, 1s by default
#   --tag TAG          a build tag, repeatable: --tag integration
#
# The budget is reported by `gotestsum tool slowest` over the run's own
# json, so it measures what this machine did and not what a table says.
set -u -o pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

base="${TASK_UI_TEST_BASE:-main}"
budget="${TASK_UI_TEST_BUDGET:-1s}"
mem="${TASK_UI_TEST_MEM:-2GiB}"
procs="${TASK_UI_TEST_PROCS:-2}"
tags=""
all=0
pattern=""

while [ $# -gt 0 ]; do
	case "$1" in
	--all) all=1 ;;
	--base) base="$2"; shift ;;
	--budget) budget="$2"; shift ;;
	--tag) tags="${tags:+${tags},}$2"; shift ;;
	-*) echo "test: no flag $1" >&2; exit 2 ;;
	*) pattern="$1" ;;
	esac
	shift
done

flags=()
if [ -n "${tags}" ]; then
	flags+=(-tags "${tags}")
fi

# changed names the packages holding a file this branch touched. A file
# that is no .go still counts: a template, a schema or an embedded asset
# is read by the package it sits in.
changed() {
	git diff --name-only "${base}"...HEAD 2>/dev/null
	git diff --name-only
	git ls-files --others --exclude-standard
}

# dependents names every package that imports one of the given ones, so a
# change to storage tests the twenty packages built on it. go list reads
# the real import graph, which is why this is not a guess.
dependents() {
	local of="$1"
	go list "${flags[@]+"${flags[@]}"}" -f '{{.ImportPath}}{{range .Deps}} {{.}}{{end}}{{range .TestImports}} {{.}}{{end}}{{range .XTestImports}} {{.}}{{end}}' ./... 2>/dev/null |
		awk -v want="${of}" '
			BEGIN { split(want, list, " "); for (i in list) if (list[i] != "") need[list[i]] = 1 }
			{ for (i = 2; i <= NF; i++) if ($i in need) { print $1; next } }
		'
}

if [ "${all}" -eq 1 ] || [ -n "${pattern}" ]; then
	mapfile -t packages < <(go list "${flags[@]+"${flags[@]}"}" "${pattern:-./...}" | sort -u)
else
	mapfile -t dirs < <(changed | grep -E '\.(go|tpl|sql|js|css|md|php)$' | xargs -r -n1 dirname | sort -u)
	if [ "${#dirs[@]}" -eq 0 ]; then
		echo "test: nothing changed against ${base}"
		exit 0
	fi
	mapfile -t direct < <(
		for dir in "${dirs[@]}"; do
			go list "${flags[@]+"${flags[@]}"}" "./${dir}" 2>/dev/null
		done | sort -u
	)
	if [ "${#direct[@]}" -eq 0 ]; then
		echo "test: nothing changed against ${base} that is in a package"
		exit 0
	fi
	mapfile -t packages < <({ printf '%s\n' "${direct[@]}"; dependents "${direct[*]}"; } | sort -u)
	printf 'test: %d package(s) changed, %d with what depends on them\n' "${#direct[@]}" "${#packages[@]}"
fi

if [ "${#packages[@]}" -eq 0 ]; then
	echo "test: no package to run"
	exit 0
fi

json="$(mktemp)"
trap 'rm -f "${json}"' EXIT

env "GOMAXPROCS=${procs}" "GOMEMLIMIT=${mem}" \
	gotestsum --jsonfile "${json}" --format pkgname -- \
	-count=1 -timeout 10m "${flags[@]+"${flags[@]}"}" "${packages[@]}"
status=$?

echo
slow="$(gotestsum tool slowest --jsonfile "${json}" --threshold "${budget}" 2>/dev/null)"
if [ -n "${slow}" ]; then
	echo "over the ${budget} budget:" >&2
	echo "${slow}" >&2
	echo >&2
	echo "A test over ${budget} is waiting on something it could be told about." >&2
	echo "docs/testing.md says what to do instead of sleeping." >&2
	status=1
else
	echo "every test is inside the ${budget} budget"
fi

exit "${status}"
