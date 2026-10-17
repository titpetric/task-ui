# Plan: task-ui as a process manager

task-ui is a push button way to run what a source tree declares, and to
watch it run. This rewrite makes runs first class: they live in a
database, they outlive the window that started them, anyone allowed may
attach, and the last frame of every terminal is one URL away.

## Packages

Following the oida rules: one struct per file, data in the `model` of its
area, dependencies one way, the front end makes no external requests.

| Package      | Holds                                                                      |
|--------------|----------------------------------------------------------------------------|
| `schema`     | the SQL migrations, the source of truth of the data model (mig)            |
| `model`      | the mig generated types and the rules on them; imports nothing of ours     |
| `storage`    | CRUD per table over pdo, request scoped clients over one pool             |
| `detector`   | the conventions a folder follows, unchanged                                |
| `folder`     | the catalogue: folders and jobs, now configured by project rows            |
| `runner`     | sessions: a process on a pty, its terminal state, subscribers, artifacts  |
| `scheduler`  | cron expressions and the loop that starts scheduled runs                   |
| `access`     | roles, passwords, sessions; users live in storage                          |
| `git`, `sshkey`, `dotenv` | unchanged services                                             |
| `api`        | the JSON endpoints: runs, frames, artifacts, schedules, projects, users    |
| `frontend`   | the pages, templates and assets, and the attach socket                     |
| `internal/assert` | the test assertions, copied from atkins                               |

## Sweeps

Each sweep ends green: `go test -race ./...`, `splint ./...` with no
findings, a commit.

1. **Foundation.** Go 1.27, pdo and mig, migrations for user, project, run,
   artifact, schedule, secret, env_file; generated model; storage with
   Create, Get, List, Update, Delete per table and the queries the app
   needs; black box tests, parallel, one in-memory database each.
2. **Runner.** A session is a process on a pty with a terminal emulator
   behind it, so the screen is known at any moment. Sessions are detached:
   a window attaches and detaches, the run goes on. Attach is read only or
   with input, by permission. Stop and Kill end a run; closing a window
   does not. The last frame renders as SVG and text. Artifacts are
   collected from `$TASK_UI_ARTIFACTS` when a run ends.
3. **Scheduler.** Cron expressions, next run computation, a loop that
   starts due schedules as detached runs under the user who scheduled them.
4. **API.** JSON for everything the pages show: runs and their frames,
   artifacts, schedules, projects, users, secrets.
5. **Front end.** The sidebar as a process manager: Folders, Runs,
   Dashboard, Artifacts, Scheduled; the folders below; Recents with user
   badges. Run all on a folder. Attach from anywhere. The dashboard of
   last frames. Settings gain Users.
6. **Docs and demo.** README, screenshots, the pull request.

## Rules

- Black box tests: `package x_test`, `t.Parallel()`, no shared state.
  Every exported symbol, constructors included, is invoked from a test and
  asserted on.
- Functions stay small: one decision per function where it can be.
- No benchmarks.
