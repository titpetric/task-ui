# Workspaces

A workspace is a directory under the root with two directories in it:
`home/`, the `HOME` of every run in the workspace, and `src/`, where its
projects live side by side, each a checkout of its own, none inside
another. The home is a checkout of the workspace's **home repository**,
what an organisation's projects share, such as `.atkins/skills`, a
`CLAUDE.md` with `.claude/`, and its policies, or an empty directory
while the workspace has none. You name the workspace's folder, and the
folder is the namespace of its projects: `dockerfiles` cloned into the
workspace `malta` lives at `<root>/malta/src/dockerfiles`, its folder ID
is `malta/dockerfiles`, and it opens at `/f/malta/dockerfiles`. The home
is the folder `malta/home`, at `/f/malta/home`.

```
<root>/malta/                 the workspace
<root>/malta/home/            the home repository's checkout, HOME of every run in malta
<root>/malta/src/dockerfiles/ a project, the folder malta/dockerfiles
<root>/malta/src/apps/        another, malta/apps
```

Every project is in a workspace. The **default workspace**, `_`, is the
root itself, laid out the same: `<root>/home` is its home, an empty
directory task-ui makes, and `<root>/src` holds the projects cloned at
the root, whose pages open under `/f/_/`, so `<root>/src/infra` is the
folder `infra` and opens at `/f/_/infra`. A link without the `_`, such
as a bookmarked `/f/infra`, redirects there for good, its query kept.
`/f/_` lists the projects at the root you can see. task-ui adds the
default workspace on its first start, when it has no workspace at all,
and makes the home and src of every workspace it has when it starts.
The default workspace deletes like any other (see [Delete the default
workspace](#delete-the-default-workspace)). No workspace, project or directory
under a workspace takes the name `_`, `internal`, `home` or `src`;
discovery skips a directory of `src` named `_` or `internal`, and a
directory of the root or of a workspace that is neither `home` nor
`src` is not a folder.

![The sidebar with the workspace acme open, its projects api and web under it](../assets/sidebar-workspace-dark.png)

## The home

Every run in a workspace gets the workspace's home as `HOME`, always:
the server's `HOME` is no fallback. The tools a job calls find the
workspace's files there:

- **atkins** reads its skills from the `.atkins/skills` of the working
  directory and of every folder above it, nearest first, then from
  `$HOME/.atkins/skills`: the home's. A project skill of the same name
  shadows the workspace's. Its pipeline is the project's own. Its
  history, `ATKINS_DB`, is `~/.atkins/atkins.db` in the home, and the
  instance's copy under the server's `HOME`, which the Tooling charts
  read, is its mirror.
- **Claude Code** gets the home as a directory it reads, with its
  `CLAUDE.md` and `.claude/` skills, beside its working directory.
- `bash -l` reads the home's `.profile` or `.bash_profile`, when it has
  one.

The convention for a home repository is a repository named like the
organisation: `malta/malta` is the home of the workspace `malta`. The
[Forgejo integration](settings.md#forgejo) follows it when it syncs a
git host, and a workspace made by hand does well to, so the tooling of
an organisation has one obvious place.

The workspace's page, `/f/malta`, lists the home's files in its **Home**
section, in the table of the Files section: a directory opens its page,
a file opens in the preview window, a dot file and `.git` list sealed
and never open. The home is a folder of its own, `malta/home`, with the
jobs of its `atkins.yml` when it has one, so `@git:pull` runs on it as
on any checkout; the home is no project: the Projects list, the sidebar and
the pins leave it out.

![The acme workspace's page: its Home section lists the home's files](../assets/workspace-home-dark.png)

Whoever commits to the workspace's home repository controls that
workspace's runs: its `.profile` runs first in every `bash -l`, and its
skills and `CLAUDE.md` shape every job and every prompt there. What the
instance keeps in the server's own `HOME` stays in use, each through a
variable the run gets when the server does not set it itself:

| What | Variable |
|------|----------|
| registry logins of `docker login` | `DOCKER_CONFIG`, `~/.docker` |
| Claude Code's settings and sessions | `CLAUDE_CONFIG_DIR`, `~/.claude` |
| the caches of XDG tools, Go's build cache among them | `XDG_CACHE_HOME`, `~/.cache` |
| Go's modules | `GOPATH`, `~/go` |

Everything else of `HOME` is the workspace's: the git identity,
`~/.gitconfig`, the shell history, the XDG configuration, data and
state. ssh reads `~/.ssh` from the user's entry in the password
database, not from `HOME`, so its configuration and known hosts stay the
server's. The run also gets `TASK_UI_HOME`, the home's directory, the
same as `HOME`.

## Skills

The skills of a workspace's home, its `.atkins/skills/`, add
rows to the [Jobs table](projects.md#skills) of every project and
directory inside it: `@git:pull` and `@git:rebase` on each checkout,
`@compose:down` and `@compose:up` beside a compose file. A skill is an atkins
pipeline file, `<name>.yml`, its jobs named `<name>:<job>`, and its
`when:` decides where it applies:

- `files:` lists the files that turn it on, any one of them, looked for
  in the directory and then in each folder above it. A name with `*`,
  `?` or `[...]` is a glob. A skill without `when:` applies everywhere.
- `here: true` looks in the directory alone, so a compose skill shows on
  the directory holding `compose.yml` and not on the folders under it.
- `contains:` is a regular expression the matched file must match, such
  as a compose file with a `build:` key.
- A job takes a `when:` of its own, with the same keys. Where it fails,
  the job has no row and a run of it fails, and the rest of the skill
  stays.
- `args: true` on a job takes arguments after `--`, which atkins appends
  to the job's `exec:` steps as argv, with no shell between. Its row
  carries `args`, and its **Run...** opens a field for them (see
  [Runs](runs.md#what-a-job-gets)).

A skill job runs in the directory of the file its `when:` matched, so
`git: pull` runs at the top of the checkout and `compose: up` beside the
compose file. One without `when:` runs in the folder that holds its
`.atkins/`.

`<root>/malta/home/.atkins/skills/compose.yml`:

```yaml
name: Docker Compose
when:
  files: [compose.yml, compose.yaml]
  here: true

jobs:
  up:
    desc: Start the services
    cmd: docker compose up -d --remove-orphans
  down:
    desc: Stop and remove the services
    cmd: docker compose down
  build:
    desc: Build the service images
    when:
      files: [compose.yml, compose.yaml]
      here: true
      contains: '(?m)^\s+build:'
    cmd: docker compose build
  custom:
    desc: Run docker compose with arguments
    args: true
    steps:
      - exec: [docker, compose]
```

In a directory whose `compose.yml` builds an image, the `@compose`
rows are `build`, `custom`, `down` and `up`; beside one that only pulls
images, they leave `build` out. `atkins -l` in a directory lists what applies
there, as the Jobs table does. The atkins [skills
guide](https://github.com/titpetric/atkins/blob/main/docs/content/usage/skills.md)
covers the rest: aliases, `help:`, the `-l --json` fields task-ui reads,
vendoring.

## Add, edit and delete

Settings, **Workspaces** manages them: Allow View on Workspaces,
`workspace/view`, opens the page, Allow Edits, `workspace/edit`, adds,
configures and orders, and Allow Deletes, `workspace/delete`, removes
one. Export and Import are an admin's: the bundle holds the workspace's
private keys and secrets. The page lists each workspace with its
home repository, or "empty", how many projects are inside it, the
checkouts under its `src/` and the project rows there, each once, the
names in the number's tooltip, and how many secrets it holds, shared
and its folders', the number opening its Secrets page, in the sidebar's
order.

![Workspaces, each with its home repository and projects](../assets/workspaces-dark.png)

- **Add workspace** makes the directory under the root, the folder you
  name, with `home/` and `src/` in it. A **Home repository** clones that
  repository into `home/`, at its **Home branch** when you name one; left
  out, the home is an empty directory. The clone works, or the page
  shows git's output and keeps nothing. task-ui refuses a folder that a
  workspace, a project or a directory under the root holds already, and
  `internal`, `_`, `home` and `src`, its own names.
- Any number of workspaces may take the same home repository, each under
  its own folder: two teams can start from the same tooling and clone
  different projects beside it.
- **Edit workspace** opens the workspace's settings page. Its
  **Settings** change the title, description, **Recursive**, which,
  as a project's does, lists the folders inside it as deep as they go,
  or, off, the workspace and its top-level folders alone, and
  **Sidebar**, how many of its projects the sidebar lists, the most
  recently committed first, empty for every one (see [In the
  sidebar](#in-the-sidebar)); the folder stays. Its **Home repository** shows the repository and branch of the
  home with what git knows of the checkout, or "The home is an empty
  directory.", and beside it **Pull now**, the home's `@git:pull`,
  **Replace**, or **Clone** while the home is empty, which removes the
  home and clones a repository in its place, and **Remove**, which
  removes the checkout and leaves the home an empty directory. Replace
  and Remove refuse, a `409`, while a run of the workspace is active:
  the home is that run's `HOME`. Its **Projects** lists the projects
  inside it in the table of the Projects page, and **Add project**, at
  the top right, clones one into its `src/`.
- **Move up** and **Move down**, in a workspace's **⋯** menu, change the order of the workspaces, in
  the sidebar and on the page. An added workspace goes last. The default
  workspace lists first and keeps its place.
- **Delete**, last in the menu, deletes every project inside the
  workspace with its checkout and its row, the env files task-ui keeps
  for them, the workspace's secrets and its keys, and the workspace's
  directory with everything in it, the home included; its confirm names
  those projects first. It refuses, a `409`, while a run of the
  workspace is active. A workspace that holds a secret has no Delete:
  the menu says **Delete its N secrets first** and opens its Secrets
  page, and `DELETE` answers `409` until you delete each one by hand,
  so a delete never takes a secret with it unseen. The default
  workspace, the root, deletes the same way: its projects at the root
  with their checkouts, their env files, its keys and its home go, and
  so does its row; the root directory stays with the named workspaces
  in it.
- **Import**, at the top right for an admin, adds a workspace from a
  bundle another instance exported (see [Move to another
  instance](#move-to-another-instance)).

### Delete the default workspace

An instance runs without a default workspace. Without `_`, the
Workspaces page lists the named workspaces alone, and task-ui makes no
`home/` or `src/` at the root again. What changes:

- **Add project** needs a workspace: the Workspace select lists the
  named ones, and `POST /api/projects` with no `workspace` answers `400`
  "pick a workspace: the instance has no default workspace".
- The Projects page, the dashboard and the sidebar list no projects at
  the root; the sidebar starts with the workspaces.
- `/f/_` still opens: the root's page, with its built-in jobs, `@host:*`,
  `@forgejo-sync` and `@cleanup`, which keep running, on their schedules
  too. A run of one of them has no workspace: no `TASK_UI_WORKSPACE`, no
  `TASK_UI_WORKSPACE_DIR`, and `HOME` stays the server's. Its keys are
  the instance keys linked to it, as before.
- A prompt at the root, and `@claude-note` on the Notes board, run in
  the root directory with the server's `HOME`.
- The Secrets page has no `_` group and no shared scope `workspace:_`;
  the integrations' settings, under the instance key, stay.
- A workflow step lists `_` as a place only for the root's own jobs.

Like any workspace, `_` deletes once it holds no secret and no run of
the root is active: delete its secrets by hand first, then Delete.

## Clone a project into one

**Add project** under Settings, Projects has a **Workspace** select.
`_`, the default, while the instance has it, clones the project into
`src/` at the root; a named workspace clones it into its `src/`, as
`malta/dockerfiles`. The Projects page lists it under that name. In the
API an empty `workspace` means `_`, and answers `400` on an instance
without a default workspace. A project may not take a workspace's name, `home`, `src` or
a directory that's there already: that answers `409`, or `400` for one
of task-ui's names, before git runs.

A checkout in the workspace's `src/` without a project row lists too,
when it holds a job source anywhere inside it, such as `malta/deploy`
with its `scripts/atkins.yml`: it opens at `/f/malta/deploy` with its
jobs and the folders inside it, as a project does (see [Where the jobs
come from](projects.md#where-the-jobs-come-from)). A checkout that a
project's `.gitignore` ignores right below it lists the same way.

**Recursive**, on the workspace's add and configure forms as on a
project's, is on by default: discovery lists the folders inside the
workspace as deep as they go. Off, it lists the workspace's top-level
folders alone, and the projects inside it by their own Recursive.

The API does the same, with a token from `task-ui token`:

```sh
curl -s -X POST https://task.example.com/api/workspaces \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"folder": "malta", "title": "Malta", "url": "git@example.com:org/workspace.git"}'
curl -s -X POST https://task.example.com/api/projects \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url": "git@example.com:org/dockerfiles.git", "workspace": "malta"}'
curl -s -X PUT https://task.example.com/api/workspaces/malta/home \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"url": "git@example.com:org/workspace.git", "branch": "main"}'
```

The create's `url` is optional: with it, the home is a clone of that
repository at `branch`, its default branch when left out; without it
the home is an empty directory. The reply is `{workspace, output}`,
git's output in `output`. A taken folder answers `409`. `PUT
/api/workspaces/malta` with `{title, description}` changes its text,
`recursive` its Recursive, and `DELETE /api/workspaces/malta` removes
it. `PUT /api/workspaces/malta/home` with `{url, branch}` replaces the
home with a clone, `DELETE /api/workspaces/malta/home` empties it, and
both answer `409` while a run of the workspace is active. A workspace's
row names its home repository in `home_url` and `home_branch`, `""` for
an empty home. A `sort` in the `PUT` places the workspace: the lowest
comes first, and workspaces of one sort go by folder. Left out, the
sort stays. A `sidebar_limit` in the `PUT` is the Sidebar setting, `0`
for every project; below `0` answers `400`, and left out it stays.

## In the sidebar

The projects at the top of the root come first, by name, and the
workspaces follow them, in their order. A workspace adds a level above
its projects: a label row with the
workspace's name, the count of its projects and a chevron that points
down while it's open. Under it come the workspace's own jobs, then its
projects, each with a fold of its own: every directory right inside the
workspace's `src/`, a checkout with jobs or without; the home never
shows in the sidebar, pinned or not; its files are on the workspace's
page. This browser remembers each fold apart, so `malta` and
`malta/dockerfiles` open and close on their own.

A workspace with many projects lists the ones in use: **Sidebar** in
its settings takes a number, and the sidebar lists that many of the
workspace's projects, the ones with the most recent commit first, a
checkout without a commit last, and ties by name; the folders inside a
kept project stay with it, and the listed projects keep the sidebar's
name order. Empty or `0` lists every project, by name. A project set
not to show takes no place in the count. The rest leave the sidebar
alone: the Projects page, the workspace's page, the dashboard and a pin
list them as before. The default workspace, `_`, takes the setting over
the projects at the root. task-ui reads each project's last commit with
one git command and keeps it a minute, so a page costs at most one git
process per project per minute, and a pull or a push shows in the order
within the minute. **Pin workspace** in
the `[...]` of the workspace's page, and **Pin** on a folder or a job
inside it, add to what the sidebar lists, for you (see [What the
sidebar lists](projects.md#what-the-sidebar-lists)).
With no named workspace, the sidebar lists the projects alone (see
[Projects and jobs](projects.md)).

## Runs in a workspace

A run of a job in the workspace or in one of its projects gets two more
variables (see [Runs](runs.md#what-a-job-gets)):

- `TASK_UI_WORKSPACE`, the workspace's folder, such as `malta`, or `_`
  for a project at the root;
- `TASK_UI_WORKSPACE_DIR`, its directory, such as `<root>/malta`, or the
  root for `_`;
- `HOME` and `TASK_UI_HOME`, the workspace's home, `<root>/malta/home`,
  or `<root>/home` for `_` (see [The home](#the-home)).

A run at the root of an instance without a default workspace gets none
of them: `HOME` stays the server's.

Before the job, the run pulls a home that holds a checkout, then the
project's checkout, so the shared tooling is current when the job
starts. When either pull fails, the run fails with it.

## Hours

**Hours** on a workspace's settings page says when runs in the
workspace may start: every job of its folders and projects, the
built-ins such as `@shell` too. One rule per line, days then a span;
the last rule that covers a day decides it:

```
mon-fri 9-17
fri 9-13
```

runs Monday to Thursday nine to five, Friday mornings, and never on the
weekend. Days are names or ranges (`mon-fri`, `sat,sun`, `weekdays`,
`daily`); spans are hours or `hh:mm`. Empty, the default, allows any
time. The workspace is the one place hours live: a project and a job
take their workspace's, and the default workspace's apply to the
projects at the root.

Out of hours **Run** fails, and **Run force** asks why: it needs a
reason, keeps it with the run, and every list of runs shows it. A
schedule and a prompt go ahead at any time: the schedule is the
decision, and a prompt runs none of the folder's jobs.

## Keys

Every workspace has SSH keys of its own, in the key folder,
`TASK_UI_SSH_DIR`, out of its directory, where a `git add -A` would
publish them: `workspace_<folder>_<name>_<algorithm>.key` and `.pub`,
the files `0600`. The **Keys** section of a workspace's settings page,
`/settings/workspaces/malta`, lists them in the keys table of the Keys
page, name, algorithm, size, public key with its copy button, purpose
and the **⋯** menu, never a private key. With none, it says "No keys
added." beside **Upload key** and **Generate key**.

**SSH keys**, in the **⋯** menu at the top right of the workspace's
page, `/f/malta`, opens `/settings/workspaces/malta/keys`, the
workspace's keys alone, edited in place: **Generate key** and **Upload
key** at the top right, and **Edit links** in a key's **⋯** menu, open
their forms in a dialog over the list, and **Remove** retires a key.
The page and the menu item take `workspace/view`, the forms
`workspace/edit`, as the key routes do.

![The workspace page's menu with SSH keys and Secrets](../assets/workspace-menu-dark.png)

![The SSH keys of acme, edited in place](../assets/workspace-keys-dark.png)

- The **SSH keys**, zero or more, are for the workspace's runs. A
  workspace has no encryption key: the instance's encrypts its
  [secrets](settings.md#secrets), as every secret.
  Generate or upload any number, each under a name of its own,
  `workspace_malta_deploy_ed25519`. A key without a link goes to every
  run in the workspace; checking roles or users narrows it to their
  runs. Remove retires the key and its links. With none, task-ui can't
  SSH to the workspace's hosts with a key of its own: its runs get only
  the instance SSH keys linked to them.

A key has one use (see [One key, one use](settings.md#one-key-one-use)):
an upload of key material the key folder holds already, under any name,
answers `409`.

![A workspace's settings page: its home repository, keys, and Export](../assets/workspace-edit-dark.png)

A run gets the instance keys linked to it (see
[Settings](settings.md#keys)), then its workspace's SSH keys, and the
first line of its terminal names each: `$ ssh-agent: deploy_rsa4096,
workspace_malta_deploy_ed25519`. A project at the root gets the keys of
`_`. The links decide what reaches a run's agent, not what a job can
open on disk (see [What a workspace keeps apart](#what-a-workspace-keeps-apart)).

```sh
curl -s -X POST https://task.example.com/api/workspaces/malta/keys/generate \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"algo": "ed25519", "name": "deploy", "policies": ["admin"]}'
curl -s https://task.example.com/api/workspaces/malta/keys \
  -H "Authorization: Bearer $TOKEN"
```

`PUT /api/workspaces/malta/keys/workspace_malta_deploy_ed25519` with
`{policies, users}` sets the links, empty for every run, and `DELETE`
retires the key. `"encrypt": true` on `generate` or `upload` answers
`400`: one encryption key per instance. Listing the keys takes
`workspace/view`; the other key routes take `workspace/edit`.

## Secrets

A workspace's secrets reach its runs as variables (see
[Settings](settings.md#secrets)). Its **shared** secrets, scope
`workspace:malta`, reach every folder in it; a project's own, scope
`malta/deploy`, reach that project and win over a shared one of the same
name. The shared secrets of `_` reach the projects at the root and no
named workspace: a project in `malta` gets `malta`'s shared secrets and
its own, never `_`'s. The instance's encryption key encrypts them all.
The Secrets page groups
them by workspace, each project's under its name.

**Secrets**, in the **⋯** menu at the top right of the workspace's
page, opens `/settings/workspaces/malta/secrets`, the workspace's
secrets alone, edited in place: **Add secret** opens a dialog with the
workspace's scopes, the shared one and each project's, **Replace the
value** in a secret's **⋯** menu opens one that takes the new value, and
**Delete** removes it. A value never shows again once saved. The page
and the menu item take `secret/view`, Add and Replace `secret/edit`,
Delete `secret/delete`.

![The secrets of acme, edited in place](../assets/workspace-secrets-dark.png)

## Move to another instance

A workspace moves to another task-ui as a bundle, `<folder>.tui.age`:
a tar.gz of `manifest.json` and `.ssh/`, encrypted with
[age](https://age-encryption.org) to a passphrase of at least 12
characters. Whoever holds the file and the passphrase holds the
workspace's secrets and SSH keys: nothing else protects them inside it,
and task-ui keeps no copy of the passphrase. To move a workspace, export
it, import the bundle on the other instance, then delete its secrets
here by hand, which Delete needs, and then the workspace.

The bundle carries:

- the workspace's row: title, description, sort and Recursive, as
  `shallow`;
- its home repository and branch in `home`, `{url, branch}`, empty
  for an empty home;
- its projects, by folder relative to the workspace, with their
  repository, branch and settings;
- its env files, by folder and path;
- its secrets in plain text in `manifest.json`, which the importing
  instance encrypts with its own encryption key as they go into its
  vault;
- the SSH keys the workspace's runs get, its own keys, as
  `.ssh/<name>.key` and `.ssh/<name>.pub`, and their links; an import
  names them after its folder, `workspace_<folder>_...`.

It never carries the instance's encryption key, the instance's SSH keys
or another workspace's keys. Schedules, workflows, runs and artifacts
stay on the instance, as do the integrations. The bundle of `_` leaves
out the named workspaces and `internal`.

**Export** at the top right of a workspace's settings page, which an
admin sees, opens a dialog for the passphrase, typed twice, or made by **Generate**: 24
random characters, shown once with a copy button. Store it before you
export, since nothing recovers it. **Export** in the dialog then
downloads the bundle; the passphrase goes in the request body, never
in the URL or the logs. **Import** at the top right of the Workspaces
page opens a dialog for the file, the passphrase and the folder to
import it as, the bundle's own when left out. The folder must be free,
as for Add workspace: a folder held answers `409`. The import makes the
workspace's directory with its home and src, writes the keys `0600`
into the key folder, adds the rows under the folder, clones the home
repository into the home, clones each project into src, then writes
the env files into the folders the clones made. A bundle of another
format answers `400`: the format is 4, and an older bundle does not
import. A project that does not clone keeps its row,
and the import reports git's output for it. Any failure before the
projects' clones keeps nothing: no directory, no keys, no rows. Env files and
secrets that a removed workspace of the same folder left behind take
the bundle's content.

The import refuses a key it would hold twice, and renames none. A bundle
whose SSH key this instance holds, under any name, or whose key name the
key folder holds answers `409` with the key's name before it clones
anything, so a workspace exported and imported again on the same
instance needs the original removed first, or another instance.

The bundle of the default workspace, `_`, imports as a named workspace,
so it needs a folder, such as `legacy`. Its projects move under that
folder, `web` becomes `legacy/web`, and its shared secrets become
`legacy`'s; the workspace's title is the folder.

```sh
curl -s -X POST https://task.example.com/api/workspaces/malta/export \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"passphrase": "correct horse battery"}' -o malta.tui.age
curl -s -X POST https://other.example.com/api/workspaces/import \
  -H "Authorization: Bearer $OTHER_TOKEN" \
  -F bundle=@malta.tui.age -F passphrase='correct horse battery' -F folder=malta
```

The import answers `{workspace, home, projects, unknown_users}`: the
home and each project with `cloned` and git's `output`, and the users
the grants name that the instance doesn't have. Their grants stay and reach a run once a user of
that name exists. An export's passphrase under 12 characters, a wrong
passphrase on import, a file that is not a bundle, and an entry other
than `manifest.json` or a `.key` or `.pub` file of `.ssh/` answer `400`;
bundled keys held here answer `409`, and a file over the upload limit
`413`. Both routes take an admin, `403` for anyone else, `workspace/edit`
or not: the bundle alone, of everything task-ui hands out, holds private
keys, inside its passphrase, and the import writes them into the key
folder.

| Limit | Size |
|-------|------|
| The bundle file | 64 MiB |
| `manifest.json` | 16 MiB |
| Each key file | 1 MiB |
| Entries | 1024 |

## What a workspace keeps apart

A workspace decides what task-ui puts into a run: which secrets go into
its environment and which SSH keys go into its agent. That is no
sandbox. Every run runs as the server's OS user, in every workspace, so
a job can read what that user can:

- every key file of the key folder, `TASK_UI_SSH_DIR`, the encryption
  keys among them; in the image, mount `/keys` and nothing else there;
- the database, with every workspace's secrets as ciphertext, which the
  keys beside it open;
- every checkout under the root, the homes among them, sealed entries
  such as `.git` and `.env` included: sealing keeps them off the pages
  and the API, not out of a job's reach.

A job that someone in one workspace can change can read another
workspace's secrets, and whoever commits to a workspace's home
repository controls that workspace's runs: its `.profile`, skills and
`CLAUDE.md` run in every one of them. Runner rights, and the right to
push to a workspace's repositories, its home repository first, belong
with people you would trust with every workspace on the instance; where
that does not hold, run one instance per trust boundary.

## A prompt against a workspace

A workspace's folder page, `/f/malta`, carries the **Prompt** box open
for a user with Allow Edits on Prompts, `prompt/edit`, once the
[Claude Code integration](settings.md#integrations) has a token for
them; every other folder page, and every directory inside
a folder, carries it folded until its directory holds a prompt (see
[Prompt](projects.md#prompt)). The prompt runs `@claude-prompt` of the
workspace as you, and Claude's answer shows under the question when the
run ends.

![The Prompt box on a workspace's page](../assets/workspace-prompt-dark.png)

Claude Code works in the workspace's home, so it reads the workspace's
`CLAUDE.md` and `.claude/`, in an empty home too. The run's environment
holds the thread in `TASK_UI_PROMPT`, the home in `TASK_UI_HOME`, the
workspace's directory in `TASK_UI_WORKSPACE_DIR`, and the file Claude
writes its answer to in `TASK_UI_PROMPT_REPLY`. `task-ui agent prompt`
works in `TASK_UI_PROMPT_DIR`, which a prompt of a folder or a directory
sets, or else in `TASK_UI_HOME`, or else in `TASK_UI_WORKSPACE_DIR`,
and exits with an error when the prompt or all three are empty. The run
lists under Runs as the workspace's, `malta/@claude-prompt`.

```sh
curl -s -X POST https://task.example.com/api/workspaces/malta/prompt \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"prompt": "list the skills this workspace carries"}'
```

The reply is the note, with its `id`, and its `run_id` once the run
starts; Claude's answer is a reply to it, which `GET
/api/prompts?folder=malta` lists. The run starts in the workspace's
home and holds it as `POST /api/prompts` holds one: a second prompt
there waits for it, and a prompt in a project of the workspace runs
beside it.

---

Previous: [Projects and jobs](projects.md) · Next: [Recents](recents.md) · [Contents](../README.md)
