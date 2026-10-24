# Sitemap

Every page of task-ui, by area: what it lists, the actions at the top
right of its head, the actions of a row, its forms and what an empty
list shows. [Development](development.md#lists-and-actions) has the
conventions the pages follow. A row action in `[...]` is an item of the
row's **⋯** menu; "Delete" and "Remove" are in red, last.

## Runs and work

| Page | Path | Lists | Head, top right | Row actions | Forms | Empty |
|------|------|-------|-----------------|-------------|-------|-------|
| Dashboard | `/` | running terminals as tiles, the control centre of jobs by folder | none | Run, Attach on a tile | none | "No jobs yet." in the control centre |
| Runs | `/runs` | every run, newest first, with filters; Activity and Tooling folds | none | **View** or **Attach**, `[...]`: open, replay, Summary, Stop, Kill | the filter bar | "No runs match." |
| Run | `/runs/<id>` | the terminal, its output, the run's details, its artifacts | **View** or **Attach**, Stop, Kill, Report, Replay | artifacts: `[...]` | the prompt about the run | "None: a job keeps files by writing them to $TASK_UI_ARTIFACTS." |
| Report | `/runs/<id>/report` | the question and Claude's answer | Run, Claude's run, Job | none | none | none |
| Workflows | `/workflows` | the workflows, their steps and last run; Recent runs, each with its state dot and its answer folded under a chevron | **Add workflow** | **Run** or **Stop**, `[...]`: Open, Edit, the run going; a recent run: the chevron of its answer | the Prompt box | "No workflows yet." |
| Add workflow | `/workflows/new` | none | none | none | name and steps, Cancel and Save | none |
| Workflow | `/workflows/<id>` | its pipeline and its runs, each with its state dot and its answer folded under a chevron | **Run** or **Stop**, Edit | runs: the chevron, **View**, `[...]` | the runs' filter | "Not run yet." |
| Edit workflow | `/workflows/<id>/edit` | the steps | `[...]`: Open, Delete | a step's `[...]`: Move up, Move down, Duplicate, Remove | the step editor, Save | none |
| Workflow run | `/workflows/<id>/runs/<run>` | the run's thread, its question and answer, the steps of the run, each its run, and Prompts, the threads of its prompt and summary steps, open while it runs and folded once it ends | none | **View** on a step | none | none |
| Artifacts | `/artifacts` | Space, Files of the tree, History | `[...]`: Delete all | Files and History: `[...]`: open, download, copy the link, Delete | the History filter | "This folder is empty.", "No artifacts yet." |
| Artifact file | `/artifacts/files/<path>?version=` | the file itself, inline; `?download` saves it; a directory goes to its listing | none | none | none | none |
| Artifact archive | `/artifacts.tgz`, `/artifacts/<directory>.tgz` | the tree, or a directory of it, as a `.tgz` of the files you see | none | none | none | none |
| Scheduled | `/schedules` | every schedule | none | `[...]`: Enable or Disable, Delete | **Add schedule**: folder, job, args, expression, Run as | "No schedules yet." |
| Notes | `/notes` | notes and their replies | none | a note's `[...]`: Edit, Open its run, Delete | the note box, **Post note** | "No notes yet." |
| Inventory | `/inventory` | hosts by cluster, Run on the hosts | none | the ssh login to copy inline, `[...]`: open, copy ssh | Run on the hosts | "Beszel isn't monitoring any hosts." |
| Host | `/inventory/hosts/<name>` | the host's charts and details | none | none | none | none |

## Folders

| Page | Path | Lists | Head, top right | Row actions | Forms | Empty |
|------|------|-------|-----------------|-------------|-------|-------|
| Workspace, project or folder | `/f/<workspace>/<folder>` | the git line, Jobs (a row per job source under its file, as `atkins.yml` or `task/atkins.yml`, then the files, skills and built-ins, each with a named button per job), Files, on a workspace Home, its home project's files, the folder's details, Environment when a user filled an env file, its block head with **Add env file** | `[...]`: Pin or Unpin, Workspace or Project settings, a workspace's SSH keys and Secrets, Add env file, Pull now, Remove checkout | a job: its named button, Run, Attach or `#!`; a Jobs row: `[...]`: Run and summarize, View the last run, Summary, History, Pin or Unpin, Run force, Kill; a file: `[...]`; an env file: `[...]` Edit | the Prompt box | "No targets in <file>."; no env file, no Environment section |
| Directory | `/f/<folder>/<dir>` | the git line, Jobs, Files | `[...]`: Pin or Unpin | as a folder's | the Prompt box | none |
| Job | `/f/<folder>/<job>` | the job, History, Schedules | **Run**, or **Attach** and Kill | History: **View** or **Attach**, `[...]`; Schedules: `[...]` | **Add schedule** | "No runs yet.", "Not scheduled." |
| Env file | `/f/<folder>/env/<path>`, `/f/<folder>/env` to add | the file, the secrets appended | none | none | the path when adding, the content, Cancel and **Save** | "No secrets reach this folder." |
| Checkout file | `/files/<path>` | the file in the preview window | none | none | none | none |

## Settings

| Page | Path | Lists | Head, top right | Row actions | Forms | Empty |
|------|------|-------|-----------------|-------------|-------|-------|
| Settings | `/settings` | the setup steps in order | none | **Open** | none | none |
| Instance | `/settings/instance` | General, Runs, Parent environment, Paths, Integrations, Keys | none | Keys: the keys table; Parent environment: the server's variables by name, forwarded or not, sensitive flagged, for an admin | the backlog length, the parent environment's None, Auto or All | a missing key's Generate and Upload |
| Data retention | `/settings/retention` | the rules and what each keeps | Clean up now | none | the rules, Save | none |
| Keys | `/settings/keys` | Scopes, Every key, Who gets which key | **Upload key**, **Generate key** | the keys table | none | "No keys added." |
| Upload key, Generate key | `/settings/keys/upload`, `/generate` | none | none | none | the key, Cancel and Upload or Generate | none |
| Key | `/settings/keys/<name>` | none | none | none | its links, Cancel and Save | none |
| Secrets | `/settings/secrets` | a block per workspace, its secrets by project | **Add secret** | a block: **Add secret**; a row: `[...]`: Replace the value, Reads, Delete | none | "No secrets in <workspace> yet." beside **Add secret** |
| Add secret, Replace | `/settings/secrets/new`, `/edit` | none | none | none | scope, name, value, Cancel and Save | none |
| Reads | `/settings/secrets/<id>/reads` | the runs that got a secret | Secrets | none | none | "No run got <name> yet." |
| Integrations | `/settings/integrations` | a block per service | none | Show | the service's settings, Save and Clear | none |
| Users | `/settings/users` | the users, Connect an MCP client | **Add user** | `[...]`: Profile, Runs, Permissions, Delete | none | none |
| Add user | `/settings/users/new` | none | none | none | name, password, Cancel and **Add user** | none |
| Permissions | `/settings/permissions` | a matrix per role | none | none | the matrices, Save | none |
| Workspaces | `/settings/workspaces` | the workspaces, their home projects and projects | Import, **Add workspace** | `[...]`: Open, Workspace settings, Move up, Move down, Delete | Import in a dialog | none: `_` always lists |
| Add workspace | `/settings/workspaces/new` | none | none | none | folder, home project URL and folder, title, branch, Cancel and Add workspace | none |
| Workspace settings | `/settings/workspaces/<folder>`, `_` too | Settings, Home project, Projects, Keys | **Add project**, Export, `[...]`: Open, Delete | Home project: Change, Add or Replace, Remove; Projects: the projects table; Keys: the keys table | Settings, Cancel and Save; Change and Replace in dialogs; Export in a dialog | "No home project." beside **Change** and **Add**; "No projects in <folder> yet."; "No keys added." beside **Upload key** and **Generate key** |
| Workspace keys | `/settings/workspaces/<folder>/keys` | the workspace's SSH keys | **Upload key**, **Generate key**, `[...]`: Open, Workspace settings, Secrets | the keys table, Edit links in a dialog | Upload, Generate and Edit links in dialogs | "No keys added." |
| Workspace secrets | `/settings/workspaces/<folder>/secrets` | the workspace's secrets by project | **Add secret**, `[...]`: Open, Workspace settings, SSH keys, Every workspace's secrets | `[...]`: Replace the value in a dialog, Reads, Delete | Add secret and Replace in dialogs | "No secrets in <folder> yet." |
| Workspace key | `/settings/workspaces/<folder>/keys/upload`, `/generate`, `/<name>` | none | none | none | as the instance key's forms | none |
| Projects | `/settings/projects` | a block per workspace: the folders at the top of the root, then each workspace's projects | **Add project** | a workspace's block: **Add project** into it; the projects table: `[...]`: Open, Project settings, Pull now, Remove checkout | none | "No projects yet.", "No projects in <folder> yet." |
| Add project | `/settings/projects/new`, `?workspace=<folder>` | none | none | none | repository, folder, workspace, branch, Cancel and Clone | none |
| Project settings | `/settings/projects/<folder>` | its jobs' rules | `[...]`: Remove checkout | none | title, who can run, workspace, sources, hours, per job, Cancel and Save | none |

## You

| Page | Path | Lists | Head, top right | Row actions | Forms | Empty |
|------|------|-------|-----------------|-------------|-------|-------|
| Profile | `/users/<name>` | Activity, Roles, Folders, Permissions, Keys, MCP access | Runs, `[...]`: Delete | Activity: **View**, `[...]`; Keys: the keys table, **Generate key** | Roles, Folders, Permissions, each Save; MCP Enable or Rotate | "Nothing you can see yet."; "No key linked to <name>." |
| Your SSH keys | `/keys` | the keys your runs get | none | the keys table | none | "No keys for you or your role yet." |
| Sign in | `/login` | none | none | none | name, password | none |

## Shared parts

| Part | Template | Where |
|------|----------|-------|
| Keys table | `keyhead`, `keyrow` | Keys, Instance, a workspace's page and its SSH keys page, Your SSH keys, a profile |
| Secrets table | `secretgroup`, `secretrow`, `secretfields` | Secrets, Add secret and Replace, a workspace's secrets page |
| Files table | `tree`, `filetable` | a folder's Files and a workspace's Home, a directory's page |
| Projects table | `projecttable` | each block of Projects, a workspace's page |
| Block head | `blockhead`, from `head` | a block's title, its note, its buttons and the line of an empty list |
| Buttons | `buttons` | the buttons of a block head |
| Copy | `copyable` | a public key, an ssh login, the MCP links and command, a token shown once |
| Row menu | `menu` | every row and page head with more than its main action |

---

[Contents](README.md) · [Development](development.md)
