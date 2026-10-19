# Settings

Not in the sidebar but in the user menu, behind your initials in the
header, for admins: what the instance brings to every run.

![Settings](../assets/settings-dark.png)

## Projects

The folders under the root, with their sources and where they came from.

![Projects](../assets/projects-dark.png)

**Add project** clones a git repository into the root. It either works,
and the folder appears in the sidebar, or git's output is shown and
nothing is kept. Private repositories are reached with the instance SSH
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

![Configure a folder](../assets/project-edit-dark.png)

## Users

Who may sign in comes from the password file on the server, not from this
page: `task-ui passwd <user>` adds or resets one, removing the line
removes them, and the file is read again whenever it changes. The roles
come from the server's `TASK_UI_ADMINS` and `TASK_UI_VIEWERS`; everyone
else is a runner.

![Users](../assets/users-dark.png)

What is set here is which folders a user sees: open a user and pick
them, or leave it empty for all.

## Secrets

Variables added to the environment of every run. A secret for all folders
reaches every job; a secret for one folder reaches that folder's jobs and
wins over the shared one. A value is never shown again once saved.

![Secrets](../assets/secrets-dark.png)

Secrets are encrypted at rest with the instance SSH keys: every ed25519
and RSA key below encrypts them, so a secret needs such a key first.
**Re-encrypt** writes every secret again for the keys there are now.

## SSH keys

The key pairs the jobs and git use to reach other hosts, and what the
secrets are encrypted with.

![SSH keys](../assets/keys-dark.png)

- **Generate key** makes a new pair (ed25519, ECDSA or RSA), **Upload key**
  takes a private key you already use, unencrypted.
- The copy button copies a public key, to add where it should be
  accepted.
- **Rotate** makes a new pair of the same type; the old one is kept as a
  backup, so the secrets still open, and the secrets are written again
  for the new key. The last key that opens the secrets cannot be deleted.
- **Policies** are the roles whose runs get a key: a run started by
  someone with one of the key's roles has it in the run's ssh agent,
  beside the keys of its project's `.ssh/` folder. Pick them when you
  generate or upload a key, or later with the edit button on its row. A
  key can be in several roles' sets, a role can have several keys, and a
  key in none stays the instance's: git and the secrets use it, runs do
  not get it. A rotation keeps a key's policies. The key files are still
  readable by any job, as jobs run as the server's user; policies decide
  what a run is handed, they do not hide a file.
- Below the instance keys, each project with keys in its `.ssh/` folder
  has a section of its own: the keys its runs get through an ssh agent.
  The project manages them, so they can be copied but not changed here.

## Env files

A folder's page lists the dotenv files its pipeline reads. An admin edits
one from there: on saving, and again before every run, the file in the
folder is written from it with the secrets appended. A file that was
never edited is read from the folder as it is.

![An env file](../assets/envfile-dark.png)

---

Previous: [Recents](recents.md) · Back to the [contents](../README.md)
