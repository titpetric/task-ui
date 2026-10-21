# Workspaces

A workspace is a checkout under the root that holds projects. It carries
what an organisation's projects share, such as an `atkins.yml` with
`.atkins/`, a `CLAUDE.md` with `.claude/`, and its policies, and the
projects clone inside it. You name its folder, and the folder is the
namespace of its projects: `dockerfiles` cloned into the workspace
`malta` lives at `<root>/malta/dockerfiles`, its folder ID is
`malta/dockerfiles`, and it opens at `/f/malta/dockerfiles`.

Every project is in a workspace. The **default workspace**, `_`, is the
root itself: it holds the projects cloned at the top of the root, and
their pages open under `/f/_/`, so `infra` opens at `/f/_/infra`. A link
without the `_`, such as a bookmarked `/f/infra`, redirects there for
good, its query kept. `/f/_` lists the projects at the root you can see.
task-ui adds the default workspace when it starts, and nobody can
remove it. No workspace, project or directory under the root takes the
name `_`, nor `internal`; discovery skips a directory of the root with
either name.

![The sidebar with the workspace acme open, its projects api and web under it](../assets/sidebar-workspace-dark.png)

## Tooling from the folders above

A run in a workspace keeps the server's `HOME`. The tools a job calls
find the workspace's files by looking in the folders above the project:

- **Claude Code** reads the `CLAUDE.md` and `.claude/` of the folders
  above its working directory, so a run in `malta/dockerfiles` reads the
  workspace's as well as the project's own.
- **atkins** looks for its configuration from the working directory up,
  and stops at the first folder with an `atkins.yml`, an `.atkins.yml` or
  an `.atkins/` folder. A project without one uses the workspace's
  pipeline and its `.atkins/skills`. A project with its own `atkins.yml`
  stops there: atkins reads the skills in the project's `.atkins/skills`
  and in `$HOME/.atkins/skills`, and not the workspace's.

## Add, edit and delete

Admins manage workspaces under Settings, **Workspaces**. The page lists
each workspace with its repository and the projects cloned inside it, in
the sidebar's order.

![Workspaces, each with its repository and projects](../assets/workspaces-dark.png)

- **Add workspace** clones a repository into the root, under the folder
  you name. The clone works, or the page shows git's output and keeps
  nothing. task-ui refuses a folder that a workspace, a project or a
  directory under the root holds already, and `internal` and `_`, its
  own.
- Any number of workspaces may come from one repository, each under its
  own folder: two teams can start from the same tooling and clone
  different projects into it.
- **Configure** changes a workspace's title and description; the
  repository and the folder stay.
- **Move up** and **Move down** change the order of the workspaces, in
  the sidebar and on the page. An added workspace goes last. The default
  workspace lists first and keeps its place.
- The remove button deletes the checkout, every project inside it and
  their rows, and its confirm names those projects first. The default
  workspace has none: `DELETE /api/workspaces/_` answers `409`.

## Clone a project into one

**Add project** under Settings, Projects has a **Workspace** select.
`_`, the default, clones the project at the top of the root; a named
workspace clones it inside, as `malta/dockerfiles`. The Projects page
lists it under that name. In the API an empty `workspace` means `_`. A
project cloned at the root may not take a workspace's name or a
directory that's there already: that answers `409` before git runs.

The project lists with the folders inside it whatever the workspace's
`.gitignore` says, so a workspace that ignores its projects' folders, as
one that clones them itself should, still shows them.

The API does the same, with a token from `task-ui token`:

```sh
curl -s -X POST https://task.example.com/api/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url": "git@example.com:org/workspace.git", "folder": "malta", "title": "Malta"}'
curl -s -X POST https://task.example.com/api/projects \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url": "git@example.com:org/dockerfiles.git", "workspace": "malta"}'
```

A taken folder answers `409`. `PUT /api/workspaces/malta` with
`{title, description}` changes its text, and `DELETE /api/workspaces/malta`
removes it. A `sort` in the `PUT` places the workspace: the lowest comes
first, and workspaces of one sort go by folder. Left out, the sort
stays.

## In the sidebar

The projects at the top of the root come first, by name, and the
workspaces follow them, in their order. A workspace adds a level above
its projects: a label row with the
workspace's name, the count of its projects and a chevron that points
down while it's open. Under it come the workspace's own jobs, then its
projects, each with a fold of its own. This browser remembers each fold
apart, so `malta` and `malta/dockerfiles` open and close on their own.
Without a workspace, the sidebar lists the projects alone (see
[Projects and jobs](projects.md)).

## Runs in a workspace

A run of a job in the workspace or in one of its projects gets two more
variables (see [Runs](runs.md#what-a-job-gets)):

- `TASK_UI_WORKSPACE`, the workspace's folder, such as `malta`, or `_`
  for a project at the root;
- `TASK_UI_WORKSPACE_DIR`, its checkout, such as `<root>/malta`, or the
  root for `_`.

Before the job, the run pulls a named workspace's checkout, then the
project's, so the shared tooling is current when the job starts. When
either pull fails, the run fails with it. The default workspace's root
isn't pulled.

## A prompt against a workspace

A workspace's folder page, `/f/malta`, carries a **Prompt** block for an
admin once the [Claude Code integration](settings.md#integrations) has a
token for them. The prompt runs `internal/claude-prompt` as you, and its
window opens on the run.

![The Prompt block on a workspace's page](../assets/workspace-prompt-dark.png)

Claude Code works in the workspace's checkout, so it reads the
workspace's `CLAUDE.md` and `.claude/`. The run's environment holds the
prompt in `TASK_UI_PROMPT` and the checkout in `TASK_UI_WORKSPACE_DIR`,
and `task-ui agent prompt` exits with an error when either is empty. The
run lists under Runs like any other, in the internal folder only admins
see.

```sh
curl -s -X POST https://task.example.com/api/workspaces/malta/prompt \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"prompt": "list the skills this workspace carries"}'
```

The reply is the run, with its `id`; `GET /api/runs/{id}` and the
recording follow it. A second prompt while one runs answers `409`.

---

Previous: [Projects and jobs](projects.md) · Next: [Recents](recents.md) · [Contents](../README.md)
