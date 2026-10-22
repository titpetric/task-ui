# Dashboard

The first item in the sidebar, and where the name in the header goes. It
is the instance at a glance, the way a control room shows its screens.

![The dashboard](../assets/dashboard-dark.png)

## The numbers

A row of five: how many folders and jobs there are, how many runs are
running, how many schedules exist, and your role.

## Running now

Every run in progress as a tile: the folder and job, how many windows
watch it, who started it, and its terminal as it looks right now,
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
same step of every project lines up, `default` first; the names that do not repeat sit
behind the ⋯ button under **Other**, a menu of the folder's other jobs
with the action of each. The button takes the running colour while one
of them runs. A folder with more than one source shows the kind on the
button (`make`, `npm`) where a name is there more than once.

A button runs the job, or, while it runs, attaches to it. A role that
may not run the job sees a `#!` mark that links to the job's page. The
button and the mark take the colour of the job's last run: green when it
succeeded, red when it failed or someone stopped or killed it, grey when
the job has no runs. A run that ends recolours its button with no reload.

Under the pointer the whole column lights up, so one job reads down every
folder at once.

---

Next: [Inventory](inventory.md) · Back to the [contents](../README.md)
