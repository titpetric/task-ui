#!/usr/bin/env bash
# Take the screenshots the README shows.
#
#   scripts/demo.sh
#
# Builds task-ui, serves a copy of testdata/demo from throwaway instances,
# records a few real runs over the websocket with scripts/run.sh, adds a
# secret and an env file, and captures every view with scripts/chromium.sh
# in both themes. Writes .github/assets/*.png.
#
# Needs chromium, curl, atkins and task in PATH. imagemagick is optional and
# shrinks the files.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

for tool in chromium curl atkins task git; do
  command -v "$tool" >/dev/null || { echo "$tool not found in PATH" >&2; exit 1; }
done

work="$(mktemp -d)"
pids=()
cleanup() {
  for pid in "${pids[@]}"; do kill "$pid" 2>/dev/null || true; done
  rm -rf "$work"
}
trap cleanup EXIT

go build -o "$work/task-ui" .

# Runs generate env files inside the folders, so the demo is served from
# a copy and the checkout stays clean.
cp -r testdata/demo "$work/demo"

# Login is mandatory: three users, one per role. The instances sign in as
# one of them with TASK_UI_AUTOLOGIN, since chromium's screenshot flag has
# no way to log in; the sign in screen comes from an instance without it.
{
  echo "users:"
  for user in admin:admin ops:runner guest:viewer; do
    echo "  - name: ${user%%:*}"
    echo "    password: \"$("$work/task-ui" hash demo)\""
    echo "    role: ${user##*:}"
  done
  cat "$work/demo/task-ui.yml.example"
} >"$work/demo/task-ui.yml"

# The infra folder comes from a git repository, so the Projects page has
# a checkout to show and every run of it pulls first.
git() { command git -c user.name=demo -c user.email=demo@example -c init.defaultBranch=main "$@"; }
mkdir -p "$work/repos"
mv "$work/demo/infra" "$work/repos/infra"
printf '.env\n' >"$work/repos/infra/.gitignore"
git -C "$work/repos/infra" init -q
git -C "$work/repos/infra" add .
git -C "$work/repos/infra" commit -q -m "Provisioning and deploy jobs"
git clone -q --bare "$work/repos/infra" "$work/repos/infra.git"

# serve starts an instance on a port, sharing the history, keys and env
# store, and waits for it to answer. Extra arguments are KEY=value.
serve() {
  local port="$1"
  shift
  env TASK_UI_ASSETS="$root" \
    TASK_UI_ROOT="$work/demo" \
    TASK_UI_LISTEN="127.0.0.1:${port}" \
    TASK_UI_HISTORY_PATH="$work/history" \
    TASK_UI_SSH_DIR="$work/ssh" \
    TASK_UI_ENV_PATH="$work/env" \
    "$@" "$work/task-ui" >"$work/task-ui-${port}.log" 2>&1 &
  pids+=("$!")
  for _ in $(seq 1 50); do
    curl -fs -o /dev/null "http://127.0.0.1:${port}/healthz" && return 0
    sleep 0.1
  done
  echo "task-ui on ${port} did not start" >&2
  cat "$work/task-ui-${port}.log" >&2
  exit 1
}

# post submits a form to the admin instance, the way the pages do.
post() {
  local path="$1"
  shift
  curl -fsS -o /dev/null -H "Origin: ${TASK_UI_BASE}" "$@" "${TASK_UI_BASE}${path}"
}

# The admin's instance.
admin=3101
serve "$admin" TASK_UI_AUTOLOGIN=admin
export TASK_UI_BASE="http://127.0.0.1:${admin}"

# Add the infra project from its repository: the clone is the test.
post /settings/projects/clone -d "url=file://$work/repos/infra.git" -d id=infra

# Secrets and a managed env file, so the runs and the pages have them.
post /settings/secrets/save -d scope=all -d key=CDN_PURGE_URL -d value=https://cdn.example/purge
post /settings/secrets/save -d scope=infra -d key=REGION -d value=eu-central-1
post /settings/secrets/save -d scope=infra -d key=DEPLOY_TOKEN -d value=s3cret
post /f/infra/env/.env --data-urlencode "content=$(cat "$work/repos/infra/.env")"
post /settings/keys/generate -d type=ed25519 -d comment=task-ui

# Real runs, so the history has something to show: a forced run that kills
# the one before it, a plain one, and runs in the other folders.
scripts/run.sh infra deploy >/dev/null &
first=$!
sleep 1
scripts/run.sh infra deploy force >/dev/null
scripts/run.sh infra plan >/dev/null
scripts/run.sh backups verify >/dev/null
scripts/run.sh tools make:test >/dev/null
scripts/run.sh tools script:cleanup >/dev/null
scripts/run.sh infra deploy >/dev/null
wait "$first"

# The newest deploy run, for the dialog shots.
run="$(basename "$(ls "$work/history/infra/deploy/"*.json | sort | tail -1)" .json)"

export TASK_UI_SHOTS="$work"
shots=()
shot() {
  local name="$1"
  shift
  shots+=("$(scripts/chromium.sh "$1" "$name" "${@:2}")")
}

for theme in dark light; do
  shot "overview-${theme}" / "$theme"
  shot "folder-${theme}" /f/infra "$theme"
  shot "tools-${theme}" /f/tools "$theme" 1440x1000
  shot "job-${theme}" /f/infra/deploy "$theme"
  # Part way through the replay, so the dialog shows a run in progress.
  shot "run-replay-${theme}" "/f/infra/deploy#replay=${run}" "$theme" 1440x900 1800
  shot "keys-${theme}" /settings/keys "$theme"
  shot "secrets-${theme}" /settings/secrets "$theme"
  shot "envfile-${theme}" /f/infra/env/.env "$theme"
  shot "projects-${theme}" /settings/projects "$theme"
done
shot project-new-dark /settings/projects/new dark
shot settings-dark /settings dark
shot keys-generate-dark /settings/keys/generate dark
shot keys-upload-dark /settings/keys/upload dark
shot secret-new-dark "/settings/secrets/new?scope=infra" dark
shot mobile-folder-dark /f/infra dark 390x844

# A viewer's instance over the same history.
viewer=3102
serve "$viewer" TASK_UI_AUTOLOGIN=guest
TASK_UI_BASE="http://127.0.0.1:${viewer}" shot viewer-job-dark /f/infra/deploy dark

# An instance nobody is signed in to, for the sign in screen.
login=3103
serve "$login"
for theme in dark light; do
  TASK_UI_BASE="http://127.0.0.1:${login}" shot "login-${theme}" /login "$theme"
done

mkdir -p .github/assets
rm -f .github/assets/*.png
for file in "${shots[@]}"; do
  name="$(basename "$file" .png)"
  out=".github/assets/${name#task-ui-}.png"
  if command -v magick >/dev/null 2>&1; then
    magick "$file" -strip -colors 256 -define png:compression-level=9 "$out"
  else
    cp "$file" "$out"
  fi
  echo "$out"
done
cp .github/assets/overview-dark.png .github/assets/task-ui.png
