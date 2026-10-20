# MCP

task-ui is also an MCP server, so an agent such as Claude Code can see
the jobs and workflows, run them and read how they went. It is at
`/api/mcp`, and the agent signs in with a token of a user: it may see and
run what that user may in the browser, with the same roles, allow lists
and hours.

## A token

An admin gives a user a token on the server, by a name of their choosing:

```bash
task-ui token claude claude-code
```

The token is printed once and only its hash is kept. Giving a token the
same name again replaces it, and the old one stops working.
`task-ui token claude` lists the names and when each was last used, and
`task-ui token claude rm claude-code` removes one.

A token signs in to the rest of `/api` too, in an `Authorization: Bearer`
header instead of the session cookie.

## Claude Code

```bash
claude mcp add --transport http --scope user task https://task.example.com/api/mcp \
  --header "Authorization: Bearer $TOKEN"
```

`claude mcp list` shows whether it connects.

## Tools

| Tool                 | Arguments                                        | What it does |
|----------------------|--------------------------------------------------|--------------|
| `list_jobs`          | `folder`                                         | The folders and their jobs, whether you may run each, how it last went |
| `list_workflows`     |                                                  | The workflows, their steps and their last run |
| `run_job`            | `folder`, `job`, `force`, `reason`, `wait`, `lines` | Starts a job; with `wait`, up to 900 seconds, returns how it ended and the end of its output |
| `run_workflow`       | `workflow`, `wait`                               | Starts a workflow by ID or name; with `wait`, returns how each step went |
| `list_runs`          | `folder`, `job`, `state`, `limit`                | The latest runs, newest first |
| `get_run`            | `id`, `lines`                                    | A run and the end of its output, as the terminal showed it |
| `list_workflow_runs` | `workflow`, `limit`                              | A workflow's latest runs and their steps |

A run started over MCP is like any other: it is in the history, anyone
allowed can watch it, and it goes on if the agent stops waiting.

## A frozen surface

These seven tools, like the JSON API behind them, are the whole surface,
and it is frozen. An agent can see what there is, run it and read how it
went; ending runs, schedules, artifacts, secrets, keys, users and
changing workflows stay with people, in the browser. A tool or an
endpoint is added only for a reason that the ones there cannot serve,
and the reason goes with the change.

---

Previous: [Settings](settings.md) · Back to the [contents](../README.md)
