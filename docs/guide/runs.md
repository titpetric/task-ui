# Runs

Every run, by anyone, newest first. A dot beside **Runs** in the sidebar
pulses while something runs.

![The runs](../assets/runs-dark.png)

## Activity

Above the list, two charts, in the server's time zone, following the
filter:

- **Last 30 days**, a bar a day, today on the right;
- **Last 24 hours**, a bar an hour, this hour on the right, with a time
  marker every six hours.

Each chart's head reads `min/avg/max/total:`, the fastest, the mean and
the slowest finished run of its window and their time together.

A bar is as tall as its runs against the busiest bar, split into the runs
that ended ok (amber), failed or got killed (red), and the rest, running
or stopped (faint). Hover a bar for its counts. Under each, the same days or
hours as a **line graph of their durations**: the average in the text
colour, the slowest dimmer and the fastest faint, against the slowest run
of the chart, a smooth curve through the middle of each column. A day or
an hour without a finished run sits on the baseline. Hover a column for the average, the
fastest, the slowest and how many runs they cover. A line above shows the
run count, how many failed and the median duration.

## The list

Each row is a run: when it started, the folder and job, who started it,
how long it took and how it ended. **Filter** narrows it by folder, job
and state. A run in progress has **Attach**, a finished one **View**.

With the Claude Code integration set up for you (see
[Settings](settings.md#integrations)), a finished run also has
**Summary**: Claude reads the end of its output, up to 2000 lines, and
writes what it did or why it failed, in a window. task-ui keeps the
summary by run, so the next **Summary** of that run, by anyone, shows
the kept one at once, without asking Claude again.

## The run window

![A run in its window](../assets/run-dark.png)

Starting a job, or attaching to a run, opens the run window over the
page: the run's terminal, live.

- **The run goes on without the window.** Closing it detaches; the run
  keeps running, and the sidebar, the dashboard or this list open a window
  on it again. A window that comes back gets the tail of what it missed.
- **Runners type into it.** Input goes to the job, and the terminal takes
  the window's size. A viewer can only watch in the window.
- **Stop** signals the job to end (SIGHUP and SIGTERM, then SIGKILL after
  `TASK_UI_GRACE`, ten seconds) and records it as `stopped`. **Kill** ends it now, as
  `killed`.
- **Run again** starts the job once more when it has ended.

A finished run stays attachable, its last frame in memory, for
`TASK_UI_KEEP`, fifteen minutes; after that it opens as history: its
whole recording played into the window, with how it ended, and **Run
again** for whoever may run it.

### One run per job

A job runs once at a time across the instance. Starting a job that runs
already attaches to the run in the way. **Run force** kills that run and
starts again; out of a folder's hours it also asks why, and keeps the
reason with the run (see [Projects and jobs](projects.md#hours)).

### From an agent

An agent starts runs over [MCP](mcp.md) as the user of its token:
`run_job` starts one and can wait for it to end, `list_runs` and
`get_run` read the history and the end of a run's output. Such a run is
like any other: this list shows it under that user, and anyone allowed
attaches to it.

## A run's page

![A run's page](../assets/run-page-dark.png)

`/runs/<id>` is one run, to link to:

- **Output**: every line the run printed, in its colours, from the first
  to the last, scrolling inside the block. The block replays the recording,
  so what redraws itself, a progress bar or a pipeline tree, shows its
  last frame.
- **Run**: folder, job, who and when, how long, how it ended, and the
  reason for a run forced out of hours.
- **Artifacts**: the files the run left (see [Artifacts](artifacts.md)).
- **Recording** downloads the run as an asciicast file, for `asciinema
  play`.

While the run is in progress the page shows its terminal as it looks now,
and **Attach**, **Stop** and **Kill** instead.

## What a job gets

A run starts in its folder, as a login shell, with:

- `TASK_UI=1`, `TASK_UI_USER`, `TASK_UI_FOLDER`, `TASK_UI_JOB`,
  `TASK_UI_RUN_ID`, and `TASK_UI_ARTIFACTS`, a directory for its files;
- the secrets of its folder (see [Settings](settings.md#secrets));
- an ssh agent, through `SSH_AUTH_SOCK`, with the instance keys whose
  policies name the role or the name of whoever started the run (see
  [Projects and jobs](projects.md#keys-for-a-run)); the terminal says
  which, or `(none)`.

Before the job, a folder in a git checkout pulls it, in the run's own
terminal; when the pull fails, the run fails with it.

---

Previous: [Workflows](workflows.md) · Next: [Artifacts](artifacts.md) · [Contents](../README.md)
