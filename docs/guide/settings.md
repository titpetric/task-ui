# Settings

Not in the sidebar but in the user menu, behind your initials in the
header, for admins: what the instance brings to every run.

![Settings](../assets/settings-dark.png)

## Projects

The folders under the root, with their sources, where they came from, and
the space each takes: its files, and the git history beside them.

![Projects](../assets/projects-dark.png)

**Add project** clones a git repository into the root. It either works,
and the folder appears in the sidebar, or the page shows git's output and
keeps nothing. git reaches private repositories with the instance SSH
keys.

**Configure** on a folder sets what the instance adds to it; its files
stay as they are:

- a title and a description;
- **who may run it**: a box for each role that runs, admin and runner,
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
- the **workspace** a project at the top of one is in: another one moves
  the project there, as
  [Move a project](projects.md#move-a-project-to-another-workspace)
  describes.

The API sets three more on a project: `commands`, the program a kind of
job runs with (`make: gmake`), and `job_allow` and `job_hours`, who may
run and when, per job.

![Configure a folder](../assets/project-edit-dark.png)

## Workspaces

Checkouts under the root that hold projects: an organisation's tooling,
with the projects cloned inside it. **Add workspace** clones one under
the folder you name, **Configure** changes its title and description,
and the remove button deletes it with the projects inside it.
[Workspaces](workspaces.md) covers them, their runs and the prompt
against a workspace.

## Users

Who may sign in comes from the password file on the server. **Create
user** adds one to it, with a name and a password of at least 8
characters, and they can sign in at once; `task-ui passwd <user>` on the
server adds or resets one too. Removing the line removes them, and
task-ui reads the file again whenever it changes. The roles
come from the server's `TASK_UI_ADMINS` and `TASK_UI_VIEWERS`; everyone
else is a runner.

![Users](../assets/users-dark.png)

This page sets which folders a user sees: open a user, then pick **All
projects**, or **Custom** and check the folders. A user's name and role
show as text: the password file and the role lists set them.

![Create user](../assets/user-new-dark.png)

A client that is not a browser, such as an agent over [MCP](mcp.md),
signs in with a token instead: `task-ui token <user> <name>` on the
server prints one, once.

## Secrets

Variables added to the environment of every run, kept per
[workspace](workspaces.md). A secret has a scope:

- `workspace:<name>`, a workspace's shared secret, reaches every folder
  in that workspace and no folder in another one. **Every folder in _**,
  `workspace:_`, reaches the projects at the root; a project in a named
  workspace doesn't get it.
- a folder, `malta/deploy`, reaches that folder's jobs and wins over its
  workspace's shared secret of the same name.

A value is never shown again once saved. The page groups the secrets by
workspace, each with an **Add secret** of its own, and the form's scope
lists each workspace's shared scope and its folders.

![Secrets](../assets/secrets-dark.png)

Each workspace's encryption key encrypts its secrets at rest (see
[Keys](workspaces.md#keys)): the shared ones, and those of every folder
in it. A folder's workspace is its first segment when that's a named
workspace, else `_`. A secret needs its workspace's key, which task-ui
makes for every workspace. **Re-encrypt** writes every secret again for
its workspace's current key, and rotating a workspace's key does it for
that workspace. Removing a workspace removes its secrets with its keys.

The instance key, `id_ed25519` or `id_rsa` when there is no ed25519 key,
encrypts only the settings of the [integrations](#integrations): with it
gone the workspaces' secrets still open.

At start, task-ui moves the secrets of an instance that kept them all on
the instance key: a secret for all folders, scope `all`, becomes `_`'s
shared secret, `workspace:_`, and task-ui writes every other secret
again for its workspace's key. Before it writes anything it copies the
database to `task-ui.db.pre-workspace-keys`; a secret the instance key
can't open stays untouched, and the log names it. The log line `secrets on
their workspace's key` counts what moved.

## SSH keys

The key pairs the jobs and git use to reach other hosts, and the instance
key that encrypts the integrations' settings, marked **integrations**.

![SSH keys](../assets/keys-dark.png)

- **Generate key** makes a new pair (ed25519, ECDSA or RSA), **Upload key**
  takes a private key you already use, unencrypted.
- The copy button copies a public key, to add on the hosts that
  accept it.
- **Rotate** makes a new pair of the same type; task-ui keeps the old one as a
  backup, so the integrations' settings still open, and writes them again
  for the new key. Nobody can delete the last key that opens them.
- **Who gets it** picks whose runs get a key, in a **Roles** row and a **Users** row of boxes. A run
  started by someone with one of the key's roles, or by one of its users,
  has it in the run's ssh agent, and nothing else does: that agent is the
  only way a key reaches a run. Pick them when you generate or upload a key, or later with the
  edit button on its row. A key can have more than one, and a key with none
  stays the instance's: git and the integrations use it, runs do not get it.
- **A key for someone**: pick the user under **For** when you generate or
  upload it, `alice`, and task-ui stores it as `id_ed25519_alice` beside
  the type's own key and grants it to alice. She finds it under **SSH keys** in her user
  menu, the keys her runs get, and copies the public half to her git host
  or servers. A rotation keeps a key's policies. The key files are still
  readable by any job, as jobs run as the server's user; policies decide
  what a run gets, they do not hide a file. Keep `TASK_UI_SSH_DIR`
  out of `~/.ssh`, so `ssh` in a job does not find them by default.
- **Who gets which key** lists each role and each user, and the keys
  their runs get, by comment: a user's own keys, then, dashed, those of
  their role. A key links to its policies.
- A project with private keys in its `.ssh/` folder gets a warning: runs
  do not load them. Keys belong here: upload a key it needs and grant it
  by policy, then delete the folder.
- **Server user's keys** lists the pairs in `~/.ssh` of the user task-ui
  runs as, when that is not `TASK_UI_SSH_DIR`. `ssh` and git take them by
  default, in every run, for git remotes and for the inventory's logins.
  The page shows them read-only, with the copy button: they come with the
  server or its image, and change there.

## Integrations

![Integrations](../assets/integrations-dark.png)

The services task-ui talks to, with their API keys and tokens. task-ui
encrypts them with the instance key, which the workspaces' secrets don't
use, and never
shows them again: a form marks a secret with a value as **set**, and an empty secret
field keeps it on save. The Secrets page leaves them out, and no run sees
them as secrets. A switch beside each integration's name shows or hides
its form, and the browser keeps the choice. **Clear** deletes every
setting of an integration.

- **Claude Code**: the token Claude Code signs in with, from
  `claude setup-token`, for the whole instance, and **Who**: a box for
  each role and each user, checked for those whose runs get it as
  `CLAUDE_CODE_OAUTH_TOKEN`, such as the admin role and the user
  `claude`. Check the agent's user too: the agent job of the
  [Notes](notes.md) board runs as that user. With no box checked, every
  run gets it. The token alone sets the integration up.
- **Beszel**: the hub's URL, as task-ui reaches it, and a token, for the
  [Inventory](inventory.md) page.

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

## Env files

A folder's page lists the dotenv files its pipeline reads. An admin edits
one from there: on saving, and again before every run, task-ui writes the file in the
folder from it with the secrets appended. For a file nobody has
edited, task-ui reads the one in the folder unchanged.

![An env file](../assets/envfile-dark.png)

---

Previous: [Recents](recents.md) · Next: [MCP](mcp.md) · [Contents](../README.md)
