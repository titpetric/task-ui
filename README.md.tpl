# Task UI

Run your pipelines from the browser. Task UI reads the `atkins.yml` and
`Taskfile.yml` files under a root folder, lists each folder with the jobs
its pipeline declares, and runs them in a terminal window over a websocket.

![The overview: folders on the left, active runs and recent history on the right](https://raw.githubusercontent.com/titpetric/task-ui/main/.github/assets/task-ui.png)

- **Folders and jobs.** A folder is a working directory with a pipeline
  file. Everything on the page comes from that file: the pipeline `name`
  and `help`, each job's `desc` and steps, the files it includes.
- **Run in a window.** Run opens a modal terminal. Closing the window stops
  the run.
- **One run per job.** A job runs once at a time across the instance,
  different jobs run side by side. Run force stops the active run and starts
  again.
- **Users and roles.** A login screen, and three roles: `admin`, `runner` and
  `viewer`. Viewers read jobs and history and run nothing.
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

Mount your folders at `/app`. Each subfolder with an `atkins.yml` or
`Taskfile.yml` becomes a folder in the UI, and a pipeline file in `/app`
itself is one too. The image bundles `task`, `atkins`, `docker` and
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
| `TASK_UI_DISCOVER`     | `true`             | List every folder under the root with a pipeline file              |
| `TASK_UI_CONFIG`       | `task-ui.yml`      | Users and folder permissions, looked up in the root                |
| `TASK_UI_HISTORY`      | `true`             | Record runs                                                        |
| `TASK_UI_HISTORY_PATH` | `history`          | Recordings folder                                                  |
| `TASK_UI_SSH_DIR`      | `$HOME/.ssh`       | Folder the SSH key settings manage                                 |
| `TASK_UI_ENV_PATH`     | `env`              | Secrets and managed env files                                      |
| `TASK_UI_SECRET`       | random per process | Signs the session cookie, set it to keep sessions over restarts    |
| `TASK_UI_SESSION_TTL`  | `12h`              | Session lifetime                                                   |
| `TASK_UI_ANONYMOUS`    | `admin`            | Role of visitors when no users exist: admin, runner, viewer, none  |
| `TASK_UI_ASSETS`       |                    | Serve templates and static files from this folder, for development |

The flags `-l`, `-f`, `--root`, `--config`, `--history-enable` and
`--history-output` override the matching variables.

## Users and roles

Users live in `task-ui.yml` next to the folders. The file is re-read when
it changes, so adding or removing a user needs no restart. With no users,
every visitor gets the `TASK_UI_ANONYMOUS` role and there is no login.

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

## Pipeline files

Task UI reads atkins pipelines and Taskfiles as they are, and nothing it
needs is an extension: the pipeline `name` and `help`, `include` and
`env.include` (Taskfile `dotenv`), the jobs under `tasks:` or `jobs:` with
`desc`, `summary`, `aliases`, `interactive`, `depends_on` and their steps,
including the short forms `job: command` and `job: [commands]`. Jobs with
a description show up; `internal: true` or `show: false` hides one. Who may
run what is decided in `task-ui.yml`, and the pipeline files are never
written.

```yaml
name: Infrastructure
help: Provisioning and deploys for the production cluster.

env:
  include: .env

jobs:
  deploy:
    desc: "Roll out the current release"
    depends_on: plan
    steps:
      - name: Pull the release image
        run: ./pull.sh
      - run: ./deploy.sh
```

Atkins files run with `atkins -f <file> <job>`, Taskfiles with
`task -t <file> <job>`. Set `command` on a folder in `task-ui.yml` to run
something else. Jobs get `TASK_UI_USER`, `TASK_UI_FOLDER`, `TASK_UI_JOB`
and `TASK_UI_RUN_ID` in the environment.

## Secrets and env files

Settings hold the secrets: variables added to the environment of every run,
for all folders or for one folder, where the folder's value wins. They are
stored under `TASK_UI_ENV_PATH` and never shown again in full.

A folder page lists the env files its pipeline reads with `env.include`.
An admin edits the content from the UI; the file in the folder is generated
from it with the secrets appended, when it is saved and again before every
run. A file that was never edited is read from the folder as it is. The
pipeline file is never changed.

## Screenshots

| | |
|-|-|
| ![A folder: its jobs, steps, pipeline and env files](.github/assets/folder-dark.png) | ![A run replayed in the dialog](.github/assets/run-replay-dark.png) |
| ![A job and its history](.github/assets/job-dark.png) | ![Overview, light theme](.github/assets/overview-light.png) |
| ![Secrets](.github/assets/secrets-dark.png) | ![An env file](.github/assets/envfile-dark.png) |
| ![SSH keys](.github/assets/keys-dark.png) | ![Login](.github/assets/login-dark.png) |

`scripts/demo.sh` regenerates them: it serves a copy of `testdata/demo`,
records real runs over the websocket with `scripts/run.sh`, and captures each
view with headless chromium through `scripts/chromium.sh`. A recorded run
opens in the dialog from a link: `/f/<folder>/<job>#history=<id>`, or
`#replay=<id>`.
