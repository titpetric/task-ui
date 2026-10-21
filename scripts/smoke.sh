#!/usr/bin/env bash
# Exercise a fresh task-ui end to end: runs, hours, secrets and keys, env
# files, schedules, users, projects, artifacts, the pages, and the browser's
# console. Every check names what it expected; the first failure stops it.
#
#   scripts/smoke.sh
#
# Needs curl, chromium, atkins, task and git in PATH. Serves a copy of
# testdata/demo from a throwaway instance on a free port.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
for tool in curl chromium atkins task git; do
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
cp -r testdata/demo "$work/demo"
for user in admin ops guest; do
  TASK_UI_ROOT="$work/demo" TASK_UI_HTPASSWD="$work/demo/.htpasswd" "$work/task-ui" passwd "$user" demo >/dev/null
done

port=3301
export TASK_UI_BASE="http://127.0.0.1:${port}"
env TASK_UI_ROOT="$work/demo" TASK_UI_LISTEN="127.0.0.1:${port}" TASK_UI_DATA="$work/data" \
  TASK_UI_SSH_DIR="$work/ssh" TASK_UI_ADMINS=admin TASK_UI_VIEWERS=guest \
  TASK_UI_AUTOLOGIN=admin TASK_UI_NOW=2026-10-16T22:00:00+02:00 \
  "$work/task-ui" >"$work/task-ui.log" 2>&1 &
pids+=("$!")
for _ in $(seq 1 50); do
  curl -fs -o /dev/null "${TASK_UI_BASE}/healthz" && break
  sleep 0.1
done

passed=0
# check <name> <expected substring> <command...>: runs the command, expects
# its output to hold the substring.
check() {
  local name="$1" want="$2"
  shift 2
  local got
  got="$("$@" 2>&1 || true)"
  if [[ "$got" != *"$want"* ]]; then
    echo "FAIL $name: expected $want" >&2
    echo "  got: ${got:0:400}" >&2
    echo "--- server log:" >&2
    tail -20 "$work/task-ui.log" >&2
    exit 1
  fi
  passed=$((passed + 1))
  echo "ok   $name"
}

# api <method> <path> [json]: calls the API as the pages do, printing the
# status and the body.
api() {
  local method="$1" path="$2" body="${3:-}"
  curl -sS -w '\n%{http_code}' -X "$method" -H "Origin: ${TASK_UI_BASE}" -H "Content-Type: application/json" \
    ${body:+-d "$body"} "${TASK_UI_BASE}/api${path}"
}
field() { grep -o "\"$1\":\"[^\"]*\"" | head -1 | cut -d'"' -f4; }
# The checks that run several calls do so in a sub shell.
export -f api field

# ---- runs
check "a run starts" '"state":"running"' api POST /runs '{"folder":"infra","job":"plan","cols":100,"rows":30}'
id="$(api GET '/runs?folder=infra&job=plan&limit=1' | field id)"
for _ in $(seq 1 100); do
  [ "$(api GET "/runs/${id}" | field state)" != "running" ] && break
  sleep 0.2
done
check "the run ends ok" '"state":"ok"' api GET "/runs/${id}"
check "its frame is text" 'aws_instance' curl -sS "${TASK_UI_BASE}/api/runs/${id}/frame.txt"
check "its frame is an image, trimmed" 'viewBox="0 0 840.0 136.0"' curl -sS "${TASK_UI_BASE}/api/runs/${id}/frame.svg"
check "its recording is an asciicast" '"version":2' curl -sS "${TASK_UI_BASE}/api/runs/${id}/cast"
check "an interactive job starts and waits" '"state":"running"' api POST /runs '{"folder":"infra","job":"shell"}'
check "a second run of a running job is refused" '"error":"already running"' api POST /runs '{"folder":"infra","job":"shell"}'
shell="$(api GET '/runs?folder=infra&job=shell&state=running&limit=1' | field id)"
for _ in $(seq 1 40); do
  curl -sS "${TASK_UI_BASE}/api/runs/${shell}/frame.txt" | grep -q 'infra\$' && break
  sleep 0.25
done
check "its frame shows the prompt" 'infra$' curl -sS "${TASK_UI_BASE}/api/runs/${shell}/frame.txt"
check "killing it ends it" '204' curl -sS -o /dev/null -w '%{http_code}' -X POST -H "Origin: ${TASK_UI_BASE}" "${TASK_UI_BASE}/api/runs/${shell}/kill"
sleep 1
check "killed is recorded" '"state":"killed"' api GET "/runs/${shell}"
check "a cross site write is refused" '403' curl -sS -o /dev/null -w '%{http_code}' -X POST -H "Origin: https://evil.example" -H "Content-Type: application/json" -d '{"folder":"infra","job":"plan"}' "${TASK_UI_BASE}/api/runs"

# ---- hours
check "hours are set on a folder" '"hours":"mon-fri 9-17\nfri 9-13"' api PUT /projects/website '{"hours":"mon-fri 9-17\nfri 9-13"}'
check "out of hours a run is refused" '"out_of_hours":true' api POST /runs '{"folder":"website","job":"lint"}'
check "forced without a reason it is refused" 'a reason is needed' api POST /runs '{"folder":"website","job":"lint","force":true}'
check "forced with a reason it runs and keeps it" '"reason":"hotfix for the outage"' api POST /runs '{"folder":"website","job":"lint","force":true,"reason":"hotfix for the outage"}'
check "bad hours are refused" 'invalid hours' api PUT /projects/website '{"hours":"someday 9-17"}'

# ---- secrets and keys
check "no key, no secret" 'generate an ed25519 or RSA key' api POST /secrets '{"scope":"all","name":"TOKEN","content":"hunter2"}'
check "a key is generated" '"name":"id_ed25519"' api POST /keys/generate '{"type":"ed25519","comment":"smoke"}'
check "a secret is kept" '"name":"TOKEN"' api POST /secrets '{"scope":"all","name":"TOKEN","content":"hunter2"}'
check "the list never shows the content" '"secrets":[{"id"' bash -c "api GET /secrets | grep -v hunter2"
check "rotating the key rekeys the secret" '"rekeyed":1' api POST /keys/generate '{"type":"ed25519","comment":"rotated"}'
check "the last key cannot go" 'generate or upload another' api DELETE /keys/id_ed25519
backup="$(api GET /keys | grep -o '"name":"id_ed25519.bak[^"]*"' | head -1 | cut -d'"' -f4)"
check "the backup can go" '204' curl -sS -o /dev/null -w '%{http_code}' -X DELETE -H "Origin: ${TASK_UI_BASE}" "${TASK_UI_BASE}/api/keys/${backup}"
check "the secret still opens: it reaches an env file" 'TOKEN=hunter2' bash -c "
  api PUT /env/infra/.env '{\"content\":\"RELEASE=v1\"}' >/dev/null
  cat '$work/demo/infra/.env'"
check "the env file content is managed" '"content":"RELEASE=v1\n"' api GET /env/infra/.env
check "an env file cannot escape the folder" '400' curl --path-as-is -sS -o /dev/null -w '%{http_code}' -X PUT -H "Origin: ${TASK_UI_BASE}" -H "Content-Type: application/json" -d '{"content":"x"}' "${TASK_UI_BASE}/api/env/infra/../x"

# ---- schedules
check "a schedule is made" '"expression":"@daily"' api POST /schedules '{"folder":"backups","job":"verify","expression":"@daily"}'
sid="$(api GET /schedules | field id)"
check "it can be switched off" '"enabled":0' api PUT "/schedules/${sid}" '{"enabled":false}'
check "a bad expression is refused" '400' curl -sS -o /dev/null -w '%{http_code}' -X PUT -H "Origin: ${TASK_UI_BASE}" -H "Content-Type: application/json" -d '{"expression":"every day"}' "${TASK_UI_BASE}/api/schedules/${sid}"
check "it can be deleted" '204' curl -sS -o /dev/null -w '%{http_code}' -X DELETE -H "Origin: ${TASK_UI_BASE}" "${TASK_UI_BASE}/api/schedules/${sid}"

# ---- users, from the password file
check "the users come from the file" '"name":"guest"' api GET /users
check "the roles come from the environment" '"role":"viewer"' api GET /users
uid="$(api GET /users | grep -o '{"id":"[^"]*","name":"ops"' | cut -d'"' -f4)"
check "folders are set here" '"folders":"[\"infra\"]"' api PUT "/users/${uid}" '{"folders":["infra"]}'
check "users are not made here" '405' curl -sS -o /dev/null -w '%{http_code}' -X POST -H "Origin: ${TASK_UI_BASE}" -H "Content-Type: application/json" -d '{"name":"x"}' "${TASK_UI_BASE}/api/users"
check "a new line in the file signs in without a restart" '"name":"late"' bash -c "
  TASK_UI_ROOT='$work/demo' TASK_UI_HTPASSWD='$work/demo/.htpasswd' '$work/task-ui' passwd late demo >/dev/null
  touch -d '+2 seconds' '$work/demo/.htpasswd'
  for _ in \$(seq 1 80); do api GET /users | grep -q '\"name\":\"late\"' && break; sleep 0.25; done
  api GET /users"
check "signing in with the file's password works" '303' bash -c "curl -sS -o /dev/null -w '%{http_code}' -d 'username=late&password=demo' '${TASK_UI_BASE}/login'"
check "a wrong password does not" '401' bash -c "curl -sS -o /dev/null -w '%{http_code}' -d 'username=late&password=nope' '${TASK_UI_BASE}/login'"

# ---- projects
git() { command git -c user.name=smoke -c user.email=smoke@example -c init.defaultBranch=main "$@"; }
mkdir -p "$work/repo" && printf 'build:\n\techo built\n' >"$work/repo/Makefile"
git -C "$work/repo" init -q && git -C "$work/repo" add . && git -C "$work/repo" commit -q -m init
check "a repository clones into the root" '"runnable":true' api POST /projects "{\"url\":\"file://$work/repo\",\"id\":\"cloned\"}"
check "the clone is a folder with its targets" '"id":"cloned"' bash -c "curl -sS '${TASK_UI_BASE}/f/cloned' | grep -o 'data-folder=\"cloned\"' | head -1 | sed 's/data-folder=/\"id\":/'"
check "a bad URL is refused" 'enter a git URL' api POST /projects '{"url":"not a url"}'
check "the clone can be removed" '204' curl -sS -o /dev/null -w '%{http_code}' -X DELETE -H "Origin: ${TASK_UI_BASE}" "${TASK_UI_BASE}/api/projects/cloned"
check "and is gone" '404' curl -sS -o /dev/null -w '%{http_code}' "${TASK_UI_BASE}/f/cloned"

# ---- artifacts
# tools reads its Makefile alone by default; its sources add the scripts.
check "sources add the scripts folder" '"sources":"[\"make\",\"script\"]"' api PUT /projects/tools '{"sources":["make","script"]}'
check "a job leaves an artifact" '"name":"report.txt"' bash -c "
  api POST /runs '{\"folder\":\"tools\",\"job\":\"script:report\"}' >/dev/null
  for _ in \$(seq 1 40); do api GET '/artifacts?folder=tools' | grep -q report.txt && break; sleep 0.25; done
  api GET '/artifacts?folder=tools'"
aid="$(api GET '/artifacts?folder=tools' | field id)"
check "it downloads" 'tools usage report' curl -sS "${TASK_UI_BASE}/api/artifacts/${aid}"
check "it can be deleted" '204' curl -sS -o /dev/null -w '%{http_code}' -X DELETE -H "Origin: ${TASK_UI_BASE}" "${TASK_UI_BASE}/api/artifacts/${aid}"

# ---- the pages, and the browser's console
check "every page answers" 'every link answers' scripts/crawl.sh
console() {
  timeout 45 chromium --headless --no-sandbox --disable-gpu --enable-logging=stderr --v=0 --virtual-time-budget=3000 \
    --dump-dom "${TASK_UI_BASE}$1" 2>&1 | grep -E 'CONSOLE|Uncaught' | grep -vi 'deprecat' || true
}
for path in / /runs /f/infra /f/infra/deploy /settings/projects/infra /schedules '/f/website#start=website/lint'; do
  check "no console errors on ${path}" '' bash -c "out=\"\$(console '$path')\"; [ -z \"\$out\" ] && echo ok || echo \"\$out\""
done
check "the sidebar runs a job in place" 'class="play" data-run="" data-folder="infra" data-job="plan"' bash -c "timeout 45 chromium --headless --no-sandbox --disable-gpu --virtual-time-budget=2000 --dump-dom '${TASK_UI_BASE}/f/infra' 2>/dev/null"
check "the out of hours window asks why" 'id="run-reason-text"' bash -c "timeout 45 chromium --headless --no-sandbox --disable-gpu --virtual-time-budget=3000 --dump-dom '${TASK_UI_BASE}/f/website#start=website/lint' 2>/dev/null | grep -A2 'run-reason\"' | grep -v hidden"

# The window itself: started from the page and attached over the socket
# in the browser. The hint is what the window says once the start answered
# and the socket opened; the run's end is read back from the API, since
# the browser's virtual time runs out before real time delivers it.
check "the window starts the run and attaches over the socket" 'Stop ends it, Kill ends it now.' bash -c "timeout 60 chromium --headless --no-sandbox --disable-gpu --virtual-time-budget=6000 --dump-dom '${TASK_UI_BASE}/f/infra#start=infra/plan' 2>/dev/null"
last="$(api GET '/runs?folder=infra&job=plan&limit=1' | field id)"
for _ in $(seq 1 40); do
  [ "$(api GET "/runs/${last}" | field state)" != "running" ] && break
  sleep 0.25
done
check "the run the window started ended ok" '"state":"ok"' api GET "/runs/${last}"

echo "smoke: ${passed} checks passed"
