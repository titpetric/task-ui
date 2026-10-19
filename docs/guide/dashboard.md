# Dashboard

The first item in the sidebar, and where the name in the header goes. It
is the instance at a glance, the way a control room shows its screens.

![The dashboard](../assets/dashboard-dark.png)

## The numbers

A row of five: how many folders and jobs there are, how many runs are
running, how many schedules are set, and your role.

## Running now

Every run in progress as a tile: the folder and job, how many windows
are attached, who started it, and its terminal as it looks right now,
redrawn every two seconds. Click a tile to open a window on the run:
runners can type into it and end it, viewers watch.

The tiles follow the runs as they happen, with no reload: a run that
starts adds its tile, one that ends marks it. When nothing runs, the
section says so.

Each tile's terminal is also a URL of its own, `/api/runs/{id}/frame.svg`
as a picture or `/api/runs/{id}/frame.txt` as text, for a wall screen or
a script.

## Control centre

Every folder's jobs in one table, a button each. A job name that repeats
across folders gets a column of its own (`build`, `push`, `plan`), so the
same step of every project lines up; the names that do not repeat are
under **Other**. A folder with several sources shows the kind on the
button (`make`, `npm`) where a name is there more than once.

A button runs the job, or, while it runs, attaches to it. Its `#!` mark
says how the job last ended.

---

Next: [Workflows](workflows.md) · Back to the [contents](../README.md)
