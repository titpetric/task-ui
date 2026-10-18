# Task UI

Run what your source trees declare, from the browser, and watch it run.
Task UI looks at each folder under a root, detects the conventions it
follows, and lists their targets as jobs: atkins pipelines, Taskfiles,
Makefile targets, package.json and composer.json scripts, and a
`scripts/` folder. A job runs as a session on the server: a process on a
terminal that anyone allowed can attach to, from any window, for as long
as it runs.

![The dashboard: every terminal in progress, as it looks right now](https://raw.githubusercontent.com/titpetric/task-ui/main/.github/assets/task-ui.png)

- **Folders and jobs.** A folder is a working directory and whatever it
  holds that can be run. Everything on the page comes from the files
  themselves: names, descriptions, steps, the files they include. A
  folder page is its jobs as cards, each with Run; the sidebar lists
  every folder's jobs with a play button, so a job is one click from
  anywhere.
- **Runs are sessions.** Run opens a window on the run. Closing the window
  leaves the run running; Stop ends it, Kill ends it now. Anyone who may
  see the job attaches from the sidebar, the dashboard or the run list;
  runners type into it, viewers watch.
- **The dashboard.** Every terminal in progress as its last frame, for
  everyone, the way a control room shows its screens. Each frame is one
  URL, as SVG or text.
- **One run per job.** A job runs once at a time across the instance,
  different jobs run side by side. Run force kills the active run and
  starts again.
- **Workflows.** A named chain of jobs from any folders, a one button
  deploy: build, test, push, plan, apply. Run starts each step when the
  one before it ended ok; the first that does not end ok ends the
  workflow and the rest are skipped. A step can run another workflow in
  place of a job, so a release is a chain of chains; a workflow may not
  run itself, through any depth. Anyone can make a workflow from the
  jobs they see; running it needs the right to run every step.
- **Users and roles.** Login is mandatory. Users come from a password
  file, htpasswd style, read again whenever it changes; `TASK_UI_ADMINS`
  and `TASK_UI_VIEWERS` name the admins and the viewers, everyone else is
  a runner.
- **Hours.** A folder runs in office hours only when told to. Out of
  hours Run force asks why, and the reason is kept with the run.
- **Projects from git.** Admins add a folder by cloning a repository; the
  clone is the test. A checkout is pulled before every run.
- **History, artifacts, schedules.** Every run is recorded with who ran it,
  how it ended and the full terminal output. Files a job leaves under
  `$TASK_UI_ARTIFACTS` are kept. A job runs on a cron expression under the
  name of whoever scheduled it.
- **Settings.** Admins manage the instance SSH keys, the secrets added to
  every run, encrypted at rest with those keys, the env files the
  pipelines read, and the folders each user sees.

The look follows [oida](https://github.com/titpetric/oida): hairlines,
one amber accent, dark and light from `prefers-color-scheme`. The type is
a dashboard pair, [Inter](https://rsms.me/inter/) for the words and
[JetBrains Mono](https://www.jetbrains.com/lp/mono/) for the data, both
vendored under `frontend/view/assets/fonts/` with their licenses. All
assets are embedded, the front end makes no external requests.

# Running

Start the Docker image with `task docker:run`, or use the compose setup:

```yaml
services:
  runner:
    image: titpetric/task-ui
    restart: always
    environment:
      TASK_UI_SECRET: change-me
    ports:
    - 3000:3000
    volumes:
    - $PWD/app:/app
    - $PWD/data:/data
    - $PWD/root/.ssh:/root/.ssh
    - /var/run/docker.sock:/var/run/docker.sock:ro
```

Mount your folders at `/app`. Each subfolder with something to run
becomes a folder in the UI, and `/app` itself is one too when it has.
The database, the recordings and the artifacts live under `/data`, and
the users sign in from `/app/.htpasswd`. The image bundles `task`,
`atkins`, `docker` and `docker compose`.

- If you don't want to use Docker, remove the volume for `docker.sock`.
- Mount `/root/.ssh` to keep the SSH keys managed from the UI.

## Configuration

Everything is configured from the environment:

| Variable              | Default              | Meaning                                                           |
|-----------------------|----------------------|-------------------------------------------------------------------|
| `TASK_UI_LISTEN`      | `:3000`              | Listen address                                                    |
| `TASK_UI_ROOT`        | `.`                  | Root folder holding the folders                                   |
| `TASK_UI_DISCOVER`    | `true`               | List every folder under the root with something to run            |
| `TASK_UI_HTPASSWD`    | `.htpasswd`          | The password file, looked up in the root                          |
| `TASK_UI_ADMINS`      |                      | Users with the admin role, comma separated                        |
| `TASK_UI_VIEWERS`     |                      | Users with the viewer role, comma separated; the rest are runners |
| `TASK_UI_DATA`        | `data`               | Database, recordings and artifacts                                |
| `TASK_UI_DB`          | `<data>/task-ui.db`  | Database file (sqlite)                                            |
| `TASK_UI_SSH_DIR`     | `$HOME/.ssh`         | Folder the SSH key settings manage                                |
| `TASK_UI_SECRET`      | random per process   | Signs the session cookie, set it to keep sessions over restarts   |
| `TASK_UI_SESSION_TTL` | `12h`                | Session lifetime                                                  |
| `TASK_UI_KEEP`        | `15m`                | How long a finished run stays attachable with its frame in memory |
| `TASK_UI_GRACE`       | `10s`                | How long a stopped process gets before it is killed               |
| `TASK_UI_AUTOLOGIN`   |                      | Sign every request in as this user, for development only          |
| `TASK_UI_NOW`         |                      | Pin the clock the hours read (RFC 3339), for the screenshots only |

The flags `-l`, `--root`, `--htpasswd` and `--data` override the
matching variables; `-v` prints the build.

## Users and roles

Login is mandatory, through the sign in screen; there is no HTTP basic
auth. Who may sign in is the password file, `.htpasswd` in the root or
wherever `TASK_UI_HTPASSWD` points: one `name:hash` per line, the
htpasswd format, bcrypt hashes only. The file is the truth. It is read
on startup and again whenever it changes, so adding, resetting or
removing a user needs no restart: a user not in the file cannot sign in
and is removed from the instance; one in the file signs in with the
password it holds. A file that is not there yet is a warning, not an
error: the instance starts and reads it once `task-ui passwd` makes it.

Adding a user, or resetting a password, is one line either way:

```sh
task-ui passwd alice            # asks nothing, reads the password from stdin
task-ui passwd alice s3cret     # or takes it as an argument
htpasswd -B .htpasswd alice     # the same with apache2-utils; -B for bcrypt
```

`task-ui passwd` writes the file `TASK_UI_HTPASSWD` names (or
`.htpasswd` in `TASK_UI_ROOT`), adding the line or replacing it. To
remove a user, delete the line. `task-ui hash <password>` prints a hash
for editing the file by hand.

What a user signs in with is kept apart from the user row, so nothing
that reads a user, the API included, carries a credential:
`user_auth_password` holds the password hashes and `user_email` the
email addresses. Sign in takes the user name or any of the user's
addresses, and the password or a backup password:

```sh
task-ui email alice                          # the addresses alice signs in with
task-ui email alice add alice@example.com    # add one or more, `rm` removes
task-ui backup alice                         # ten new backup passwords, printed once
```

A backup password signs in once and is then used up; `task-ui backup`
replaces the previous set. Both commands work on the database of a user
the server has read from the password file.

The roles are labels from the environment: `TASK_UI_ADMINS=alice,bob`
makes those two admins, `TASK_UI_VIEWERS=support` a viewer, and everyone
else in the file is a runner. Which folders a user sees is set per user
under Settings, Users; empty shows them all.

| Role     | See jobs, watch runs, read history | Run, Run force, Stop, Kill, schedule, workflows | Settings |
|----------|------------------------------------|-----------------------------------------------|----------|
| `admin`  | yes                                | everything                                    | yes      |
| `runner` | yes                                | what the allow lists permit                   | no       |
| `viewer` | yes                                | no                                            | no       |


## Conventions

A folder may follow several conventions at once; each one is a source
with its own table on the folder page. Nothing task-ui needs is an
extension of any file, and no pipeline file is ever written.

| Source | File | Runs with | Targets |
|--------|------|-----------|---------|
| atkins | `atkins.yml`, `.atkins.yml` | `atkins -f <file> <job>` | jobs with a `desc`; `show: false` hides one |
| task | `Taskfile.yml` | `task -t <file> <task>` | tasks with a `desc`; `internal: true` hides one |
| make | `GNUmakefile`, `makefile`, `Makefile` | `make <target>` | plain targets; `target: ## text` or a comment above describes one |
| npm | `package.json` | `npm run <script>`, or `pnpm`, `yarn`, `bun` after the lock file | the scripts, hooks folded into theirs |
| composer | `composer.json` | `composer run-script <script>` | the scripts, with `scripts-descriptions` |
| script | `scripts/*` | `./scripts/<file>`, or `sh` for a `.sh` that is not executable | one per file; the first comment describes it |

A pipeline file stands for its folder: a folder with `atkins.yml` or a
Taskfile lists that pipeline's targets only, and the Makefile, package
scripts or `scripts/` beside it are the pipeline's business. Without a
pipeline file every convention is a source. A job's ID is its target
name, or `kind:name` when the folder lists more than one convention, so
`make build` and `npm run build` stay apart.

Discovery goes as deep as the folders do: every directory under the root
with something to run is a folder, at its path (`services/api`), and the
sidebar shows them as a tree, each folder indented under its parent with
its targets below it. Under a folder with a pipeline file only folders
with a pipeline file of their own are listed, so a Go tree's Makefiles
do not become folders; under a plain folder they do. Hidden directories,
`node_modules`, `vendor` and `testdata` are never looked into.

Pipeline sources also give the folder its name and description (atkins
`name` and `help`), the files under `include`, and the dotenv files under
`env.include` or `dotenv`, which task-ui manages (below).

A convention is a `Detector` in the `detector` package: one type that
reads a folder and returns a `Source`, added to the `Detectors` list. The
rest of task-ui runs on sources and jobs and never asks which kind they
are.

Settings, Projects, Configure sets what the instance adds to a folder: a
title and description, who may run it (`allow`: user names or roles,
empty means every runner), and which `sources` to read, which is also
how a pipeline folder gets its Makefile or scripts back as jobs. Through the API a
project also takes `commands` (the runner per kind, `make: gmake`) and
`job_allow` (per job). Pipeline files stay as they are.

Jobs get `TASK_UI=1`, `TASK_UI_USER`, `TASK_UI_FOLDER`, `TASK_UI_JOB`,
`TASK_UI_RUN_ID` and `TASK_UI_ARTIFACTS` in the environment, along with
the secrets.

## Runs, windows and the dashboard

A run is a session: the process on a pty, a terminal emulator following
its output so the screen is known at every moment, the windows attached,
and how it ended. The window that started it is one of many. Closing a
window does nothing to the run; the sidebar, the dashboard and the run
list open a window on it again. A window that was away gets the tail of
what it missed.

A viewer's window is read only. A runner's window takes input, resizes
the terminal, and ends the run: **Stop** sends the process group SIGHUP
and SIGTERM, then SIGKILL after `TASK_UI_GRACE`, and records the run as
`stopped`; **Kill** does the same and records `killed`, as does **Run
force** when it takes a job over. A finished session stays attachable for
`TASK_UI_KEEP`, with its last frame; after that the row and the recording
remain.

The dashboard shows the frame of every live session, refreshed every two
seconds. `/api/runs/{id}/frame.svg` is the frame as an image, with the
terminal colors, `/api/runs/{id}/frame.txt` as text, and
`/api/runs/{id}/cast` is the recording, an asciicast v2 file. The runs
page lists every run with who started it, how many windows are on it and
how it ended, filtered by folder, job and state.

## Hours

A folder can be limited to office hours: one rule per line in Settings,
Projects, Configure, days then a span. The last rule that covers a day
decides it, so

```
mon-fri 9-17
fri 9-13
```

runs Monday to Thursday nine to five and Friday mornings, and nothing on
the weekend. Days are names or ranges (`mon-fri`, `sat,sun`, `weekdays`,
`daily`), spans are hours or `hh:mm`. The API also sets hours per job
(`job_hours`). The folder and job pages say whether it is open now.

Out of hours a run is refused with "out of hours". Run force is still
there, but it asks why: the reason is required, kept on the run, and
shown wherever the run is listed, as the log entry of who ran what out
of hours and why. Scheduled runs are not gated; the schedule is the
decision.

## Projects from git

Settings, Projects lists the folders under the root with their source.
Add project clones a repository into the root: it either succeeds and the
folder appears in the sidebar, or git's output is shown and nothing is
kept. Private repositories are reached with the instance SSH keys, which
git uses through `GIT_SSH_COMMAND` with a `known_hosts` file next to them.

A folder in a checkout is pulled with `git pull --ff-only` before every
run, in the run's own terminal, so the job always runs the latest
version of the project; a folder inside a checkout pulls the checkout.
A pull that fails fails the run.

A checkout with a `.ssh/` folder at its top has its keys loaded for its
runs: every unencrypted private key in it goes into an ssh agent that
task-ui serves for the run alone, and the job gets `SSH_AUTH_SOCK`, so
`ssh`, `git`, `rsync` and terraform's agent authentication use them. The
keys stay in memory and the agent ends with the run; the terminal says
which keys were loaded.

## Artifacts

A job that writes files to the directory named in `$TASK_UI_ARTIFACTS`
leaves artifacts: when the run ends they are recorded with their size and
type, listed on the run and under Artifacts, and served for download.
Admins delete them.

## Scheduled

A job runs on a schedule: a five field cron expression (`0 3 * * *`,
`*/15 * * * 1-5`, names and `@hourly`, `@daily`, `@weekly`, `@monthly`
work too), in the server's time zone. The run starts in the background
under the name of whoever scheduled it, with their permissions, and shows
up like any other. A job that is still running when its time comes is
skipped; a schedule whose job is gone is disabled.

## Secrets and env files

Settings hold the secrets: variables added to the environment of every run,
for all folders or for one folder, where the folder's value wins. They are
never shown again in full, and they are encrypted at rest with the
instance SSH keys: every ed25519 and RSA key under Settings, SSH keys
encrypts (as an [age](https://age-encryption.org) recipient), and the
private key files decrypt. Without such a key no secret can be added.
A rotation keeps the old pair as a backup, so everything still opens, and
writes the secrets again for the new key; Re-encrypt on the secrets page
does the same by hand, after which the backup can go. The last key that
opens the secrets cannot be deleted until another is there. ECDSA keys
reach hosts but do not encrypt.

A folder page lists the dotenv files its sources declare. An admin edits
the content from the UI; the file in the folder is generated from it with
the secrets appended, when it is saved and again before every run. A file
that was never edited is read from the folder as it is. The pipeline file
is never changed.

## API

Everything the pages do goes through `/api`, JSON over the session cookie.
Writes need the same origin. The browser pages are the reference client;
`scripts/run.sh` is the curl one.

| Method | Path | Does |
|--------|------|------|
| `GET` | `/api/runs?folder=&job=&state=&limit=` | the runs, newest first, with `live`, `attached` and `can_run` |
| `POST` | `/api/runs` `{folder, job, force, reason, cols, rows}` | starts a run; `409` with the run in the way unless forced; `403` out of hours unless forced with a reason |
| `GET` | `/api/runs/{id}` | one run |
| `POST` | `/api/runs/{id}/stop`, `/kill` | ends a run |
| `GET` | `/api/runs/{id}/frame.svg`, `/frame.txt`, `/cast` | the last frame, the recording |
| `WS` | `/api/runs/{id}/attach` | binary frames are output; text frames are `{"type":"input"\|"resize"\|"stop"\|"kill"}` in and `{"type":"status"}` out |
| `GET` `DELETE` | `/api/artifacts`, `/api/artifacts/{id}` | list, download, delete |
| `GET` `POST` `PUT` `DELETE` | `/api/schedules`, `/api/schedules/{id}` | `{folder, job, expression, enabled}` |
| `GET` `POST` `PUT` `DELETE` | `/api/workflows`, `/api/workflows/{id}` | `{name, steps: [{folder, job} or {workflow}]}`; owner or admin changes; 409 deleting one another workflow runs |
| `GET` `POST` | `/api/workflows/{id}/runs` | runs, newest first, with their steps; `POST` starts one, 409 while one runs |
| `POST` | `/api/workflow-runs/{id}/stop` | stops the step running, skips the rest |
| `GET` `POST` `PUT` `DELETE` | `/api/projects`, `/api/projects/{folder}` | clone `{url, id, branch}`; configure `{title, description, allow, sources, commands, job_allow, hours, job_hours}` |
| `GET` `PUT` | `/api/users`, `/api/users/{id}` | `{folders}`; names, passwords and roles come from the password file and the role lists |
| `GET` `POST` `DELETE` | `/api/secrets`, `/api/secrets/{id}` | `{scope, name, content}`; content is never returned; `409` without a key to encrypt with |
| `POST` | `/api/secrets/rekey` | writes every secret again for the current keys |
| `GET` `POST` `DELETE` | `/api/keys`, `/api/keys/generate`, `/api/keys/upload`, `/api/keys/{name}` | `{type, comment}`, `{content, comment}`; a new key rekeys the secrets, the last key that opens them cannot go |
| `GET` `PUT` | `/api/env/{folder}`, `/api/env/{folder}/{path}` | the env files; `{content}` saves and generates |

## Screenshots

| | |
|-|-|
| ![The dashboard: every live terminal as its last frame](.github/assets/dashboard-dark.png) | ![A run in its window, attached](.github/assets/run-dark.png) |
| ![A folder: its jobs, steps, pipeline and env files](.github/assets/folder-dark.png) | ![A job, its history and its schedules](.github/assets/job-dark.png) |
| ![The runs](.github/assets/runs-dark.png) | ![A run's page](.github/assets/run-page-dark.png) |
| ![Dashboard, light theme](.github/assets/dashboard-light.png) | ![A folder with a Makefile, package scripts and a scripts folder](.github/assets/tools-dark.png) |
| ![Artifacts](.github/assets/artifacts-dark.png) | ![Scheduled](.github/assets/schedules-dark.png) |
| ![Projects](.github/assets/projects-dark.png) | ![Configure a folder](.github/assets/project-edit-dark.png) |
| ![Users, from the password file](.github/assets/users-dark.png) | ![Secrets, encrypted with the SSH key](.github/assets/secrets-dark.png) |
| ![Secrets](.github/assets/secrets-dark.png) | ![An env file](.github/assets/envfile-dark.png) |
| ![SSH keys](.github/assets/keys-dark.png) | ![Login](.github/assets/login-dark.png) |
| ![Out of hours: Run force asks why](.github/assets/out-of-hours-dark.png) | ![The phone: the drawer with the folders and their jobs, and the jobs as cards](.github/assets/mobile-menu-dark.png) |

`scripts/demo.sh` regenerates them: it serves a copy of `testdata/demo`,
sets the instance up through the API with curl, starts real runs with
`scripts/run.sh`, walks every link of every page with `scripts/crawl.sh`
and captures each view with headless chromium through
`scripts/chromium.sh`, the phone views included. A run opens in its
window from a link: `/runs/<id>#run=<id>`; `#start=<folder>/<job>`
starts one; `#menu` opens the drawer.

Under 860px the sidebar is a drawer behind the menu button, with the
folders opening into their jobs, and the tables of jobs and runs fold
into cards.

## Layout

Packages are the bounded contexts, one per domain, singular. The
dependency runs one way: `app` to everything, `api` and `frontend` to the
domains, the domains to `model` and `storage`, `model` to nothing of ours.

| Package     | Holds                                                                        |
|-------------|------------------------------------------------------------------------------|
| `schema`    | the SQL migrations, the source of truth of the data model (mig)              |
| `model`     | the mig generated types and the rules on them                                |
| `storage`   | CRUD per table over pdo, one in-memory database per test                     |
| `detector`  | the conventions: a `Detector` per file kind, returning a `Source` of `Target`s |
| `folder`    | the catalogue: a `Folder` with its `Job`s, discovery under the root, the project configuration |
| `runner`    | sessions: a `Process` on a pty, the terminal `Frame`, subscribers, artifacts |
| `scheduler` | cron `Expression`s and the `Loop` that starts due schedules                  |
| `hours`     | the office hours `Rules` a folder or job may run in                          |
| `recording` | the asciicast files                                                          |
| `access`    | who may do what: the password file, `Sessions`, the `Guard`                  |
| `secret`    | the secrets, encrypted at rest with the SSH keys through the `Vault`         |
| `envfile`   | the dotenv files a folder reads, as task-ui manages and generates them       |
| `sshkey`, `git`, `dotenv` | the key pairs, the checkouts, the file format                 |
| `api`       | the JSON endpoints and the attach socket                                     |
| `frontend`  | the pages; `frontend/view` renders them from embedded templates and assets   |
| `app`       | the composition root: configuration, wiring, serving                         |

`splint ./...` (`task lint`) is the bar: every file has a test beside it,
every exported symbol a test that reaches it, every symbol a file named
for it, and every dependency one import per package. Tests are black box,
parallel, and share nothing. `scripts/smoke.sh` (`task smoke`) then
exercises a fresh instance end to end with curl and headless chromium:
runs, hours, secrets and keys, env files, schedules, users, projects,
artifacts, every page, and the browser's console.

# Development

task: Available tasks for this project:

* default:            Run everything
* fix:                Fix code
* install:            Install task-ui
* lint:               Lint with splint, zero findings is the bar
* run:                Run task-ui over the demo projects
* smoke:              Exercise a fresh instance end to end, the browser console included
* test:               Test task-ui
* docker:build:       Build task-ui docker image
* docker:gen:         Generate ssh key for docker image
* docker:run:         Run task-ui in docker env
* docs:shots:         Regenerate the README screenshots

## task: default

Run everything

commands:
- `Task: fix`
- `Task: lint`
- `Task: install`
- `Task: test`

## task: test

Test task-ui

commands:
- `CGO_ENABLED=1 go test -race -count=1 -cover ./...`
- `CGO_ENABLED=0 go test -count=1 -cover ./...`

## task: run

Run task-ui over the demo projects

commands:
- `go run .`

## task: docs:shots

Regenerate the README screenshots

commands:
- `scripts/demo.sh`

## task: smoke

Exercise a fresh instance end to end, the browser console included

commands:
- `scripts/smoke.sh`

## task: lint

Lint with splint, zero findings is the bar

commands:
- `splint ./...`

## task: install

Install task-ui

commands:
- `CGO_ENABLED=0 go install .`

## task: fix

Fix code

dependencies:
- `deps:splint`

commands:
- `splint fix ./...`
- `go vet ./...`
- `go mod tidy`
- `./README.md.sh > README.md`

