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

A step is a job of a folder, or another workflow, which then runs in its
place, all its steps. In the editor, a switch leads each step. Switch
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

A step of `/api/workflows` runs a command on hosts of the
[inventory](inventory.md) as `@host` of the root, `_`, with a `target`,
a host or `*` for every host with an ssh login, and the command as its
`args`: `{"folder": "_", "job": "@host", "target": "*", "args":
["uptime"]}`. A step on `*` runs on every host at once and ends ok when
every host's run did; running it takes the `hosts` permission.

---

Previous: [Inventory](inventory.md) · Next: [Runs](runs.md) · [Contents](../README.md)
