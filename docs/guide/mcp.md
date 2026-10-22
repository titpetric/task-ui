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

## Claude Code inside task-ui

The Claude Code that task-ui runs, the agent job of the [Notes](notes.md)
board and a Claude terminal, reaches the same tools when its own
configuration has the server. Add it with a token of the agent's user and
the loopback address, from inside the server:

```bash
claude mcp add --transport http --scope user task http://127.0.0.1:3000/api/mcp \
  --header "Authorization: Bearer $(task-ui token claude task-agent)"
```

The configuration sits in `CLAUDE_CONFIG_DIR`; keep that on a volume so
it outlives the container. `claude mcp list` shows the server as
connected.

## Tools

| Tool                 | Arguments                                        | What it does |
|----------------------|--------------------------------------------------|--------------|
| `list_jobs`          | `folder`                                         | The folders and their jobs, whether you may run each, how it last went |
| `list_workflows`     |                                                  | The workflows, their steps and their last run |
| `run_job`            | `folder`, `job`, `force`, `reason`, `wait`, `lines` | Starts a job; with `wait`, up to 900 seconds, returns how it ended and the end of its output |
| `run_workflow`       | `workflow`, `wait`                               | Starts a workflow by ID or name; with `wait`, returns how each step went |
| `list_runs`          | `folder`, `job`, `state`, `limit`                | The latest runs, newest first, a directory's run as `folder/dir/job` |
| `get_run`            | `id`, `lines`                                    | A run and the end of its output, as the terminal showed it |
| `list_workflow_runs` | `workflow`, `limit`                              | A workflow's latest runs and their steps |
| `list_notes`         | `mention`, `state`, `limit`                      | The notes with their replies; by default the open ones that name you |
| `note_comment`       | `id`, `body`, `state`                            | Replies to a note and sets its instruction to working, done or failed |
| `post_note`          | `body`, `parent`                                 | Leaves a note on the board as you, or a reply to `parent`; takes `note/edit` |
| `ask_prompt`         | `folder`, `dir`, `prompt`, `wait`                | Asks Claude in a folder, or `dir` inside it, as the folder's **Prompt** box does; with `wait`, up to 900 seconds, returns Claude's answer; takes `prompt/edit` |

Each tool takes the permissions of the token's user, the ones a page
takes in the browser. A note from `post_note` and a prompt from
`ask_prompt` carry that user's name. A user with `prompt/edit` asks in
the folders the user sees, and reads the runs of the prompts the user
asked with `get_run`, though they run as `internal/claude-prompt`.

A run started over MCP is like any other: it shows in the history, anyone
allowed can watch it, and it goes on if the agent stops waiting.

`run_job` takes the job IDs the folder's page runs: a job of
`list_jobs`, `./<path>` for a file of the folder that runs, and
`@<skill>:<job>` for a job of an atkins skill that applies in the
folder's own directory, such as `@git:status`. It has no `dir` and no
`args`: the targets of a directory inside a folder and a job's arguments
start from the browser or `POST /api/runs`, and `list_runs` and
`get_run` name such a run `folder/dir/job`, as `malta/dockerfiles/blog/build`. `@shell`, the interactive
shell of **Enter Shell**, answers `403` to every token but an admin's.

## A frozen surface

These eleven tools, like the JSON API behind them, are the whole surface,
and the surface stays frozen. An agent can see what there is, run it, read
how it went, work through the notes that name it, leave notes and ask
Claude in a folder; ending runs, schedules, artifacts, secrets, keys, users and
changing workflows stay with people, in the browser. A new tool or
endpoint needs a reason that the current ones cannot serve,
and the reason goes with the change.

---

Previous: [Settings](settings.md) · Back to the [contents](../README.md)
