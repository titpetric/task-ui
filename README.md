# Task UI

Run what your source trees declare, from the browser, and watch it run.
Task UI looks at each folder under a root, detects the conventions it
follows, and lists their targets as jobs: atkins pipelines, Taskfiles,
Makefile targets, package.json and composer.json scripts, and a
`scripts/` folder. A job runs as a session on the server: a process on a
terminal that anyone allowed can attach to, from any window, for as long
as it runs.

Projects live in workspaces: a checkout of shared tooling with the
projects cloned inside it, its own secrets and SSH keys, and an export
that moves it, keys and all, to another instance. A project's Files
panel browses its checkout like a file manager: a file opens in a
preview window, and a file that runs gets a Run button. Every folder
and directory has an actions bar: the git state of its checkout, a chip
per atkins skill that applies there, and, for an admin, a shell and a
Claude Code prompt in that directory.

![The dashboard: every terminal in progress, as it looks right now](https://raw.githubusercontent.com/titpetric/task-ui/main/docs/assets/task-ui.png)

The [usage guide](docs/README.md) shows how to use it, a chapter per
section of the sidebar: the [dashboard](docs/guide/dashboard.md), the
[inventory](docs/guide/inventory.md), [workflows](docs/guide/workflows.md),
[runs](docs/guide/runs.md), [artifacts](docs/guide/artifacts.md),
[schedules](docs/guide/scheduled.md), [notes](docs/guide/notes.md), the
[projects and their jobs](docs/guide/projects.md), the
[workspaces](docs/guide/workspaces.md) that hold projects, the
[recents](docs/guide/recents.md), the [settings](docs/guide/settings.md),
and [MCP](docs/guide/mcp.md), for an agent that runs jobs. This README
links to the rest.

The look follows [oida](https://github.com/titpetric/oida): hairlines,
one amber accent, dark and light from `prefers-color-scheme`. The type is
a dashboard pair, [Inter](https://rsms.me/inter/) for the words and
[JetBrains Mono](https://www.jetbrains.com/lp/mono/) for the data, both
vendored under `frontend/view/assets/fonts/` with their licenses. The
binary embeds all assets, and the front end makes no external requests.

## Documents

| Document | What is in it |
|----------|---------------|
| [Usage guide](docs/README.md) | The screen, a chapter per sidebar section, and MCP |
| [Running the server](docs/running.md) | The Docker image, the environment, users, roles and the server commands |
| [API](docs/api.md) | The JSON endpoints and the MCP server |
| [Development](docs/development.md) | The packages, the lint and test bar, and the screenshots |

A quick start, with the compose file in [Running the server](docs/running.md):

```bash
task docker:run
```

# Development

The packages, the lint and test bar and the screenshots are in [docs/development.md](docs/development.md). The tasks:

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

