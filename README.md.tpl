# Task UI

Run what your source trees declare, from the browser, and watch it run.
Task UI looks at each folder under a root, detects the conventions it
follows, and lists their targets as jobs: atkins pipelines, Taskfiles,
Makefile targets, package.json and composer.json scripts, and a
`scripts/` folder. A job runs as a session on the server: a process on a
terminal that anyone allowed can attach to, from any window, for as long
as it runs.

![The dashboard: every terminal in progress, as it looks right now](https://raw.githubusercontent.com/titpetric/task-ui/main/docs/assets/task-ui.png)

How to use it is in the [usage guide](docs/README.md), a chapter per
section of the sidebar: the [dashboard](docs/guide/dashboard.md),
[workflows](docs/guide/workflows.md), [runs](docs/guide/runs.md),
[artifacts](docs/guide/artifacts.md), [schedules](docs/guide/scheduled.md),
the [projects and their jobs](docs/guide/projects.md), the
[recents](docs/guide/recents.md), the [settings](docs/guide/settings.md),
and [MCP](docs/guide/mcp.md), for an agent that runs jobs. This README
covers running the server and working on it.

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
      TASK_UI_SSH_DIR: /keys
    ports:
    - 3000:3000
    volumes:
    - $PWD/app:/app
    - $PWD/data:/data
    - $PWD/keys:/keys
    - /var/run/docker.sock:/var/run/docker.sock:ro
```

Mount your folders at `/app`. Each subfolder with something to run
becomes a folder in the UI, and `/app` itself is one too when it has.
The database, the recordings and the artifacts live under `/data`, and
the users sign in from `/app/.htpasswd`. The image bundles `task`,
`atkins`, `docker` and `docker compose`.

- If you don't want to use Docker, remove the volume for `docker.sock`.
- Mount `/keys` with `TASK_UI_SSH_DIR` to keep the SSH keys managed from
  the UI. Keep it out of `/root/.ssh`, so `ssh` in a job does not find
  them by default; a run gets the keys its policies give it.

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
and is removed from the instance. A file that is not there yet is a
warning, not an error: the instance reads it once `task-ui passwd` makes
it.

```sh
task-ui passwd alice                         # add or reset; the password from stdin
task-ui passwd alice s3cret                  # or as an argument
htpasswd -B .htpasswd alice                  # the same with apache2-utils; -B for bcrypt
task-ui hash s3cret                          # a hash, for editing the file by hand
task-ui email alice add alice@example.com    # an address alice signs in with; `rm` removes
task-ui backup alice                         # ten backup passwords, each signs in once
task-ui token alice cli                      # a token for a client such as an MCP agent
```

What a user signs in with is kept apart from the user row, so nothing
that reads a user, the API included, carries a credential:
`user_auth_password` holds the password hashes, `user_email` the
addresses and `user_auth_token` the token hashes. The `email`, `backup`
and `token` commands work on the database of a user the server has read
from the password file. To remove a user, delete the line.

`TASK_UI_ADMINS=alice,bob` makes those two admins, `TASK_UI_VIEWERS=support`
a viewer, and everyone else in the file is a runner; what each role may
do is in [the guide](docs/guide/projects.md#who-may-run-what).

## API

Everything the pages do goes through `/api`, JSON over the session cookie,
or over a token in an `Authorization: Bearer` header (`task-ui token
<user> <name>` prints one). Writes need the same origin. The browser
pages are the reference client; `scripts/run.sh` is the curl one.

The surface below, and the MCP tools on it, is frozen: a route or a tool
is added only for a need these cannot serve, with the reason in the
change.

| Method | Path | Does |
|--------|------|------|
| `GET` | `/api/folders?folder=` | the folders you can see and their jobs, with `can_run` and `last_run` |
| `GET` | `/api/runs?folder=&job=&state=&limit=` | the runs, newest first, with `live`, `attached` and `can_run` |
| `POST` | `/api/runs` `{folder, job, force, reason, cols, rows}` | starts a run; `409` with the run in the way unless forced; `403` out of hours unless forced with a reason |
| `GET` | `/api/runs/{id}` | one run |
| `POST` | `/api/runs/{id}/stop`, `/kill` | ends a run |
| `GET` | `/api/runs/{id}/frame.svg`, `/frame.txt`, `/cast` | the last frame, the recording |
| `WS` | `/api/runs/{id}/attach` | binary frames are output; text frames are `{"type":"input"\|"resize"\|"stop"\|"kill"}` in and `{"type":"status"}` out |
| `GET` `DELETE` | `/api/artifacts`, `/api/artifacts/{id}` | list, download, delete |
| `GET` `POST` `PUT` `DELETE` | `/api/schedules`, `/api/schedules/{id}` | `{folder, job, expression, enabled}` |
| `GET` `POST` `PUT` `DELETE` | `/api/workflows`, `/api/workflows/{id}` | `{name, steps: [{folder, job}]}`; owner or admin changes |
| `GET` `POST` | `/api/workflows/{id}/runs` | runs, newest first, with their steps; `POST` starts one, 409 while one runs |
| `POST` | `/api/workflow-runs/{id}/stop` | stops the step running, skips the rest |
| `GET` `POST` `PUT` `DELETE` | `/api/projects`, `/api/projects/{folder}` | clone `{url, id, branch}`; configure `{title, description, allow, sources, commands, job_allow, hours, job_hours}` |
| `GET` `PUT` | `/api/users`, `/api/users/{id}` | `{folders}`; names, passwords and roles come from the password file and the role lists |
| `GET` `POST` `DELETE` | `/api/secrets`, `/api/secrets/{id}` | `{scope, name, content}`; content is never returned; `409` without a key to encrypt with |
| `POST` | `/api/secrets/rekey` | writes every secret again for the current keys |
| `GET` `POST` `PUT` `DELETE` | `/api/keys`, `/api/keys/generate`, `/api/keys/upload`, `/api/keys/{name}` | `{type, name, comment, policies, users}`, `{content, name, comment, policies, users}`; `name` makes `id_<type>_<name>` for someone; `PUT` sets a key's `{policies, users}`, the roles and users whose runs get it; a new key rekeys the secrets, the last key that opens them cannot go |
| `GET` `PUT` | `/api/env/{folder}`, `/api/env/{folder}/{path}` | the env files; `{content}` saves and generates |

### MCP

`/api/mcp` is an MCP server over the streamable HTTP transport: seven
tools to list jobs and workflows, run them, and read the runs, as the
token's user. Its use is in [the MCP chapter](docs/guide/mcp.md).

## Screenshots

Every view, in both themes and on a phone, is in
[docs/screenshots.md](docs/screenshots.md); the pictures are under
`docs/assets/`.

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
| `access`    | who may do what: the password file, `Sessions`, tokens, the `Guard`          |
| `secret`    | the secrets, encrypted at rest with the SSH keys through the `Vault`         |
| `envfile`   | the dotenv files a folder reads, as task-ui manages and generates them       |
| `sshkey`, `git`, `dotenv` | the key pairs, the checkouts, the file format                 |
| `api`       | the JSON endpoints and the attach socket                                     |
| `mcp`       | the MCP server: JSON-RPC tools that call the API as the token's user         |
| `frontend`  | the pages; `frontend/view` renders them from embedded templates and assets   |
| `app`       | the composition root: configuration, wiring, serving                         |

A convention is a `Detector` in the `detector` package: one type that
reads a folder and returns a `Source`, added to the `Detectors` list. The
rest of task-ui runs on sources and jobs and never asks which kind they
are.

`splint ./...` (`task lint`) is the bar: every file has a test beside it,
every exported symbol a test that reaches it, every symbol a file named
for it, and every dependency one import per package. Tests are black box,
parallel, and share nothing. `scripts/smoke.sh` (`task smoke`) then
exercises a fresh instance end to end with curl and headless chromium:
runs, hours, secrets and keys, env files, schedules, users, projects,
artifacts, every page, and the browser's console.
