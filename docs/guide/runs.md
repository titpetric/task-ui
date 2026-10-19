# Runs

Every run, by anyone, newest first. A dot beside **Runs** in the sidebar
pulses while something runs.

![The runs](../assets/runs-dark.png)

## Activity

Above the list, two charts of the runs of the last thirty days, in the
server's time zone, following the filter:

- **by day**, a bar a day, today on the right;
- **by hour of the day**, over the same days, so a schedule at seven
  stands out from the work spread over the afternoon.

A bar is as tall as its runs against the busiest bar, split into the runs
that ended ok (amber), failed or were killed (red), and the rest, running
or stopped (faint). Hover a bar for its counts. Under each, a grey row
of the same days or hours is the **average duration** of their finished
runs against the slowest, so a day or an hour that ran slow stands out
however many runs it had; hover one for the average and how many runs it
is over. A line above says how
many runs there were, how many failed and the median duration.

## The list

Each row is a run: when it started, the folder and job, who started it,
how long it took and how it ended. **Filter** narrows it by folder, job
and state. A run in progress has **Attach**, a finished one **View**.

## The run window

![A run in its window](../assets/run-dark.png)

Starting a job, or attaching to a run, opens the run window over the
page: the run's terminal, live.

- **The run goes on without the window.** Closing it detaches; the run
  keeps running, and the sidebar, the dashboard or this list open a window
  on it again. A window that was away gets the tail of what it missed.
- **Runners type into it.** Input goes to the job, and the terminal takes
  the window's size. A viewer's window is read only.
- **Stop** asks the job to end (SIGHUP and SIGTERM, then SIGKILL after a
  grace period) and records it as `stopped`. **Kill** ends it now, as
  `killed`.
- **Run again** starts the job once more when it has ended.

A finished run opens as history: its whole recording played into the
window, with how it ended, and **Run again** for whoever may run it.

### One run per job

A job runs once at a time across the instance. Starting a job that runs
already attaches to the run in the way. **Run force** kills that run and
starts again; out of a folder's hours it also asks why, and keeps the
reason with the run (see [Projects and jobs](projects.md#hours)).

## A run's page

![A run's page](../assets/run-page-dark.png)

`/runs/<id>` is one run, to link to:

- **Output**: every line the run printed, in its colours, from the first
  to the last, scrolling inside the block. It is the recording replayed,
  so what redraws itself, a progress bar or a pipeline tree, shows as it
  was last drawn.
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
- an ssh agent, through `SSH_AUTH_SOCK`, with the keys of its project's
  `.ssh/` folder (see [Projects and jobs](projects.md#keys-from-the-project))
  and the instance keys whose policies name the role of whoever started
  the run (see [Settings](settings.md#ssh-keys)); the terminal says which.

Before the job, a folder in a git checkout pulls it, in the run's own
terminal; a pull that fails fails the run.

---

Previous: [Workflows](workflows.md) · Next: [Artifacts](artifacts.md) · [Contents](../README.md)
