#!/usr/bin/env bash
# Take the screenshots the README shows.
#
#   scripts/demo.sh
#
# Builds task-ui, serves a copy of testdata/demo from throwaway instances
# over one database, sets the instance up through the API with curl,
# starts real runs with scripts/run.sh, and captures every view with
# scripts/chromium.sh in both themes. Writes docs/assets/*.png.
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
# the environment, which the first start imports into the database. The instances sign in as one of them with
# TASK_UI_AUTOLOGIN, since chromium's screenshot flag has no way to log
# in; the sign in screen comes from an instance without it.
for user in admin ops guest; do
  TASK_UI_WORKSPACE="$work/demo" TASK_UI_DATA="$work/data" "$work/task-ui" passwd "$user" demo >/dev/null
done

# The infra folder comes from a git repository, so the Projects page has
# a checkout to show and every run of it pulls first.
git() { command git -c user.name=demo -c user.email=demo@example -c init.defaultBranch=main "$@"; }
# repo commits $work/repos/<name> and serves it bare beside it.
repo() {
  git -C "$work/repos/$1" init -q
  git -C "$work/repos/$1" add .
  git -C "$work/repos/$1" commit -q -m "$2"
  git clone -q --bare "$work/repos/$1" "$work/repos/$1.git"
}
mkdir -p "$work/repos"
mv "$work/demo/src/infra" "$work/repos/infra"
printf '.env\n' >"$work/repos/infra/.gitignore"
repo infra "Provisioning and deploy jobs"

# The acme workspace's home repository and its two projects are
# repositories of their own: the home cloned into acme/home, the
# projects side by side in acme/src.
mv "$work/demo/acme" "$work/repos/acme"
mv "$work/repos/acme/api" "$work/repos/acme-api"
mv "$work/repos/acme/web" "$work/repos/acme-web"
repo acme "The tooling every Acme project shares"
repo acme-api "The API's jobs"
repo acme-web "The front end's targets"

# serve starts an instance on a port over the shared database and waits
# for it to answer. Extra arguments are KEY=value. Every instance starts
# before the first run: a starting instance marks runs in progress killed.
serve() {
  local port="$1"
  shift
  env TASK_UI_WORKSPACE="$work/demo" \
    TASK_UI_LISTEN="127.0.0.1:${port}" \
    TASK_UI_DATA="$work/data" \
    TASK_UI_SSH_DIR="$work/ssh" \
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
prompter=3104
# The clock is pinned to a Friday evening, out of the website's hours,
# for the Run force capture: the coming Friday, so the runs the demo
# makes now fall in the activity charts' thirty days before it.
friday="$(date -d friday +%F)"
serve "$admin" TASK_UI_AUTOLOGIN=admin TASK_UI_NOW="${friday}T22:00:00+02:00"
serve "$viewer" TASK_UI_AUTOLOGIN=guest
serve "$login"
# The prompt instance answers the Prompt box with a Claude Code stand-in,
# which the acme workspace's home holds while it is empty, below.
serve "$prompter" TASK_UI_AUTOLOGIN=admin
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

# An SSH key for the keys page, which git pulls with; the instance's
# encryption key encrypts the secrets. Then secrets and a managed env
# file, so the runs and the pages have them.
api POST /keys/generate -d '{"name":"deploy","type":"ed25519","comment":"task-ui","git":true}'
api POST /secrets -d '{"scope":"all","name":"CDN_PURGE_URL","content":"https://cdn.example/purge"}'
api POST /secrets -d '{"scope":"infra","name":"REGION","content":"eu-central-1"}'
api POST /secrets -d '{"scope":"infra","name":"DEPLOY_TOKEN","content":"s3cret"}'
api PUT /env/infra/.env -d '{"content":"# Managed from task-ui; the secrets are appended on generation.\nRELEASE=v1.42.0\n"}'

# Two schedules, one of them off.
api POST /schedules -d '{"folder":"backups","job":"verify","expression":"0 3 * * *"}'
api POST /schedules -d '{"folder":"infra","job":"plan","expression":"@hourly","enabled":false}'

# Two workflows: a release chain, and one that runs it as a step, asks
# Claude about the build and ends with a summary, which the editor shows.
api POST /workflows -d '{"name":"release","steps":[{"folder":"infra","job":"plan"},{"folder":"infra","job":"deploy"},{"folder":"tools","job":"make:build"}]}'
release="$(grep -o '"id":"[a-z0-9]*"' "$work/api.out" | head -1 | cut -d'"' -f4)"
api POST /workflows -d "{\"name\":\"ship\",\"description\":\"Tests the tools, releases them, and says how it went.\",\"steps\":[{\"kind\":\"job\",\"folder\":\"tools\",\"job\":\"make:test\"},{\"kind\":\"workflow\",\"workflow\":\"${release}\"},{\"kind\":\"prompt\",\"folder\":\"tools\",\"prompt\":\"List the images the build pushed, with their tags.\"},{\"kind\":\"summary\",\"folder\":\"_\"}]}"
ship="$(grep -o '"id":"[a-z0-9]*"' "$work/api.out" | head -1 | cut -d'"' -f4)"
# The ship workflow on a schedule, for the Schedule block of its edit page.
api POST /schedules -d "{\"workflow\":\"${ship}\",\"expression\":\"0 7 * * 1-5\"}"

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
# More runs than a page holds, so the runs shot shows Next under the
# table, in a window tall enough for it.
for _ in 1 2 3; do scripts/run.sh backups verify >/dev/null 2>&1; done
shot runs-dark /runs dark 1440x1500
scripts/run.sh tools make:build >/dev/null 2>&1

for theme in dark light; do
  shot "folder-${theme}" /f/_/infra "$theme"
  shot "job-${theme}" /f/_/infra/deploy "$theme"
  shot "run-${theme}" "/runs/${run}#run=${run}" "$theme" 1440x900 2500
done
shot tools-dark /f/_/tools dark 1440x1000
shot run-page-dark "/runs/${run}" dark
shot artifacts-dark /artifacts dark
shot schedules-dark /schedules dark
shot schedules-light /schedules light
shot workflow-dark "/workflows/${release}" dark
shot workflow-edit-dark "/workflows/${ship}/edit" dark 1440x1200
shot workflow-edit-light "/workflows/${ship}/edit" light 1440x1200
shot settings-dark /settings dark
shot retention-dark /settings/retention dark 1440x1300
shot retention-light /settings/retention light 1440x1300
shot instance-dark /settings/instance dark 1440x2400
shot projects-dark /settings/projects dark
shot project-new-dark /settings/projects/new dark
shot project-edit-dark /settings/projects/infra dark
shot keys-generate-dark /settings/keys/generate dark
shot keys-upload-dark /settings/keys/upload dark
shot secret-new-dark "/settings/secrets/new?scope=infra" dark
shot users-dark /settings/users dark
# The roles' permissions, the admin's profile, with the activity of the
# runs above, and the edit page of ops, with a policy of its own over
# the runner role's: shell denied, the secrets readable.
shot permissions-dark /settings/permissions dark 1440x2400
api PUT /permissions/users/ops -d '{"shell":"deny","secret/view":"allow"}'
for theme in dark light; do
  shot "profile-${theme}" /users/admin "$theme" 1440x1200
  shot "user-edit-${theme}" /users/ops/edit "$theme" 1440x2000
done
shot user-new-dark /settings/users/new dark
api PUT /integrations/forgejo -d '{"values":{"url":"https://git.example","token":"demo"}}'
shot notes-dark /notes dark
shot envfile-dark /f/_/infra/env/.env dark
shot out-of-hours-dark "/f/_/website#start=website/lint" dark 1440x900 2000
shot mobile-dashboard-dark / dark 390x844
shot mobile-menu-dark "/f/_/infra#menu" dark 390x844
shot mobile-folder-dark /f/_/infra dark 390x1100

# The viewer's instance over the same database: it attaches read only.
TASK_UI_BASE="http://127.0.0.1:${viewer}" shot viewer-job-dark /f/_/infra/deploy dark

# The instance nobody is signed in to, for the sign in screen.
for theme in dark light; do
  TASK_UI_BASE="http://127.0.0.1:${login}" shot "login-${theme}" /login "$theme"
done

# The workspace comes last, so the shots above keep a sidebar without
# one: the acme directory with an empty home, its two projects cloned
# into its src, and a Claude token for the admin, which the Prompt box
# needs: open on the workspace's page, folded on every other. The prompt
# instance asks one prompt there, so the box shows a question and its
# answer. The home repository comes after the prompt: while the home is
# empty, its .profile puts a claude stand-in first in the PATH of the
# runs, which prints the answer in markdown; the clone replaces it.
api POST /workspaces -d '{"folder":"acme","title":"Acme","description":"The tooling every Acme project shares"}'
mkdir -p "$work/demo/acme/home/bin"
printf 'PATH="$HOME/bin:$PATH"\n' >"$work/demo/acme/home/.profile"
cat >"$work/demo/acme/home/bin/claude" <<'ANSWER'
#!/bin/sh
cat <<'MD'
The tooling every Acme project shares:

- **Skills** in `.atkins/skills/`: `git`, `compose` and `dockerfile`, which apply in each project's checkout.
- **Jobs** in `atkins.yml`: `bootstrap` clones the projects beside it, `policy` checks them before a release.
- **Projects**: `api` and `web`, each with a `release` job of its own.
MD
ANSWER
chmod +x "$work/demo/acme/home/bin/claude"
api POST /projects -d "{\"url\":\"file://$work/repos/acme-api.git\",\"id\":\"api\",\"workspace\":\"acme\"}"
api POST /projects -d "{\"url\":\"file://$work/repos/acme-web.git\",\"id\":\"web\",\"workspace\":\"acme\"}"
api PUT /integrations/claude -d '{"values":{"token":"sk-ant-oat01-demo","who":"admin,runner"}}'
# Integrations after the token, with the admin and runner roles checked:
# the Claude row's Who shows admin getting it by its own box, ops as a
# runner, and guest without.
for theme in dark light; do
  shot "integrations-${theme}" /settings/integrations "$theme" 1440x1600
done
# The workspace's own: a shared secret and one of a project in it, an SSH key every run in acme gets and one only the
# admins' runs get, and a local .env in the api checkout, which its
# Files panel lists sealed.
api POST /secrets -d '{"scope":"workspace:acme","name":"REGISTRY_URL","content":"registry.acme.example"}'
api POST /secrets -d '{"scope":"acme/api","name":"DATABASE_URL","content":"postgres://api@db/api"}'
api POST /workspaces/acme/keys/generate -d '{"type":"ed25519","name":"git","comment":"acme git"}'
api POST /workspaces/acme/keys/generate -d '{"type":"ed25519","name":"deploy","comment":"acme deploy","policies":["admin"]}'
printf 'DATABASE_URL=postgres://localhost/api\n' >"$work/demo/acme/src/api/.env"

TASK_UI_BASE="http://127.0.0.1:${prompter}" api POST /workspaces/acme/prompt -d '{"prompt":"what does this workspace hold, in a few lines"}'
for _ in $(seq 1 100); do
  curl -fsS "${TASK_UI_BASE}/api/prompts?folder=acme" | grep -q '"state":"done"' && break
  sleep 0.2
done
# A workflow of one prompt in acme, run from the prompt instance while
# the stand-in answers, so the Workflows page has a run with an answer
# to show open under Recent runs.
api POST /workflows -d '{"name":"acme brief","description":"Asks Claude what the Acme workspace holds.","steps":[{"kind":"prompt","folder":"acme","prompt":"what does this workspace hold, in a few lines"}]}'
brief="$(grep -o '"id":"[a-z0-9]*"' "$work/api.out" | head -1 | cut -d'"' -f4)"
TASK_UI_BASE="http://127.0.0.1:${prompter}" api POST "/workflows/${brief}/runs"
for _ in $(seq 1 200); do
  curl -fsS "${TASK_UI_BASE}/api/workflows/${brief}/runs?limit=1" | grep -q '"state":"running"' || break
  sleep 0.3
done
shot workflows-dark /workflows dark 1440x1500
shot workflows-light /workflows light 1440x1500

# The tooling repository as the home repository, the HOME of acme's
# runs, where atkins finds the skills of the actions bar: the clone
# replaces the empty home with the stand-in.
api PUT /workspaces/acme/home -d "{\"url\":\"file://$work/repos/acme.git\"}"
# A directory beside the projects in src with nothing to run, which the
# sidebar lists under acme all the same; the home it leaves out.
mkdir -p "$work/demo/acme/src/design"
printf '# Design\n\nMockups and the style guide.\n' >"$work/demo/acme/src/design/README.md"

# Every link on every page answers, the workspace's included.
scripts/crawl.sh

shot workspaces-dark /settings/workspaces dark 1440x1000
shot workspace-edit-dark /settings/workspaces/acme dark 1440x1500
shot workspace-edit-light /settings/workspaces/acme light 1440x1500
shot mobile-workspace-edit-dark /settings/workspaces/acme dark 390x1600
# The workspaces list after the projects at the root: a window tall
# enough for the sidebar to reach acme and its projects.
shot sidebar-workspace-dark /f/acme/api dark 1440x1900
shot workspace-prompt-dark /f/acme dark
# The workspace page's [...] menu open, with its SSH keys and Secrets,
# its Home section, the home's files, and the two pages that
# edit the workspace's keys and secrets in place.
shot workspace-menu-dark "/f/acme#actions=head" dark
shot workspace-home-dark /f/acme dark 1440x1700
shot workspace-keys-dark /settings/workspaces/acme/keys dark
shot workspace-secrets-dark /settings/workspaces/acme/secrets dark
# The default workspace's page as a dashboard of the root: each project
# in its Files list.
shot folder-dashboard-dark /f/_ dark 1440x800
# The api project's Files: .git and .env sealed, ./serve marked to run,
# and its README in the preview window.
shot files-dark /f/acme/api dark
shot file-dark "/f/acme/api#file=README.md" dark
# The Jobs table of a directory inside the api project under the git
# line of its checkout: the git, compose and dockerfile skills of acme's
# .atkins/skills that apply there, and @shell, which the admin gets.
# Then the field of @compose:custom with a line typed in, and what it
# runs as argv under it.
shot actions-dark /f/acme/api/deploy dark 1440x1100
shot args-dark "/f/acme/api/deploy#args=@compose:custom;ps%20-a%20%27;%20id%27" dark 1440x1100

# A second workspace, labs, with no SSH key, so the keys page shows a
# workspace without one.
mkdir -p "$work/repos/labs"
printf '# Labs\n\nExperiments that share nothing yet.\n' >"$work/repos/labs/README.md"
repo labs "Experiments"
api POST /workspaces -d "{\"url\":\"file://$work/repos/labs.git\",\"folder\":\"labs\",\"title\":\"Labs\"}"
shot keys-dark /settings/keys dark 1440x1100
# A key's edit form: Who gets it, the element every such form shares.
shot key-edit-dark /settings/keys/deploy_ed25519 dark 1440x1000
shot secrets-dark /settings/secrets dark 1440x1300
# A row's [...] menu open, #actions=<n> opening row n's: a key's, its
# details on top and Remove last, in both themes; a run of a job beside
# View;
# on a phone, a sheet along the bottom.
for theme in dark light; do
  shot "keys-menu-${theme}" "/settings/keys#actions=1" "$theme" 1440x1500
  shot "job-menu-${theme}" "/f/_/infra/deploy#actions=1" "$theme" 1440x1400
done
shot mobile-actions-dark "/settings/users#actions=2" dark 390x844

# Two pins of the admin: the api's docs directory, which the sidebar
# lists under acme/api, and its ./serve file, a run item under the
# project.
api POST /pins -d '{"folder":"acme/api/docs"}'
api POST /pins -d '{"folder":"acme/api","job":"./serve"}'
for theme in dark light; do
  shot "sidebar-pin-${theme}" /f/acme/api "$theme" 1440x1900
done

mkdir -p docs/assets
rm -f docs/assets/*.png
for file in "${shots[@]}"; do
  name="$(basename "$file" .png)"
  out="docs/assets/${name#task-ui-}.png"
  if command -v magick >/dev/null 2>&1; then
    magick "$file" -strip -colors 256 -define png:compression-level=9 "$out"
  else
    cp "$file" "$out"
  fi
  echo "$out"
done
cp docs/assets/dashboard-dark.png docs/assets/task-ui.png

# The Inventory page's shots, from a hub and hosts of their own.
scripts/inventory-shots.sh
