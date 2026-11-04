#!/usr/bin/env bash
# Compare a page of the Go front end against the same page of the one being
# ported, and report what differs.
#
#   scripts/compare.sh /notes
#   scripts/compare.sh /runs /runs/01hz /schedules
#   scripts/compare.sh /m/titpetric/refinement=/refinement   paths that differ
#
# A pair is <go-path>=<php-path>, and the Go path is the one the prefix goes on.
#
# Several paths report one line each and the diffs below them, and the exit
# code is the count that differ, so a round's verification gates on it.
#
# Fetches <base>/legacy<path> and <base><path> from the running instance,
# normalises both and diffs them. Exits 1 when they differ, so a round's
# verification gates on it.
#
# The ported pages answer at the root, because they are what the instance
# serves, and the Go pages answer under app.GoBase. It was the other way round
# until the port took the root, so the prefix moved from the PHP fetch to the
# Go one and nothing else here changed: neither front end writes a prefix into
# a link, so both sides read the same either way.
#
# Normalising is what makes the diff readable, and it is deliberately narrow:
# the mount prefix comes off every link, runs of whitespace between tags
# collapse, the asset version is dropped, an identifier that changes per run is
# masked, and the static files answer under one name. Everything else is a real
# difference.
#
# The static files are one copy under two URLs while both front ends are alive:
# Go writes /static and the theme writes /assets/theme, and internal/assets
# serves the same bytes at both. /static goes at cutover, so the two read as
# one here.
#
# The quote entities are two spellings of one character: html/template writes
# &#34; and &#39;, and MiniTPL compiles htmlspecialchars with ENT_QUOTES, which
# writes &quot; and &#039;. A browser reads the same text from either, and the
# escaper is upstream's, so they read as one here too.
#
# The apostrophe and the plus are read as the characters themselves, on both
# sides, because one side escapes them where the other does not: a page that
# holds the prose in the template leaves the apostrophe of "Claude's" alone,
# and the same sentence built in PHP and escaped on the way out writes the
# entity; html/template escapes a + in text and htmlspecialchars does not.
# Neither character delimits an attribute in either front end, so the raw
# character cannot split a tag and hide a difference.
#
# A page that differs is read a second time, and only what both readings hold
# is reported: the instance under a sweep is running jobs, so a mtime and a
# "1m ago" move between the two fetches of one page, and a value that moved is
# not a difference between the two front ends.
#
# Works against an instance signed in with TASK_UI_AUTOLOGIN, since curl has
# no session. Target another with TASK_UI_BASE.
#
# A status that is no page, no redirect and not the shared 404 — a 401 from the
# Guard, a 5xx, a connection that did not answer — is reported as not compared
# and counted as a difference: nothing was drawn, so the two front ends were
# not read. Two instances with different autologin users must not share a data
# directory for the same reason — the second one's password-file sync removes
# the first one's user, and every page of the first then answers 401.
set -euo pipefail

BASE="${TASK_UI_BASE:-http://localhost:3000}"
PREFIX="${TASK_UI_GO_BASE:-/legacy}"

if [ $# -eq 0 ]; then
  echo "usage: compare.sh <path>[=<php-path>] [path...]" >&2
  exit 2
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# fetch writes a page and reports its status.
fetch() {
  curl -sS -o "$2" -w '%{http_code}' "${BASE}$1"
}

# normalise reads a page and writes it in the shape the diff compares: the
# prefix off the links, one tag per line, the per-run values masked.
normalise() {
  sed -e "s#\"${PREFIX}/#\"/#g" -e "s#\"${PREFIX}\"#\"/\"#g" "$1" |
    sed -e 's#"/assets/theme/#"/static/#g' |
    sed -e 's/&quot;/\&#34;/g' -e "s/&#039;/'/g" -e "s/&#39;/'/g" -e 's/&#43;/+/g' |
    sed -e 's/?v=[0-9a-f]\{12\}//g' |
    sed -e 's/\b[0-9a-z]\{26\}\b/ID/g' |
    sed -e 's/>/>\n/g' |
    sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' |
    grep -v '^$'
}

# reading diffs the two pages as they were last fetched and writes both the
# diff and the lines of it that differ: <stem>.diff for a reader and
# <stem>.lines, sorted, for comparing one reading against another.
reading() {
  local stem="$1" left="$2" right="$3"

  normalise "$work/go.html" | sed -e "s#\"${left}/#\"${right}/#g" -e "s#\"${left}\"#\"${right}\"#g" >"$work/go.txt"
  normalise "$work/php.html" >"$work/php.txt"

  if diff -u "$work/go.txt" "$work/php.txt" >"${stem}.diff"; then
    : >"${stem}.lines"
    return 0
  fi

  # Byte order, because comm reads the two files as sorted and the locale's
  # collation puts punctuation where comm does not expect it: a mismatch makes
  # comm skip lines and the intersection comes out short.
  grep -E '^[-+]' "${stem}.diff" | grep -Ev '^(---|\+\+\+) ' | LC_ALL=C sort -u >"${stem}.lines"

  return 1
}

# compare reports one path and returns whether it differs. The argument is a
# path both front ends answer, or <go-path>=<php-path> for a page that moved:
# the Go path's own links are rewritten to the PHP path's, so the page a
# module served under /m compares against the page it became.
#
# The = that pairs them is the one before the query, because a query holds its
# own: /login?back=%2Fruns is one path and not a pair. The rewrite reads the
# paths without their queries, since a link is to a path.
compare() {
  local pair="$1"
  local path php left right go_status php_status lines moved first

  case "${pair%%\?*}" in
  *=*)
    path="${pair%%=*}"
    php="${pair#*=}"
    ;;
  *)
    path="$pair"
    php="$pair"
    ;;
  esac

  left="${path%%\?*}"
  right="${php%%\?*}"

  go_status="$(fetch "${PREFIX}${path}" "$work/go.html")"
  php_status="$(fetch "$php" "$work/php.html")"

  if [ "$go_status" != "$php_status" ]; then
    echo "${pair}: go answered ${go_status}, php answered ${php_status}" >&2
    return 1
  fi

  # A page, a redirect, the shared 404 and the refusal page are a comparison.
  # Everything else is identical on both sides whatever the page would have
  # held: the Guard writes "unauthorized" in front of either front end, a 500
  # writes nothing, and a connection that failed writes nothing twice.
  # Reporting the same for those reads as a clean comparison of a page nobody
  # drew, which is how a whole sweep passes against an instance whose autologin
  # user is gone.
  #
  # 403 is in and 401 is out, and the two are not the same answer. A 403 on a
  # GET is the application's own: either front end drew the page that says the
  # reader needs a permission, and access.SameOrigin, the only other source of
  # one, refuses nothing a GET asks. A 401 is the Guard's bare "unauthorized"
  # in front of both. Leaving 403 out hid a refusal page that differed from
  # Go's on every gated page in the tree, for four rounds.
  case "$go_status" in
  2?? | 3?? | 403 | 404) ;;
  *)
    echo "${pair}: not compared, both answered ${go_status}: no page was drawn" >&2
    return 1
    ;;
  esac

  if reading "$work/first" "$left" "$right"; then
    echo "${pair}: the same, ${go_status}"
    return 0
  fi

  # A page is read twice before it is called different, because a value on it
  # moves between the two fetches: a file a run of the instance is rewriting
  # carries a new mtime, and "1m ago" becomes "2m ago" while the sweep walks.
  # A real difference is in both readings and one that moved is in neither, so
  # what the two diffs hold in common is the finding.
  #
  # The second reading fetches the other way round, because the two fetches are
  # not simultaneous and view.Ago reads the clock as the page is drawn: a stamp
  # of 179 seconds ago served to the first fetch is 181 to the second, which is
  # "2m ago" against "3m ago" twice over, in the same order, on every reload.
  # Swapping the order puts the later reading on the other side, so the skew
  # lands in the other diff and the two have nothing in common.
  fetch "$php" "$work/php.html" >/dev/null
  fetch "${PREFIX}${path}" "$work/go.html" >/dev/null
  reading "$work/second" "$left" "$right" || true

  LC_ALL=C comm -12 "$work/first.lines" "$work/second.lines" >"$work/both"
  lines="$(wc -l <"$work/both")"
  moved="$(($(wc -l <"$work/second.lines") - lines))"

  if [ "$lines" -eq 0 ]; then
    first="$(wc -l <"$work/first.lines")"
    echo "${pair}: the same, ${go_status}, read twice: the ${first} lines that differed moved between the fetches"
    return 0
  fi

  echo "${pair}: ${lines} lines differ, ${go_status}" >&2
  if command -v colordiff >/dev/null; then
    colordiff <"$work/second.diff" >&2
  else
    cat "$work/second.diff" >&2
  fi
  if [ "$moved" -gt 0 ]; then
    echo "${pair}: ${moved} more lines of that diff moved between the fetches and are not a difference" >&2
  fi

  return 1
}

differ=0
for path in "$@"; do
  compare "$path" || differ=$((differ + 1))
done

if [ "$differ" -gt 0 ]; then
  echo "${differ} of $# differ" >&2
fi

# The count is the status, and a status wraps past 255: a sweep of 263 pages
# that differed on every one of them would exit 7, and one that differed on 256
# would exit 0 and read as a clean run. It saturates instead.
if [ "$differ" -gt 255 ]; then
  differ=255
fi

exit "$differ"
