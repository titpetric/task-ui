# Running the server

Start the Docker image with `task docker:run`, or use the compose setup:

```yaml
services:
  runner:
    image: titpetric/task-ui
    restart: always
    environment:
      TASK_UI_SSH_DIR: /keys
    ports:
    - 3000:3000
    volumes:
    - $PWD/app:/app
    - $PWD/data:/data
    - $PWD/keys:/keys
    - /var/run/docker.sock:/var/run/docker.sock:ro
```

Mount your folders at `/app`. Each subfolder with something to run
becomes a folder in the UI, and `/app` itself is one too when it has.
The database, the recordings and the artifacts live under `/data`, and
the users sign in from `/app/.htpasswd`. The image bundles `task`,
`atkins`, `docker` and `docker compose`.

- To run without Docker, remove the volume for `docker.sock`.
- Mount `/keys` with `TASK_UI_SSH_DIR` to keep the SSH keys managed from
  the UI. Keep it out of `/root/.ssh`, so `ssh` in a job does not find
  them by default; a run receives the keys its policies grant it.
- The keys that encrypt live under `/data`, apart from the SSH keys:
  `TASK_UI_DATA/instance/encrypt/id_ed25519` encrypts the integrations'
  settings, and `TASK_UI_DATA/workspaces/<id>/encrypt` each workspace's
  secrets. task-ui makes them at start, directories `0700` and files
  `0600`. Back up `/data` with the database: without these keys the
  secrets don't open.

## Configuration

The environment sets the whole configuration:

| Variable              | Default              | Meaning                                                           |
|-----------------------|----------------------|-------------------------------------------------------------------|
| `TASK_UI_LISTEN`      | `:3000`              | Listen address                                                    |
| `TASK_UI_ROOT`        | `.`                  | Root folder holding the folders                                   |
| `TASK_UI_DISCOVER`    | `true`               | List every folder under the root with something to run            |
| `TASK_UI_HTPASSWD`    | `.htpasswd`          | The password file, looked up in the root                          |
| `TASK_UI_ADMINS`      |                      | Users with the admin role, comma separated                        |
| `TASK_UI_VIEWERS`     |                      | Users with the viewer role, comma separated; the rest are runners |
| `TASK_UI_DATA`        | `data`               | Database, recordings, artifacts, workspace keys and the encryption key |
| `TASK_UI_DB`          | `<data>/task-ui.db`  | Database file (sqlite)                                            |
| `TASK_UI_SSH_DIR`     | `$HOME/.ssh`         | Folder the SSH key settings manage; its keys only authenticate    |
| `TASK_UI_SESSION_TTL` | `12h`                | How long a sign in lasts                                          |
| `TASK_UI_KEEP`        | `15m`                | How long a finished run stays attachable with its frame in memory |
| `TASK_UI_GRACE`       | `10s`                | How long a stopped process gets before task-ui kills it           |
| `TASK_UI_ARTIFACT_HISTORY` | `50`            | Commits the artifact history keeps before it squashes; 0 keeps all |
| `TASK_UI_AGENT`       |                      | The user a note names to send it an instruction, such as `claude` |
| `TASK_UI_AGENT_JOB`   | `@claude-note` | The job an instruction runs: `@claude-note`, the Notes agent in the root, or a job of a folder as `folder/job`; set empty, it waits for MCP. Another built-in, or one under a folder, as `_/@claude-note` or `internal/claude-note`, stops task-ui at start with an error |
| `TASK_UI_COMPOSE_PROJECT` |                  | The compose project task-ui runs in, whose `compose: up` and `down` ask first; empty reads its container's compose label |
| `TASK_UI_URL`         |                      | The public base link, such as `https://task.example.com`, that the links to copy start with; empty reads it from each request and the proxy's `X-Forwarded-*` headers |
| `TASK_UI_AUTOLOGIN`   |                      | Sign every request in as this user, for development only          |
| `TASK_UI_NOW`         |                      | Pin the clock the hours read (RFC 3339), for the screenshots only |

The flags `-l`, `--root`, `--htpasswd` and `--data` override the
matching variables; `-v` prints the build.

## Users and roles

Login is mandatory, through the sign in screen; there is no HTTP basic
auth. Who may sign in is the password file, `.htpasswd` in the root or
wherever `TASK_UI_HTPASSWD` points: one `name:hash` per line, the
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
task-ui channel --url https://task.example.com --token-file task.token  # pushes alice's notes into a Claude Code session
task-ui cleanup --dry-run                    # what the Data retention rules would remove; without the flag, removes it
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

`TASK_UI_ADMINS=alice,bob` makes those two admins, `TASK_UI_VIEWERS=support`
a viewer, and everyone else in the file is a runner; what each role may
do is in [Settings, Permissions](guide/settings.md#permissions).

---

[Contents](README.md) · [API](api.md) · [Development](development.md)
