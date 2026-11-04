#!/usr/bin/env bash
# Compare every page of the running instance against the ported front end.
#
#   scripts/compare-site.sh                 # crawl from / and compare each page
#   scripts/compare-site.sh /settings       # start the crawl somewhere else
#
# The page list is the site's own: scripts/crawl.sh walks every link from the
# start path and reports the paths that answered 200 with HTML, and each one is
# then handed to scripts/compare.sh. Nothing is kept by hand, so a page added
# to the sidebar is compared the run after it is linked, and a folder's every
# job page is compared rather than the one a list would have named. Both front
# ends are crawled and the lists are read together, so a page only one of them
# links is compared rather than missed.
#
# The exit code is the count that differ, as scripts/compare.sh exits, plus one
# for a broken link the crawl found: both are the port's problem, and a round's
# verification gates on the number being zero.
#
# Works against an instance signed in with TASK_UI_AUTOLOGIN, since curl has no
# session. Target another with TASK_UI_BASE, and run it a second time against
# an instance signed in as a reader who is not an admin: a permission gate one
# front end has and the other does not is invisible to an admin's sweep.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}"

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

# The paths a page list holds that this does not compare. A byte stream is not
# a page and is not ported — docs/phpfront.md says so of the three of them —
# and the static files answer under their own router on both sides.
#
# The modules of /m are the other PHP tree, served by the module runtime rather
# than by this front end.
readonly STREAMS='^/(files|artifacts/files|static|assets|api)/|\.tgz$|^/(healthz|logout)$'

# Both front ends are crawled and the page lists are read together, because
# each crawl is blind to a page only the other draws. The ported pages hold the
# root and the Go pages answer under a prefix, and a page list is one front
# end's own links: a crawl of the root alone would never reach a page the Go
# sidebar offers and the ported one does not, and the page that was never
# fetched is the page that reads as compared.
#
# The link check is the crawl of the root alone. The ported pages are what the
# instance serves, so a broken link there is what a reader meets. The crawl
# under the prefix cannot answer that question: it puts the prefix on every
# path, and a page links the root as well as itself — the telemetry dashboard,
# the byte streams, a module — so each of those reads as broken when it is
# only not under the prefix. That crawl is read for its page list and its
# status is not.
pages="${work}/pages"
ported="${work}/ported"
legacy="${work}/legacy"
crawled=0
if ! TASK_UI_CRAWL_OUT="${ported}" scripts/crawl.sh "${1:-/}"; then
  crawled=1
fi
TASK_UI_CRAWL_PREFIX="${TASK_UI_GO_BASE:-/legacy}" TASK_UI_CRAWL_OUT="${legacy}" \
  scripts/crawl.sh "${1:-/}" >/dev/null 2>&1 || true
sort -u "${ported}" "${legacy}" >"${pages}"

# A page of /m is a page of a module and not of this front end, so nothing of
# it is compared. The refinement board is the one that moved here, and the move
# changed the page: the module chrome around it is gone, its badges and its
# render footer with it, and the thread page it links became a page with a
# heading, a crumb trail and the thread's files in its actions. Its own
# fixtures are what cover it, as for any module's PHP.
list=()
while read -r path; do
  case "${path}" in
  /m/*) echo "compare-site: not compared, a module serves it: ${path}" >&2 ;;
  *) list+=("${path}") ;;
  esac
done < <(grep -Ev "${STREAMS}" "${pages}")

if [ "${#list[@]}" -eq 0 ]; then
  echo "compare-site: the crawl found no page to compare" >&2
  exit 1
fi

# compare.sh reports a page that differs on stderr with its diff under it, and
# a page that is the same on stdout: the diffs are what a run is read for, so
# they pass through and the clean pages are counted.
echo "comparing ${#list[@]} pages"
differ=0
scripts/compare.sh "${list[@]}" >"${work}/out" || differ=$?
same="$(grep -c ': the same' "${work}/out" || true)"
echo "the same: ${same} of ${#list[@]}"

# A status wraps past 255 and saturates rather than wrapping, as in compare.sh:
# the count is read, but a run that failed must not exit 0.
status="$((differ + crawled))"
if [ "${status}" -gt 255 ]; then
  status=255
fi

exit "${status}"
