# Settings

Not in the sidebar but in the user menu, behind your initials in the
header, for whoever holds Allow View on any of its sections: Workspaces,
Projects, Instance Keys, Secrets, Integrations or Users. It holds what
the instance brings to every run. The page
lists the steps in the order you set up an instance, numbered, and
each step's page ends with a footer: **Previous** on the left,
**Next** on the right, and the steps between them as links, with the
current one in bold. A step's sub-pages, such as a key's grants, a
user's profile or a project's settings page, end with the footer of
their step. The trail lists only the steps you may open.

1. [Instance](#instance): what the server runs with.
2. [Data retention](#data-retention): how long the instance keeps what it
   records.
3. [Keys](#keys): the encryption key and the SSH keys.
4. [Secrets](#secrets), on the instance's encryption key.
5. [Integrations](#integrations): Claude Code and Beszel.
6. [Users](#users): who may sign in.
7. [Permissions](#permissions): what each role may do.
8. [Workflows](workflows.md): chains of jobs.
9. [Workspaces](#workspaces): directories of a home and projects under src.
10. [Projects](#projects): the folders under the root.

![Settings](../assets/settings-dark.png)

## Instance

`/settings/instance` shows what the server runs with, in sections, each
a table of the setting, its value, the variable that sets it and what it
does:

- **General**: the base link (`TASK_UI_URL`), the version, the listen
  address, the time zone of the server's clock, how users sign in, the
  session lifetime, the password file, the roles and the MCP endpoint.
- **Runs**: the runs that go at once, the backlog of runs that wait for
  a slot, how long a finished run stays attachable, the stop grace, the
  artifact history and a link to Data retention.
- **Parent environment**: how much of the server's own environment a run
  starts with, and, for an admin, the server's variables by name.
- **Paths**: the workspace folder, discovery, the data folder and how
  the host sees it, the key folder and the compose project.
- **Storage**, for an admin: what the instance keeps on disk, an item a
  row with its place, its size and its files, and a total: the
  database with its journal, the recordings, the artifacts, the tool
  history of each workspace's home and the instance's own copy, and the
  checkouts under each workspace's `src`. The instance measures it now
  and then and shows the last measure; an ellipsis stands where a
  measure is still under way.
- **Integrations**: whether the Beszel integration lists the
  Inventory's hosts.
- **Keys**: the instance's encryption key and the git key, in the
  columns of the Keys page: name, algorithm, size, public key with its
  copy button, purpose, `encryption` or `git + ssh`. The **⋯** menu of
  a key shows its fingerprint and dates, copies its public key,
  downloads it as `<name>.pub` and rotates the key (see [Keys](#keys));
  nothing shows the private key.

The environment sets most of these: change one on the server, then
restart task-ui. The backlog length and the parent environment are the
settings the page saves.
The page shows paths and names, never a password, token or key. Anyone
who may open a step of the settings may open it.

![Instance](../assets/instance-dark.png)

### Runs

**Parallel runs** is how many runs do work at once, the CPUs of the
machine. A run over that waits for a slot, in a line that holds the
**Backlog length**: 20 runs per CPU by default, **Automatic backlog**,
80 on a host with 4 CPUs. An admin sets another length here, a whole
number up to 100000, and **Save** applies it to the next start; empty
goes back to the automatic one. A start while the line is full answers
`429` (see [Runs](runs.md#runs-at-once)).

### Parent environment

What of the server's own environment a run starts with, one of three:

- **None**, secure, the default of a fresh instance: the base variables
  alone, `PATH`, `HOME`, `USER`, `LOGNAME`, `SHELL`, `HOSTNAME`,
  `TMPDIR`, `LANG`, `LANGUAGE`, `LC_*`, `TZ`, `TERM`, `TASK_UI_*`,
  `DOCKER_HOST`, and the ones a workspace's home keeps in use,
  `DOCKER_CONFIG`, `CLAUDE_CONFIG_DIR`, `XDG_CACHE_HOME` and `GOPATH`,
  less any sensitive one;
- **Auto**, filtered: every variable but the sensitive ones;
- **All**: every variable, sensitive ones too.

A name is sensitive when it holds `KEY`, `TOKEN`, `SECRET`, `PASS`,
`CREDENTIAL` or `PRIVATE`, in any case: `WORKSPACE_KEY`,
`SECRETSMANAGER_KEY`, `GITHUB_TOKEN`, `DB_PASSWORD`. Under every choice
a run gets what task-ui sets for it, its `TASK_UI_*` facts, its `HOME`,
the `SSH_AUTH_SOCK` of its own agent, the secrets of its folder and the
integrations' variables, which are grants of their own.

Below the choice, an admin sees every variable of the server by name,
never its value, whether a run gets it under the choice and a
**sensitive** badge on a sensitive name. Picking another choice shows at
once what it would pass; **Save** applies it to the next run. An
instance that had runs before the setting existed starts on **All**, as
runs got the whole environment then; switching to None or Auto is the
admin's call, once the scripts read their secrets as
[Secrets in scripts](#secrets-in-scripts) shows. The setting lives in
the `setting` table as `runs.parent_env`.

## Data retention

`/settings/retention` is one table of what the instance keeps, a row per
area, with the columns **Area**, **What goes**, **Keep**, **Size
limit** and **Holds now**. The line above the table names `@cleanup`,
the job of the root folder that applies the rules, linking its runs,
and says when it last ran and how it ended, linking the run. Every area
the instance records has a row, so nothing grows without a rule:

| Area | What goes | What stays | Rule |
|------|-----------|------------|------|
| **Run recordings** | the `.cast` file of a finished run, its output | the run, who started it and how it ended; its page says the recording expired | age and size |
| **Run history** | the rows of runs older than the age, with their recordings, summaries and the prompt threads about them | runs in progress or queued, and the cleanup's own | age and size, the size their recordings' |
| **Prompt threads** | prompt threads whose last reply is older than the age, with their replies | the runs; the board's notes, which people write, whatever their age; a thread with an open or working instruction | age |
| **Workflow run history** | workflow runs older than the age, with their steps, answers and the threads of their prompts | the workflows; a workflow run in progress | age |
| **Artifact history** | earlier versions of the artifact tree's files, with their records | the latest version of every file in the tree | age and size |
| **Secret read log** | which run got which secret | the last 1000 reads of each secret, which the log keeps whatever the rule | age |
| **Expired sign-ins** | tokens and browser sessions past their expiry | the ones still signing in | fixed: removed every cleanup, no setting |
| **Tool history** | the runs and commands older than the age in `atkins.db` of each workspace's home and the instance's own copy, which atkins prunes and vacuums | the newer runs and commands, the databases | age |

Every rule keeps 31 days at the start. **Keep** picks 24 hours, 7 days,
31 days, 90 days, 365 days or Never; **Size limit**, in the rows whose
items have a size, picks one of these, and a limit saved through the
API that is none of them shows as an option of its own, so the closed
select always shows the saved limit:

| Size limit |
|------------|
| No limit |
| 100 MB |
| 500 MB |
| 1 GB |
| 5 GB |
| 10 GB |
| 50 GB |

The rows without a size show a dash there. **Holds now** is the count,
with the bytes where the items have them. The fixed rows say their
rule across the two columns.

Past either rule the oldest go first, until both hold. `@cleanup`
applies the rows in the order of the table: run recordings, then run
history, which takes a run's recording, summary and threads with it,
then prompt threads, workflow runs, artifact history and the secret
read log, and the expired sign-ins last. A thread run history removes
counts to it, not to the Prompt threads row. A recording of a run in
progress, or of a step of a workflow run in progress, stays and counts
towards the limit. A run of `internal/*`, from before prompts had
folders of their own, falls under the same rules. A file deleted from
the tree keeps its versions in the history until they expire.

Anyone who may open a step of the settings reads the page; an admin
saves the rules, with `PUT /api/retention`, and **Clean up now** starts
`@cleanup` at once, as a run whose output shows what went. A run of
`@cleanup` takes the workspace folder from its own workspace directory,
`TASK_UI_WORKSPACE_DIR`, because `TASK_UI_WORKSPACE` names a run's
workspace. A save takes effect at the next run of `@cleanup`: task-ui
schedules it `@daily`, at midnight, as the first admin, once, when it
first starts; the [Schedules](scheduled.md#cleanup) page changes or
deletes that schedule.

![Data retention: a table, a row per area](../assets/retention-dark.png)

## Keys

Keys have two purposes, and no key serves both:

- the **encryption key** stores secrets. The instance has exactly one,
  and it encrypts every secret, the workspaces' and the integrations'
  settings. Its Purpose reads `encryption`.
- **SSH keys** connect runs and git to hosts and remotes. A workspace has
  any number, for its own runs; the instance's go to the runs of the
  roles and users they link. Git runs over ssh, so an SSH key's Purpose
  reads `git + ssh`, the git key's too.

![Keys](../assets/keys-dark.png)

The page opens on one table, a row for the instance and one per
workspace:

- **Scope**: the instance, or the workspace with its projects under its
  name. The name links to the workspace's settings page.
- **Secrets**: how many the workspace stores, linked to
  `/settings/secrets?workspace=<name>`. For the instance, the
  integrations.
- **Encryption key**: for the instance, **set**, a link to the key's
  row in Every key, whose **⋯** menu rotates it, and the key's
  fingerprint in small type; for a workspace, the instance's, which
  encrypts its secrets too.
- **SSH keys**: how many, and their names, each linked to its edit
  form. A workspace with none says its runs get only the instance SSH
  keys linked to them, with **Generate** and **Upload**.

The only key task-ui makes on its own is the instance's encryption key,
at start when it's missing. The workspace's settings page has its SSH
keys in its [Keys](workspaces.md#keys) section.

Every key lives in one folder, the key folder, `TASK_UI_SSH_DIR`,
`/keys` in the image: flat, `<name>.key` and `<name>.pub`, the
folder `0700` and each file `0600`, beside `known_hosts`. The name ends
with the algorithm and the size, `deploy_rsa4096`,
`task_encryption_ed25519`, `workspace_malta_build_ed25519`,
`user_alice_ed25519`, and is unique across the folder. The database
says what each key is: the encryption key, or an SSH key of the
instance or of a workspace, and whose runs get an SSH key. Nothing but the key
folder reads a key file: git, the inventory's logins, the secrets and
the runs all go through it, and none reads `~/.ssh`.

Under the table, **Every key** lists them all in the columns every keys
table has, the Instance page's too: **Name**, **Algorithm** (ed25519,
ecdsa, rsa), **Size** (256 bits, 4096 bits), **Public key**, shortened
to its type and the end of its material with the whole line in its
tooltip and a copy button that copies the whole line, **Purpose**,
`encryption` or `git + ssh`, and the row's **⋯** menu. The menu opens
on the key's details, its fingerprint, its creation and rotation dates,
its backup, scope and links, then Copy public key, Download
public key, Edit links, Rotate Key or Delete backup, and Remove last.

- **Generate key** makes a key: a name, the algorithm and the size,
  ed25519 by default, ECDSA P-256, P-384 or P-521, RSA 2048, 3072 or
  4096, and the purpose. **Upload key** takes a private key you already
  use, unencrypted, under a name; task-ui reads the algorithm and the
  size from it. A key the folder holds already, under any name or
  purpose, answers `409`: a key has one purpose.
- The copy button copies a public key, to add on the hosts that
  accept it. **Remove**, in the **⋯** menu, retires an SSH key: its files stay in the folder
  as `<name>.key.bak-<time>`, nothing lists or loads them.
- **Who gets it** picks whose runs get an SSH key, in a **Roles** row and
  a **Users** row of boxes. A run started by someone with one of the
  key's roles, or by one of its users, has it in the run's ssh agent, and
  nothing else does: that agent is the only way a key reaches a run,
  from memory, with no key file on disk. A workspace's SSH key with no
  link goes to every run of the workspace. An instance SSH key with no
  link stays the instance's: the inventory's logins use it, runs don't.
- **git pulls and clones with this key**, on a key's edit form, makes it
  the one key git uses for every checkout; the Instance page shows it.
- **A key for someone**: **Generate key** in the **Keys** of her profile
  starts `user_alice_ed25519`, linked to her. She finds it under **SSH
  keys** in her user menu, and copies the public half to her git host or
  servers. The key files are still readable by any job, as jobs run as
  the server's user; links decide what a run gets, they do not hide a
  file. Keep the key folder out of `~/.ssh`, so `ssh` in a job does not
  find the keys by default.
- **Who gets which key** lists each role and each user, and the instance
  SSH keys their runs get: a user's own keys, then, dashed, those of
  their role.
- A project with private keys in its `.ssh/` folder gets a warning: runs
  do not load them. Keys belong in the key folder: upload a key it needs
  and link it, then delete the folder.

The instance's encryption key, `task_encryption_ed25519`, encrypts the
integrations' settings. No run, git remote or SSH connection gets it,
and with it gone the workspaces' secrets still open.

### Rotate Key

**Rotate Key**, in the **⋯** menu of a key on the Instance page and in
Every key here, makes a new key of the same
algorithm and size under the key's name and moves the previous one to
`backup/<name>.key` and `.pub` in the key folder, `0600` in a `0700`
folder. The key's row records the backup's fingerprint and when it
rotated, and shows `BACKUP ACTIVE: backup/<name>` while the backup is
there. Both keys serve meanwhile:

- the encryption key decrypts with the new key, then the backup, and
  the rotation writes the integrations' settings again for the new key;
  the page says how many it re-encrypted.
- an SSH key goes into the ssh agent of runs, git pulls and host probes
  with its backup after it, so a host or a remote that knows the old
  public key alone still lets them in. Add the new public key there,
  then delete the backup.

A key keeps one backup: while it has one, **Rotate Key** stays off with
"Delete the backup first". **Delete backup** removes the backup's files
and record, after a confirm that names its fingerprint; the key then
rotates again. The routes are `POST /api/keys/{name}/rotate` and
`DELETE /api/keys/{name}/backup`; `POST /api/keys/encrypt/rotate`
rotates the instance's encryption key the same way, and a key with a
backup answers `409`.

### One key, one use

A key either encrypts secrets or authenticates SSH connections, never
both, and the folder holds a key once, by its fingerprint.

- No encryption key goes into a run's ssh agent, into git's ssh command,
  into a link, or to the inventory's logins; no route hands one out.
- No SSH key is an age recipient or opens a secret.
- An upload of key material the folder holds already, under any name or
  purpose, answers `409` and writes nothing. A scope holds one
  encryption key.

```sh
curl -s https://task.example.com/api/keys -H "Authorization: Bearer $TOKEN"
curl -s -X POST https://task.example.com/api/keys/generate -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "deploy", "algo": "rsa", "bits": 4096, "policies": ["admin"], "users": ["claude"]}'
curl -s -X PUT https://task.example.com/api/keys/task_ssh_ed25519 -H "Authorization: Bearer $TOKEN" \
  -d '{"git": true}'
```

`GET /api/keys` lists every key, the instance's encryption key again as
`encrypt` and the git key's name as `git`, public halves and
fingerprints only. The rotation answers the key and `rekeyed`, the
count of integration settings task-ui wrote again.

### Moving the keys into the key folder

An instance from before the key folder kept its keys apart:
`TASK_UI_DATA/instance/encrypt`, `TASK_UI_DATA/workspaces/<id>/encrypt`
and `ssh`, and `id_*` files in `TASK_UI_SSH_DIR`. The first start takes
them in once and logs each by its old file, its name and its
fingerprint:

- the instance's encryption key becomes `task_encryption_<algo>`;
- a workspace's becomes `workspace_<folder>_encryption_<algo>`, the
  default workspace's folder as `default`, and its SSH keys
  `workspace_<folder>_<name>_<algo>`, with their grants;
- `id_ed25519` of the key folder becomes `task_ssh_ed25519`, which git
  pulls with, and `id_<type>_<name>`, as `id_rsa_deploy`, becomes
  `<name>_<algo>`, `deploy_rsa4096`, with the roles and users its
  policy named.

Backups come along, to decrypt what they encrypted, and the secrets and
the integrations open as before. The old files stay in place;
nothing reads them, and the log says so: remove them once the keys work.

## Secrets

Variables added to the environment of the runs in a
[workspace](workspaces.md). The page manages them and groups them by
workspace: first the secrets for every project in it,
then each project's own under the project's name. A row shows the
secret's name, its scope, when it changed, its **Reads** and **Last
read**, and two buttons, replace the value and delete. A value is never
shown again once saved.

Every run records the secrets it gets in its environment, one read per
secret per run. **Reads** counts them, and **Last read** says when the
last run got the secret and names its job as `folder/dir/job`, a link to
that run. The count links to the secret's read log,
`/settings/secrets/<id>/reads`: the runs that got it, newest first, 50 a
page, each with when, the run, the job and who started it. task-ui keeps
the last 1000 reads of each secret and drops older ones. The log on
the secret's page is the only trace: task-ui writes nothing of a run's
environment into the run's page, its output or its recording, neither
values nor names. What a job prints itself, it records.

A secret has a scope:

- `workspace:<name>`, a workspace's shared secret, reaches every project
  in that workspace and no project in another one. "every project in
  _", `workspace:_`, reaches the projects at the root; a project in a
  named workspace doesn't get it.
- a folder, `malta/deploy`, reaches that folder's jobs and wins over its
  workspace's shared secret of the same name.

**Add secret** at the top opens the form with a scope picker; each
workspace has an **Add secret** of its own that starts on its shared
scope. In the API, the scope `all`, or none, is `workspace:_`.
`/settings/secrets?workspace=malta` shows one workspace alone; the keys
page links each workspace's count of secrets there.
`/settings/workspaces/malta/secrets`, **Secrets** in the **⋯** menu of
the workspace's page, lists one workspace's secrets and edits them in
place, Add secret and Replace the value in dialogs (see
[Workspaces](workspaces.md#secrets)).

![Secrets, grouped by workspace](../assets/secrets-dark.png)

The instance's encryption key encrypts every secret, the workspaces'
and the integrations': one encryption key per instance, which
[Rotate Key](#rotate-key) rotates, re-encrypting them all. Runs see no
change.

The scopes keep a workspace's secrets out of the environment of
another workspace's runs, not out of their reach on disk (see
[What a workspace keeps apart](workspaces.md#what-a-workspace-keeps-apart)).
A folder's workspace is its first segment when that's a named
workspace, else `_`. Removing a workspace removes its secrets with its
SSH keys. The integrations' settings aren't on this page: the
[Integrations](#integrations) page holds them.

### Secrets in scripts

A run gets its folder's secrets in its environment. A script reads one
where it needs it with `task-ui get-secret NAME`, which prints the value
and nothing after it, and fails with a message on stderr when no secret
of the name reaches the folder:

```bash
GITHUB_TOKEN=$(task-ui get-secret GITHUB_TOKEN) gh release list
```

A value the server's environment holds, which [Parent
environment](#parent-environment) None or Auto keeps from the runs,
belongs in a secret of the folders or workspace that use it. For
secretsmanager, which opens a project's `.secrets.json` with
`SECRETSMANAGER_KEY`, save the key as a secret of that scope and set it
for the one command that needs it:

```bash
SECRETSMANAGER_KEY=$(task-ui get-secret SECRETSMANAGER_KEY) secretsmanager ...
```

A Taskfile or an atkins step takes the same line:

```yaml
# Taskfile.yml
tasks:
  release:
    cmds:
      - GITHUB_TOKEN=$(task-ui get-secret GITHUB_TOKEN) gh release create {{.TAG}}
```

```yaml
# atkins.yml
jobs:
  decrypt:
    steps:
      - SECRETSMANAGER_KEY=$(task-ui get-secret SECRETSMANAGER_KEY) secretsmanager ...
```

`task-ui get-secret` reads `TASK_UI_API`, the instance's API as the run
reaches it, and `TASK_UI_RUN_TOKEN`, a token of the run's own that
signs in `GET /api/run/secrets/{name}` alone and stops working when the
run ends. It returns the secrets that reach the run's folder, the set
its environment holds, never a variable of the server's environment,
and each read lands in the secret's read log under the run. Outside a
run of task-ui it fails: neither variable is there.

At start, task-ui writes every secret the instance's encryption key
can't open again for that key: a secret of a workspace's encryption key
from before one key encrypted them all, or of an instance that kept them
all on its SSH key, the `id_ed25519` or `id_rsa` of the key folder. A
secret for all folders, scope `all`, becomes `_`'s shared secret,
`workspace:_`. Before it writes anything it copies the database to
`task-ui.db.pre-instance-key`, and for the integrations to
`task-ui.db.pre-keysplit`; it reads each secret back before it keeps
it. A secret no key opens stays untouched, and the log names it. The
log lines `secrets on the instance's encryption key` and `integrations
on the instance's encryption key` count what moved. Then each
workspace's encryption key whose secrets the instance's key opens goes,
its row and its files in the key folder, logged as `workspace
encryption key removed`; one with a secret the instance's key doesn't
open stays, with a warning.

## Integrations

![Integrations](../assets/integrations-dark.png)

The services task-ui talks to, with their API keys and tokens, a row per
integration: its name with its set up badge, its description and its
form on the left, and on the right the numbered steps to its key, the
short form of the how-tos below. task-ui encrypts the keys with the
instance's encryption key, as every secret, and never shows them again: a
form marks a secret with a value as **set**, and an empty secret field
keeps it on save. The Secrets page leaves them out, and no run sees them
as secrets. A switch beside each integration's name shows or hides its
row below the name, and the browser keeps the choice. **Clear** deletes
every setting of an integration. The API serves the steps with the kind,
`steps` on `GET /api/integrations`.

- **Claude Code**: the token Claude Code signs in with, from
  `claude setup-token`, for the whole instance, and **Who**: a box for
  each role and each user, checked for those whose runs get it as
  `CLAUDE_CODE_OAUTH_TOKEN`, such as the admin role and the user
  `claude`. Check the agent's user too: the agent job of the
  [Notes](notes.md) board runs as that user. With no box checked, every
  run gets it. The token alone sets the integration up.
- **Forgejo**: the git host's URL and a token, for the sync that makes
  workspaces of its organisations and projects of their repositories,
  see [Forgejo](#forgejo) below.
- **Beszel**: the hub's URL, as task-ui reaches it, and a token, for the
  [Inventory](inventory.md) page. task-ui reads the collections
  `systems`, `system_stats`, the latest minute of each host and its
  records of eight or two hours for the charts, `containers`, the
  containers each host runs, and `container_stats`, their records,
  the newest minute on a hub without `containers`. A read-only user
  reads them all.

### Who gets the token

The **Who** field of Claude Code lists the roles, each with how many
users act with it, and the users, each with the role a run checks for
them: an admin-flagged user with the runner role counts as `admin`.
Beside each user stands their state, which follows the boxes as you
check them, before the save:

- **gets it**: the user's own box has a check.
- **gets it as `role`**: the user's box is clear, and a checked role is
  the role they act with; the row carries the same tint as a checked
  one.
- **doesn't**: neither, so a run of the user starts without the token;
  the row is faint.
- **gets it (everyone)**: no box at all has a check, roles or users, so
  every run gets the token.

The line under the list sums it up: *Nothing checked: every run gets
the token*, or how many of the users get it, naming the ones who don't
when they are five or fewer.

### Beszel's token

Beszel has no API keys, so task-ui takes a token of a read-only user that
no one signs in as. Make one in Beszel with a password nobody keeps, then
mint a token for it as a superuser, with how long it lasts in seconds:

```bash
super=$(curl -s -X POST http://beszel:8090/api/collections/_superusers/auth-with-password \
  -H 'Content-Type: application/json' -d '{"identity":"<admin email>","password":"<admin password>"}' | jq -r .token)
curl -s -X POST http://beszel:8090/api/collections/users/impersonate/<user id> \
  -H "Authorization: $super" -H 'Content-Type: application/json' -d '{"duration":31536000}' | jq -r .token
```

The token works until it expires, and changing the user's password in
Beszel revokes it.

### Forgejo

The Forgejo integration mirrors a git host into task-ui: every
organisation the token's user belongs to is a workspace, with the
organisation's lower case name as its folder; the repository named like
the organisation, `malta/malta`, is the workspace's home repository, the
convention a home follows (see [Workspaces](workspaces.md#the-home));
every other repository of the organisation is a project in the
workspace's `src`, cloned on its default branch. The token user's own
repositories, outside any organisation, stay out.

The settings:

- **URL**: the instance, such as `https://git.cdn.si`.
- **Token**: a personal access token of a user who is a member of the
  organisations to sync. Mint it in Forgejo under **Settings ->
  Applications -> Generate token**, with the scopes `read:organization`,
  `read:repository` and `write:repository`; the last is for the hooks
  alone, which sync installs on the repositories.
- **Clone over**: `ssh`, the default, clones with the instance's git key
  as every project does (see [Keys](#keys)); `https` works for public
  repositories, as task-ui puts no token in a clone URL.
- **Webhook secret**: what signs the pushes Forgejo posts. Leave it
  empty and the first save makes one, which later saves keep.

**Sync now** runs `@forgejo-sync` of the root, `task-ui forgejo-sync`,
and opens the run window on it; **Dry run** runs it with `--dry-run`,
which reports what a sync would do and changes nothing. The run's output
is the report, one line per action, `workspace malta: created`, `home
Malta/malta: cloned main`, `project malta/apps: cloned main`, `hook
Malta/apps: installed`, `skipped Malta/old-thing: archived`, `not on
Forgejo: malta/vhosts`, and a summary line; the line under the form
names the last run and links to it. Admins run the job, by hand, on a
[schedule](scheduled.md) or as a workflow step, as `@cleanup` runs.

Sync adds and updates, and never removes:

- A missing workspace comes into being with its `home` and `src`; one
  that exists stays as you set it.
- The home: sync clones the repository into an empty home directory,
  pulls a home that is the home repository's checkout, and leaves a
  home that is a checkout of some other repository alone, with a line
  in the report: replace it on the workspace's settings page if the
  convention should apply. Without a `<org>/<org>` repository the home
  stays, empty or not.
- A project row that exists keeps its title, who may run it and its
  other settings; its source URL follows the repository's. A checkout
  without a row gets one; a directory in `src` that is no checkout
  stays alone.
- Sync skips an archived repository, and one whose name task-ui keeps,
  `home`, `src`, `internal` or `_`; a mirror or a fork is a project
  like any other.
- A project whose repository Forgejo lacks shows as `not on
  Forgejo` and stays; remove it on the Projects page if it should go.

For every repository it syncs, sync lists the repository's webhooks and,
when none posts to the instance's `<URL>/api/webhooks/forgejo`, installs
one: a push hook, JSON, signed with the webhook secret. On a push,
Forgejo posts to the instance, task-ui checks the signature and pulls
the checkouts of that repository, the project and the home alike, under
the lock a run pulls with, so a job in progress is not pulled under. An
instance without a public URL (`TASK_UI_URL`) installs no hook, and the
report says so. Sync turns a hook of ours that is off or hears other
events back into a push hook; the listing shows no secret, so a changed
webhook secret reaches the hooks only when you delete them on Forgejo
and sync again.

Troubleshooting: a hook whose delivery Forgejo shows as `401` has a
secret other than the integration's: delete the hook in the
repository's settings and run **Sync now**. A `404` means the
integration is not set up on the instance the hook reaches. A `200` with
`no such checkout` means no project or home has that repository's clone
URL: sync, or check the project's source URL. A sync that ends with
`listing the organisations: forgejo answered 401` has a token Forgejo
does not take any more.

## Users

Who may sign in comes from the password file on the server. **Create
user** adds one to it, with a name and a password of at least 8
characters, and they can sign in at once; `task-ui passwd <user>` on the
server adds or resets one too. **Delete**, in a user's row's **⋯** menu or on the profile,
removes one who is no admin and not you: their line leaves the file, and
their sessions, tokens and policy end at once, while runs, notes and
history keep their name. Deleting takes Allow Deletes on Users,
`user/delete`. Removing the line on the server does the same,
and task-ui reads the file again whenever it changes. A new user
comes in as a viewer, the lowest role; the first users of an empty
database come in as admins (see [Running the
server](../running.md#users-and-roles)). The **Flags** column shows
`mcp` for an MCP agent user and `agent` for one whose mentions run
instance Claude.

What a role may do comes from permission policies over the rules of
`config/permissions.yml`, which [Permissions](#permissions) sets. An
admin may do everything. A runner runs jobs, reads and writes notes,
workflows, schedules and the artifact tree, and reads the inventory; a
viewer does the same without running and without writing the artifact
tree. A user's own allow or deny decides first; where the user has
none, the role's rules decide, and a deny there wins over an allow.
[Permissions](#permissions) lists what each role starts with.

![Users](../assets/users-dark.png)

A user's name opens their **profile**, at `/users/<name>`: the name,
when they joined, and what the user did lately, as far as you see it,
in the order you look at a person: **Runs** in folders you see, each
with its result and **View** or **Attach**, **Workflow runs**, **Notes**
with Allow View on Notes, each as one line of plain text, and
**Prompts** with Allow View on Prompts in places you see. A section
with nothing says so in one line. Anyone with Allow View on Users opens
any profile; everyone opens their own, from **Profile** in the user
menu. The profile shows no roles, permissions or forms; **Edit**, top
right, for your own page or with Allow Edits on Users, opens them, and
the **⋯** menu holds **Runs by this user** and **Delete**.

![A profile: the activity of a user](../assets/profile-dark.png)

The **edit page**, at `/users/<name>/edit`, from **Edit** on the profile
or in a user's row's **⋯** menu, holds what the user is and may do, with
**View profile** top right and **Cancel** under a form back to the
profile. It opens as the profile does; each form saves through the API,
which takes Allow Edits on Users, and shows its values as text to anyone
else. Its sections, in order:

- **Roles**, which an admin sets:
  - **Administrative privileges (all)**: every permission, whatever the
    role says. The last admin keeps it: unchecking it answers "the last
    admin".
  - **Role**: runner or viewer, what the user's permissions start from
    without the flag.
  - **MCP agent user**: the user's token reaches the MCP endpoint, and
    an agent learns of work from the notes that name the user. Off, it
    ends the user's MCP token.
  - **Run instance Claude when mentioned**: a note naming the user
    starts the Notes agent as them. It needs MCP agent user.
- **Folders**: which folders the user sees, **All projects**, or
  **Custom** and the folders checked.
- **Permissions**: the user's own rules over the role's (see
  [Permissions](#permissions)).
- **Keys**: the SSH keys linked to the user, with **Generate key** (see
  [Keys](#keys)).
- **MCP access**: **Enable MCP** makes a token for an MCP client that
  lasts 30 days and shows it once, with the `claude mcp add` command
  filled in. Once on, the section shows its dates, the expiry in amber
  within a week and red once past, with **Rotate** and **Disable**.
  Enabling takes Allow Edits on Users; you rotate or disable your own
  from your own edit page. See [MCP](mcp.md).

`/settings/users/<id>` opens the edit page too.

![The edit page of a user: Roles first, then Folders, Permissions, Keys and MCP access](../assets/user-edit-dark.png)

Next to a user's name, on every page that shows one, a **⋯** button
opens a menu: **Profile**, **Runs by this user**, which opens Runs
filtered to them, and **Edit** with Allow Edits on Users. A name that
no user holds, such as the agent a note names, has no profile.

![Add user](../assets/user-new-dark.png)

A client that is not a browser, such as an agent over [MCP](mcp.md),
signs in with a token instead: `task-ui token <user> <name>` on the
server prints one, once, without an expiry. Under the table, **Connect an
MCP client** shows the instance's base link and its MCP endpoint, each
with a copy button, the link to your own MCP access, and the command
that connects Claude Code.

## Permissions

What each role may do, rule by rule, over `config/permissions.yml`: each
section, such as Notes, Secrets or Users, has the rules **Allow All**,
**Allow View**, **Allow Edits** and **Allow Deletes**, and the
application has **Run Jobs**, **Enter Shell**, **Run on Hosts** and
**Allow interactive mode**. A matrix per role holds a row per section,
a column per rule, and under them the application's permissions. No
rule reads a private key: none leaves task-ui.

![Permissions: a matrix per role](../assets/permissions-dark.png)

A cell inherits, empty, allows, a green ✓, or denies, a red ✕; a click
moves it to the next, and **Save** keeps the cells that changed. **Allow
All** sets every rule of its section; where it and a rule set on its
own disagree, the deny wins. The admin role shows every cell allowed and none changes:
an admin may do everything. The page opens with Allow View on Users and
changes with Allow Edits on Users; your own role's matrix changes only
when your role grants Allow Edits on Users.

A user's edit page has the same matrix under **Permissions**, for their
own policy over their role's, with the result after each cell: a muted ✓
where the user may do it, ✕ where not. The user's own allow or deny
decides first, whatever the role says; an empty cell inherits the role's
rules, where a deny wins over an allow; what neither allows, the user may
not do. An admin's matrix shows
everything allowed and does not change, and your own changes only with
Allow Edits on Users through your role, so no one grants themselves
more.

![A user's permissions on their edit page: shell denied, the secrets readable](../assets/user-edit-light.png)

The application's permissions gate what the sections do not:

| Permission   | Rule    | Gates |
|--------------|---------|-------|
| Run Jobs     | `run`   | starting a folder's jobs and ending their runs, from the pages, the API and `run_job`; a workflow takes it for every step, and a schedule for its job |
| Enter Shell  | `shell` | **Enter Shell**, the job `@shell`, an interactive `bash -l`, and **Enter Claude**, the job `@claude`, which also takes a Claude token; task-ui checks it besides their allow list, which names the admin role and no folder rule widens, so admins alone start them |
| Run on Hosts | `hosts` | the command form of [Inventory](inventory.md), which runs a command on its hosts over ssh, and its **Run and summarize** |
| Allow interactive mode | `input` | **Take input** in the run window: typing into a running run of a job that takes no input, to answer a prompt it waits on (see [Runs](runs.md#the-run-window)) |

The admin role may do everything and holds no policy. The other roles,
and the `claude` user, start with these policies; everything else they
lack until a policy allows it:

| Subject       | Allowed at the start |
|---------------|----------------------|
| runner role   | Run Jobs; view, edit and delete on Notes, Workflows and Schedules; Allow View and Allow Edits on Artifacts; Allow View on Inventory |
| viewer role   | the runner's, without Run Jobs and Allow Edits on Artifacts |
| `claude` user | Run Jobs; Allow View and Allow Edits on Notes and on Prompts |

So no role but admin starts with Enter Shell, Run on Hosts, Allow
interactive mode, the prompts or any section of the settings.

Folder visibility and the allow lists of jobs still apply over the
permissions: a user runs a job with Run Jobs, in a folder they see, when
its allow lists name them or their role.

## Workflows

The workflows step opens the [Workflows](workflows.md) page, the
chains of jobs that run one after another. Its footer links to
Permissions before it and Workspaces after it.

## Workspaces

Directories under the root, each with a `home`, the checkout of an
organisation's tooling or an empty directory, which the workspace's
runs get as `HOME`, and a `src` of projects side by side. **Add
workspace** makes one under the folder you name, its home cloned from a
**Home repository** when the form has one, **Move up** and **Move
down** order them, and **Delete** deletes one with everything inside
it. **Workspace settings** opens a workspace's settings page: its
title, description and **Recursive**, the toggle a project has, its
**Home repository**, with **Pull now**, **Replace** or **Clone**, and
**Remove**, its **Projects**, in the table of the Projects page, and
its **Keys**, the SSH keys of its runs, which the [Keys](#keys) page
also lists.
`/settings/workspaces/<folder>/keys` and `/secrets`, **SSH keys** and
**Secrets** in the **⋯** menu of the workspace's page, edit its keys
and its secrets in place. At the top right of that page,
**Add project** clones a project into the workspace, and, for an admin,
**Export** downloads the workspace, encrypted to a passphrase, which
**Import** at the top right of the Workspaces page reads on another
instance, for an admin there.
[Workspaces](workspaces.md) covers them, their keys and runs, the move
between instances, what they keep apart, and the prompt against a
workspace.

## Projects

The folders under the root, a block per workspace: the folders at the
top of the root first, then each workspace's projects, with their
sources, where they came from, and the space each takes: its files, and
the git history beside them. A workspace's block has its own **Add
project**, which clones into it; the Origin column holds a long remote
on one line, whole on a hover.

![Projects](../assets/projects-dark.png)

**Add project** clones a git repository into the root. It either works,
and the folder appears in the sidebar, or the page shows git's output and
keeps nothing. git reaches private repositories with the instance SSH
keys.

A row's `[...]` menu opens the folder, its **Project settings**, and
**Pull now**, which runs `@git:pull` there, on a checkout with the `git`
skill for a user who may run the folder's jobs. **Remove checkout**, last
in the menu, in this list, on the settings page and on the folder's
page, deletes the folder, its files and its configuration, for a user
with Allow Deletes on Projects, `project/delete`. It names the folder and
the remote to clone it again from before it does.

**Project settings** on a folder sets what the instance adds to it; its files
stay as they are:

- a title and a description;
- **who can run it**: a box for each role that runs, admin and runner,
  and each user who is no viewer; none checked is every runner. Each job
  has the same boxes in the per-job table, over the folder's, and none
  checked there inherits;
- which **sources** to read, which is also how a folder with a pipeline
  gets its Makefile or scripts back as jobs;
- **hours**, one rule per line, days then a span. The last rule that
  covers a day decides it:

  ```
  mon-fri 9-17
  fri 9-13
  ```

  runs Monday to Thursday nine to five, Friday mornings, and never on the
  weekend. Days are names or ranges (`mon-fri`, `sat,sun`, `weekdays`,
  `daily`); spans are hours or `hh:mm`;
- **Recursive**, on by default, lists the folders inside the project as
  deep as they go; off lists the project and its top-level folders
  alone, to keep a tree's nested folders out of the catalogue (see
  [Projects and jobs](projects.md#where-the-jobs-come-from));
- the **workspace** a project at the top of one is in: another one moves
  the project there, as
  [Move a project](projects.md#move-a-project-to-another-workspace)
  describes.

The API sets three more on a project: `commands`, the program a kind of
job runs with (`make: gmake`), and `job_allow` and `job_hours`, who may
run and when, per job.

![Project settings](../assets/project-edit-dark.png)

## Env files

An env file is a dotenv file a user fills for a folder. On saving, and
again before every run, task-ui writes it into the folder with the
folder's secrets appended. task-ui reads no env file and no source on
its own: atkins and the job tools read what they need, and every run
gets the folder's secrets in its environment, with or without an env
file.

A user with Allow Edits on Env Files, `env/edit`, adds one with **Add
env file** in a project's or a workspace's **⋯** menu: a path relative
to the folder, such as `.env`, and its content. The folder's page then
lists it under Environment, a section it shows only for a folder with
an env file. A user with Allow View, `env/view`, opens one from there,
and saving it empty deletes it, which takes Allow Deletes,
`env/delete`.

![An env file](../assets/envfile-dark.png)

---

Previous: [Recents](recents.md) · Next: [MCP](mcp.md) · [Contents](../README.md)
