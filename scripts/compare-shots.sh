#!/usr/bin/env bash
# Shoot a page on both front ends, in both palettes, and compare the pixels.
#
#   scripts/compare-shots.sh /notes
#   scripts/compare-shots.sh / /runs /settings
#
# Each path reports one line per palette and the exit code is the count of
# shots that differ, so a round's verification gates on it.
#
# This is the half of the comparison scripts/compare.sh cannot make. That one
# proves the two front ends write the same markup; a stylesheet rule that
# matches a class one of them stopped writing, a font that loads from one path
# and not the other, an SVG that draws outside its box — none of those are in
# the HTML. A browser is what reads them.
#
# The metric is AE, the count of pixels that are not the same colour, and the
# bar is zero: two renderings of one page in one browser at one size are the
# same image or the port moved something.
#
# A shot that differs is taken again, the two front ends the other way round,
# and the smaller of the two counts is the finding. A screenshot is not an
# instant: a page carries "2m ago" cells that the Go front end and the port
# both read off the process clock, and a minute that turns between the two
# browsers repaints a cell in both renderings of the next pair. A layout
# difference is in every pair; a cell that turned is in one.
#
# What that does not reach, and what no budget reaches either, is a block the
# browser fills itself after the load event: the run page's output is a
# terminal its script writes out of the recording it fetches, and the shot is
# taken before or after that write depending on the moment. Both front ends
# write the same placeholder and load the same script, so the difference is in
# the capture and never in the port — two shots of one front end differ as far
# as a shot of each. Such a page is compared as markup and is not in a shot
# set.
#
# Works against an instance signed in with TASK_UI_AUTOLOGIN, since the
# browser has no session of its own. Target another with TASK_UI_BASE.
set -euo pipefail

BASE="${TASK_UI_BASE:-http://localhost:3000}"
# The ported pages answer at the root, the Go pages under app.GoBase, so the
# prefix goes on the Go shot. compare.sh says why that is the way round.
PREFIX="${TASK_UI_GO_BASE:-/legacy}"
SIZE="${TASK_UI_SHOT_SIZE:-1440x900}"

# Where the three images of a shot that differs are kept: the two renderings
# and the difference between them. A count of pixels names nothing a reader
# can act on, and the instance a sweep found it on is torn down by the time
# the sweep reports. Default is the system temporary directory.
KEEP="${TASK_UI_SHOT_DIFF_DIR:-${TMPDIR:-/tmp}/task-ui-shot-diffs}"

if [ "$#" -eq 0 ]; then
  echo "usage: compare-shots.sh <path> [path...]" >&2
  exit 2
fi

command -v magick >/dev/null || {
  echo "compare-shots: magick not in PATH" >&2
  exit 1
}

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# shot renders one page of one front end and writes it under a name of its own.
shot() {
  local path="$1" name="$2" theme="$3"

  TASK_UI_SHOTS="$work" scripts/chromium.sh "$path" "$name" "$theme" "$SIZE" >/dev/null
}

# slug names a path the way a file can hold it.
slug() {
  printf '%s' "$1" | sed -e 's#^/##' -e 's#[^A-Za-z0-9]#-#g' -e 's#^$#root#'
}

# keep writes the two renderings and the difference between them where a
# reader can open them, and names where they went.
keep() {
  local path="$1" theme="$2" go="$3" php="$4"
  local stem

  mkdir -p "$KEEP"
  stem="${KEEP}/$(slug "$path")-${theme}"
  cp "$go" "${stem}-go.png"
  cp "$php" "${stem}-php.png"
  magick compare -metric AE "$go" "$php" "${stem}-diff.png" >/dev/null 2>&1 || true
  echo "${stem}-{go,php,diff}.png" >&2
}

# pixels is how many pixels of two shots are not the same colour, which magick
# reports on stderr as "<count> (<share>)" and answers 1 for. A pair magick
# refuses to compare — two sizes, a file it could not read — answers -1: that
# is the whole page moved and it is reported as such rather than as a count.
pixels() {
  local left="$2" right="$3" out count
  out="$(magick compare -metric AE "$left" "$right" null: 2>&1 || true)"
  count="${out%% *}"

  case "$count" in
  "" | *[!0-9.e+]*)
    echo "-1"
    return
    ;;
  esac

  printf '%.0f\n' "$count"
}

# reading shoots one palette of one page on both front ends and answers the
# pixels that differ. The order is a parameter because the second reading
# reverses it.
reading() {
  local path="$1" theme="$2" php_first="$3"
  local go="$work/task-ui-go.png" php="$work/task-ui-php.png"

  if [ "$php_first" = yes ]; then
    shot "$path" php "$theme"
    shot "${PREFIX}${path}" go "$theme"
  else
    shot "${PREFIX}${path}" go "$theme"
    shot "$path" php "$theme"
  fi

  pixels "$path" "$go" "$php"
}

differ=0
shots=0
for path in "$@"; do
  for theme in light dark; do
    shots=$((shots + 1))
    first="$(reading "$path" "$theme" no)"

    if [ "$first" = 0 ]; then
      echo "${path} ${theme}: the same"
      continue
    fi

    second="$(reading "$path" "$theme" yes)"
    least="$first"
    if [ "$second" != -1 ] && [ "$second" -lt "$least" ]; then
      least="$second"
    fi

    if [ "$least" = 0 ]; then
      echo "${path} ${theme}: the same, read twice: ${first} pixels moved between the shots"
      continue
    fi

    if [ "$least" = -1 ]; then
      echo "${path} ${theme}: the two shots are not the same size" >&2
    else
      echo "${path} ${theme}: ${least} pixels differ" >&2
    fi

    keep "$path" "$theme" "$work/task-ui-go.png" "$work/task-ui-php.png"
    differ=$((differ + 1))
  done
done

if [ "$differ" -gt 0 ]; then
  echo "${differ} of ${shots} shots differ" >&2
fi

# The count is the status, and a status wraps past 255, as in compare.sh.
if [ "$differ" -gt 255 ]; then
  differ=255
fi

exit "$differ"
