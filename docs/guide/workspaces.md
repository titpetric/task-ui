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
- **atkins** looks for its pipeline from the working directory up, and
  stops at the first folder with an `atkins.yml`, an `.atkins.yml` or an
  `.atkins/` folder: a project without one runs the workspace's pipeline.
  Its skills come from further: the `.atkins/skills` of the working
  directory and of every folder above it, nearest first, then
  `$HOME/.atkins/skills`. A project with a pipeline of its own still gets
  the workspace's skills, and a project skill of the same name shadows
  the workspace's.

## Skills

The skills of a workspace's `.atkins/skills/` put chips on the [actions
bar](projects.md#actions) of every project and directory inside it:
`[git: pull · rebase]` on each checkout, `[compose: down · up]` beside
a compose file. A skill is an atkins
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
  the job has no chip and a run of it fails, and the rest of the skill
  stays.
- `args: true` on a job takes arguments after `--`, which atkins appends
  to the job's `exec:` steps as argv, with no shell between. Its chip
  reads `custom...` and opens a field for them (see
  [Runs](runs.md#what-a-job-gets)).

A skill job runs in the directory of the file its `when:` matched, so
`git: pull` runs at the top of the checkout and `compose: up` beside the
compose file. One without `when:` runs in the folder that holds its
`.atkins/`.

`<workspace>/.atkins/skills/compose.yml`:

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

In a directory whose `compose.yml` builds an image, the chip reads
`[compose: build · custom... · down · up]`; beside one that only pulls
images, the chip leaves `build` out. `atkins -l` in a directory lists what applies
there, as the bar does. The atkins [skills
guide](https://github.com/titpetric/atkins/blob/main/docs/content/usage/skills.md)
covers the rest: aliases, `help:`, the `-l --json` fields task-ui reads,
vendoring.

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
  their rows, the env files task-ui keeps for them, the workspace's
  secrets and its keys, and its confirm names those projects first. The
  default workspace has none: `DELETE /api/workspaces/_` answers `409`.
- **Import**, under the list, adds a workspace from a bundle another
  instance exported (see [Move to another
  instance](#move-to-another-instance)).

## Clone a project into one

**Add project** under Settings, Projects has a **Workspace** select.
`_`, the default, clones the project at the top of the root; a named
workspace clones it inside, as `malta/dockerfiles`. The Projects page
lists it under that name. In the API an empty `workspace` means `_`. A
project cloned at the root may not take a workspace's name or a
directory that's there already: that answers `409` before git runs.

The project lists with the folders inside it whatever the workspace's
`.gitignore` says, so a workspace that ignores its projects' folders, as
one that clones them itself should, still shows them. A checkout the
workspace ignores without a project row lists too, when it holds a job
source at its top, such as `malta/deploy` with its `atkins.yml`: it
opens at `/f/malta/deploy` with its jobs and the folders inside it, as a
project does (see [Where the jobs come from](projects.md#where-the-jobs-come-from)).

**Recursive**, on the workspace's add and configure forms as on a
project's, is on by default: discovery lists the folders inside the
workspace as deep as they go. Off, it lists the workspace's top-level
folders alone, and the projects inside it by their own Recursive.

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
`{title, description}` changes its text, `recursive` its Recursive, and
`DELETE /api/workspaces/malta` removes it. A `sort` in the `PUT` places the workspace: the lowest comes
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
With no named workspace, the sidebar lists the projects alone (see
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

## Keys

Every workspace has keys of its own, under the data directory and out
of its checkout, where a `git add -A` would publish them:
`TASK_UI_DATA/workspaces/<workspace id>/encrypt` and `/ssh`. The
directories are `0700` and the key files `0600`. The **Keys** section of
a workspace's configure page, `/settings/workspaces/malta`, lists them
with their fingerprints, never a private key.

- The **encryption key**, an ed25519 key, exactly one, encrypts the
  workspace's [secrets](settings.md#secrets): its shared ones and its
  projects'. A workspace starts without one and stores no secrets until
  an admin adds it: **Generate** makes one, **Upload** takes one you
  have, on this page or on the [Keys](settings.md#keys) page. Rotate
  makes a new one, here, on the Keys page or from the ⋯ menu of the
  workspace on the Secrets page; the previous key stays as a backup, and
  task-ui re-encrypts the workspace's secrets for the new key. No run
  gets it.
- The **SSH keys**, zero or more, are for the workspace's runs.
  Generate or upload any number, each named `id_<type>_<name>`. A key
  without a grant goes to every run in the workspace; checking roles or
  users narrows it to their runs. A rotation keeps a key's grants, and
  delete takes the key and its grants. With none, task-ui can't SSH to
  the workspace's hosts with a key of its own: its runs get only the
  instance SSH keys their policies grant.

A key has one use (see [One key, one use](settings.md#one-key-one-use)).
An upload refuses key material that another store holds for the other
use, in this workspace, another one or the instance, with `409`: an SSH
key that is an encryption key anywhere, an encryption key that is an SSH
key anywhere. An encryption key has one holder, so an upload of an
encryption key another workspace or the instance holds answers `409`
too; a backup of the workspace's own may come back.

![A workspace's configure page: its keys, and Export](../assets/workspace-edit-dark.png)

A run gets the instance keys its policies name it for (see
[Settings](settings.md#keys)), then its workspace's SSH keys, and the first
line of its terminal names each: `$ ssh-agent: id_ed25519,
ws:id_ed25519_deploy`. A project at the root gets the keys of `_`. The grants
decide what reaches a run's agent, not what a job can open on disk (see
[What a workspace keeps apart](#what-a-workspace-keeps-apart)).

```sh
curl -s -X POST https://task.example.com/api/workspaces/malta/keys/generate \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type": "ed25519", "name": "deploy", "policies": ["admin"]}'
curl -s https://task.example.com/api/workspaces/malta/keys \
  -H "Authorization: Bearer $TOKEN"
```

`PUT /api/workspaces/malta/keys/id_ed25519_deploy` with `{policies,
users}` sets the grants, empty for every run, and `DELETE` removes the
key. `"encrypt": true` on `generate` or `upload` rotates or replaces the
encryption key; `rekeyed` counts the secrets task-ui wrote again for
it. The key routes are for admins.

## Secrets

A workspace's secrets reach its runs as variables (see
[Settings](settings.md#secrets)). Its **shared** secrets, scope
`workspace:malta`, reach every project in it; a project's own, scope
`malta/deploy`, reach that project and win over a shared one of the same
name. The shared secrets of `_` reach the projects at the root and no
named workspace: a project in `malta` gets `malta`'s shared secrets and
its own, never `_`'s. The workspace's encryption key encrypts them all,
and a workspace without one stores none: the Secrets page turns its
**Add secret** off and the API answers `409`. The Secrets page groups
them by workspace, each project's under its name.

## Move to another instance

A workspace moves to another task-ui as a bundle, `<folder>.taskui-ws`:
a tar.gz of `manifest.json` and `keys/`, encrypted with
[age](https://age-encryption.org) to a passphrase. Every bundle needs
one, and whoever holds the file and the passphrase holds the
workspace's keys.

The bundle carries:

- the workspace's row: title, description, repository, branch and sort;
- its projects, by folder relative to the workspace, with their
  repository, branch and settings;
- its env files, by folder and path;
- its secrets as ciphertext: the bundled encryption key opens them, so
  their values never leave it;
- its key files, backups among them, and the grants of its SSH keys.

Schedules, workflows, runs and artifacts stay on the instance, as do the
integrations, which the instance's encryption key encrypts. The bundle of `_` leaves
out the named workspaces and `internal`.

**Export** on a workspace's configure page takes the passphrase and
downloads the bundle. **Import** on the Workspaces page takes the file,
the passphrase and the folder to import it as, the bundle's own when
left out. The folder must be free, as for Add workspace: a folder held
answers `409`. The import clones the workspace's repository, or makes
its directory when it has none, writes the keys `0600`, adds the rows
under the folder, clones each project, then writes the env files into
the folders the clones made. A project that does not clone keeps its
row, and the import reports git's output for it. Any failure before the
projects keeps nothing: no checkout, no keys, no rows. Env files and
secrets that a removed workspace of the same folder left behind take
the bundle's content.

The import keeps a key to one use too. A bundle whose encryption key a
workspace on this instance holds answers `409` before it clones anything,
so a workspace exported and imported again on the same instance needs
the original removed first, or another instance. Two workspaces never
share an encryption key, and rotating one never leaves the other on the
old key. The same goes for a bundled key that this instance holds for
the other use. The bundle's SSH keys may also be SSH keys here.

The bundle of the default workspace, `_`, imports as a named workspace,
so it needs a folder, such as `legacy`. Its projects move under that
folder, `web` becomes `legacy/web`, and its shared secrets become
`legacy`'s; the workspace's title is the folder.

```sh
curl -s -X POST https://task.example.com/api/workspaces/malta/export \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"passphrase": "correct horse battery"}' -o malta.taskui-ws
curl -s -X POST https://other.example.com/api/workspaces/import \
  -H "Authorization: Bearer $OTHER_TOKEN" \
  -F bundle=@malta.taskui-ws -F passphrase='correct horse battery' -F folder=malta
```

The import answers `{workspace, projects, unknown_users}`: each project
with `cloned` and git's `output`, and the users the grants name that the
instance doesn't have. Their grants stay and reach a run once a user of
that name exists. A wrong passphrase, a file that is not a bundle, an
entry outside `manifest.json`, `keys/encrypt` and `keys/ssh`, and a
secret the bundled key does not open answer `400`; bundled keys refused
for their use answer `409`, and a file over the upload limit `413`. Both routes are for admins.

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

- every workspace's key files under `TASK_UI_DATA/workspaces/`, the
  encryption keys among them, the instance's encryption key under
  `TASK_UI_DATA/instance/encrypt`, and the instance's SSH keys;
- the database, with every workspace's secrets as ciphertext, which the
  keys beside it open;
- every checkout under the root, sealed entries such as `.git` and
  `.env` included: sealing keeps them off the pages and the API, not
  out of a job's reach.

A job that someone in one workspace can change can read another
workspace's secrets. Runner rights, and the right to push to a
workspace's repositories, belong with people you would trust with every
workspace on the instance; where that does not hold, run one instance
per trust boundary.

## A prompt against a workspace

A workspace's folder page, `/f/malta`, carries the **Prompt** box open
for an admin once the [Claude Code integration](settings.md#integrations)
has a token for them; every other folder page, and every directory inside
a folder, carries it folded until its directory holds a prompt (see
[Prompt](projects.md#prompt)). The prompt runs `@claude-prompt` of the
workspace as you, and Claude's answer shows under the question when the
run ends.

![The Prompt box on a workspace's page](../assets/workspace-prompt-dark.png)

Claude Code works in the workspace's checkout, so it reads the
workspace's `CLAUDE.md` and `.claude/`. The run's environment holds the
thread in `TASK_UI_PROMPT`, the checkout in `TASK_UI_WORKSPACE_DIR`, and
the file Claude writes its answer to in `TASK_UI_PROMPT_REPLY`.
`task-ui agent prompt` works in `TASK_UI_PROMPT_DIR`, which a prompt of a
folder or a directory sets, or else in `TASK_UI_WORKSPACE_DIR`, and exits
with an error when the prompt or both directories are empty. The run
lists under Runs as the workspace's, `malta/@claude-prompt`.

```sh
curl -s -X POST https://task.example.com/api/workspaces/malta/prompt \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"prompt": "list the skills this workspace carries"}'
```

The reply is the note, with its `id`, and its `run_id` once the run
starts; Claude's answer is a reply to it, which `GET
/api/prompts?folder=malta` lists. The run starts in the workspace's
checkout and holds that directory as `POST /api/prompts` holds one: a
second prompt there waits for it, and a prompt in a project of the
workspace runs beside it.

---

Previous: [Projects and jobs](projects.md) · Next: [Recents](recents.md) · [Contents](../README.md)
