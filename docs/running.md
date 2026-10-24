# Running the server

Start the Docker image with `task docker:run`, or use the compose setup:

```yaml
services:
  runner:
    image: titpetric/task-ui
    restart: always
    ports:
    - 3000:3000
    volumes:
    - $PWD/data:/data
    - $PWD/keys:/keys
    - /var/run/docker.sock:/var/run/docker.sock:ro
```

Two mounts hold everything the instance keeps:

- `/data`, `TASK_UI_DATA`: the database, the recordings, the artifacts,
  the password file, `/data/.htpasswd`, and the workspace folder,
  `/data/workspaces`, `TASK_UI_WORKSPACE`. Each subfolder of the
  workspace folder with something to run becomes a folder in the UI,
  and the workspace folder itself is one too when it has.
- `/keys`, `TASK_UI_SSH_DIR`: the keys of the instance. Keep it out of
  `/root/.ssh`, so `ssh` in a job does not find them by default; a run
  receives the keys its policies grant it.

`/app`, the working directory, holds the image's own configuration and
nothing a run writes; mount nothing there. The image bundles `task`,
`atkins`, `docker` and `docker compose`.

- To run without Docker, remove the volume for `docker.sock`. Without
  `/data` and `/keys`, task-ui uses `./data` and `./keys`.
- A job's `docker run -v $PWD:/src` goes to the host's docker daemon,
  which resolves `$PWD` on the host, where `/data` has another path.
  Every run gets `TASK_UI_HOST_PWD`, its directory as the host sees it,
  and `TASK_UI_HOST_DATA`, the data directory there: write `docker run
  -v "${TASK_UI_HOST_PWD:-$PWD}:/src"`. task-ui reads the host paths
  from the mounts of its container; `TASK_UI_HOST_DATA` set in its
  environment overrides the data directory's.
- The key folder holds every key, `<name>.key` and `<name>.pub`, the
  encryption key among them: `task_encryption_ed25519` encrypts every
  secret, the workspaces' and the integrations' settings. A rotated
  key keeps its backup in `backup/` until an admin deletes it. task-ui
  makes the folder `0700` and each file `0600`. Back up `/keys` with
  `/data`: without the encryption key the secrets don't open. A copy of `/data` alone holds no key.

## Configuration

The environment sets the whole configuration:

| Variable              | Default              | Meaning                                                           |
|-----------------------|----------------------|-------------------------------------------------------------------|
| `TASK_UI_LISTEN`      | `:3000`              | Listen address                                                    |
| `TASK_UI_DATA`        | `/data`, else `data` | Database, recordings, artifacts, the password file `.htpasswd`, the workspace folder |
| `TASK_UI_WORKSPACE`   | `<data>/workspaces`  | The workspace folder: the folders with jobs, the default workspace `_` |
| `TASK_UI_DISCOVER`    | `true`               | List every folder under the workspace folder with something to run |
| `TASK_UI_DB`          | `<data>/task-ui.db`  | Database file (sqlite)                                            |
| `TASK_UI_SSH_DIR`     | `/keys`, else `keys` | The key folder, every key of the instance                         |
| `TASK_UI_HOST_DATA`   |                      | The data directory on the docker host; empty reads it from the container's mounts |
| `TASK_UI_SESSION_TTL` | `12h`                | How long a sign in lasts                                          |
| `TASK_UI_KEEP`        | `15m`                | How long a finished run stays attachable with its frame in memory |
| `TASK_UI_GRACE`       | `10s`                | How long a stopped process gets before task-ui kills it           |
| `TASK_UI_ARTIFACT_HISTORY` | `50`            | Commits the artifact history keeps before it squashes; 0 keeps all |
| `TASK_UI_COMPOSE_PROJECT` |                  | The compose project task-ui runs in, whose `compose: up` and `down` ask first; empty reads its container's compose label |
| `TASK_UI_URL`         |                      | The public base link, such as `https://task.example.com`, that the links to copy start with; empty reads it from each request and the proxy's `X-Forwarded-*` headers |
| `TASK_UI_AUTOLOGIN`   |                      | Sign every request in as this user, for development only          |
| `TASK_UI_NOW`         |                      | Pin the clock the hours read (RFC 3339), for the screenshots only |

The flags `-l`, `--workspace` and `--data` override the matching
variables; `-v` prints the build. `TASK_UI_ROOT` and `TASK_UI_HTPASSWD`
set stop task-ui at start, with where the folders and the password file
go. `/healthz` answers `ok` to anyone, for a proxy or a compose
health check.

## Users and roles

Login is mandatory, through the sign in screen; there is no HTTP basic
auth. Who may sign in is the password file, `.htpasswd` in the data
directory: one `name:hash` per line, the
htpasswd format, bcrypt hashes only. The file is the truth. task-ui reads
it on startup and again whenever it changes, so adding, resetting or
removing a user needs no restart: a user not in the file cannot sign in
and the instance removes them. A file that is not there yet is a
warning, not an error: the instance reads it once `task-ui passwd` makes
it.

A sign in sets a session cookie that holds a random token and nothing
else. The database keeps the hash of the token, the user and the expiry,
so a session survives a restart of the server and ends when it expires,
on sign out, when the file removes the user, or when the user gets a new
password.

```sh
task-ui passwd alice                         # add or reset; the password from stdin
task-ui passwd alice s3cret                  # or as an argument
htpasswd -B .htpasswd alice                  # the same with apache2-utils; -B for bcrypt
task-ui hash s3cret                          # a hash, for editing the file by hand
task-ui email alice add alice@example.com    # an address alice signs in with; `rm` removes
task-ui backup alice                         # ten backup passwords, each signs in once
task-ui token alice cli                      # a token for a client such as an MCP agent
task-ui token alice                          # the names of alice's tokens, and when each was last used
task-ui token alice rm cli                   # removes a token
task-ui user list                            # every user with the role and the flags
task-ui user admin alice on                  # the admin flag; off for the last admin needs --force
task-ui user role alice viewer               # runner or viewer, the role without the flag
task-ui user mcp alice on                    # MCP agent user; off ends alice's MCP token
task-ui user agent alice on                  # Run instance Claude when mentioned, with MCP
task-ui channel --url https://task.example.com --token-file task.token  # pushes alice's notes into a Claude Code session
task-ui cleanup --dry-run                    # what the Data retention rules would remove; without the flag, removes it
task-ui get-secret GITHUB_TOKEN              # in a run: prints a secret of the run's folder, for a script that reads it where it needs it
task-ui -v                                   # the version and the commit
```

`task-ui agent note|prompt|shell [args]` are the host commands of the
Claude built-ins, `@claude-note`, `@claude-prompt` and `@claude`: the
server runs them in a run's terminal, and they read the run's environment
(see [Notes](guide/notes.md#the-claude-jobs)).

task-ui keeps what a user signs in with apart from the user row, so nothing
that reads a user, the API included, carries a credential:
`user_auth_password` holds the password hashes, `user_email` the
addresses and `user_auth_token` the token hashes. The `email`, `backup`
and `token` commands work on the database of a user the server has read
from the password file. To remove a user, delete the line, or
**Delete** the user on Settings, **Users**, or call `DELETE
/api/users/{id}`; both take Allow Deletes on Users, `user/delete`, and
remove the line and the user's sessions, tokens and policy at once.

The database holds what each user is: the admin flag, **Administrative
privileges (all)**, which allows everything, and the role, runner or
viewer, whose policy decides the rest; what each role may do is in
[Settings, Permissions](guide/settings.md#permissions). The password
file brings users in:

- The first read into an empty database makes every user in the file an
  admin.
- A file of one user makes that user an admin while no user is one.
- A name new to a database with users comes in as a viewer, the lowest
  role, with no flags.

After that an admin sets the flags on the user's profile, or with
`task-ui user`, which works on the database without the server and
keeps the flag on the last admin unless `--force`. Two more
flags make a user an agent: **MCP agent user** lets the user's token
reach the MCP endpoint, and **Run instance Claude when mentioned** has a
note naming the user start the Notes agent as them (see
[Notes](guide/notes.md)).

An instance set up before the database held the roles had them in
`TASK_UI_ADMINS`, `TASK_UI_VIEWERS` and `TASK_UI_AGENT`. The first start
over its database applies them once: the admins get the flag, the
viewers the role, and the agent user both agent flags. It logs the
import and records it, and reads the variables no more; remove them from
the environment after. `TASK_UI_AGENT_JOB` is no setting: a note naming
an agent user runs the Notes agent, `@claude-note`.

---

[Contents](README.md) · [API](api.md) · [Development](development.md)
