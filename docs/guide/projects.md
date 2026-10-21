# Projects and jobs

Below the views, the sidebar lists what there is to run, as a tree:

- a **project** (the branch icon) is a folder at the top of the root,
  most often a git checkout;
- under it, its own **jobs**, then the **folders** inside it (the folder
  icon), each with its jobs.

Every job has a play button, so a job is one click from anywhere; a job
that runs has an eye instead, to watch it. A project folds on its
chevron, and this browser remembers the folded ones. The project of the
page you are on stays open.

A [workspace](settings.md#workspaces) adds a level above its projects:
a label row with the workspace's name, the count of its projects and a
chevron that points down while it's open. Under it come the workspace's
own jobs, then its projects, each a project as above with a fold of its
own. Their links carry the workspace, as `/f/malta/dockerfiles`. Without
a workspace, the sidebar lists the projects alone.

![A folder: its jobs, files and environment](../assets/folder-dark.png)

An admin adds a project from a git repository on **Add project** (see
[Settings](settings.md#projects)). With a workspace, the form has a
**Workspace** select: none clones the project at the top of the root,
and a workspace clones it inside that workspace, as `malta/dockerfiles`
(see [Settings](settings.md#workspaces)).

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
folder lists that pipeline's jobs, and the files beside it are the
pipeline's business. Without one, the folder lists every source, and a job name
that two of them share gets its kind in front (`make:build`,
`npm:build`). A folder that is itself a `scripts` directory with a
pipeline also lists the scripts beside it that the pipeline does not
run, under their own names, such as a `bash.sh` with `exec bash` and
`# interactive: true`, a terminal to work in.

Every directory under the root that has something to run is a folder, as
deep as they go. Under a folder with a pipeline, the sidebar lists only folders with a
pipeline of their own, so the Makefiles of a Go tree do not
become folders; under a plain folder they do. Hidden directories, `node_modules`, `vendor` and
`testdata` are never looked into.

An agent sees the same list over [MCP](mcp.md): `list_jobs` returns the
folders, their jobs, whether its user may run each, and how each last
went.

## A folder's page

Open a folder from the sidebar:

- **Jobs**: a card each, with what it does, its first steps, how it last
  went, **Run** and **Run force**. **Run all** starts every job of the
  folder in the background. A folder with more than one source has a section
  per source.
- **Files**: what the folder holds, each with its size and when it
  changed; a directory opens.
- **Folder**: its path, its sources, who can run it, its hours, the
  files it includes, its schedules. **Show** beside a pipeline file opens
  it, coloured, in a window. A folder in a git checkout also shows the
  checkout's size: its files, and the history in `.git`.
- **Environment**: the dotenv files the pipeline reads (see
  [Settings](settings.md#env-files)).

## A job's page

![A job: what it runs, its history and schedules](../assets/job-dark.png)

Open a job from the sidebar or a card: the command it runs, its steps
and dependencies, who can run it and when, its **History** of runs, and
its **Scheduled** section (see [Scheduled](scheduled.md)).

The History opens with the job's activity: its runs of the last thirty
days by day and of the last 24 hours by hour, as on the [Runs](runs.md#activity)
page, and a line of how long its latest thirty runs took, a red mark at
each that failed. Hover a point for the run's time, outcome and
duration.

## Pulled before every run

A folder in a git checkout pulls it with `git pull --ff-only` before every
run, in the run's own terminal, so a job always runs the latest version.
A folder inside a checkout pulls the checkout. When the pull fails, the
run fails with it.

## Keys for a run

A run gets SSH keys only through an ssh agent served for that run alone:
the instance keys whose policies name the role of whoever started it, or
them by name (see [Settings](settings.md#ssh-keys)). A scheduled run gets
the keys of the schedule's creator. The job gets `SSH_AUTH_SOCK`, and
`ssh`, `git`, `rsync` and terraform's agent authentication use it. The
keys stay in memory and the agent ends with the run; the terminal says
which keys it loaded, or `(none)`. A run with no key has an empty
`SSH_AUTH_SOCK`, never the server's own.

`ssh` and `git` in a job also try the server's own keys in `~/.ssh`,
the identity of the host task-ui runs on, such as the key it pulls with.
terraform's agent authentication uses only the agent.

A checkout's `.ssh/` folder is not loaded. The SSH keys page warns about
one with private keys in it: upload a key the project needs as an
instance key and grant it by policy.

## Hours

An admin can limit a folder or a job to office hours (see
[Settings](settings.md#projects)). Out of hours **Run** fails.
**Run force** still works, but asks why: it needs a reason, keeps it
with the run, and every list of runs shows it.

![Out of hours, Run force asks why](../assets/out-of-hours-dark.png)

## Who may run what

An admin runs everything. A runner runs a folder's jobs unless its allow
list names other users or roles. A viewer runs nothing, but sees the jobs
and watches the runs. An admin sets which folders each user sees at all under
Settings, Users.

---

Previous: [Notes](notes.md) · Next: [Recents](recents.md) · [Contents](../README.md)
