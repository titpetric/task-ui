#!/usr/bin/env bash
#
# test-sleeps reports every sleep in a test.
#
# A sleep is poor management of a lifetime. It says "something is
# happening elsewhere and I do not know when it ends", and that is a
# question the code can almost always answer: a channel, a Done(), a
# subscription, a healthcheck, `docker compose up --wait`. A sleep that
# is long enough is slow; one that is short enough is flaky; there is no
# length that is both.
#
#   scripts/test-sleeps.sh            # report, exit 1 on an unjustified one
#   scripts/test-sleeps.sh --list     # report every one, justified or not
#
# A sleep that is the subject of the test rather than a wait inside it
# carries its reason on the line above, as a comment holding "lifetime:".
# A timeout test needs something that outlives the timeout, and saying so
# is the difference between a decision and an accident:
#
#   // lifetime: outlives the 200ms timeout, which is what this asserts
#   "slow": "sleep 5",
set -u -o pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

list=0
[ "${1:-}" = "--list" ] && list=1

mark="lifetime:"
found=0
justified=0

report() {
	local file="$1" line="$2" text="$3"
	# The reason sits on the line itself, which suits a fixture entry, or
	# anywhere in the comment block above it, which suits a statement and
	# leaves room to say why in a sentence.
	local block="${text}" at=$((line - 1)) above
	while [ "${at}" -gt 0 ]; do
		above="$(sed -n "${at}p" "${file}")"
		printf '%s' "${above}" | grep -qE '^[[:space:]]*(//|#)' || break
		block="${block}
${above}"
		at=$((at - 1))
	done
	if printf '%s' "${block}" | grep -q "${mark}"; then
		justified=$((justified + 1))
		[ "${list}" -eq 1 ] && printf 'ok   %s:%s %s\n' "${file}" "${line}" "$(printf '%s' "${text}" | sed 's/^[[:space:]]*//' | cut -c1-70)"
		return
	fi
	found=$((found + 1))
	printf '%s:%s:1: a sleep waits on a lifetime nothing manages: %s\n' \
		"${file}" "${line}" "$(printf '%s' "${text}" | sed 's/^[[:space:]]*//' | cut -c1-70)"
}

while IFS=: read -r file line text; do
	[ -n "${file}" ] || continue
	report "${file}" "${line}" "${text}"
done < <(
	git ls-files '*_test.go' |
		xargs -r grep -nE 'time\.Sleep\(|\bsleep [0-9.]+' 2>/dev/null
)

echo
if [ "${found}" -gt 0 ]; then
	printf '%d sleep(s) with no reason given, %d justified.\n' "${found}" "${justified}" >&2
	echo >&2
	echo "Wait on what the code already offers: a channel, Done(), a subscription," >&2
	echo "or docker compose up --wait. To prove something did not happen, read the" >&2
	echo "channel with a default and assert it was empty. docs/testing.md has the" >&2
	echo "rules; a sleep that is the subject of a test says so in a lifetime: comment." >&2
	exit 1
fi

printf 'no unjustified sleeps; %d carry a reason.\n' "${justified}"
