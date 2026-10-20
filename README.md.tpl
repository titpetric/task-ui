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
links to the rest.

The look follows [oida](https://github.com/titpetric/oida): hairlines,
one amber accent, dark and light from `prefers-color-scheme`. The type is
a dashboard pair, [Inter](https://rsms.me/inter/) for the words and
[JetBrains Mono](https://www.jetbrains.com/lp/mono/) for the data, both
vendored under `frontend/view/assets/fonts/` with their licenses. All
assets are embedded, the front end makes no external requests.

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
