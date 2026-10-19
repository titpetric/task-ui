# Projects and jobs

Below the views, the sidebar lists what there is to run, as a tree:

- a **project** (the branch icon) is a folder at the top of the root,
  usually a git checkout;
- under it, its own **jobs**, then the **folders** inside it (the folder
  icon), each with its jobs.

Every job has a play button, so a job is one click from anywhere; a job
that runs has an eye instead, to watch it. A project folds on its
chevron, and this browser remembers the folded ones. The project of the
page you are on stays open.

![A folder: its jobs, files and environment](../assets/folder-dark.png)

## Where the jobs come from

task-ui reads the files a folder has and lists their targets. It never
changes them.

| Source   | File                                  | A job is                                              |
|----------|---------------------------------------|-------------------------------------------------------|
| atkins   | `atkins.yml`, `.atkins.yml`           | a job with a `desc`; `show: false` hides one          |
| task     | `Taskfile.yml`                        | a task with a `desc`; `internal: true` hides one      |
| make     | `Makefile`, `makefile`, `GNUmakefile` | a target; `target: ## text` or a comment describes it |
| npm      | `package.json`                        | a script, run with npm, pnpm, yarn or bun by the lock file |
| composer | `composer.json`                       | a script, described by `scripts-descriptions`         |
| script   | `scripts/*`                           | a file; its first comment describes it, `# interactive: true` makes it take input |

A pipeline file, `atkins.yml` or a Taskfile, speaks for its folder: the
folder lists that pipeline's jobs, and what is beside it is the
pipeline's business. Without one, every source is listed, and a job name
that two of them share gets its kind in front (`make:build`,
`npm:build`). A folder that is itself a `scripts` directory with a
pipeline also lists the scripts beside it that the pipeline does not
run, under their own names, such as a `bash.sh` with `exec bash` and
`# interactive: true`, a terminal to work in.

Every directory under the root that has something to run is a folder, as
deep as they go. Hidden directories, `node_modules`, `vendor` and
`testdata` are never looked into.

## A folder's page

Open a folder from the sidebar:

- **Jobs**: a card each, with what it does, its first steps, how it last
  went, **Run** and **Run force**. **Run all** starts every job of the
  folder in the background. A folder with several sources has a section
  per source.
- **Files**: the pipeline's files; **Show** opens one, coloured, in a
  window.
- **Folder**: its path, its sources, who can run it, its hours, the
  files it includes, its schedules.
- **Environment**: the dotenv files the pipeline reads (see
  [Settings](settings.md#env-files)).

## A job's page

![A job: what it runs, its history and schedules](../assets/job-dark.png)

Open a job from the sidebar or a card: the command it runs, its steps
and dependencies, who can run it and when, its **History** of runs, and
its **Scheduled** section (see [Scheduled](scheduled.md)).

The History opens with the job's activity: its runs of the last thirty
days by day and by hour of the day, as on the [Runs](runs.md#activity)
page, and a line of how long its latest thirty runs took, a red mark at
each that failed. Hover a point for the run's time, outcome and
duration.

## Pulled before every run

A folder in a git checkout pulls it with `git pull --ff-only` before every
run, in the run's own terminal, so a job always runs the latest version.
A folder inside a checkout pulls the checkout. A pull that fails fails
the run.

## Keys for a run

A run gets SSH keys only through an ssh agent served for that run alone:
the instance keys whose policies name the role of whoever started it, or
them by name (see [Settings](settings.md#ssh-keys)). A scheduled run gets
the keys of the schedule's creator. The job gets `SSH_AUTH_SOCK`, and
`ssh`, `git`, `rsync` and terraform's agent authentication use it. The
keys stay in memory and the agent ends with the run; the terminal says
which keys it loaded, or `(none)`. A run with no key has an empty
`SSH_AUTH_SOCK`, never the server's own.

`git` in a job uses `GIT_SSH_COMMAND` with `IdentityFile=none`, so it
offers the agent's keys and no key file, and the instance `known_hosts`.

A checkout's `.ssh/` folder is not loaded. The SSH keys page warns about
one with private keys in it: upload a key the project needs as an
instance key and give it by policy.

## Hours

A folder or a job can be limited to office hours, set by an admin (see
[Settings](settings.md#projects)). Out of hours **Run** is refused.
**Run force** still works, but asks why: the reason is required, kept
with the run, and shown wherever the run is listed.

![Out of hours, Run force asks why](../assets/out-of-hours-dark.png)

## Who may run what

An admin runs everything. A runner runs a folder's jobs unless its allow
list names other users or roles. A viewer runs nothing, but sees the jobs
and watches the runs. Which folders each user sees at all is set under
Settings, Users.

---

Previous: [Scheduled](scheduled.md) · Next: [Recents](recents.md) · [Contents](../README.md)
