#!/usr/bin/env bash
#
# test-duplicate reports what two tests both cover.
#
# A coverage number says how much of a package is reached. It does not say
# that four tests reach the same four hundred lines, which is four times
# the wall clock for one test's worth of evidence. This runs each test on
# its own with a profile of its own and compares the sets.
#
#   scripts/test-duplicate.sh                 # every package
#   scripts/test-duplicate.sh ./runner/...    # one
#   scripts/test-duplicate.sh --overlap 90    # only pairs over 90%
#
# A pair over the threshold is a candidate, not a verdict: two tests may
# cover one function and assert different things about it. Read them
# before deleting one. What the report is for is finding the pairs worth
# reading.
set -u -o pipefail

# sort and comm have to agree on what order is, and only the C locale
# makes them: a profile line holds slashes, colons and dots, which a
# locale collates its own way and comm then calls unsorted.
export LC_ALL=C

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

overlap="${TASK_UI_TEST_OVERLAP:-80}"
pattern="./..."

while [ $# -gt 0 ]; do
	case "$1" in
	--overlap) overlap="$2"; shift ;;
	-*) echo "test-duplicate: no flag $1" >&2; exit 2 ;;
	*) pattern="$1" ;;
	esac
	shift
done

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

mapfile -t packages < <(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' "${pattern}" | sort -u)

for package in "${packages[@]}"; do
	mapfile -t names < <(go test -list '^Test' "${package}" 2>/dev/null | grep -E '^Test')
	[ "${#names[@]}" -le 1 ] && continue

	dir="${work}/$(echo "${package}" | tr / _)"
	mkdir -p "${dir}"

	# One profile a test. -coverpkg is the package itself: what a test of
	# another package happens to reach is not what this is measuring.
	for name in "${names[@]}"; do
		go test -count=1 -run "^${name}\$" -coverprofile="${dir}/${name}.cov" \
			-coverpkg "${package}" "${package}" >/dev/null 2>&1
	done

	# A profile line is "file:from.col,to.col statements count". The lines
	# a test reached are the ones whose count is not zero.
	for file in "${dir}"/*.cov; do
		[ -e "${file}" ] || continue
		awk 'NR > 1 && $NF != "0" { print $1 }' "${file}" | sort -u > "${file%.cov}.lines"
	done

	printf '\n== %s\n' "${package}"
	for a in "${dir}"/*.lines; do
		[ -e "${a}" ] || continue
		for b in "${dir}"/*.lines; do
			[ "${a}" \< "${b}" ] || continue
			both="$(comm -12 "${a}" "${b}" | wc -l)"
			mine="$(wc -l < "${a}")"
			theirs="$(wc -l < "${b}")"
			[ "${mine}" -eq 0 ] || [ "${theirs}" -eq 0 ] && continue
			smaller="${mine}"
			[ "${theirs}" -lt "${smaller}" ] && smaller="${theirs}"
			share=$(( both * 100 / smaller ))
			if [ "${share}" -ge "${overlap}" ]; then
				printf '%3d%%  %s (%d) and %s (%d) share %d lines\n' \
					"${share}" "$(basename "${a}" .lines)" "${mine}" \
					"$(basename "${b}" .lines)" "${theirs}" "${both}"
			fi
		done
	done
done
