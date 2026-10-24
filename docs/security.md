# Security

<!-- vale off -->
Least access is an important security management principle, where access of credentials is limited to what they require, and not more. TaskUI may give a wide surface of execution, or may be limited to read-only or exec-only users that can only invoke something in a secured system. The barriers between container and machine can be a little bit blurred. Credential management is simplified and the SRE usability for the tool and its Claude bindings has proven to be pretty good. Integrations with Beszel give access to a deployment cluster, and even without specific automation in place it can sort of be used as a mix of terminal interface and claude prompts. It tells me I should water my plants.
<!-- vale on -->

This page says how the configuration of task-ui moves it between wide
reach and least access: what least access means here, the
[decisions](#decisions) an operator makes and what each one opens up,
and a [checklist](#checklist). A review follows, a tour of the features
that handle identity, keys, encryption and roles: each section says what
the code does, names the file that does it, and says where the
protection ends. The [findings](#findings) table at the end lists what
to fix or watch, by severity.

task-ui runs jobs from source trees as the user who asks, records each
run, and hands each run the secrets and SSH keys its user and its
workspace allow. task-ui works as a CI/CD system you can type into: a
run is a terminal that people and agents watch, attach to and answer. It manages
who may do what, which keys a run carries, and how secrets sit at rest.
It does not sandbox the jobs it runs.

## Managing task-ui security is managing Linux security

task-ui is a server in a container, and every run is a process in that
container: `bash -l`, `make`, `docker`, `ssh`, Claude Code. A configured
instance is as secure as the Linux system and the container it runs in:
the mounts it gets, the daemons it reaches, the keys in its key folder,
the network around it and the people who sign in. Managing its security
is managing Linux security, with the same two moves:

- **Limit the attack surface.** Leave out the mounts, sockets and
  integrations the jobs don't need, scope keys and secrets to the runs
  that use them, and grant each user the narrowest role.
- **Block access.** A policy deny, a folder limit or an allow list
  removes a path; an instance of its own for each trust boundary
  removes every path between them.

Admin is root inside task-ui. An admin passes every permission check,
sees every folder, opens `@shell` and `@claude`, reads and rotates every
key, and edits users and policies. Anyone who holds admin credentials, a
password, a session or an MCP token of an admin, can act anywhere in the
system and can disrupt task-ui itself: end runs, rotate or delete keys,
delete users, projects and workspaces, or stop the container through the
docker socket. Guard an admin account as you guard a root login on the
host.

## What least access means here

Least access grants each credential the permissions its work needs and no
more. A credential in task-ui is a user: a password, a session or an MCP
token. What it may do is its role, its own policy over the role, the
folders an admin limits it to, and the allow lists of folders and jobs
(see [Roles and policies](#roles-and-policies)). The default roles are a
starting point, not least access: the viewer role starts with edit and
delete on notes, workflows and schedules.

| Profile | Set up as | What it does | What it costs |
|---------|-----------|--------------|------------------|
| Read-only | the viewer role; the user's policy denies `note/edit`, `note/delete`, `workflow/edit`, `workflow/delete`, `schedule/edit` and `schedule/delete`; a folder limit to the folders to watch | sees those folders' jobs, runs, live terminals and recordings; reads their files in the preview, the artifacts and the inventory | running jobs, notes and replies, prompts, workflows and schedules, every setting |
| Exec-only | the runner role; a folder limit to a secured folder, one whose code the user cannot push to; the folder's allow list names the user or the role; denies on the workflow, schedule and artifact rules the user doesn't need | starts the folder's jobs and ends their runs; types into the interactive jobs they may run; sees the folder's runs | `@shell` and `@claude`, Claude prompts, notes to the agent, editing pipelines (no push to the checkout, no workflow edits), env files, secrets, keys, projects, workspaces and their export, commands on hosts, **Take input** on jobs that take none |
| Agent | a user of its own with **MCP agent user**, `run` on chosen folders and notes; `prompt/edit` only where Claude may run commands | the MCP tools as that user, for 30 days a token | the limits of its user; a leaked token is the user until it expires or rotates |
| Admin | the admin role or flag | everything | nothing: an admin holds the host |

What the profiles leave open:

- **Read-only is not blind.** A viewer sees what a run prints, live and
  in its recording: a secret a job echoes, and an admin's `@shell` in a
  folder the viewer sees (see [Terminals](#terminals)). The file preview
  serves any file of a folder the viewer sees, an env file at a path
  that doesn't start with a dot included (see the findings). The folder
  limit is what keeps a viewer out.
- **Exec-only holds while the code holds.** `run` on a folder whose
  `atkins.yml`, `Taskfile.yml`, `Makefile` or `scripts/*` the user can
  change through git is a command of their choosing. A job's allow list
  narrows that job alone: `./<path>`, any executable file of the
  folder, and `@skill:<job>`, an atkins skill that applies there, run
  under the folder's allow list (`folder/folder_path.go`, `FileJob`;
  `folder/skill.go`, `SkillJob`). Name in the folder's allow list only
  those who may run every one of them.
- **Least access costs reach.** Without admin and `prompt/edit` a user
  can't look around in a terminal or ask Claude; without `run` on a
  folder they can't start its jobs; without the Claude Code integration,
  docker or Beszel, no run builds images, asks Claude or reaches the
  inventory. Each [decision](#decisions) names what its less secure
  setting enables, which the more secure one forgoes.

## Decisions

Each row is a choice an operator makes. **Default** says where the
image, the compose files or the database start: **on** in the compose
files (`docker/docker-compose.yml` in this repository and
`malta/dockerfiles/task/compose.yml` in production), **built in** to the
image or the code, **default** for a setting with a starting value, or
**opt-in**, off until someone turns it on.

| # | Decision | Default | The less secure choice enables | The more secure alternative, and what it costs |
|---|----------|---------|--------------------------------|---------------------------------------------------|
| D1 | The host's docker socket mounted at `/var/run/docker.sock` | on | To build and push docker images. To run docker images on the host, such as `docker compose up` of a deployment beside task-ui. The socket is root on the host: `docker run -v /:/host` reads and writes everything. A `:ro` mount, as in this repository's compose file, limits nothing: a client still connects and uses the whole API. | **A docker-in-docker (dind) sidecar**, with `DOCKER_HOST` in task-ui's environment, which runs inherit: builds, pushes and containers stay in the sidecar's daemon, which runs privileged. Costs: deploying on the host, and a `docker run -v` resolves its paths in the sidecar, which needs the data folder at the host's path. **A socket proxy** that allows build and push alone: costs running containers. **No socket**: no docker at all, and no image builds. |
| D2 | One container and one OS user, root, for task-ui and every run | built in | To run any tool a job needs, beside `/keys`, the database, the registry logins and Claude Code's settings. Any run reads every key file, the encryption key included. | task-ui has no setting for a run user of its own. Run one instance per trust boundary; the instances share nothing but the host. |
| D3 | The container's network | on: a bridge network in both compose files; production publishes no port and its proxy forwards to it | Host networking (`network_mode: host`) lets runs reach the host's loopback services and every port the host listens on. Neither compose file uses it. | A bridge network, with the port behind a TLS proxy. Costs: services bound to the host's loopback. |
| D4 | Home folders mounted from the host: `/root/.docker`, the registry logins, and `/root/.claude`, Claude Code's settings and sessions | on in production | To push docker images with a stored login, and to keep Claude Code's sign-in and sessions across deploys. Every run of every user reads both. | A registry token as a scoped secret and `docker login --password-stdin` in the job that pushes; no mount for Claude Code's settings. Costs: a login shared by every job, and Claude's sessions after a redeploy. |
| D5 | Secrets in the server's environment, such as `SECRETSMANAGER_KEY` | on in production | To let every job open the secretsmanager stores without a secret of its own. Every run inherits it, and no read log records it (`runner/host.go`, `inherited`). | A secret scoped to the folders or workspace that use it, with its read log. Costs nothing but the move. |
| D6 | Host keys: `StrictHostKeyChecking accept-new` | built in | To clone and push over SSH and to reach new hosts with no "continue connecting?" question: the first answer for a host becomes its key. | Seed `/keys/known_hosts` with the git remotes and inventory hosts: ssh checks a key the file holds, and `accept-new` adds only hosts it lacks. For the jobs' own ssh, `StrictHostKeyChecking yes` in the image's `ssh_config`; task-ui's git and inventory probe keep `accept-new`. Costs: reaching an unseeded host. |
| D7 | Which runs get an SSH key (**Who gets it**) and the git key | opt-in per key; a workspace key linked to nobody goes to every run of the workspace | To clone and push over SSH, and to deploy from any run of a workspace. A job pushes anywhere the key may. | Link each key to the roles and users that need it; scope keys on the git host to the repositories they serve; make the git key a read-only deploy key where task-ui only pulls. Costs: keys in the runs of users not linked. |
| D8 | The Beszel integration | opt-in | To list the hosts on Inventory, and to reach inventory hosts: `@host:<name>` and `@host:*` run a command over ssh with every instance SSH key, for users with `hosts`. | Leave it off: no inventory and no host commands (`hostrun/starter.go`, `beszel.ErrOff`). Or keep it on with `hosts` for admins alone and a token of a read-only Beszel user. |
| D9 | The Claude Code integration and its **Who** | opt-in; with **Who** empty, every run gets the token | To let Claude act in runs: `@claude`, `@claude-prompt` and `@claude-note` sign in with it. | **Who** naming the agent user and the roles that use Claude; or no integration, and no Claude in runs. |
| D10 | Prompts and notes to the agent (`prompt/edit`, the user flag **Run instance Claude when mentioned**) | the `claude` user holds `prompt/edit`; no role does | To let Claude act in runs on a prompt or a note. Claude Code runs with `--dangerously-skip-permissions`, any command as root. | `prompt/edit` for admins and the agent user; the flag on the agent user alone. Costs: prompts for everyone else. |
| D11 | **MCP agent user** | opt-in per user | To let an agent outside task-ui run jobs, read runs and artifacts, and post notes, with a 30-day token. | An account of its own for each agent, with `run` on chosen folders. Costs: an account to keep for each agent. |
| D12 | `key/read`, an agent's read of an SSH private key | admins alone; outside `key/all` | To return an SSH key to an agent for a task, encrypted to its recipient. | Keep it with admins. Costs: agents that fetch keys. |
| D13 | Interactive input: the `input` rule and jobs marked interactive | `input` for admins alone; `@shell`, `@claude` and jobs marked `interactive` take keystrokes | To answer a job's prompt from the run window, and to watch and type into shells together. | Keep `input` with admins; mark interactive only the jobs that need it. Costs: answering a waiting job from the browser. |
| D14 | Who sees a folder's runs | default: every user who sees the folder | To watch terminals live, an admin's `@shell` included, as a shared remote desktop. | Folder limits per user; admins work in folders viewers don't see. Costs: watching those runs. |
| D15 | Recording retention | default: 31 days, no size cap | To replay a run's terminal after it ends. | A shorter age and a size cap where jobs touch secrets. Costs: older replays; the run's row stays. |
| D16 | Secret scope | per secret | To reach every folder of a workspace with one secret (`workspace:<name>`). | A folder's own secret. Costs: one secret for many folders. |
| D17 | Env file path | per env file | To write an env file where a tool expects it, such as `deploy/app.env`. Secret values sit there in plain text, and the file preview serves it. | A dot path, such as `.env`, which task-ui seals. Costs: tools that read a fixed path without a dot. |
| D18 | Workflow webhooks: a git host calls `POST /api/workflows/{id}/runs` with a token | opt-in | To start a workflow on a push. A leaked hook token starts the workflow. | A user of its own with `run` on the workflow's folders alone, and the token's 30-day expiry; or no webhook. Costs: push-started workflows. |
| D19 | Workspace export and import (`workspace/edit`) | admins alone | To move a workspace with its secrets and SSH keys between instances. | Keep `workspace/edit` with admins; a passphrase of 12 characters or more; delete the file after the import. Costs non-admins nothing. |
| D20 | `TASK_UI_AUTOLOGIN` | opt-in; unset | To skip sign in, for development and screenshots. | Unset. |
| D21 | `TASK_UI_SESSION_TTL` | default: 12 hours | To stay signed in for a working day. | A shorter session. Costs more sign-ins. |

## Checklist

- [ ] Admins are people you trust with the host; each admin signs in with
      their own account ([D2](#decisions)).
- [ ] `prompt/edit`, `hosts`, `env/edit`, `project/edit` and
      `workspace/edit` go only to people trusted with the host
      ([D10](#decisions), [D8](#decisions), [D19](#decisions)).
- [ ] Runners run folders whose code they cannot change, or you trust
      them with the host; a folder's allow list names only those who
      may run every job, file and skill job of it
      ([Least access](#what-least-access-means-here)).
- [ ] Viewers that should only read have denies on the edit and delete
      rules of notes, workflows and schedules, and folder limits
      ([D14](#decisions)).
- [ ] The compose file mounts the docker socket only where jobs build or deploy;
      otherwise a dind sidecar, a socket proxy, or no docker
      ([D1](#decisions)).
- [ ] The container runs on a bridge network behind the TLS proxy
      ([D3](#decisions)).
- [ ] The compose file mounts registry logins and Claude Code's settings
      only where jobs need them ([D4](#decisions)).
- [ ] No secret sits in the server's environment; `SECRETSMANAGER_KEY`
      and the like live in scoped secrets ([D5](#decisions)).
- [ ] `/keys/known_hosts` holds the git remotes and inventory hosts
      before the first run ([D6](#decisions)).
- [ ] Each SSH key links to the roles and users that need it; the git
      key exists; no production key goes to a role it doesn't need
      ([D7](#decisions)).
- [ ] Beszel, when on, signs in as a read-only user, and `hosts` stays
      with admins ([D8](#decisions)).
- [ ] The Claude Code integration's **Who** names the agent user and
      the roles that use Claude ([D9](#decisions)).
- [ ] Each MCP agent user is an account of its own, with only what the
      agent needs, and its token rotates within 30 days
      ([D11](#decisions), [D18](#decisions)).
- [ ] `key/read` and `input` stay with admins ([D12](#decisions),
      [D13](#decisions)).
- [ ] Recording retention fits what the jobs print; jobs do not echo
      secrets ([D15](#decisions)).
- [ ] Each secret's scope is the folder or workspace that uses it
      ([D16](#decisions)).
- [ ] Env files use dot-file paths, such as `.env` ([D17](#decisions)).
- [ ] Backups of the database live outside the data folder, `0700`; a
      copy of the database and `/keys` together opens every secret, so
      no backup off the host holds both.
- [ ] Workspace exports use a long passphrase; delete the file after the
      import ([D19](#decisions)).
- [ ] `TASK_UI_AUTOLOGIN` is unset ([D20](#decisions)).
- [ ] `/login` has a rate limit at the proxy (see the findings).

## Stance and threat model

task-ui serves two jobs: development, where people and agents work
in checkouts with the tools of a full machine at hand, and the
orchestration of restricted production environments, where task-ui
holds the keys and credentials and the people who use it hold roles. The
first wants reach; the second wants control over who reaches what. The
image chooses reach and draws the line at who may run code.

What task-ui trusts:

- **The host.** The host's root reads the data and key folders, the
  database and the recordings.
- **The container.** task-ui and every run share one container and one
  OS user, root: the image sets no `USER`
  (`malta/dockerfiles/task/Dockerfile`).
- **Admins.** An admin passes every permission check
  (`permission/set.go`, `decide`), sees every folder and runs every job.

What the image grants by design (`malta/dockerfiles/task/compose.yml`):

- `/var/run/docker.sock`, the host's docker daemon, so jobs build and
  push images and start containers;
- `/keys`, every private key of the instance, encryption key included;
- `/root/.docker`, the registry logins jobs push with;
- `/root/.claude`, Claude Code's settings and sessions for every user;
- `SECRETSMANAGER_KEY` in the server's environment, which every run
  inherits (`runner/host.go`, `inherited`).

The consequence: **anyone who can make task-ui run a command of their
choosing is root on the host**, through `docker run -v /:/host`, and can
read every key and every secret of the instance. Running a command of
your choosing takes less than admin. Each of these does it, in the
container or, for `hosts`, on the inventory's hosts:

| Permission | How it runs a command of the user's choosing |
|------------|----------------------------------------------|
| admin | `@shell` and `@claude`, a login shell and Claude Code in a terminal: their allow list names the admin role, and no folder or job rule changes it (`folder/shell.go`, `folder/claude.go`); task-ui checks `shell` besides |
| `prompt/edit` | a prompt starts `@claude-prompt`, Claude Code with `--dangerously-skip-permissions` (`agent/claude.go`) |
| `run` with push access to a checkout | the job runs whatever the pushed `atkins.yml`, `Taskfile.yml`, `Makefile` or `scripts/*` says; `./<path>` runs any executable file of the folder |
| `run` and `note/edit` | a note naming the agent starts `@claude-note` on the thread |
| `hosts` | `@host:<name>` runs a command on an inventory host over ssh, with every instance login key |
| `project/edit` | adding a project clones a repository of the user's choosing, whose jobs then list |
| `env/edit` | an env file writes content to any relative path in a folder, such as `Makefile` |
| `workspace/edit` | a workspace import brings SSH keys, secrets and projects |

Grant these permissions only to people you would trust with root on the host.
Viewers watch and read; runners on folders whose code they cannot change
start what others wrote. Where people of different trust share one
host, run one instance per trust boundary, and keep production keys on
an instance that only its operators reach.

## Identity and access

### Users, passwords and sessions

Who may sign in is the password file, `.htpasswd` in the data directory,
bcrypt hashes only (`access/htpasswd.go`, `access/password.go`).
`task-ui backup <user>` makes ten single-use backup passwords of about 59
bits each, of which the database keeps only the hashes
(`access/backup_passwords.go`).

A sign in issues a session cookie: 256 random bits, `HttpOnly`,
`SameSite=Strict`, `Secure` behind TLS or `X-Forwarded-Proto: https`
(`access/guard.go`, `SetCookie`). The database keeps its SHA-256 hash, the
user and the expiry, `TASK_UI_SESSION_TTL`, 12 hours by default
(`access/session.go`). A new password or the user's removal ends their
sessions. Requests that change state from another site fail the origin
check (`access/handler.go`, `SameOrigin`), and the run window's
websocket refuses a foreign `Origin` (`api/attach.go`).

`TASK_UI_AUTOLOGIN` signs every request in as one user with no password.
It exists for development and the screenshot scripts; production leaves
it unset.

### Roles and policies

A user is an admin, a runner or a viewer (`model/user.go`). What a role
may do is a policy: one entry per rule, `allow`, `deny` or nothing, which
inherits. A user's own policy sits on top of the role's
(`permission/set.go`, `decide`):

1. an admin may do everything, whatever any policy says;
2. the user's own entry, when it has one, decides;
3. else the role's entry decides; nothing means no.

A section's `all` entry stands for its rules, and a `deny` on `all` wins
over an `allow` on a rule (`permission/policy.go`, `Effect`). A rule a
section adds, `key/read`, takes an entry of its own: `key/all` does not
grant it.

The rules live in one file, `config/permissions.yml`: `view`, `edit`,
`delete` for each section (workspaces, projects, notes, prompts,
workflows, schedules, artifacts, env files, secrets, instance keys,
integrations, inventory, users), `key/read`, and four application
rules, `run`, `shell`, `hosts` and `input`. `go run ./internal/cmd/rulelint`
fails the build when a check names a rule the file lacks, or a rule of
the file goes unchecked, so the matrix on Settings, Permissions is the
whole list.

By default (`schema/permission_policy.up.sql`) a runner may `run`,
and view, edit and delete notes, workflows and schedules, view the artifacts and
the inventory, and write the artifact tree; a viewer the same without
`run` and without writing artifacts. Neither holds `shell`, `hosts`,
`input`, `prompt/edit` or any key, secret, env, project or workspace
rule. The `claude` user holds `run`, notes and prompts in its own
policy.

On top of the rules, an admin may limit a user to some folders, and a
folder or a job may carry an allow list of users and roles
(`permission/set.go`, `Run`). A schedule runs as its maker; only an
admin picks another user, who must be able to run the job
(`api/schedule.go`, `runAs`). A user with `user/edit` changes policies
only when the role grants it, not when their own policy does
(`api/permission.go`).

### Tokens

| Token | Who holds it | Lifetime | Storage | Scope |
|-------|--------------|----------|---------|-------|
| Session cookie | a browser | `TASK_UI_SESSION_TTL` | SHA-256 hash | the user |
| MCP token, `tui_...` | a user with **MCP agent user** | 30 days, then a rotation (`access/token.go`, `MCPTokenTTL`) | SHA-256 hash | the user, over `/api` and `/api/mcp` |
| Run token | a Claude run that task-ui starts | the run | memory alone (`access/run_tokens.go`) | the run's user, its thread |

An admin turns on **MCP agent user**; a user then makes, rotates and
turns off their own token, and the token shows once, in the response
that makes it (`api/user_mcp.go`). `/api/mcp` refuses a user without the
flag unless a run token signed the request in (`access/handler.go`,
`RequireMCP`).

A run token exists for `@claude-prompt` and `@claude-note` runs. The
board mints it at start, writes it into an MCP configuration file in a
`0700` directory of its own, and revokes it and removes the directory
when the run ends (`note/board_mcp.go`). It signs in as the run's user
with that user's permissions, and nothing more. It cannot start a
prompt: `POST /api/prompts` answers `409` to it (`api/prompt.go`), so a
Claude run does not spawn more Claude runs.

### Interactive input

A window on a run shows output to everyone who may see the run. Whether
it types depends on the job:

- an interactive job, `@shell`, `@claude` or a job marked interactive,
  takes keystrokes from every window whose user may run that job
  (`api/attach.go`, `ReadOnly: !canRun`): any admin for `@shell` and
  `@claude`, any runner the allow lists name for a job marked
  interactive;
- a job that takes no input types only after **Take input**, which takes
  the `input` permission, which only admins hold by default
  (`api/attach.go`, `takesInput`).

## Keys

### Instance keys

Every key lives in one folder, `/keys` in the image (`TASK_UI_SSH_DIR`),
`0700`, each file `0600`, as `<name>.key` and `<name>.pub`
(`keystore/local.go`, `write`). The name ends in the algorithm and size,
such as `deploy_rsa4096`. The database says what each key is for, its
scope (the instance or a workspace) and whose runs get it. The page is
Settings, Keys (`/settings/keys`); the API is `/api/keys`.

A key has one of two purposes (`keystore/local.go`, `ErrPurpose`):

- **encrypt**: the instance's encryption key, one per instance, which
  encrypts every secret. task-ui makes it at the first start when the
  instance has none, as `task_encryption_ed25519`. A scope holds one encryption
  key (`keystore/local.go`, `free`).
- **ssh**: keys that authenticate runs, git and the inventory to hosts
  and remotes. A workspace holds any number for its own runs.

task-ui signs nothing with an instance key: there is no signing
purpose. A job may still use a key from its agent to sign, such as a git
commit with `gpg.format ssh`; that is the job's choice.

**Rotate Key** writes new material of the same algorithm and size under
the key's name and moves the old pair to `backup/<name>.key` and `.pub`
(`keystore/local_rotate.go`). The key's row shows
`BACKUP ACTIVE: backup/<name>` while the backup exists, and both serve:

- the encryption key decrypts with the new key and then the backup,
  and the rotation re-encrypts every secret for the new key;
- an SSH key goes into an agent with its backup after it, so a host that
  knows only the old public key still lets the run in.

A key keeps one backup. While it has one, Rotate Key answers `409`; an
admin adds the new public key where it belongs, then deletes the
backup, and the key rotates again.

### One key, one use

A key either encrypts secrets or authenticates SSH, never both, and the
folder holds a key material once, by its fingerprint: an upload of
material it already holds, under any name or purpose, answers `409`
(`keystore/local.go`, `add`). From there:

- no encryption key goes into a run's agent, git's ssh command, a link
  or the inventory's logins (`keystore/local_runs.go`, `AgentKeys`
  refuses one with `ErrNotSSH`);
- no SSH key is an age recipient or opens a secret (`secret/vault.go`);
- no route hands out an encryption key: `GET /api/keys/{name}/private`
  answers `403` for one (`api/key_read.go`), and a workspace export
  leaves them out (`keystore/bundle.go`).

This limits the blast radius. A leaked deploy key opens the hosts that
trust it and no secret; a leaked encryption key opens the database's
secrets and no host.

### Who gets an SSH key

An SSH key links to roles and users (Settings, Keys, **Who gets it**;
`PUT /api/keys/{name}`). A run gets (`keystore/local_runs.go`, `ForRun`):

- the instance SSH keys linked to the run's role or to its user;
- its workspace's SSH keys linked the same way, or linked to nobody,
  which goes to every run of the workspace.

An instance key linked to nobody stays with the instance: the inventory
signs in with it, runs don't. A host command, `@host:<name>` or
`@host:*`, gets every instance SSH key besides, as the inventory does
(`runner/keys.go`). One SSH key may be the git key, which task-ui's own
clones and pulls use (`keystore/local_git.go`). A typical layout: the
git key and a deploy key linked to the admin role, a key per user for
their own hosts, and a workspace build key linked to nobody.

### The run's ssh agent

A run gets its keys in an ssh agent of its own, served from memory on a
socket, and `SSH_AUTH_SOCK` names it (`runner/manager.go`, `execute`;
`sshagent/agent.go`). A run with no keys gets an empty `SSH_AUTH_SOCK`,
never the server's agent. The terminal's first line names the keys:
`$ ssh-agent: deploy_rsa4096, task_ssh_ed25519`, or `(none)`. The agent
closes when the run ends.

The agent decides which keys task-ui loads for a run. It does not decide
what a run can reach: a job runs as root beside `/keys` and reads every key file
there. Links are a statement of intent and keep honest jobs to their
keys; they are not a wall against a job that goes looking.

### Host keys

Every ssh in the image accepts a host's key on first contact and keeps
it in `/keys/known_hosts`: `StrictHostKeyChecking accept-new` in
`/etc/ssh/ssh_config.d/task-ui.conf` (Dockerfile), in git's ssh command
(`git/git.go`) and in the inventory's probe (`sshprobe/prober.go`). No
run waits on a "continue connecting?" question, and a key that changes
later fails the connection. The cost is trust on first use: whoever
answers for a host the first time is the host from then on. Seed
`known_hosts` with the keys of the git remotes and the inventory before
the first run reaches them.

### Private keys leave encrypted

A user with `key/read` and **MCP agent user** may fetch the private half
of an SSH key linked to them or their role, any SSH key for an admin,
with `GET /api/keys/{name}/private?recipient=age1...` or the MCP tool
`get_ssh_key` (`api/key_read.go`). The response holds the key encrypted
to the X25519 recipient the caller made for the request, never the key
in clear, with `Cache-Control: no-store`. Each read goes to the key read
log: the key, the user, the token and whether it came over MCP or the
API. `key/read` is outside `key/all`, so only admins hold it until a
policy grants it.

## Secrets

### Encryption at rest

A secret is a name and a value that runs get as an environment variable.
task-ui encrypts the value with age to the instance's encryption key, an
ed25519 or RSA key used through `agessh`, and stored armored
(`secret/vault.go`; `secret/store.go`, `Set`). The public key
encrypts; the private key in `/keys` decrypts. A copy of the database
alone opens nothing; a copy of the database and `/keys` opens
everything. Rotating the encryption key re-encrypts every secret
(`secret/store.go`, `Rekey`). Neither Settings, Secrets nor the API
shows a value after a save.

### Scope

A secret's scope decides which runs get it (`secret/store.go`,
`Reaching`):

- `workspace:<name>` reaches every folder of the workspace, and no
  folder of another workspace; `workspace:_` reaches the folders of the
  default workspace;
- a folder's own secret reaches that folder's runs and wins over a
  shared secret of the same name.

The scope keeps a secret out of another workspace's environment. It
does not keep it out of reach on disk: the database and the key that
opens it sit beside every run.

### The read log

Each run records one read per secret it gets: the secret, the run, the
folder, the directory, the job and the user (`runner/command.go`;
`storage/secret_read_storage.go`). Settings, Secrets shows each
secret's **Reads** and **Last read**, and `/settings/secrets/<id>/reads`
pages the log, newest first. task-ui keeps the last 1000 reads of each
secret.

### A run records nothing of its environment

A run's environment is sensitive, so task-ui writes nothing of it into
the run: not a value and not a name. The run's page, its output, its
recording and its row hold no secret and no variable name. The only
trace of which run got which secret is the read log on the secret's
page, which takes `secret/view`. The terminal's first line names the SSH
keys in the run's agent, by name; no key material and no fingerprint.

What a job prints, the recording keeps. A job that echoes a secret puts
it in the recording, readable by everyone who sees the run (see
[Terminals](#terminals)).

### Env files

An env file is a dotenv file a user fills for a folder (`envfile/`).
Before every run task-ui writes it into the folder, `0600`, with the
folder's secrets appended in plain text after a marker line
(`envfile/store.go`, `Generate`). The env file's page lists the names of
the appended secrets, never their values. The file on disk holds the
values: task-ui seals a path that starts with a dot, such as `.env`, which
stays off the folder's file listing and the file preview; any other
path is not (see the findings).

### Workspace export

A workspace exports as `<folder>.tui.age`: an age file encrypted to a
passphrase of at least 12 characters, through age's scrypt recipient
(`workspace/bundle.go`). Inside are the manifest, with the workspace,
its projects, its env files and its secrets in plain text, and the SSH
keys its runs get (`workspace/manifest.go`). No instance key and no
encryption key goes: the importing instance encrypts the secrets for its
own key. The passphrase travels in a `POST` body, out of URLs and logs
(`api/workspace_export.go`). The bundle is as strong as its passphrase,
and it carries the workspace's secrets and SSH keys: no file task-ui
produces needs more care.

## Terminals

Every run is a terminal on the server, a PTY, that belongs to no
browser window (`runner/`). The dashboard shows each terminal in
progress as a tile with its live screen; a click opens a window on it,
and `/api/runs/{id}/frame.svg` and `frame.txt` serve the screen as a
picture or text. Together this works as a shared remote desktop for
terminals: an admin opens `@shell` in a folder, a colleague attaches and
sees the same screen, and an agent's run shows next to it.

Identity carries into the terminal:

- `@shell`, a login shell, and `@claude`, Claude Code in a terminal, run
  for admins alone: their allow list names the admin role, and the
  task-ui checks the `shell` permission besides (`folder/shell.go`,
  `api/target.go`, `canRun`); `@claude` also takes a Claude token from
  the Claude Code integration.
- The shell runs with the opener's SSH keys in its agent, and with
  `TASK_UI_USER`, `TASK_UI_ROLE` and `TASK_UI_PERMISSIONS` in its
  environment (`runner/command.go`).
- Every run records who started it; schedules run as their user.

What a terminal shows, everyone who sees the run sees:

- a run shows to every user who sees its folder (`model/user.go`,
  `SeesRun`), viewers included, so a viewer watches an admin's `@shell`
  in that folder as it happens;
- an interactive run takes keystrokes from every window whose user may
  run the job (see [Interactive input](#interactive-input));
- the recording is an asciicast of the terminal's output
  (`recording/recorder.go`): what the screen showed, keystrokes included
  as the terminal echoes them, a password typed at a prompt that does
  not echo left out. Data retention keeps recordings 31 days by default
  (`retention/settings.go`); the run's row stays after its recording
  goes.

A secret printed in a shell, a token pasted on a command line, a
`cat .env`: each lands in the recording and on the screens of
everyone attached.

## Workspaces

A workspace is a directory under the root that holds projects, with its
own SSH keys, its own secrets and an optional home project
([Workspaces](guide/workspaces.md)). A run in a workspace with a home
project gets that project's directory as `HOME`, so atkins finds the
workspace's skills and Claude Code reads its `CLAUDE.md`
(`runner/home.go`, `HomeEnv`). What the server keeps in its own `HOME`
stays in use, through variables pointed back at it: atkins' stats, the
registry logins (`DOCKER_CONFIG`), the git identity, the XDG caches,
`GOPATH`, the shell history (`HISTFILE`) and Claude Code's settings
(`CLAUDE_CONFIG_DIR`). ssh reads `~/.ssh` from the password database,
not from `HOME`, so its configuration stays the server's.

What a workspace keeps apart: which secrets enter a run's environment
and which SSH keys enter its agent. What it does not keep apart: the
OS user, the container, the key folder, the database, every checkout
under the root, the registry logins, the shell history and Claude
Code's sessions. A job someone in one workspace can change can read
another workspace's secrets. Workspaces organise trusted people; they
do not separate untrusted ones.

## Agents and MCP

Claude Code runs inside task-ui in three ways: `@claude`, an interactive
terminal; `@claude-prompt`, a prompt from a page, the API or
`ask_prompt`; and `@claude-note`, a note to the agent. All three start
Claude Code with `--dangerously-skip-permissions` (`agent/shell.go`,
`agent/claude.go`): Claude runs commands without asking. What bounds it:

- **The user.** A prompt runs as the asking user, with that user's
  folders, keys and secrets. Claude's system prompt names the user, the
  role and the permissions (`agent/claude.go`, `context`), and its MCP
  connection is the run token, with that user's permissions.
- **The token.** The run token lives for the run, signs in as one user,
  and cannot start another prompt.
- **The container.** Claude runs as root in the task container, with
  `docker.sock`. Nothing inside the container bounds it beyond the user's
  own keys and secrets, and those only by convention (see
  [Stance](#stance-and-threat-model)).

The MCP server at `/api/mcp` serves its tools by calling the JSON API in
process as the token's user (`mcp/server.go`), so a tool checks what the
page checks. The tools: `whoami`, `list_jobs`, `list_workflows`,
`run_job`, `run_workflow`, `get_workflow`, `create_workflow`,
`update_workflow`, `list_runs`, `get_run`, `list_workflow_runs`,
`list_notes`, `note_comment`, `post_note`, `ask_prompt`,
`list_artifacts`, `read_artifact`, `write_artifact` and `get_ssh_key`.
`run_job` opens `@shell` or `@claude` for admins alone, and cannot
start `@claude-prompt` or `@claude-note`, which start from their own
routes ([MCP](guide/mcp.md)). A refused call names the permission it
needs.

An agent outside task-ui, such as Claude Code on a laptop, holds an MCP
token for 30 days. Treat it as the user's password: a dedicated user
with only the permissions the agent needs, `run` on chosen folders and
notes, is the usual setup.

## Findings

| Severity | Finding | Where | Recommendation |
|----------|---------|-------|----------------|
| High | The task container mounts the host's docker socket, so any user who can run a command of their choosing (admin, `prompt/edit`, `run` on code they can change, `env/edit`, `project/edit`, `workspace/edit`) is root on the host. | `malta/dockerfiles/task/compose.yml` | Grant those permissions only to people trusted with the host. Where builds need docker but users don't need the host, put a socket proxy that allows build and push alone in front of the daemon, or run one instance per trust boundary. |
| High | Every run is root in the container beside `/keys`, the database, `/root/.docker` and `/root/.claude`. Key links, ssh agents and secret scopes decide what task-ui loads for a run, not what it can read: any run reads every private key, the encryption key included, and so every secret. | `malta/dockerfiles/task/Dockerfile` (no `USER`), `keystore/docs.go` | State it to operators (this page, [Workspaces](guide/workspaces.md#what-a-workspace-keeps-apart)). For real separation, run jobs as another uid without read access to `/keys` and the data folder, or in a container of their own that gets only the run's agent socket. |
| Medium | An env file saved at a path that does not start with a dot, such as `deploy/app.env`, carries the folder's secret values in plain text, and the file preview serves the file to every user who sees the folder, viewers included. | `envfile/store.go` (`SafeRelative`), `api/file.go` (`viewFile`), `frontend/file.go` (`checkoutFile`) | Refuse an env file path whose last segment does not start with a dot on save, so task-ui seals the file; have both file viewers refuse a file that holds `envfile.Marker`. |
| Medium | A workspace export takes `workspace/edit` alone, yet returns the workspace's secrets in plain text and its SSH private keys, under a passphrase the caller picks. It goes around `secret/view` and `key/read`. | `api/workspace_export.go` | Require `key/read` and `secret/edit` besides `workspace/edit`, or an admin. |
| Medium | Runs inherit the server's environment. `SECRETSMANAGER_KEY`, set in production, reaches every run of every user and workspace, and no read log records it. | `runner/host.go` (`inherited`), `compose.yml` | Move it into a scoped secret, so scopes and the read log apply, or strip it from the inherited environment. |
| Medium | ssh trusts host keys on first use: `accept-new` everywhere ssh runs. The first answer for a host, right or wrong, becomes its key. | Dockerfile (`ssh_config.d/task-ui.conf`), `git/git.go`, `sshprobe/prober.go` | Seed `/keys/known_hosts` for the git remotes and inventory hosts; review it after adding hosts. |
| Low | The Claude Code token reaches every run of the users and roles **Who** names, every run when it names nobody, not only the Claude runs. | `integration/store.go` (`Env`) | Check a **Who** box; better, set the token only for `@claude`, `@claude-prompt` and `@claude-note`. |
| Low | An interactive run takes keystrokes from any window whose user may run the job: any admin types into another admin's `@shell`, which carries that admin's SSH agent, and any runner the allow lists name types into a job marked interactive that someone else started. Nobody records who typed. | `api/attach.go` (`serveAttach`) | Let only the starter type into `@shell` and `@claude`; log the users who attach with input. |
| Low | A job's allow list narrows that job alone. `./<path>`, any executable file of the folder, and `@skill:<job>` run under the folder's allow list, so a user the folder names runs them though the folder's listed jobs name others. | `folder/folder_path.go` (`FileJob`), `folder/skill.go` (`SkillJob`) | Name in a folder's allow list only users who may run every job of it, or have file and skill jobs take the strictest job allow list of the folder. |
| Low | Terminal output reaches every user who sees the folder and stays in the recording, 31 days by default. A recording is `0644` in a `0755` tree on the host. | `model/user.go` (`SeesRun`), `recording/store.go`, `retention/settings.go` | Limit folders per user where admins work in shells; shorten retention where jobs touch secrets; write recordings `0600` in `0700` folders. |
| Low | The production data folder holds 32 copies of the database and 58 of its WAL and SHM files, `task-ui.db*.pre-*`, `0644` in a `0755` folder. They hold secret ciphertext, password and token hashes. `/keys` still holds the old-layout `id_ed25519` and `id_rsa_deploy`, which may open secrets in the older copies. | `/mnt/apps/task/data`, `/mnt/apps/task/keys` | Move the copies to `/mnt/apps/task/backup/<date>-<name>/` under `0700`, or delete them; remove the old-layout key files once the migrated keys work. |
| Low | Sign in has no limit on attempts; bcrypt slows each guess, nothing stops many. | `access/authenticator.go`, the login handler | Rate limit `/login` at the proxy or in the handler. |
| Low | Every run's shell history goes to the server's `~/.bash_history`, across users and workspaces. | `runner/home.go` (`homeKept`) | Point `HISTFILE` per user, or nowhere for `@shell`. |
| Info | A run token's MCP configuration file sits in a `0700` temporary directory, owned by root, which every other run can read while the run lasts. | `note/board_mcp.go` | Covered by the High findings; separate uids would close it. |
| Info | `TASK_UI_AUTOLOGIN` signs every request in as one user. Production leaves it unset (checked). | `app/config.go` | Keep it out of production configuration. |

---

[Contents](README.md) · [Running the server](running.md) · [API](api.md)
