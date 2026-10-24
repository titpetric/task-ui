# Task UI

Run what your source trees declare, from the browser, and watch it run.
Task UI looks at each folder under a root, detects the conventions it
follows, and lists their targets as jobs: atkins pipelines, Taskfiles,
Makefile targets, package.json and composer.json scripts, and a
`scripts/` folder. A job runs as a session on the server: a process on a
terminal that anyone allowed can attach to, from any window, for as long
as it runs.

Projects live in workspaces: a directory with a home, the checkout of
shared tooling that the workspace's runs get as `HOME`, its checkouts
side by side under `src`, its own secrets and SSH keys, and an export
that moves it, keys and all, to another instance. A project's Files
panel browses its checkout like a file manager: a file opens in a
preview window. Every folder and directory has the git state of its
checkout and one Jobs table of what runs there: its targets, the files
that run, the jobs of each atkins skill that applies, and, for an
admin, a shell and Claude Code in a terminal in that directory.

Claude works beside the jobs. A prompt asks Claude Code in a folder, a
directory, about a run, from the header on any page, or at the end of a
workflow, and Claude's reply lands under the question. **|>summarize** runs
a job and has Claude summarise the run. The Notes board takes messages
and replies, and a note naming a user with `@name` is an instruction to
that user: to the agent, which works on it at once, or to
an agent elsewhere over MCP. The MCP server at `/api/mcp` serves an
agent the same jobs, runs, notes and prompts, as a user with a token of
its own. Permissions decide, per role and per user, what each may view,
edit, delete and run. The inventory lists the hosts Beszel watches and
runs a command on every one at once.

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
| [Security](docs/security.md) | The threat model, identity and roles, keys, secrets, terminals and agents, and the review's findings |
| [Development](docs/development.md) | The packages, the lint and test bar, and the screenshots |
| [Testing](docs/testing.md) | Unit tests by default, integration tests behind a build tag |

A quick start, with the compose file in [Running the server](docs/running.md):

```bash
task docker:run
```
