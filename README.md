# Task UI

Run what your source trees declare, from the browser, and watch it run. Task UI looks at each folder under a root, detects the conventions it follows, and lists their targets as jobs: atkins pipelines, Taskfiles, Makefile targets, package.json and composer.json scripts, and a `scripts/` folder. A job runs as a session on the server: a process on a terminal that anyone allowed can attach to, from any window, for as long as it runs.

Projects live in workspaces: a directory with a home, the checkout of shared tooling that the workspace's runs get as `HOME`, its checkouts side by side under `src`, its own secrets and SSH keys, and an export that moves it, keys and all, to another instance. A project's Files panel browses its checkout like a file manager: a file opens in a preview window. Every folder and directory has the git state of its checkout and one Jobs table of what runs there: its targets, the files that run, the jobs of each atkins skill that applies, and, for an admin, a shell and Claude Code in a terminal in that directory.

Claude works beside the jobs. A prompt asks Claude Code in a folder, a directory, about a run, from the header on any page, or at the end of a workflow, and Claude's reply lands under the question. **|> Summarize** runs a job and has Claude summarise the run. The Notes board takes messages and replies, and a note naming a user with `@name` is an instruction to that user: to the agent, which works on it at once, or to an agent elsewhere over MCP. The MCP server at `/api/mcp` serves an agent the same jobs, runs, notes and prompts, as a user with a token of its own. Permissions decide, per role and per user, what each may view, edit, delete and run. The inventory lists the hosts Beszel watches and runs a command on every one at once.

![The dashboard: every terminal in progress, as it looks right now](https://cdn.si/titpetric/task-ui/docs/assets/task-ui.png)

The [usage guide](docs/README.md) shows how to use it, a chapter per section of the sidebar: the [dashboard](docs/guide/dashboard.md), the [inventory](docs/guide/inventory.md), [workflows](docs/guide/workflows.md), [runs](docs/guide/runs.md), [artifacts](docs/guide/artifacts.md), [schedules](docs/guide/scheduled.md), [notes](docs/guide/notes.md), the [projects and their jobs](docs/guide/projects.md), the [workspaces](docs/guide/workspaces.md) that hold projects, the [recents](docs/guide/recents.md), the [modules](docs/guide/modules.md), the [settings](docs/guide/settings.md), and [MCP](docs/guide/mcp.md), for an agent that runs jobs. This README links to the rest.

A module is a plugin written in PHP, run inside the binary by [phpscript](https://github.com/titpetric/phpscript) and served at `/m/<vendor>/<name>` as an ordinary page: the layout is the instance's, the body is the module's, and nothing loads in an iframe. A directory with a `modinfo.php` in it is a module. `titpetric/refinement` ships with it: a thread of comments, each one a Claude run that writes artifacts and the `README.md` that indexes them.

The look follows [oida](https://github.com/titpetric/oida): hairlines, one amber accent, dark and light from `prefers-color-scheme`. The type is a dashboard pair, [Inter](https://rsms.me/inter/) for the words and [JetBrains Mono](https://www.jetbrains.com/lp/mono/) for the data, both vendored under `frontend/view/assets/fonts/` with their licenses. The binary embeds all assets, and the front end makes no external requests.

## Documents

| Document                               | What is in it                                                                                        |
|----------------------------------------|------------------------------------------------------------------------------------------------------|
| [Usage guide](docs/README.md)          | The screen, a chapter per sidebar section, and MCP                                                   |
| [Running the server](docs/running.md)  | The Docker image, the environment, users, roles and the server commands                              |
| [API](docs/api.md)                     | The JSON endpoints and the MCP server                                                                |
| [Security](docs/security.md)           | The threat model, identity and roles, keys, secrets, terminals and agents, and the review's findings |
| [Development](docs/development.md)     | The packages, the lint and test bar, the prose gate and the screenshots                              |
| [Modules](docs/modules.md)             | Writing a plugin in PHP: the directory, the bindings, and what the runtime refuses                   |
| [Corrections](docs/todo-correction.md) | Where the PHP runtime lacks what the module system needs, and what the code does instead             |
| [Testing](docs/testing.md)             | Unit tests by default, integration tests behind a build tag                                          |

A quick start, with the compose file in [Running the server](docs/running.md):

```bash
task docker:run
```

# Development

The packages, the lint and test bar and the screenshots are in [docs/development.md](docs/development.md). The tasks:

task: Available tasks for this project:

* default: Run everything
* fix: Fix code
* install: Install task-ui
* lint: Lint with splint, rulelint and vale, zero findings is the bar
* run: Run task-ui over the demo projects
* test: Run the unit tests
* docker:build: Build task-ui docker image
* docker:gen: Generate an ssh key for the key folder, ./keys
* docker:run: Run task-ui in docker env
* docs:shots: Regenerate the README screenshots
* lint:docs: Report the prose findings of the markdown, Go, PHP, JavaScript and CSS
* test:cover: Report unit test coverage per package
* test:duplicate: Report the coverage two tests share
* test:integration: Run the integration tests: processes, databases and the network
* test:scope: Run the tests of what changed, inside the 1s per test budget

## task: default

Run everything

commands:
- `Task: fix`- `Task: lint`- `Task: install`- `Task: test`

## task: test

Run the unit tests

commands:
- `go test ./...`

## task: test:scope

The packages a change touched and the packages that import them, from git and go list, with every test held to a second. Takes the flags of scripts/test.sh: task test:scope -- --all

commands:
- `scripts/test.sh`

## task: test:duplicate

Report the coverage two tests share

commands:
- `scripts/test-duplicate.sh`

## task: test:integration

One go test per package, batched by test name, under a memory cap, so a package that holds more than the cap fits cannot take the suite down. Takes a package pattern: task test:integration -- ./storage/...

commands:
- `scripts/test-integration.sh`

## task: test:cover

Report unit test coverage per package

commands:
- `go test -cover ./...`

## task: run

Run task-ui over the demo projects

commands:
- `go run .`

## task: docs:shots

Regenerate the README screenshots

commands:
- `scripts/demo.sh`

## task: lint

Lint with splint, rulelint and vale, zero findings is the bar

commands:
- `splint ./...`- `go run ./internal/cmd/rulelint`

## task: lint:docs

Report the prose findings of the markdown, Go, PHP, JavaScript and CSS

commands:
- `git ls-files --cached --others --exclude-standard | grep -E '\.(md|go|php|js|css)$' | xargs -r vale`

## task: install

Install task-ui

commands:
- `CGO_ENABLED=0 go install .`

## task: fix

Fix code

dependencies:
- `deps:splint` commands:
- `splint fix ./...`- `go vet ./...`- `go mod tidy`- `./README.md.sh > README.md`
