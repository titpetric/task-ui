# Settings

Not in the sidebar but in the user menu, behind your initials in the
header, for admins: what the instance brings to every run.

![Settings](../assets/settings-dark.png)

## Projects

The folders under the root, with their sources and where they came from.

![Projects](../assets/projects-dark.png)

**Add project** clones a git repository into the root. It either works,
and the folder appears in the sidebar, or the page shows git's output and
keeps nothing. git reaches private repositories with the instance SSH
keys.

**Configure** on a folder sets what the instance adds to it; its files
stay as they are:

- a title and a description;
- **who may run it**, user names or roles, empty for every runner;
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
  `daily`); spans are hours or `hh:mm`.

The API sets three more on a project: `commands`, the program a kind of
job runs with (`make: gmake`), and `job_allow` and `job_hours`, who may
run and when, per job.

![Configure a folder](../assets/project-edit-dark.png)

## Users

Who may sign in comes from the password file on the server, not from this
page: `task-ui passwd <user>` adds or resets one, removing the line
removes them, and task-ui reads the file again whenever it changes. The roles
come from the server's `TASK_UI_ADMINS` and `TASK_UI_VIEWERS`; everyone
else is a runner.

![Users](../assets/users-dark.png)

This page sets which folders a user sees: open a user and pick
them, or leave it empty for all.

A client that is not a browser, such as an agent over [MCP](mcp.md),
signs in with a token instead: `task-ui token <user> <name>` on the
server prints one, once.

## Secrets

Variables added to the environment of every run. A secret for all folders
reaches every job; a secret for one folder reaches that folder's jobs and
wins over the shared one. A value is never shown again once saved.

![Secrets](../assets/secrets-dark.png)

The instance SSH keys encrypt the secrets at rest: every ed25519
and RSA key below encrypts them, so a secret needs such a key first.
**Re-encrypt** writes every secret again for the keys there are now.

## SSH keys

The key pairs the jobs and git use to reach other hosts, and the keys that
encrypt the secrets.

![SSH keys](../assets/keys-dark.png)

- **Generate key** makes a new pair (ed25519, ECDSA or RSA), **Upload key**
  takes a private key you already use, unencrypted.
- The copy button copies a public key, to add on the hosts that
  accept it.
- **Rotate** makes a new pair of the same type; task-ui keeps the old one as a
  backup, so the secrets still open, and writes the secrets again
  for the new key. Nobody can delete the last key that opens the secrets.
- **Policies** are whose runs get a key: roles, and users by name. A run
  started by someone with one of the key's roles, or by one of its users,
  has it in the run's ssh agent, and nothing else does: that agent is the
  only way a key reaches a run. Pick them when you generate or upload a key, or later with the
  edit button on its row. A key can have more than one, and a key with none
  stays the instance's: git and the secrets use it, runs do not get it.
- **A key for someone**: set a **Name** when you generate or upload
  it, `alice`, and task-ui stores it as `id_ed25519_alice` beside the type's own
  key, then grant it to alice. She finds it under **SSH keys** in her user
  menu, the keys her runs get, and copies the public half to her git host
  or servers. A rotation keeps a key's policies. The key files are still
  readable by any job, as jobs run as the server's user; policies decide
  what a run gets, they do not hide a file. Keep `TASK_UI_SSH_DIR`
  out of `~/.ssh`, so `ssh` in a job does not find them by default.
- A project with private keys in its `.ssh/` folder gets a warning: runs
  do not load them. Upload a key it needs here and grant it by policy.

## Env files

A folder's page lists the dotenv files its pipeline reads. An admin edits
one from there: on saving, and again before every run, task-ui writes the file in the
folder from it with the secrets appended. For a file nobody has
edited, task-ui reads the one in the folder unchanged.

![An env file](../assets/envfile-dark.png)

---

Previous: [Recents](recents.md) · Next: [MCP](mcp.md) · [Contents](../README.md)
