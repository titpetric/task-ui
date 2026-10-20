# MCP

task-ui is also an MCP server, so an agent such as Claude Code can see
the jobs and workflows, run them and read how they went. The endpoint
`/api/mcp` serves it, and the agent signs in with a token of a user: it may see and
run what that user may in the browser, with the same roles, allow lists
and hours.

## Install

Each step runs on the server, in the task-ui container when task-ui runs
in Docker (`docker exec task task-ui ...`).

1. Pick the user the agent signs in as. A runner suits an agent that runs
   jobs; a viewer only reads. Add one to the password file if the agent
   gets a user of its own, and wait for task-ui to read the file:

   ```bash
   task-ui passwd claude
   ```

   `TASK_UI_ADMINS` and `TASK_UI_VIEWERS` set the role; every other user
   is a runner (see [Running the server](../running.md)).

2. Create a token for that user, under a name of your choosing, and keep
   it in a file only root reads:

   ```bash
   docker exec task task-ui token claude claude-code > /root/task.token
   chmod 600 /root/task.token
   ```

   task-ui prints the token once and keeps only its hash. A new token
   under the same name replaces it, and the old one stops working.

3. Add the server to the agent, on the machine the agent runs on. For
   Claude Code:

   ```bash
   claude mcp add --transport http --scope user task https://task.example.com/api/mcp \
     --header "Authorization: Bearer $(cat /root/task.token)"
   ```

   Any other MCP client takes the same two things: the URL of `/api/mcp`
   over HTTP, and the header `Authorization: Bearer <token>`.

4. Check it: `claude mcp list` shows `task` as connected, and a new
   session lists the tools below. Without a token, or with a wrong one,
   `/api/mcp` answers 401.

To list the tokens of a user with the last use of each, run
`task-ui token claude`; `task-ui token claude rm claude-code` removes one.
A token signs in to the rest of `/api` too, in an `Authorization: Bearer`
header in place of the session cookie.

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

A run started over MCP is like any other: it shows in the history, anyone
allowed can watch it, and it goes on if the agent stops waiting.

## A frozen surface

These seven tools, like the JSON API behind them, are the whole surface,
and the surface stays frozen. An agent can see what there is, run it and read how it
went; ending runs, schedules, artifacts, secrets, keys, users and
changing workflows stay with people, in the browser. A new tool or
endpoint needs a reason that the current ones cannot serve,
and the reason goes with the change.

---

Previous: [Settings](settings.md) · Back to the [contents](../README.md)
