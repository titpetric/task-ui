# Using task-ui

task-ui runs what your source trees declare, and lets everyone watch it
run. This guide walks through the screen the way the sidebar lists it:
the whole first, then one chapter per section, top to bottom.

![The dashboard, with the sidebar on the left](assets/dashboard-dark.png)

## The screen

The page has three parts.

- **The header.** The name and its prompt logo go home to the dashboard.
  Your initials, on the right, open the user menu: who you are, your role,
  **Settings** for admins, and **Sign out**. On a phone a menu button on
  the left opens the sidebar as a drawer.
- **The sidebar.** Everything you can open, in a fixed order:
  1. the views across all folders: **Dashboard**, **Workflows** (with the
     workflows listed under it), **Runs**, **Artifacts** and **Scheduled**;
  2. the **projects**, each a folder at the top of the root, with their
     jobs and the folders inside them, as a tree;
  3. **Recents**, the latest runs and workflow runs by anyone.
- **The page.** What you opened in the sidebar. A run you start or attach
  to opens in a **window** over the page: a terminal on the run, which
  you close without ending the run.

Three ideas carry the rest of the guide.

- **A job** is a target a folder declares: an atkins or Taskfile job, a
  Makefile target, a package script, a file under `scripts/`. task-ui
  reads the files and lists what they declare; it never writes them.
- **A run** is a job running on the server, on a terminal of its own. It
  belongs to no window: anyone allowed attaches to it, and it goes on when
  you close every window. One job runs once at a time across the
  instance; different jobs run side by side.
- **A role** decides what you can do. Viewers watch and read history,
  runners start the jobs the allow lists let them, admins also manage the
  instance under Settings.

Every list marks a job, a workflow or a run with `#!`: faint
until it has run, then in the colour of how it last ended, green for ok
and red for failed or killed, and pulsing amber while it runs.

## Chapters

| Chapter                              | The sidebar section                    | What it covers                                              |
|--------------------------------------|----------------------------------------|-------------------------------------------------------------|
| [Dashboard](guide/dashboard.md)      | Dashboard                              | The terminals running now, and the control centre           |
| [Workflows](guide/workflows.md)      | Workflows, and the list under it       | Chains of jobs, run with one button                         |
| [Runs](guide/runs.md)                | Runs                                   | The run window, the history, a run's page and its output    |
| [Artifacts](guide/artifacts.md)      | Artifacts                              | Files the runs leave behind                                 |
| [Scheduled](guide/scheduled.md)      | Scheduled                              | Jobs on a cron expression                                   |
| [Projects and jobs](guide/projects.md) | The tree of projects, folders, jobs  | Folders, the conventions, a job's page, pulls, keys, hours  |
| [Recents](guide/recents.md)          | Recents                                | The latest runs, by anyone                                  |
| [Settings](guide/settings.md)        | The user menu, Settings                | Projects, users, secrets, SSH keys, env files               |
| [MCP](guide/mcp.md)                  |                                        | Tokens, and the tools an agent runs jobs with               |
| [Screenshots](screenshots.md)        |                                        | Every view, in both themes and on a phone                   |

Beside the guide:

- [Running the server](running.md): the Docker image, the environment
  variables, the password file and the server commands;
- [API](api.md): the JSON endpoints and the MCP server;
- [Development](development.md): the packages, the lint and test bar, and
  the screenshots.
