# MCP

task-ui is also an MCP server, so an agent such as Claude Code can see
the jobs and workflows, run them and read how they went. The endpoint
`/api/mcp` serves it, and the agent signs in with a token of a user: it may see and
run what that user may in the browser, with the same roles, allow lists
and hours.

## Install

1. Pick the user the agent signs in as. A runner suits an agent that runs
   jobs; a viewer only reads. Add one to the password file if the agent
   gets a user of its own: **Create user** on Settings, Users, or on the
   server:

   ```bash
   task-ui passwd claude
   ```

   A new user comes in as a viewer; an admin sets the role on the user's
   profile, under **Roles**, or with `task-ui user role claude runner`
   (see [Running the server](../running.md#users-and-roles)).

2. Turn on **MCP agent user** under **Roles** on the user's profile, as
   an admin, or with `task-ui user mcp claude on`: the endpoint takes the
   tokens of such a user alone, and answers `403` "MCP is off for this
   user" to any other, an admin's too. Turning it off ends the user's MCP
   token.

3. Open the user's profile, from Settings, Users or the **⋯** menu next to
   their name, and press **Enable MCP** under **MCP access**. task-ui makes a token that lasts 30 days and shows it once,
   beside the ready `claude mcp add` command; copy both before you leave
   the page, which keeps only the token's hash. The section shows the
   token's creation, its expiry, in amber within a week and red once
   past, and its last use. **Rotate** makes a new token and
   the old one stops working at once; **Disable** removes it.

   Enabling takes Allow Edits on Users, `user/edit`, which the admin role
   holds and a policy can grant (see
   [Settings](settings.md#permissions)); a user rotates
   or disables their own token once on, from their own profile. A
   token past its expiry gets `401` "the token has expired; rotate it on
   your profile".

4. Add the server to the agent, on the machine the agent runs on. For
   Claude Code:

   ```bash
   claude mcp add --transport http --scope user task https://task.example.com/api/mcp \
     --header "Authorization: Bearer <token>"
   ```

   Any other MCP client takes the same two things: the URL of `/api/mcp`
   over HTTP, and the header `Authorization: Bearer <token>`. Settings,
   Users shows the instance's base link and the endpoint to copy, and this
   command with both filled in; `TASK_UI_URL` sets the base link (see
   [Running the server](../running.md)).

5. Check it: `claude mcp list` shows `task` as connected, and a new
   session lists the tools below. Without a token, or with a wrong one,
   `/api/mcp` answers 401.

On the server, the CLI makes tokens without an expiry, under a name of
your choosing, for an agent that runs unattended: in the task-ui
container when task-ui runs in Docker:

```bash
docker exec task task-ui token claude claude-code > /root/task.token
chmod 600 /root/task.token
```

task-ui prints the token once and keeps only its hash. A new token under
the same name replaces it, and the old one stops working.

To list the tokens of a user with the last use of each, run
`task-ui token claude`; `task-ui token claude rm claude-code` removes one.
A token signs in to the rest of `/api` too, in an `Authorization: Bearer`
header in place of the session cookie.

## Claude Code inside task-ui

Every prompt task-ui runs, `@claude-prompt`, and the Notes agent,
`@claude-note`, get the tools as the user they run for, with no setup:
task-ui mints a token for the run when it starts and revokes it when the
run ends, writes an MCP configuration with it to a directory of the run's
own, mode 0700, that `TASK_UI_MCP_CONFIG` names, and Claude starts with
`--mcp-config` on it. The token has that user's permissions and no more,
shows nowhere, and is none of the user's tokens: it needs no MCP agent user
on the user's profile. The configuration goes with the run. So a
prompt can build a workflow with `create_workflow`, run it with
`run_workflow` and read how it went.

A prompt's run also gets its source in `TASK_UI_PROMPT_SOURCE`, which
goes to Claude's system prompt: the thread's note ID, the reply that asks
when a reply asks, its place or area and its link, and the instruction to
reply there with `note_comment` once the result is in. The token
comments on that thread alone, with the asker's `prompt/edit`, and any
other note answers 403. The reply comes as Claude, the author of the
thread's answers, and a state given with it changes nothing: the thread
ends with its run. Claude's final message is the fallback: task-ui posts
it only when the run posted no reply. A run's token can't start a prompt,
`ask_prompt` answers 409: a builder would start another builder.

A Claude terminal, **Enter Claude**, reaches the same tools when its own
configuration has the server. Add it with a token of the agent's user and
the loopback address, from inside the server:

```bash
claude mcp add --transport http --scope user task http://127.0.0.1:3000/api/mcp \
  --header "Authorization: Bearer $(task-ui token claude task-agent)"
```

The configuration sits in `CLAUDE_CONFIG_DIR`; keep that on a volume so
it outlives the container. `claude mcp list` shows the server as
connected.

Every Claude task-ui starts reads `config/context.md` after its own
system prompt, with a line naming the user it works for, the role and
what the user may do, from the run's `TASK_UI_USER`, `TASK_UI_ROLE` and
`TASK_UI_PERMISSIONS`, and the run's `$TASK_UI_ARTIFACTS`: the prompts,
the Notes agent, **Enter Claude**, and the summaries of runs, workflow
runs and the inventory.

## An agent elsewhere

An MCP client on another machine, with a token of its own user, such as
`consuela`, takes the notes that name it, as the agent job does for
`@claude`:

1. `list_notes` with no arguments lists the open notes that name the
   token's user, in the note or in a reply, with their threads. Poll it.
2. `note_comment` with `state` `working` claims a note, so the board shows
   someone is on it.
3. `note_comment` with a `body`, what changed and where to look, and
   `state` `done`, or `failed` with the reason, ends it.

A reply that names the user again opens the note again, and it shows in
the next `list_notes`. The state is for the users the note names, its
author and an admin; anyone who may post replies.

## An SSH key for a task

An agent elsewhere that needs an instance SSH key, such as the deploy
key the hosts accept, borrows it for one task, and the key never
crosses in clear. It takes **Allow Read Private** on Instance Keys,
`key/read`: admins hold it, and no other role does unless an admin
allows it in [Settings, Permissions](settings.md#permissions). The
user has MCP agent user on, and the key is an SSH key linked to the
user or their role; an admin reads any SSH key. An encryption key never
leaves task-ui. `whoami` names the keys you may read.

1. Make an age identity for this request alone:

   ```bash
   umask 077
   age-keygen -o /tmp/task-key.id 2>/dev/null
   age-keygen -y /tmp/task-key.id    # the recipient, age1...
   ```

2. Call `get_ssh_key` with the key's name and the recipient, or `GET
   /api/keys/{name}/private?recipient=age1...`. It returns the key's
   name, algorithm and fingerprint and the private key armored and
   encrypted to the recipient.

3. Decrypt into `/tmp`, use it, and delete both files after the task:

   ```bash
   age -d -i /tmp/task-key.id -o /tmp/deploy.key reply.age
   ssh -i /tmp/deploy.key root@host uptime
   rm -f /tmp/deploy.key /tmp/task-key.id
   ```

A missing or wrong recipient answers `400`, and each read lands in the
key's read log: who, by which token, over MCP or the API. The Keys page
shows a key's **Reads** and **Last read**, and [Data
retention](settings.md#data-retention) keeps the log.

## Live notes for an agent elsewhere

`task-ui channel` pushes the notes to an open Claude Code session as they
come, in place of polling `list_notes`: a Claude Code
[channel](https://code.claude.com/docs/en/channels-reference.md), a
research preview, and a stdio MCP server, in the task-ui binary, that the
session starts. It holds the instance's notes stream open with the
token, connects again with a growing wait when the stream drops, and
reads the open notes each time it connects. Every open instruction that
names the token's user, from someone else, reaches the session once
while it stays open, as:

```text
<channel source="task-notes" note_id="01J..." author="black" url="https://task.example.com/notes#note-01J...">
Note 01J... from black: @consuela tidy the images

Reply with note_comment (state working, then done or failed).
</channel>
```

A note to any other user, and the user's own, never reaches the session:
anyone who may post can name the user, and only the user's notes go
through. The channel has no tools; the session answers with the `task`
server's `note_comment`, so add that server too (see [Install](#install)).

Add the channel on the machine the agent runs on, with the token in a
file:

```bash
claude mcp add --scope user task-notes -- \
  task-ui channel --url https://task.example.com --token-file /root/task.token
```

and start the session with the channel loaded. Custom channels need the
development flag in the preview, in an interactive session:

```bash
claude --dangerously-load-development-channels server:task-notes
```

The notes reach a session only while it runs; a note left between
sessions arrives with the open notes when the next session starts the
channel. Without channels, such as in a session started
with `-p`, `/loop` with `list_notes` polls in its place:
`/loop 5m list_notes, and work through the notes that name you`.

## Tools

| Tool                 | Arguments                                        | What it does |
|----------------------|--------------------------------------------------|--------------|
| `whoami`             |                                                  | Who you are: your role, the permissions you hold, each area of task-ui with a line on its purpose and what you may do there, and the folders you see |
| `list_jobs`          | `folder`                                         | The folders and their jobs, `default` first, whether you may run each, how it last went |
| `list_workflows`     |                                                  | The workflows, their steps and their last run |
| `run_job`            | `folder`, `job`, `force`, `reason`, `wait`, `lines` | Starts a job; with `wait`, up to 900 seconds, returns how it ended and the end of its output |
| `run_workflow`       | `workflow`, `input`, `wait`                      | Starts a workflow by ID or name, `input` going to its prompt; with `wait`, returns how each step went and the run's answer, Claude's reply to its last prompt |
| `get_workflow`       | `workflow`                                       | A workflow by ID or name: its steps, each a job, a workflow or a prompt, its final prompt and where it runs, and its last run |
| `create_workflow`    | `name`, `steps`, `prompt`, `prompt_folder`, `sidebar` | Makes a workflow as you: `steps` in order, each `{folder, job, args}`, `{workflow}` or `{prompt, folder, dir}`, and `enabled`; takes `workflow/edit`, and `prompt/edit` for a prompt |
| `update_workflow`    | `workflow`, and the fields of `create_workflow`  | Changes a workflow you own, or any as an admin: a field left out stays, `steps` given replace them all; there is no delete |
| `list_runs`          | `folder`, `job`, `state`, `limit`                | The latest runs, newest first, a directory's run as `folder/dir/job` |
| `get_run`            | `id`, `lines`                                    | A run and the end of its output, as the terminal showed it |
| `list_workflow_runs` | `workflow`, `limit`                              | A workflow's latest runs and their steps |
| `list_notes`         | `mention`, `state`, `limit`                      | The notes with their replies; by default the open ones that name you |
| `note_comment`       | `id`, `body`, `state`                            | Replies to a note and sets its instruction to working, done or failed; the state is for a user the note names, its author or an admin. A prompt's run replies in its own thread alone, as Claude |
| `post_note`          | `body`, `parent`                                 | Leaves a note on the board as you, or a reply to `parent`; takes `note/edit` |
| `ask_prompt`         | `folder`, `dir`, `prompt`, `area`, `wait`        | Asks Claude in a folder, or `dir` inside it, as the folder's **Prompt** box does; with `wait`, up to 900 seconds, returns Claude's answer; `area` `workflows` asks as the Workflows page's box does, in the root unless `folder` names one: Claude builds or changes the workflow the prompt describes, runs it and answers with its link; takes `prompt/edit` |
| `list_artifacts`     | `path`                                           | A directory of the artifact tree, the top unless given: each file's size, time, and the run or the user that wrote it; takes `artifact/view` |
| `read_artifact`      | `path`, `version`                                | A file of the tree, the latest or `version`: text up to 256 KiB, saying where it cut, and for anything else its type, size and download link; takes `artifact/view` |
| `write_artifact`     | `path`, `content`, `encoding`                    | Writes a file into the tree as you, at most 16 MiB, `encoding` `base64` for a binary one, over the version there, which the history keeps; a relative path with no `..` and no dot file; takes `artifact/edit` |
| `get_ssh_key`        | `name`, `recipient`                              | An instance SSH key for the length of a task, its private half encrypted to your age recipient; takes `key/read` and MCP agent user (see [An SSH key for a task](#an-ssh-key-for-a-task)) |

Each tool takes the permissions of the token's user, the ones a page
takes in the browser: the role's policy and the user's own, which
Settings, Permissions and the user's profile set (see
[Settings](settings.md#permissions)). An agent that should post notes or
ask prompts gets Allow Edits on Notes or on Prompts in its own policy;
the `claude` user starts with both. A note from `post_note` and a prompt from
`ask_prompt` carry that user's name. A user with `prompt/edit` asks in
the folders the user sees, and reads the runs of the prompts the user
asked with `get_run`: `@claude-prompt` of the folder asked in. Other
users' prompt runs list in `list_runs` and `get_run` with `prompt/view`.

A call your permissions refuse fails with the permission it needs and a
pointer to `whoami`, as `you need the prompt/edit permission (403); call
whoami to see what you can do`.

A run started over MCP is like any other: it shows in the history, anyone
allowed can watch it, and it goes on if the agent stops waiting.

`run_job` takes the job IDs the folder's page runs: a job of
`list_jobs`, `./<path>` for a file of the folder that runs, and
`@<skill>:<job>` for a job of an atkins skill that applies in the
folder's own directory, such as `@git:status`. It has no `dir` and no
`args`: the targets of a directory inside a folder and a job's arguments
start from the browser or `POST /api/runs`, and `list_runs` and
`get_run` name such a run `folder/dir/job`, as `malta/dockerfiles/blog/build`. `@shell`, the interactive
shell of **Enter Shell**, and `@claude`, Claude Code in one, answer `403`
to every token whose user lacks the `shell` permission, Enter Shell; `@claude-note` and `@claude-prompt` start
from the Notes board and the prompt routes alone. The Notes agent's runs,
`@claude-note`, list with `note/view`, and to the agent's user always.

## What task-ui is

`initialize` returns the server's instructions: `config/context.md`,
which task-ui embeds, then what the tools are for and "Call whoami to
see what you may do". `context.md` describes task-ui for an agent, an
area at a time, such as artifacts, notes or secrets: its purpose, its
page and its API route or tool, and the permission that governs it. The
same text is the resource `taskui://context`, `text/markdown`, for
`resources/list` and `resources/read`, and `whoami` returns one line of
it per area beside what you may do there. `GET /api/me` answers the same as JSON (see
[API](../api.md)).

## A frozen surface

These nineteen tools, like the JSON API behind them, are the whole surface,
and the surface stays frozen. An agent can see what it may do and what there is, run it, read
how it went, work through the notes that name it, leave notes, ask
Claude in a folder, read and write the artifact tree, and make and change
workflows; ending runs, schedules, deleting workflows, artifacts,
secrets, keys and users stay with people, in the browser. A new tool or
endpoint needs a reason that the current ones cannot serve,
and the reason goes with the change.

---

Previous: [Settings](settings.md) · Back to the [contents](../README.md)
