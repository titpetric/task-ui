# Workflows

A workflow is a named chain of steps: a one button deploy, such as
build, test, push, plan, apply, or a chain of prompts to Claude that
fetches something and reasons over it, ending with a summary. A step
is a command, a job of any folder or another workflow, a prompt to
Claude, or a summary by Claude of the steps before it, in any number
and order. The sidebar lists the workflows under **Workflows**, each
with a play button.

![Workflows](../assets/workflows-dark.png)

## Building one with Claude

With `prompt/edit` and the Claude Code integration set up for you, the
Workflows page has a **Prompt** box: describe a workflow to build or
change, and Claude Code builds it, as a run of `@claude-prompt` as you
in the root, or the folder **Run in** picks. Claude works with the task
tools of [MCP](mcp.md#claude-code-inside-task-ui), as you: it reads the
workflows and jobs there are, saves the workflow with prompt steps and
existing jobs before it writes any file, runs it and waits, reads how it
went, and fixes and runs it again until it works, at most 3 runs. A
prompt step of the workflow runs beside the builder's own run, as every
prompt does, so the test run goes while the builder waits, and the
builder waits for a run until it ended. It replies in the thread with
`note_comment` (see [MCP](mcp.md#claude-code-inside-task-ui)): the
workflow's link, the run's link, its steps, and the output of the final
step verbatim. Its threads list under the box, apart from the root's
prompts, and a reply in one asks again.

Over MCP, `ask_prompt` with `area` `workflows` posts the same prompt.

## The steps

Each step is of a kind, which the editor shows on its card and the API
names in `kind`:

- **Command** (`job`): a job of a folder, picked from the jobs you can
  see, with its arguments for a job that takes them; a command on hosts
  of the inventory (see below); or another workflow (`workflow`), which
  then runs in its place, all its steps.
- **Prompt** (`prompt`): Claude answers a prompt in a folder, the root by
  default, or a directory inside it, with the output of the steps before
  it. Claude's answer is the step's output: the prompts after it read it.
- **Summary** (`summary`): Claude summarises the steps before it, with
  no prompt of yours: task-ui asks Claude the question "Summarize this
  workflow run." and the instruction of **Run and summarize**, one or
  two short paragraphs and a table of the values worth a glance. A
  summary is optional; it ends most chains.

A longer chain of prompts is fine: fetch with one prompt, reason over
its answer with the next, and summarise at the end.

## Running one

**Run** starts the first step. Each next step starts when the one before
it ended ok. The first step that does not end ok ends the workflow, and
the rest do not run, but for a summary: a summary step runs after a
failed step too, to say what failed and why. A **Stop** skips every step
after it, a summary too. While a workflow runs, the play button in the
sidebar becomes an eye: open it to watch.

A step that is off stays in the chain and a run skips it: its row reads
skipped, and the steps after it run.

Running a workflow needs the right to run every step in it, and
`prompt/edit` for a prompt or a summary. A workflow runs once at a time;
task-ui turns down a second start while it runs.

**Run** on a workflow with a prompt or a summary opens a field for
input first, optional: each prompt reads it after its text, under
`Input:`, and every prompt and summary has it in the workflow context.

An agent runs one over [MCP](mcp.md) by name: `run_workflow` starts it,
with `input` for its prompts, and can wait for the end, then says how
each step went, the input and the run's answer; `list_workflows` and
`list_workflow_runs` show the chains and their last runs, each with its
input and answer.

## A run is a thread

A workflow run reads like a note and its reply:

- **The comment** is what the run started with, under who started it:
  the input of the Run field, or the workflow's first prompt when the
  run has no input. A workflow of commands alone run without input
  says that it ran.
- **The reply** is the run's answer, under Claude: the markdown of the
  last prompt or summary step that answered, Claude's final message as
  it wrote it, or its reply in the thread over MCP. A prompt whose run
  ends without one, only the report of how its run ended, answers
  nothing, and the reply of the step before it stays. A workflow of commands alone
  has no answer: its reply is how the run ended, its state, its message
  and how long it took.

While a run goes the reply shows working, and fills in when the run
ends. `/api/workflow-runs` and `/api/workflows/<id>/runs` return a run's
`input` and `answer`, and the `done` event of the workflow's topic
carries the `answer`.

## Prompt and summary steps

A prompt or a summary runs `@claude-prompt` in its place as the user
who started the run. Claude gets the workflow context (below) and each
step before it: its place in the chain, its folder and job, how it
ended, its exit code, how long it took and the end of its output,
colours stripped, or a prompt step's answer. The steps share the output
budget of a run's prompt, the newest step first, and a cut output says
so. A step that is off reads skipped. Claude changes nothing
unless the prompt names a change, as with a prompt on a run's page.

The step ends ok when Claude's run ends ok; a prompt that fails ends
the workflow, as a job does. **Stop** during a prompt stops Claude's run
too, and the workflow run ends stopped.

The pipeline shows a prompt step with the prompt icon, as a prompt in
its folder, its prompt under it, a summary with the summary icon, and
each with its run once it ran. Their threads show on the workflow run's
page under **Prompts**, live while Claude works: the question, and
Claude's answer under it. A reply in a thread asks again with the same
steps. Saving a prompt or a summary, and running a workflow with one,
takes the `prompt/edit` permission.

Through `/api/workflows`, a prompt step is `{"kind": "prompt", "prompt":
"fetch the weather", "folder": "_", "dir": ""}`, a summary step
`{"kind": "summary", "folder": "_"}`: `folder` the folder Claude runs
in, `_` or empty for the root, and `dir` a directory inside it.

## Workflow context

Every prompt a workflow runs, a prompt step's and a summary's, starts
with a **Workflow context** block, before the steps' output: the
workflow's name and link, the run's number, when it started, who
started it and its input, and what the last 3 finished runs of the
workflow replied, newest first, each with its date and how the run
ended, the run's answer (see [A run is a thread](#a-run-is-a-thread)). The
block skips a run without one, and it does not count. Each keeps its first
1.5 KiB, and a cut one says so; the block takes its share of the run's
output budget. The first run of a workflow reads that there are no
earlier runs. So a prompt can stay consistent with what it said before,
such as "yesterday I said water tomorrow". The editor says so under a
prompt: Claude also gets the last 3 answers of this workflow.

## The Workflows page

**Workflows** lists the workflows you may see, each with its chain, its
last run and **Run**. **Recent runs** under it lists the newest 15 runs
of those workflows, each with the workflow, who started it, when and how
it ended, and its thread on two lines: the input over the start of the
answer. A run links to its page. A run shows when you see every folder
its steps ran in.

## The workflow page

![A workflow: its chain and its runs](../assets/workflow-dark.png)

- **Pipeline**: the steps in the order they run, a step that is off
  dimmed, a prompt with its text under it. While a run is going, it
  follows that run live: each step's state and terminal.
- **Runs**: the workflow's activity, its runs of the last thirty days by
  day and of the last 24 hours by hour, and the duration of its latest thirty (see
  [Runs](runs.md#activity)), then every run of the workflow, newest first, with who started it,
  how it ended, its input and the start of its answer, and each step's
  run, which opens like any other run.
  **View** opens that run in a window, its thread over its steps, each
  step with its state and the screen it ended on; Ctrl and a click open
  it as a page.
  With the Claude Code integration set up for you, a finished run also
  has **Summary**: Claude reads every step, its job, how it ended and the
  end of its output, and writes what the workflow did or where and why it
  failed, in one or two short paragraphs and a table of the values worth
  a glance, such as failures per step. task-ui keeps the summary, as for a single run (see
  [Runs](runs.md#the-list)).
  The runs come a page at a time, as on [Runs](runs.md#the-list):
  **Page size** picks 15, 50 or 100 rows, and **Previous** and **Next**
  under the table page through the rest. The pipeline and its latest run
  follow the newest run on every page.
  **Stop** on a running workflow stops the step that runs and skips the
  rest.

## Making and changing one

![The workflow editor: the steps on a rail, a prompt open](../assets/workflow-edit-dark.png)

**Add workflow** on the Workflows page opens the editor on a blank
workflow: a **Name**, a **Description**, optional, which the workflow's
page and the list show under its name, and **Show in sidebar**. The
steps come next, on a rail: a numbered node a step, in the order they
run, and the step's card beside it, in the colour of its kind.

- **Add step** at the foot of the rail picks the kind: **Command**,
  **Prompt** or **Summary**, the last two with `prompt/edit`. The new
  card opens on its fields.
- A card folds to one line: its kind, and the job and its args, the
  workflow, the first line of the prompt, or where the summary asks. A
  click on it opens it on its fields; **Expand all** opens them all.
- A workflow belongs to no workspace, so every step starts with its
  **Workspace**: **Default** for the root and the projects at its top,
  then each workspace that holds a folder you see. The folder list
  next to it holds that workspace's folders with a job, every one you
  see, also the ones the sidebar leaves out: a checkout the
  workspace's `.gitignore` ignores, a folder below a project or a
  workspace with **Recursive** off, a folder hidden from the sidebar.
  A folder with no job isn't listed. Picking another
  workspace clears the folder and the job. A step you edit opens on the
  workspace of its folder.
- A **Command** picks its workspace, then its folder, the job and its
  arguments; **Workflows** in place of a workspace lists the workflows
  it may run. A **Prompt** takes its text, its workspace, **Asks in**,
  the folder Claude works in, and a directory inside it. A **Summary**
  takes the workspace, **Asks in** and the directory, and shows what
  Claude gets. The step keeps the folder alone, whose ID names its
  workspace.
- The grip at a card's left moves it: drag it, or focus it and press the
  up and down arrows. The card's [...] menu has **Move up**, **Move
  down**, **Disable** or **Enable**, **Duplicate** and **Remove**.
- **Save** sends the whole list. A step that misses something, a
  command without its job or a prompt without its text, opens with what
  is missing under it, and so does a step the server refuses, such as a
  directory Claude cannot work in.

The editor works the same at phone width: the fields stack under each
card. Anyone can make a workflow. Its owner, or an admin, edits or
deletes it; the runs so far keep the steps they had.

A step's args field goes to a job that takes arguments, after `--`.

A step runs a command on hosts of the [inventory](inventory.md) as
`@host:<name>` of the root, `_`, on the host, or `@host:*` on every host
with an ssh login, with the command as its args. With the `hosts`
permission, the folder list of the **Default** workspace has `_ hosts`,
and once you pick it the job picker lists them under **Hosts**, and the args field takes the command whole. Through
`/api/workflows`: `{"folder": "_", "job": "@host:*", "args":
["uptime"]}`. A step of `@host:*` runs on every host at once and ends ok
when every host's run did; running it takes the `hosts` permission.

---

Previous: [Inventory](inventory.md) · Next: [Runs](runs.md) · [Contents](../README.md)
