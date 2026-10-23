# Workflows

A workflow is a named chain of jobs from any folders: a one button
deploy, such as build, test, push, plan, apply. It can end with a
prompt to Claude, which reads what the steps printed, or be a prompt
alone. The sidebar lists the workflows under **Workflows**, each with a
play button.

![Workflows](../assets/workflows-dark.png)

## Running one

**Run** starts the first step. Each next step starts when the one before
it ended ok. The first step that does not end ok ends the workflow, and
the rest do not run. While a workflow runs, the play button in the
sidebar becomes an eye: open it to watch.

A step is a job of a folder, another workflow, which then runs in its
place, all its steps, or a prompt to Claude (see [Prompt
steps](#prompt-steps)). In the editor, a switch leads each step. Switch
a step off and it collapses to its name, folder / job, dimmed; it stays
in the chain and a run skips it: its row reads skipped, and the steps
after it run. Switch it on and its selects come back.

Running a workflow needs the right to run every step in it. A workflow
runs once at a time; task-ui turns down a second start while it runs.

An agent runs one over [MCP](mcp.md) by name: `run_workflow` starts it,
with `input` for its prompt, and can wait for the end, then says how
each step went; `list_workflows` and `list_workflow_runs` show the
chains and their last runs.

## A workflow with a prompt

A workflow's **Prompt** is what Claude gets once the steps end, as the
user who started the run, in the folder **Run in** names, the root by
default. In the editor, a switch leads the **Prompt** heading: on, it
shows the prompt, **Summarize** and **Run in**; off, only the heading
stays, and the workflow saves with no prompt. The text stays in the
form until you leave the page, so switching back on before you save
restores it. A new workflow starts with the prompt off. Three shapes
are valid:

- **Steps alone**: the chain as above.
- **Steps and a prompt**: Claude reads the steps' output and answers, a
  summary or work on what they printed. **Summarize** under the prompt
  fills it with the short summary of **Run and summarize**, so the
  workflow summarises itself on every run.
- **A prompt alone**: no steps, or every step toggled off; the run is
  the prompt.

task-ui refuses a workflow with no step to run and no prompt.

Claude gets the workflow's name, the prompt, the run's input, and each
step: its place in the chain, its folder and job, how it ended, its exit
code, how long it took and the end of its output, colours stripped. The
steps share the output budget of a run's prompt, the newest step first,
and a cut output says so. A step toggled off reads skipped. Claude changes nothing unless
the prompt names a change, as with a prompt on a run's page.

**Run** on a workflow with a prompt opens a field for input first, optional: Claude
reads it after the prompt, under `Input:`. The run stays running while
Claude works on the prompt. It ends with the steps' state, or failed
when the prompt fails after steps that ended ok. **Stop** at the prompt
stops Claude's run too, and the run ends stopped.

The prompt's thread shows on the workflow run's page, live while Claude
works: the prompt and the input as the question, Claude's answer under
it. A reply in the thread asks again with the same steps. Making a
workflow with a prompt, and running one, takes the `prompt/edit`
permission.

## Prompt steps

A prompt step asks Claude in place of running a job: **Add prompt
step** in the editor, beside **Add step**, adds one, a field for the
prompt and **Run in**, the folder Claude works in, the root by default.
It runs `@claude-prompt` there as the user who started the run, with
the run's input and the output of the steps before it, as the final
prompt gets them, within the same budget. Claude's reply is the step's
output: the prompt steps after it and the final prompt read it as the
step's answer. The step ends ok when Claude's run ends ok; one that
fails ends the workflow, as a job does.

The pipeline shows a prompt step with the prompt icon, as a prompt in
its folder, its prompt under it, and its run once it ran. Its switch
works as a job step's: off, a run skips it. Its thread lists on the
workflow run's page with the final prompt's. Saving a prompt step, and
running a workflow with one, takes the `prompt/edit` permission.

Through `/api/workflows`, a prompt step is `{"prompt": "fetch the
weather", "folder": "_", "dir": ""}`: `folder` the folder Claude runs
in, `_` or empty for the root, and `dir` a directory inside it.

The final prompt stays the last prompt: the one that reads every step.

## Workflow context

Every prompt a workflow runs, a prompt step's and the final prompt,
starts with a **Workflow context** block, before the steps' output: the
workflow's name and link, the run's number, when it started, who
started it and its input, and what the last 3 finished runs of the
workflow replied, newest first, each with its date and how the run
ended. A run's answer is the reply to its final prompt, or for a
workflow without one, the reply to its last prompt step. The block
skips a run without one, and it does not count. Each keeps its first
1.5 KiB, and a cut one says so; the block takes its share of the run's
output budget. The first run of a workflow reads that there are no
earlier runs. So a prompt can stay consistent with what it said before,
such as "yesterday I said water tomorrow". The editor says so under
**Prompt**: Claude also gets the last 3 answers of this workflow.

## The workflow page

![A workflow: its chain and its runs](../assets/workflow-dark.png)

- **Pipeline**: the steps in the order they run, a step toggled off
  dimmed, and the prompt under them. While a run is going, it follows
  that run live: each step's state and terminal.
- **Runs**: the workflow's activity, its runs of the last thirty days by
  day and of the last 24 hours by hour, and the duration of its latest thirty (see
  [Runs](runs.md#activity)), then every run of the workflow, newest first, with who started it,
  how it ended, and each step's run, which opens like any other run.
  **View** opens that run's pipeline in a window, each step with its
  state and the screen it ended on; Ctrl and a click open it as a page.
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

**Add workflow** on the Workflows page takes a name and the steps, in
order, picked from the jobs you can see and the other workflows, each
with a toggle that runs it or skips it; with `prompt/edit`, the
**Prompt**, its **Summarize** preset and **Run in**, the folders you see.
Anyone can make one. Its owner, or an admin, edits or deletes it; the runs so far
keep the steps they had.

A step's args field goes to a job that takes arguments, after `--`.

A step runs a command on hosts of the [inventory](inventory.md) as
`@host:<name>` of the root, `_`, on the host, or `@host:*` on every host
with an ssh login, with the command as its args. With the `hosts`
permission, the step's job picker lists them under **Hosts** once you
pick the root `_`, and the args field takes the command whole. Through
`/api/workflows`: `{"folder": "_", "job": "@host:*", "args":
["uptime"]}`. A step of `@host:*` runs on every host at once and ends ok
when every host's run did; running it takes the `hosts` permission.

---

Previous: [Inventory](inventory.md) · Next: [Runs](runs.md) · [Contents](../README.md)
