# task-ui

task-ui is a web UI and an API that runs the jobs of folders under one
root directory, as the user who asks, and keeps every run: its terminal
recording, its output and the files it leaves. People use it in the
browser; agents use the JSON API under `/api` and the MCP server at
`/api/mcp` with a token of a user. A user sees and does what the user's
role and own policy allow, in the folders the user may see. Each section
below names its permissions in brackets after its heading: a section key
such as `note` stands for its rules `note/view`, `note/edit` and
`note/delete`.

## Folders and jobs (`run`, `shell`)

A folder is a directory under the root with something to run, and a
project is a folder at the top, most often a git checkout. task-ui reads
the jobs from the folder's sources, atkins (`atkins.yml`), Taskfile,
make, npm, composer and `scripts/*`, and does not edit them itself. The
jobs it runs, and Claude in a run, work in the folder and may change its
files when asked to. A pipeline
file speaks for its folder. **Recursive**, on by default, lists the
folders inside a project as deep as they go; off lists the project and
its top-level folders alone. Discovery skips what `.gitignore` ignores,
except an ignored checkout right below a workspace or a project that
holds a job source. Pages: `/f/<workspace>/<folder>`; API: `GET
/api/folders`; MCP: `list_jobs`.

Job IDs a run takes besides the listed jobs: `./<path>` runs a file of
the folder that runs, `@<skill>:<job>` runs a job of an atkins skill
that applies there, such as `@git:status`, and a directory inside a
folder runs its own targets with `dir`. Every folder has the built-ins
`@shell`, an interactive `bash -l`, `@claude`, Claude Code in a
terminal, `@claude-prompt`, Claude Code on a prompt, and `@claude-note`,
Claude Code on a note to the agent. `run` lets a user run jobs, in a
folder the user sees, when the allow lists of the folder and the job
name the user or the role, or are empty, and inside the folder's hours
unless forced with a reason. `@shell` and `@claude` take `shell`.

## Runs

A run is one start of a job: who started it, how it ended, its output
and its recording. One run per job goes at a time; **Run force**
replaces it. **Run again** starts the same job, directory and arguments.
**|>summarize** runs a job and has Claude summarise the run on its page
when it ends. A run's page takes prompts about the run. Pages: `/runs`,
`/runs/<id>`; API: `/api/runs`; MCP: `run_job`, `list_runs`, `get_run`.
You see the runs of folders you see; ending one takes the right to run
its job.

## Workspaces (`workspace`)

A workspace is a checkout under the root that holds projects, with its
own encryption key, SSH keys and secrets, kept from other workspaces.
The default workspace, `_`, is the root itself. A workspace exports and
imports between instances. Page: `/settings/workspaces`; API:
`/api/workspaces`.

## Projects (`project`)

The configuration task-ui adds to a folder: title, sources, who may run
it, hours, Recursive and its workspace. Adding one clones a git
repository. Page: `/settings/projects`; API: `/api/projects`.

## Notes (`note`)

A board of notes with replies, in markdown. A note naming a user
with `@name`, from someone who may run jobs, is an instruction to that
user, with a state: open, working, done or failed. A note naming the
agent, `@claude` on most instances, starts `@claude-note` on the thread.
Page: `/notes`; API: `/api/notes`; MCP: `list_notes`, `note_comment`,
`post_note`.

## Prompts (`prompt`)

A prompt asks Claude Code a question in a place, as a run of
`@claude-prompt` as the asking user, and Claude answers in a thread that
**Reply** continues. Prompts come from the Prompt box of a folder or a
directory, the header's prompt, a run's page, and a workflow's prompt.
API: `/api/prompts`; MCP: `ask_prompt`. `prompt/edit` asks, `prompt/view`
reads other users' prompts and their runs.

## Workflows (`workflow`)

A workflow is a named chain of steps, each a job of a folder or another
workflow, run one after another until one fails. A switch on each step
skips it. A workflow can end with a prompt to Claude, which reads the
steps' output and the run's input, or be a prompt alone. Running one
takes the right to run every step. Page: `/workflows`; API:
`/api/workflows`; MCP: `list_workflows`, `run_workflow`,
`list_workflow_runs`.

## Schedules (`schedule`)

A schedule runs a job on a cron expression, in the server's time zone,
as its user. An admin picks who it runs as; anyone else schedules as
themselves. Page: `/schedules`; API: `/api/schedules`.

## Artifacts (`artifact`)

Artifacts are the files runs leave. Every run gets an empty directory in
`$TASK_UI_ARTIFACTS`; files written there move into the one artifact
tree at the same paths when the run ends. A later run that writes the
same path replaces the file, and the history keeps the earlier
versions. A generated file worth keeping, such as a report, a build or
an image, belongs there. An agent writes into the tree as its user,
with `write_artifact`, over the version there, which the history keeps;
`list_artifacts` lists a directory and `read_artifact` reads a file.
`artifact/edit` writes and takes a file out of the tree; the runner role
holds it. Page: `/artifacts`; API: `/api/artifacts`; MCP:
`list_artifacts`, `read_artifact`, `write_artifact`.

## Env files (`env`)

The dotenv files a folder's pipeline reads, as task-ui keeps them, with
the secrets appended on save and before every run. API: `/api/env`.

## Secrets (`secret`)

Variables added to the environment of runs, encrypted with the
workspace's key. A secret's scope is a workspace, `workspace:<name>`,
or a folder. A value is never shown again once saved. Page:
`/settings/secrets`; API: `/api/secrets`.

## Keys (`key`)

The instance's keys: an encryption key for the integrations' settings,
and SSH keys that runs get through their ssh agent by role or user
grants. A key encrypts or connects, never both. Private keys are never
shown. Page: `/settings/keys`; API: `/api/keys`.

## Integrations (`integration`)

The services task-ui talks to: Claude Code, whose token runs get as
`CLAUDE_CODE_OAUTH_TOKEN` for the users and roles it names, and Beszel,
which lists the hosts of the inventory. Their tokens are never shown.
Page: `/settings/integrations`; API: `/api/integrations`.

## Inventory (`inventory`, `hosts`)

The hosts Beszel watches, with their load and the ssh login task-ui
reaches each with. `hosts` runs a shell command on the picked hosts at
once, and can have Claude summarise the outputs. Prompts get the hosts
in `TASK_UI_HOSTS`. Page: `/inventory`; API: `/api/inventory`.

## Users (`user`)

Who may sign in, from the password file, with a role: admin, runner or
viewer. A profile, `/users/<name>`, shows what a user ran and wrote,
the folders the user sees, the user's permissions and MCP access. Page:
`/settings/users`; API: `/api/users`.

## Permissions

A policy allows or denies a rule, for a role or for a user. Every
section has the rules view, edit and delete, and the application has
`run`, `shell` and `hosts`. A deny in the role's or the user's policy
wins, and task-ui refuses what neither allows. An admin may do everything and
holds no policy. Page: `/settings/permissions`; API: `/api/permissions`;
MCP and API: `whoami`, `GET /api/me`, what the token's user may do.

## MCP

The MCP server at `/api/mcp` calls the API as the token's user, with
the same permissions. Its tools: `whoami`, `list_jobs`, `run_job`,
`list_runs`, `get_run`, `list_workflows`, `run_workflow`,
`list_workflow_runs`, `list_notes`, `note_comment`, `post_note`,
`ask_prompt`, `list_artifacts`, `read_artifact` and `write_artifact`.
A user with `user/edit` enables a token on a user's profile; it lasts
30 days, and its user rotates or disables it. A refused call names the
permission it needs.
