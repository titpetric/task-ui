#!/usr/bin/env bash
# Take the screenshots the README shows.
#
#   scripts/demo.sh
#
# Builds task-ui, serves a copy of testdata/demo from throwaway instances
# over one database, sets the instance up through the API with curl,
# starts real runs with scripts/run.sh, and captures every view with
# scripts/chromium.sh in both themes. Writes .github/assets/*.png.
#
# Needs chromium, curl, atkins, task and git in PATH.
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

# Login is mandatory: three users in the password file, the roles from
# the environment. The instances sign in as one of them with
# TASK_UI_AUTOLOGIN, since chromium's screenshot flag has no way to log
# in; the sign in screen comes from an instance without it.
for user in admin ops guest; do
  TASK_UI_ROOT="$work/demo" TASK_UI_HTPASSWD="$work/demo/.htpasswd" "$work/task-ui" passwd "$user" demo >/dev/null
done

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

# serve starts an instance on a port over the shared database and waits
# for it to answer. Extra arguments are KEY=value. Every instance starts
# before the first run: a starting instance marks runs in progress killed.
serve() {
  local port="$1"
  shift
  env TASK_UI_ROOT="$work/demo" \
    TASK_UI_LISTEN="127.0.0.1:${port}" \
    TASK_UI_DATA="$work/data" \
    TASK_UI_SSH_DIR="$work/ssh" \
    TASK_UI_SECRET=demo \
    TASK_UI_ADMINS=admin \
    TASK_UI_VIEWERS=guest \
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

admin=3101
viewer=3102
login=3103
serve "$admin" TASK_UI_AUTOLOGIN=admin TASK_UI_NOW=2026-10-16T22:00:00+02:00
serve "$viewer" TASK_UI_AUTOLOGIN=guest
serve "$login"
export TASK_UI_BASE="http://127.0.0.1:${admin}"

# api calls the admin instance the way the pages do.
api() {
  local method="$1" path="$2"
  shift 2
  curl -fsS -o "$work/api.out" -X "$method" -H "Origin: ${TASK_UI_BASE}" -H "Content-Type: application/json" "$@" "${TASK_UI_BASE}/api${path}" \
    || { echo "api ${method} ${path} failed:" >&2; cat "$work/api.out" >&2; exit 1; }
}

# Add the infra project from its repository: the clone is the test. Then
# say who may run it.
api POST /projects -d "{\"url\":\"file://$work/repos/infra.git\",\"id\":\"infra\"}"
api PUT /projects/infra -d '{"allow":["ops"],"job_allow":{"rotate-certs":["admin"]}}'
# The website folder runs in office hours only. The admin instance's
# clock is pinned to a Friday night, so a run of it asks for a reason.
api PUT /projects/website -d '{"hours":"mon-fri 9-17\nfri 9-13"}'
# The tools folder reads its Makefile alone by default; its sources name
# the package scripts and the scripts folder too.
api PUT /projects/tools -d '{"sources":["make","npm","script"]}'

# A key first: the secrets are encrypted with it. Then secrets and a
# managed env file, so the runs and the pages have them.
api POST /keys/generate -d '{"type":"ed25519","comment":"task-ui"}'
api POST /secrets -d '{"scope":"all","name":"CDN_PURGE_URL","content":"https://cdn.example/purge"}'
api POST /secrets -d '{"scope":"infra","name":"REGION","content":"eu-central-1"}'
api POST /secrets -d '{"scope":"infra","name":"DEPLOY_TOKEN","content":"s3cret"}'
api PUT /env/infra/.env -d '{"content":"# Managed from task-ui; the secrets are appended on generation.\nRELEASE=v1.42.0\n"}'

# Two schedules, one of them off.
api POST /schedules -d '{"folder":"backups","job":"verify","expression":"0 3 * * *"}'
api POST /schedules -d '{"folder":"infra","job":"plan","expression":"@hourly","enabled":false}'

# Two workflows: a release chain, and one that runs it as a step.
api POST /workflows -d '{"name":"release","steps":[{"folder":"infra","job":"plan"},{"folder":"infra","job":"deploy"},{"folder":"tools","job":"make:build"}]}'
release="$(grep -o '"id":"[a-z0-9]*"' "$work/api.out" | head -1 | cut -d'"' -f4)"
api POST /workflows -d "{\"name\":\"ship\",\"steps\":[{\"folder\":\"tools\",\"job\":\"make:test\"},{\"workflow\":\"${release}\"}]}"

# Real runs, so the history has something to show: a forced run that kills
# the one before it, a plain one, a failure, and runs in the other folders.
scripts/run.sh infra deploy >/dev/null 2>&1 &
first=$!
sleep 1.5
scripts/run.sh infra deploy force >/dev/null 2>&1
scripts/run.sh infra plan >/dev/null 2>&1
scripts/run.sh infra rotate-certs >/dev/null 2>&1 || true
scripts/run.sh backups verify >/dev/null 2>&1
scripts/run.sh tools make:test >/dev/null 2>&1
scripts/run.sh tools script:cleanup >/dev/null 2>&1
scripts/run.sh tools script:report >/dev/null 2>&1
scripts/run.sh website lint >/dev/null 2>&1 || true
scripts/run.sh website lint force "customer waiting on the fix" >/dev/null 2>&1 || true
wait "$first" || true

# A run of the release workflow, so its page has one to show.
api POST "/workflows/${release}/runs"
for _ in $(seq 1 200); do
  curl -fsS "${TASK_UI_BASE}/api/workflows/${release}/runs?limit=1" | grep -q '"state":"running"' || break
  sleep 0.3
done

# Every link on every page answers.
scripts/crawl.sh

# The newest deploy run, for the run page and the dialog.
run="$(curl -fsS "${TASK_UI_BASE}/api/runs?folder=infra&job=deploy&limit=1" | grep -o '"id":"[a-z0-9]*"' | head -1 | cut -d'"' -f4)"

export TASK_UI_SHOTS="$work"
shots=()
shot() {
  local name="$1"
  shift
  shots+=("$(scripts/chromium.sh "$1" "$name" "${@:2}")")
}

# Runs in progress for the dashboard: an interactive shell that waits for
# input, and a deploy that is part way through.
scripts/run.sh infra shell >/dev/null 2>&1 &
pids+=("$!")
scripts/run.sh infra deploy >/dev/null 2>&1 &
pids+=("$!")
sleep 1.5

for theme in dark light; do
  shot "dashboard-${theme}" / "$theme"
done
shot runs-dark /runs dark
scripts/run.sh tools make:build >/dev/null 2>&1

for theme in dark light; do
  shot "folder-${theme}" /f/infra "$theme"
  shot "job-${theme}" /f/infra/deploy "$theme"
  shot "run-${theme}" "/runs/${run}#run=${run}" "$theme" 1440x900 2500
done
shot tools-dark /f/tools dark 1440x1000
shot run-page-dark "/runs/${run}" dark
shot artifacts-dark /artifacts dark
shot schedules-dark /schedules dark
shot workflows-dark /workflows dark
shot workflow-dark "/workflows/${release}" dark
shot settings-dark /settings dark
shot projects-dark /settings/projects dark
shot project-new-dark /settings/projects/new dark
shot project-edit-dark /settings/projects/infra dark
shot keys-dark /settings/keys dark
shot keys-generate-dark /settings/keys/generate dark
shot keys-upload-dark /settings/keys/upload dark
shot secrets-dark /settings/secrets dark
shot secret-new-dark "/settings/secrets/new?scope=infra" dark
shot users-dark /settings/users dark
shot envfile-dark /f/infra/env/.env dark
shot out-of-hours-dark "/f/website#start=website/lint" dark 1440x900 2000
shot mobile-dashboard-dark / dark 390x844
shot mobile-menu-dark "/f/infra#menu" dark 390x844
shot mobile-folder-dark /f/infra dark 390x1100

# The viewer's instance over the same database: it attaches read only.
TASK_UI_BASE="http://127.0.0.1:${viewer}" shot viewer-job-dark /f/infra/deploy dark

# The instance nobody is signed in to, for the sign in screen.
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
cp .github/assets/dashboard-dark.png .github/assets/task-ui.png
