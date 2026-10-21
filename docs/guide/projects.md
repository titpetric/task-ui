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

A [workspace](workspaces.md) adds a level above its projects: a label
row with the workspace's name, and its projects folded under it. Their
links carry the workspace, as `/f/malta/dockerfiles`. The projects at
the top of the root are in the default workspace, `_`, and list first,
without a label row; their links carry `_`, as `/f/_/infra`.

![A folder: its jobs, files and environment](../assets/folder-dark.png)

An admin adds a project from a git repository on **Add project** (see
[Settings](settings.md#projects)). Its **Workspace** select clones the
project at the top of the root, `_`, or inside a named workspace (see
[Workspaces](workspaces.md#clone-a-project-into-one)).

## Move a project to another workspace

A project's configure page, **Settings**, **Projects**, then the
project, has the same **Workspace** select, on the workspace the
project is in. Pick another one and **Save** moves it: `web` in `_`
becomes `acme/web` in `acme`, and `acme/web` picked back to `_` becomes
`web`. Before the move a dialog names the new ID and every row that
stays behind.

What moves and what stays:

- the checkout moves, renamed in place, with the folders inside it;
- the project row moves, with its title, allow lists and hours;
- the users whose folder list names the project, or a folder inside it,
  get the new ID in its place, so access follows the project;
- the runs stay under the old ID: their pages, recordings and artifacts
  still open from the Runs page;
- schedules, secrets, env files and workflow steps that name the old ID
  stay behind. The dialog lists them, as does `left` in the API's
  answer; change them to the new ID, or move the project back, to
  bring them along.

A move refuses with `409` while a run of the project, or of a folder
inside it, is going, and while the project is moving already. It also
refuses a workspace itself, a workspace that does not exist, an ID
that a project, a workspace or a directory holds, and a workspace on
another file system, as a move renames and does not copy. Only a
project at the top of a workspace moves; a folder inside a project
stays with it.

To move one through the API, name the workspace on the configure
request:

```sh
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -d '{"title": "Web", "workspace": "acme"}' \
  https://task.example.com/api/projects/web
```

The request configures as well, so send the fields the project has.
It responds with the project row under its new ID, and `left`: each
row that still names the old ID, as `{kind, id, folder, name}`.

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
- **Files**: what the folder holds, dot entries too, each with its size
  and when it changed, and `..` up a level. A directory opens as a page
  of its own; a file opens in a window, an image shown, markdown
  rendered, text coloured and cut past 1 MiB, with **Download** beside
  and in the row.
  A file that runs, executable or with a shebang on its first line,
  carries `#!` and **exec**, and **Run** for those the folder lets run:
  it starts the file as the job `./<path>`, from the folder, and opens
  the run's window. Its history is at the file's path under the folder.
- **Sealed entries**: `.git`, `.env`, `.env.*`, `.htpasswd` and
  `.htaccess` list with a lock and open nothing, nor does a path through
  them, on the page or over the API. Nor does a symlink that leads out
  of the checkout. Whoever has the server reaches them there, and so
  does a job, which runs as the server's user.
- **Folder**: its path, its sources, who can run it, its hours, the
  files it includes, its schedules. **Show** beside a pipeline file opens
  it, coloured, in a window. A folder in a git checkout also shows the
  checkout's size: its files, and the history in `.git`.
- **Environment**: the dotenv files the pipeline reads (see
  [Settings](settings.md#env-files)).

![A project's Files: .git and .env sealed, ./serve marked to run](../assets/files-dark.png)

![A file of the checkout in the preview window](../assets/file-dark.png)

A link to the folder's page with `#file=<name>`, such as
`/f/acme/api#file=README.md`, opens that file of its Files in the window
on load.

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
run fails with it. A project in a workspace pulls the workspace's
checkout first (see [Workspaces](workspaces.md#runs-in-a-workspace)).

## Keys for a run

A run gets SSH keys only through an ssh agent served for that run alone:
the instance keys whose policies name the role of whoever started it, or
them by name (see [Settings](settings.md#ssh-keys)), and the SSH keys of
its workspace that its grants let through (see
[Workspaces](workspaces.md#keys)). A scheduled run gets
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

Previous: [Notes](notes.md) · Next: [Workspaces](workspaces.md) · [Contents](../README.md)
