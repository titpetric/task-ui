# Workflows

A workflow is a named chain of jobs from any folders: a one button
deploy, such as build, test, push, plan, apply. The sidebar lists the
workflows under **Workflows**, each with a play button.

![Workflows](../assets/workflows-dark.png)

## Running one

**Run** starts the first step. Each next step starts when the one before
it ended ok. The first step that does not end ok ends the workflow, and
the rest do not run. While a workflow runs, the play button in the
sidebar becomes an eye: open it to watch.

A step is a job of a folder, or another workflow, which then runs in its
place, all its steps.

Running a workflow needs the right to run every step in it. A workflow
runs once at a time; task-ui turns down a second start while it runs.

An agent runs one over [MCP](mcp.md) by name: `run_workflow` starts it
and can wait for the end, then says how each step went;
`list_workflows` and `list_workflow_runs` show the chains and their
last runs.

## The workflow page

![A workflow: its chain and its runs](../assets/workflow-dark.png)

- **Pipeline**: the steps in the order they run. While a run is going, it
  follows that run live: each step's state and terminal.
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
order, picked from the jobs you can see and the other workflows. Anyone
can make one. Its owner, or an admin, edits or deletes it; the runs so far
keep the steps they had.

---

Previous: [Inventory](inventory.md) · Next: [Runs](runs.md) · [Contents](../README.md)
