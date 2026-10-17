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
  themselves: names, descriptions, steps, the files they include.
- **Runs are sessions.** Run opens a window on the run. Closing the window
  leaves the run running; Stop ends it, Kill ends it now. Anyone who may
  see the job attaches from the sidebar, the dashboard or the run list;
  runners type into it, viewers watch.
- **The dashboard.** Every terminal in progress as its last frame, for
  everyone, the way a control room shows its screens. Each frame is one
  URL, as SVG or text.
- **One run per job.** A job runs once at a time across the instance,
  different jobs run side by side. Run force kills the active run and
  starts again. Run all starts every job of a folder in the background.
- **Users and roles.** Login is mandatory, with three roles: `admin`,
  `runner` and `viewer`. Users are seeded from `task-ui.yml` and managed
  in Settings.
- **Projects from git.** Admins add a folder by cloning a repository; the
  clone is the test. A checkout is pulled before every run.
- **History, artifacts, schedules.** Every run is recorded with who ran it,
  how it ended and the full terminal output. Files a job leaves under
  `$TASK_UI_ARTIFACTS` are kept. A job runs on a cron expression under the
  name of whoever scheduled it.
- **Settings.** Admins manage the instance SSH keys, the secrets added to
  every run, the env files the pipelines read, and the users.

The look follows [oida](https://github.com/titpetric/oida): monospace,
hairlines, one amber accent, dark and light from `prefers-color-scheme`. All
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
The database, the recordings and the artifacts live under `/data`. The
image bundles `task`, `atkins`, `docker` and `docker compose`.

- If you don't want to use Docker, remove the volume for `docker.sock`.
- Mount `/root/.ssh` to keep the SSH keys managed from the UI.

## Configuration

Everything is configured from the environment:

| Variable              | Default              | Meaning                                                           |
|-----------------------|----------------------|-------------------------------------------------------------------|
| `TASK_UI_LISTEN`      | `:3000`              | Listen address                                                    |
| `TASK_UI_ROOT`        | `.`                  | Root folder holding the folders                                   |
| `TASK_UI_DISCOVER`    | `true`               | List every folder under the root with something to run            |
| `TASK_UI_CONFIG`      | `task-ui.yml`        | Users to seed, looked up in the root                              |
| `TASK_UI_DATA`        | `data`               | Database, recordings and artifacts                                |
| `TASK_UI_DB`          | `<data>/task-ui.db`  | Database file (sqlite)                                            |
| `TASK_UI_SSH_DIR`     | `$HOME/.ssh`         | Folder the SSH key settings manage                                |
| `TASK_UI_SECRET`      | random per process   | Signs the session cookie, set it to keep sessions over restarts   |
| `TASK_UI_SESSION_TTL` | `12h`                | Session lifetime                                                  |
| `TASK_UI_KEEP`        | `15m`                | How long a finished run stays attachable with its frame in memory |
| `TASK_UI_GRACE`       | `10s`                | How long a stopped process gets before it is killed               |
| `TASK_UI_AUTOLOGIN`   |                      | Sign every request in as this user, for development only          |

The flags `-l`, `--root`, `--config` and `--data` override the matching
variables; `-v` prints the build.

## Users and roles

Login is mandatory. On startup the users in `task-ui.yml` are written to
the database: a user that is not there is created, one that is gets the
role, folders and password of the file. From then on Settings manages them.
With no users nobody can sign in.

```yaml
users:
  - name: admin
    password: $2a$10$...   # task-ui hash <password>, or plain text
    role: admin
  - name: ops
    password: $2a$10$...
    role: runner
    folders: [infra]       # optional, limits the visible folders
  - name: support
    password: $2a$10$...
    role: viewer
```

| Role     | See jobs, watch runs, read history | Run, Run force, Run all, Stop, Kill, schedule | Settings |
|----------|------------------------------------|-----------------------------------------------|----------|
| `admin`  | yes                                | everything                                    | yes      |
| `runner` | yes                                | what the allow lists permit                   | no       |
| `viewer` | yes                                | no                                            | no       |

Generate a password hash with `task-ui hash <password>`, or
`echo secret | task-ui hash`.

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

Pipeline sources also give the folder its name and description (atkins
`name` and `help`), the files under `include`, and the dotenv files under
`env.include` or `dotenv`, which task-ui manages (below). A job's ID is
its target name, or `kind:name` when the folder follows more than one
convention, so `make build` and `npm run build` stay apart.

A convention is a `Detector` in the `detector` package: one type that
reads a folder and returns a `Source`, added to the `Detectors` list. The
rest of task-ui runs on sources and jobs and never asks which kind they
are.

Settings, Projects, Configure sets what the instance adds to a folder: a
title and description, who may run it (`allow`: user names or roles,
empty means every runner), and which `sources` to read. Through the API a
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

## Projects from git

Settings, Projects lists the folders under the root with their source.
Add project clones a repository into the root: it either succeeds and the
folder appears in the sidebar, or git's output is shown and nothing is
kept. Private repositories are reached with the instance SSH keys, which
git uses through `GIT_SSH_COMMAND` with a `known_hosts` file next to them.

A folder that is a checkout is pulled with `git pull --ff-only` before
every run, in the run's own terminal, so the job always runs the latest
version of the project. A pull that fails fails the run.

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
kept in the database and never shown again in full.

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
| `POST` | `/api/runs` `{folder, job, force, cols, rows}` | starts a run; `409` with the run in the way unless forced |
| `POST` | `/api/runs/all` `{folder, force}` | starts every job of a folder the user may run |
| `GET` | `/api/runs/{id}` | one run |
| `POST` | `/api/runs/{id}/stop`, `/kill` | ends a run |
| `GET` | `/api/runs/{id}/frame.svg`, `/frame.txt`, `/cast` | the last frame, the recording |
| `WS` | `/api/runs/{id}/attach` | binary frames are output; text frames are `{"type":"input"\|"resize"\|"stop"\|"kill"}` in and `{"type":"status"}` out |
| `GET` `DELETE` | `/api/artifacts`, `/api/artifacts/{id}` | list, download, delete |
| `GET` `POST` `PUT` `DELETE` | `/api/schedules`, `/api/schedules/{id}` | `{folder, job, expression, enabled}` |
| `GET` `POST` `PUT` `DELETE` | `/api/projects`, `/api/projects/{folder}` | clone `{url, id, branch}`; configure `{title, description, allow, sources, commands, job_allow}` |
| `GET` `POST` `PUT` `DELETE` | `/api/users`, `/api/users/{id}` | `{name, password, role, folders}` |
| `GET` `POST` `DELETE` | `/api/secrets`, `/api/secrets/{id}` | `{scope, name, content}`; content is never returned |
| `GET` `POST` `DELETE` | `/api/keys`, `/api/keys/generate`, `/api/keys/upload`, `/api/keys/{name}` | `{type, comment}`, `{content, comment}` |
| `GET` `PUT` | `/api/env/{folder}`, `/api/env/{folder}/{path}` | the env files; `{content}` saves and generates |

## Screenshots

| | |
|-|-|
| ![The dashboard: every live terminal as its last frame](.github/assets/dashboard-dark.png) | ![A run in its window, attached](.github/assets/run-dark.png) |
| ![A folder: its jobs, steps, pipeline and env files](.github/assets/folder-dark.png) | ![A job, its history and its schedules](.github/assets/job-dark.png) |
| ![The runs](.github/assets/runs-dark.png) | ![A run's page](.github/assets/run-page-dark.png) |
| ![Dashboard, light theme](.github/assets/dashboard-light.png) | ![A folder with a Makefile, package scripts and a scripts folder](.github/assets/tools-dark.png) |
| ![Artifacts](.github/assets/artifacts-dark.png) | ![Scheduled](.github/assets/schedules-dark.png) |
| ![Projects](.github/assets/projects-dark.png) | ![Users](.github/assets/users-dark.png) |
| ![Secrets](.github/assets/secrets-dark.png) | ![An env file](.github/assets/envfile-dark.png) |
| ![SSH keys](.github/assets/keys-dark.png) | ![Login](.github/assets/login-dark.png) |

`scripts/demo.sh` regenerates them: it serves a copy of `testdata/demo`,
sets the instance up through the API with curl, starts real runs with
`scripts/run.sh`, and captures each view with headless chromium through
`scripts/chromium.sh`. A run opens in its window from a link:
`/runs/<id>#run=<id>`.

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
| `recording` | the asciicast files                                                          |
| `access`    | who may do what: passwords, `Sessions`, the `Guard`                          |
| `envfile`   | the dotenv files a folder reads, as task-ui manages and generates them       |
| `sshkey`, `git`, `dotenv` | the key pairs, the checkouts, the file format                 |
| `api`       | the JSON endpoints and the attach socket                                     |
| `frontend`  | the pages; `frontend/view` renders them from embedded templates and assets   |
| `app`       | the composition root: configuration, wiring, serving                         |

`splint ./...` (`task lint`) is the bar: every file has a test beside it,
every exported symbol a test that reaches it, every symbol a file named
for it, and every dependency one import per package. Tests are black box,
parallel, and share nothing.

# Development

task: Available tasks for this project:

* default:            Run everything
* fix:                Fix code
* install:            Install task-ui
* lint:               Lint with splint, zero findings is the bar
* run:                Run task-ui over the demo projects
* test:               Test task-ui
* docker:build:       Build task-ui docker image
* docker:gen:         Generate ssh key for docker image
* docker:run:         Run task-ui in docker env
* docs:shots:         Regenerate the README screenshots

## task: default

Run everything

vars:
  image: "titpetric/task-ui"
  root: sh: git rev-parse --show-toplevel || echo .

commands:
- `Task: fix`
- `Task: lint`
- `Task: install`
- `Task: test`

## task: test

Test task-ui

vars:
  image: "titpetric/task-ui"
  root: sh: git rev-parse --show-toplevel || echo .

commands:
- `CGO_ENABLED=1 go test -race -count=1 -cover ./...`
- `CGO_ENABLED=0 go test -count=1 -cover ./...`

## task: run

Run task-ui over the demo projects

vars:
  image: "titpetric/task-ui"
  root: sh: git rev-parse --show-toplevel || echo .

commands:
- `go run .`

## task: docs:shots

Regenerate the README screenshots

vars:
  image: "titpetric/task-ui"
  root: sh: git rev-parse --show-toplevel || echo .

commands:
- `scripts/demo.sh`

## task: lint

Lint with splint, zero findings is the bar

vars:
  image: "titpetric/task-ui"
  root: sh: git rev-parse --show-toplevel || echo .

commands:
- `splint ./...`

## task: install

Install task-ui

vars:
  image: "titpetric/task-ui"
  root: sh: git rev-parse --show-toplevel || echo .

commands:
- `CGO_ENABLED=0 go install .`

## task: fix

Fix code

vars:
  image: "titpetric/task-ui"
  root: sh: git rev-parse --show-toplevel || echo .

dependencies:
- `deps:splint`

commands:
- `splint fix ./...`
- `go vet ./...`
- `go mod tidy`
- `./README.md.sh > README.md`

