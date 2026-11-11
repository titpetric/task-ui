#!/usr/bin/env bash
# Walk every link of the running task-ui and report the ones that do not
# answer 200. The pages are the map: starting at /, every href on a page
# that stays on the site is fetched once.
#
#   scripts/crawl.sh [start-path]
#
# Works against an instance signed in with TASK_UI_AUTOLOGIN. Target
# another instance with TASK_UI_BASE. Exits 1 when a link is broken.
#
# TASK_UI_CRAWL_OUT writes the pages it walked, one path per line: the ones
# that answered 200 with HTML, which is the site's own page list rather than a
# list kept by hand.
set -euo pipefail

BASE="${TASK_UI_BASE:-http://localhost:3000}"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
queue="$work/queue"
seen="$work/seen"
pages="$work/pages"
broken="$work/broken"
# A queue entry is the path and the page that linked it, tab apart, because a
# broken link is read to find what draws it: a run that answers 404 says
# nothing about which list offered it.
printf '%s\t%s\n' "${1:-/}" "the start" >"$queue"
: >"$seen"
: >"$pages"
: >"$broken"

while [ -s "$queue" ]; do
  entry="$(head -1 "$queue")"
  tail -n +2 "$queue" >"$queue.next" && mv "$queue.next" "$queue"
  path="${entry%%$'\t'*}"
  from="${entry#*$'\t'}"
  grep -qxF "$path" "$seen" && continue
  echo "$path" >>"$seen"

  status="$(curl -sS -o "$work/page" -w '%{http_code}' "${BASE}${path}")"
  if [ "$status" != "200" ]; then
    echo "$status $path, linked from ${from}" >>"$broken"
    continue
  fi
  case "$(file -b --mime-type "$work/page")" in
    text/html) echo "$path" >>"$pages" ;;
    *) continue ;;
  esac
  # Links on the site: absolute paths, the fragment dropped; a href of a
  # download is a GET as well.
  grep -o 'href="/[^"]*"' "$work/page" | cut -d'"' -f2 | sed 's/#.*//; s/&amp;/\&/g' | grep -v '^/$' | sort -u |
    while read -r link; do
      [ -n "$link" ] && ! grep -qxF "$link" "$seen" && printf '%s\t%s\n' "$link" "$path" >>"$queue"
    done || true
done

count="$(wc -l <"$seen" | tr -d ' ')"
if [ -n "${TASK_UI_CRAWL_OUT:-}" ]; then
  sort -u "$pages" >"$TASK_UI_CRAWL_OUT"
fi
if [ -s "$broken" ]; then
  echo "crawled ${count} pages, broken links:" >&2
  cat "$broken" >&2
  exit 1
fi
echo "crawled ${count} pages, every link answers"
