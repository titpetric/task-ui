#!/usr/bin/env bash
#
# test-memory runs one test function per process and reports what each one
# cost at its worst.
#
# A package of tests that grows until the kernel stops it says nothing about
# which test grew it: the OOM names the binary. The two questions are separate
# and this answers the first of them. Is one test heavy, or is every test
# heavy together? One process per test answers it: a test that is heavy alone
# shows here, and a suite whose tests are each small while the package is
# large is a leak between them, which `TASK_UI_TEST_MEM_TRACE=1` and a heap
# profile at the kill then name.
#
# The peak is the resident size of the whole process, read by `time -f %M`,
# so the test binary's own 40MB of text and data is in every number. What
# matters is the spread, not the floor.
#
#   scripts/test-memory.sh                   # every package with a tagged test
#   scripts/test-memory.sh ./frontend        # one package or pattern
#   TASK_UI_TEST_MEM_MAX=1GiB scripts/test-memory.sh ./app
#
# A test over TASK_UI_TEST_MEM_MAX fails the run and is named. The default is
# generous on purpose: this gates the outlier, not the budget.
set -u -o pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

pattern="${1:-./...}"
max="${TASK_UI_TEST_MEM_MAX:-768MiB}"
procs="${TASK_UI_TEST_PROCS:-2}"
kill_at="${TASK_UI_TEST_MEM_KILL:-4GiB}"

# ceiling is max in kilobytes, which is what time -f %M prints.
ceiling="$(
	awk -v s="${max}" 'BEGIN {
		n = s + 0
		if (s ~ /[Gg]i?B?$/) n *= 1024 * 1024
		else if (s ~ /[Mm]i?B?$/) n *= 1024
		else if (s ~ /[Kk]i?B?$/) n *= 1
		else n /= 1024
		printf "%d", n
	}'
)"

mapfile -t packages < <(
	go list -tags integration -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' "${pattern}" | sort -u
)
if [ "${#packages[@]}" -eq 0 ]; then
	echo "test-memory: no package under ${pattern} holds a test" >&2
	exit 2
fi

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
status=0
over=()

for package in "${packages[@]}"; do
	name="${package##*/}"
	bin="${work}/${name}.test"
	if ! go test -c -tags integration -o "${bin}" "${package}" >"${work}/build.log" 2>&1; then
		cat "${work}/build.log" >&2
		status=1
		continue
	fi
	# A package can hold no test under the tag even when it holds test
	# files: -c still writes a binary, and -list then prints nothing.
	mapfile -t tests < <("${bin}" -test.list '.*' 2>/dev/null | grep -E '^(Test|Fuzz|Example)')
	if [ "${#tests[@]}" -eq 0 ]; then
		continue
	fi

	printf '\n== %s (%d tests)\n' "${package}" "${#tests[@]}"
	for test in "${tests[@]}"; do
		log="${work}/${name}.${test}"
		/usr/bin/time -f '%M %e' -o "${log}.rss" \
			env "GOMAXPROCS=${procs}" "TASK_UI_TEST_MEM_KILL=${kill_at}" \
			"${bin}" -test.run "^${test}\$" -test.timeout 5m >"${log}.log" 2>&1
		code=$?
		read -r kb secs <"${log}.rss" 2>/dev/null || {
			kb=0
			secs=0
		}
		if [ "${code}" -ne 0 ]; then
			status=1
			printf '%9s %7s %s FAILED\n' "$((kb / 1024))MiB" "${secs}s" "${test}"
			tail -20 "${log}.log" >&2
			continue
		fi
		if [ "${kb}" -gt "${ceiling}" ]; then
			over+=("${package} ${test} $((kb / 1024))MiB")
		fi
		printf '%9s %7s %s\n' "$((kb / 1024))MiB" "${secs}s" "${test}"
	done
done

if [ "${#over[@]}" -gt 0 ]; then
	printf '\nover the %s ceiling:\n' "${max}" >&2
	printf '  %s\n' "${over[@]}" >&2
	echo >&2
	echo "A test that holds this much alone is holding something it could drop." >&2
	echo "docs/testing.md says how to find what." >&2
	status=1
fi

exit "${status}"
