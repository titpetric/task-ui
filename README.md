# Task UI

Run what your source trees declare, from the browser. Task UI looks at
each folder under a root, detects the conventions it follows, and lists
their targets as jobs: atkins pipelines, Taskfiles, Makefile targets,
package.json and composer.json scripts, and a `scripts/` folder. A job
runs in a terminal window over a websocket.

![The overview: folders on the left, active runs and recent history on the right](https://raw.githubusercontent.com/titpetric/task-ui/main/.github/assets/task-ui.png)

- **Folders and jobs.** A folder is a working directory and whatever it
  holds that can be run. Everything on the page comes from the files
  themselves: names, descriptions, steps, the files they include.
- **Run in a window.** Run opens a modal terminal. Closing the window stops
  the run.
- **One run per job.** A job runs once at a time across the instance,
  different jobs run side by side. Run force stops the active run and starts
  again.
- **Users and roles.** Login is mandatory, with three roles: `admin`,
  `runner` and `viewer`. Viewers read jobs and history and run nothing.
- **Projects from git.** Admins add a folder by cloning a repository; the
  clone is the test. A checkout is pulled before every run, so the latest
  version of the project is what runs.
- **History.** Every run is recorded with who ran it, how it ended and the
  full terminal output, with view and replay.
- **Settings.** Admins manage the instance SSH keys, the secrets added to
  every run, and the env files the pipelines read.

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
    - $PWD/root/.ssh:/root/.ssh
    - /var/run/docker.sock:/var/run/docker.sock:ro
```

Mount your folders at `/app`. Each subfolder with something to run
becomes a folder in the UI, and `/app` itself is one too when it has. The image bundles `task`, `atkins`, `docker` and
`docker compose`.

- If you don't want to use Docker, remove the volume for `docker.sock`.
- Mount `/root/.ssh` to keep the SSH keys managed from the UI.

## Configuration

Everything is configured from the environment:

| Variable               | Default            | Meaning                                                            |
|------------------------|--------------------|--------------------------------------------------------------------|
| `TASK_UI_LISTEN`       | `:3000`            | Listen address                                                     |
| `TASK_UI_ROOT`         | `.`                | Root folder holding the pipeline folders                           |
| `TASK_UI_FILE`         |                    | Serve one pipeline file as the only folder                         |
| `TASK_UI_DISCOVER`     | `true`             | List every folder under the root with something to run             |
| `TASK_UI_CONFIG`       | `task-ui.yml`      | Users and folder permissions, looked up in the root                |
| `TASK_UI_HISTORY`      | `true`             | Record runs                                                        |
| `TASK_UI_HISTORY_PATH` | `history`          | Recordings folder                                                  |
| `TASK_UI_SSH_DIR`      | `$HOME/.ssh`       | Folder the SSH key settings manage                                 |
| `TASK_UI_ENV_PATH`     | `env`              | Secrets and managed env files                                      |
| `TASK_UI_SECRET`       | random per process | Signs the session cookie, set it to keep sessions over restarts    |
| `TASK_UI_SESSION_TTL`  | `12h`              | Session lifetime                                                   |
| `TASK_UI_AUTOLOGIN`    |                    | Sign every request in as this user, for development only          |
| `TASK_UI_ASSETS`       |                    | Serve templates and static files from this folder, for development |

The flags `-l`, `-f`, `--root`, `--config`, `--history-enable` and
`--history-output` override the matching variables.

## Users and roles

Login is mandatory. Users live in `task-ui.yml` next to the folders, and
the file is re-read when it changes, so adding or removing a user needs no
restart. With no users nobody can sign in, and the sign in screen says so.

```yaml
users:
  - name: admin
    password: $2a$10$...   # task-ui hash <password>
    role: admin
  - name: ops
    password: $2a$10$...
    role: runner
    projects: [infra]      # optional, limits the visible folders
  - name: support
    password: $2a$10$...
    role: viewer

projects:
  - id: infra
    allow: [ops]           # users or roles that may run this folder's jobs
    jobs:
      rotate-certs: [admin]
```

| Role     | Read jobs and history | Run, Run force, Kill        | Settings |
|----------|-----------------------|-----------------------------|----------|
| `admin`  | yes                   | everything                  | yes      |
| `runner` | yes                   | what the allow lists permit | no       |
| `viewer` | yes                   | no                          | no       |

Generate a password hash with `task-ui hash <password>`, or
`echo secret | task-ui hash`.

## Conventions

A folder may follow several conventions at once; each one is a source
with its own table on the folder page. Nothing task-ui needs is an
extension of any file, and no file is ever written.

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

In `task-ui.yml`, `command` replaces the runner of a folder's only source
(`gmake`), `commands` does so per kind, and `sources` limits which
conventions are read in a folder. Who may run what is decided there too.

```yaml
projects:
  - id: tools
    sources: [make, script]   # ignore package.json here
    commands:
      make: gmake
    jobs:
      "make:release": [admin]
```

Jobs get `TASK_UI_USER`, `TASK_UI_FOLDER`, `TASK_UI_JOB` and
`TASK_UI_RUN_ID` in the environment.

## Projects from git

Settings → Projects lists the folders under the root with their source.
Add project clones a repository into the root: it either succeeds and the
folder appears in the sidebar, or git's output is shown and nothing is
kept. Private repositories are reached with the instance SSH keys, which
git uses through `GIT_SSH_COMMAND` with a `known_hosts` file next to them.

A folder that is a checkout is pulled with `git pull --ff-only` before
every run, in the run's own terminal, so the job always runs the latest
version of the project. A pull that fails fails the run.

## Secrets and env files

Settings hold the secrets: variables added to the environment of every run,
for all folders or for one folder, where the folder's value wins. They are
stored under `TASK_UI_ENV_PATH` and never shown again in full.

A folder page lists the dotenv files its sources declare.
An admin edits the content from the UI; the file in the folder is generated
from it with the secrets appended, when it is saved and again before every
run. A file that was never edited is read from the folder as it is. The
pipeline file is never changed.

## Screenshots

| | |
|-|-|
| ![A folder: its jobs, steps, pipeline and env files](.github/assets/folder-dark.png) | ![A run replayed in the dialog](.github/assets/run-replay-dark.png) |
| ![A folder with a Makefile, package scripts and a scripts folder](.github/assets/tools-dark.png) | ![A job and its history](.github/assets/job-dark.png) |
| ![Overview, light theme](.github/assets/overview-light.png) | ![Projects](.github/assets/projects-dark.png) |
| ![Secrets](.github/assets/secrets-dark.png) | ![An env file](.github/assets/envfile-dark.png) |
| ![SSH keys](.github/assets/keys-dark.png) | ![Login](.github/assets/login-dark.png) |

`scripts/demo.sh` regenerates them: it serves a copy of `testdata/demo`,
records real runs over the websocket with `scripts/run.sh`, and captures each
view with headless chromium through `scripts/chromium.sh`. A recorded run
opens in the dialog from a link: `/f/<folder>/<job>#history=<id>`, or
`#replay=<id>`.

## Layout

Packages are the bounded contexts, one per domain, singular:

| Package    | Holds                                                                  |
|------------|------------------------------------------------------------------------|
| `detector` | the conventions: a `Detector` per file kind, returning a `Source` of `Target`s |
| `folder`   | the catalogue: a `Folder` with its `Job`s, discovery under the root, the configuration overrides |
| `run`      | execution: the `Run` lock registry and the `Process` on a pty         |
| `history`  | the recordings: a `Record` per run, written by a `Recorder`           |
| `access`   | who may do what: `User`, `Role`, passwords, `Sessions`                |
| `secret`   | the variables added to runs, per folder or for all                    |
| `envfile`  | the dotenv files a folder reads, as task-ui manages and generates them |
| `sshkey`   | the key pairs in the SSH folder                                       |
| `git`      | clone, pull and the facts of a checkout                               |
| `dotenv`   | the file format the two above share                                   |
| `server`   | the presentation: routes, pages, the run socket                       |

`splint ./...` (`task lint`) is the bar: every file has a test beside it,
every exported symbol a test that reaches it, every symbol a file named
for it, and every dependency one import per package.

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

