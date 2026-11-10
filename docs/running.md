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

- `/data`, `TASK_UI_DATA`: the database, the recordings, the artifacts, the password file, `/data/.htpasswd`, and the workspace folder, `/data/workspaces`, `TASK_UI_WORKSPACE`. The workspace folder holds `home/`, the `HOME` of every run at the root, `src/`, whose subfolders with something to run are the folders of the UI, and a directory per named workspace, each with a `home/` and a `src/` of its own ([Workspaces](guide/workspaces.md)).
- `/keys`, `TASK_UI_SSH_DIR`: the keys of the instance. Keep it out of `/root/.ssh`, where `ssh` in a job would find them by default; a run receives the keys its policies grant it.

`/app` is the working directory and `HOME`. It holds the image's own configuration: `entrypoint.d`, `.ssh`, and `.docker/config.json`, where `docker` reads the registry logins. Nothing a run writes goes there.

`/go`, `GOPATH`, is the third mount and the one that decides how long a restart takes. The image bundles `docker` and `docker compose` and builds nothing else in: `task` and `atkins` install at boot, from `/app/entrypoint.d`, into the default `/go/bin`, which is first on `PATH`. With `/go` mounted a restart is two seconds; without it the binaries, the module downloads and the compiled objects are all gone and a boot spends three minutes rebuilding tools that did not change. `GOCACHE` is set to `/go/cache` for that reason: it defaults under `HOME`, which is not a volume.

Every tool installs at `@latest`, or `@main` where the repository has no tag, and installs on every boot. A pinned version is a number somebody has to remember to raise, and skipping the install when the binary exists means a restart never picks up a release.

- To run without Docker, drop `TASK_UI_DOCKER` and mount no socket; a job that calls `docker` then fails on its own. Without `/data` and `/keys`, task-ui uses `./data` and `./keys`.
- A job's `docker run -v $PWD:/src` goes to the host's docker daemon, which resolves `$PWD` on the host, where `/data` has another path. Every run gets `TASK_UI_HOST_PWD`, its directory as the host sees it, and `TASK_UI_HOST_DATA`, the data directory there: write `docker run -v "${TASK_UI_HOST_PWD:-$PWD}:/src"`. task-ui reads the host paths from the mounts of its container; `TASK_UI_HOST_DATA` set in its environment overrides the data directory's.
- The key folder holds every key, `<name>.key` and `<name>.pub`, the encryption key among them: `task_encryption_ed25519` encrypts every secret, the workspaces' and the integrations' settings. A rotated key keeps its backup in `backup/` until an admin deletes it. task-ui makes the folder `0700` and each file `0600`. Back up `/keys` with `/data`: without the encryption key the secrets don't open. A copy of `/data` alone holds no key.

## Provisioning the image

`/usr/local/bin/entrypoint.sh` runs every `*.sh` of two directories, in lexical order, under `sh -x` so the run is traced in the container's log, and then execs `task-ui`:

- `/app/entrypoint.d`, baked into the image from `docker/app/entrypoint.d` of this repository. `task` and `atkins` are installed here and nowhere else.
- `/data/entrypoint.d`, on the mounted volume. Another repository drops scripts here to add software without rebuilding the image, and `apt-get`, `go install` and `pip` all work: the base image is Debian with the Go toolchain and the docker CLI, which is enough to install the rest.

A missing or empty directory is not an error, so an instance with no provisioning starts. A script that fails stops the boot, because a half-provisioned instance that serves is worse than one that does not come up: its jobs fail one at a time, somewhere else, with nothing naming the cause. `exit 1` in `run_scripts` is the one word to change to make a failure a warning.

PID 1 of the image is `docker-init`, tini, and the entrypoint runs under it with `-g`. A script that backgrounds a daemon leaves it parented to PID 1 once that script's shell exits, and tini is what reaps it; `-g` signals the process group, so `docker stop` reaches the daemon as well as the server.

## Docker and registries

A job that builds an image needs a daemon to build on and, to push, a login to the registry. The two are separate choices and neither is a step on one ladder.

Which daemon:

| Arrangement           | What the compose file does                                                   | What a job reaches                            | What it costs the host                        |
|-----------------------|------------------------------------------------------------------------------|-----------------------------------------------|-----------------------------------------------|
| The host's daemon     | mounts `/var/run/docker.sock:ro`                                             | every image, container and volume on the host | a job can read and start anything on the host |
| A daemon in the image | `TASK_UI_DOCKER=dind`, `privileged: true` and a volume for `/var/lib/docker` | its own images, containers and volumes        | more than the socket does: privileged is root |

A daemon inside the container is better isolation of the docker state and worse isolation of the host, because it needs `privileged: true`. Reach for it when a job must not see the host's containers, not when the host must be protected from the job.

`docker/docker-compose.yml` ships the second, with the socket line commented out: most hosts this runs on have no socket to mount. `docker/app/entrypoint.d/05-dockerd.sh` starts `dockerd` when `TASK_UI_DOCKER` is `dind`, waits for it to answer, and does nothing at all when a socket is already mounted on `/var/run/docker.sock` - the mount is the operator naming which daemon to use, read off `/proc/self/mountinfo` and not off the file, because the daemon's own socket from a previous boot is a file in the container's layer and `restart: always` restarts the same container. The daemon, its shim, `runc` and the two helpers are the release's static binaries, copied from `docker:29-dind`; docker 29 builds containerd into `dockerd`, so there is no separate containerd to run.

The volume on `/var/lib/docker` is required, and not as a cache. The layers the daemon writes there are overlayfs, and an overlayfs upperdir may not itself sit on overlayfs: without the volume `dockerd` starts, pulls an image, and then fails every container with `failed to mount ... fstype: overlay ... invalid argument`. With it, `docker run` works in a privileged container on a cgroup v2 host, measured on docker 29.9.3.

Which registry login:

| Arrangement       | Mount                                        | Who logs in                              | The integration                             |
|-------------------|----------------------------------------------|------------------------------------------|---------------------------------------------|
| The host's logins | `$HOME/.docker:/app/.docker`                 | the host, and task-ui into the same file | on, and its saves reach the host's file too |
| task-ui's own     | a named volume on `/app/.docker`, or nothing | task-ui alone                            | on                                          |

task-ui stores a registry, a username and a password in the integration, encrypted with the instance key, and runs `docker login --password-stdin` for each on save and at boot. It writes no JSON itself, so whatever the installed CLI writes is what is on disk, a `credsStore` included. The database holds the pairs because a form has to list and edit them and because a container rebuild wipes the filesystem.

**Mount the directory, never the file.** `-v $HOME/.docker/config.json:/app/.docker/config.json` makes every credential change a silent no-op: the CLI writes a sibling file and renames it over `config.json`, the rename answers `EBUSY` on a bind-mounted file, and the CLI discards the error and reports success. `docker logout reg.example` prints `Removing login credentials for reg.example`, exits 0, and leaves the credential in the file. `:ro` makes no difference, and neither does running as root. A directory mount renames correctly.

A job that runs `docker login` itself writes `$HOME/.docker/config.json`, and a run's `HOME` is its workspace home, which is a git checkout: the login would land as an untracked file a folder page then reports. Runs therefore get `DOCKER_CONFIG=/app/.docker`, so a job's own login and the integration's reach the same file.

## Configuration

The environment sets the whole configuration:

| Variable                   | Default              | Meaning                                                                                                                                                               |
|----------------------------|----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `TASK_UI_LISTEN`           | `:3000`              | Listen address                                                                                                                                                        |
| `TASK_UI_DATA`             | `/data`, else `data` | Database, recordings, artifacts, the password file `.htpasswd`, the workspace folder                                                                                  |
| `TASK_UI_WORKSPACE`        | `<data>/workspaces`  | The workspace folder: the default workspace `internal`, its `home` and `src`, and a directory per named workspace                                                     |
| `TASK_UI_DISCOVER`         | `true`               | List every folder under the workspace folder with something to run                                                                                                    |
| `TASK_UI_DB`               | `<data>/task-ui.db`  | Database file (sqlite)                                                                                                                                                |
| `TASK_UI_SSH_DIR`          | `/keys`, else `keys` | The key folder, every key of the instance                                                                                                                             |
| `TASK_UI_HOST_DATA`        |                      | The data directory on the docker host; empty reads it from the container's mounts                                                                                     |
| `TASK_UI_SESSION_TTL`      | `7d`                 | How far ahead of its last use a session stays valid; a session in use keeps this window ahead of it                                                                   |
| `TASK_UI_SESSION_MAX`      | `30d`                | The ceiling from the sign in, which no extension passes; at or below the TTL a session lives one TTL and slides none                                                  |
| `TASK_UI_KEEP`             | `15m`                | How long a finished run stays attachable with its frame in memory                                                                                                     |
| `TASK_UI_GRACE`            | `10s`                | How long a stopped process gets before task-ui kills it                                                                                                               |
| `TASK_UI_ARTIFACT_HISTORY` | `50`                 | Commits the artifact history keeps before it squashes; 0 keeps all                                                                                                    |
| `TASK_UI_COMPOSE_PROJECT`  |                      | The compose project task-ui runs in, whose `compose: up` and `down` ask first; empty reads its container's compose label                                              |
| `TASK_UI_URL`              |                      | The public base link, such as `https://task.example.com`, that the links to copy start with; empty reads it from each request and the proxy's `X-Forwarded-*` headers |
| `TASK_UI_AUTOLOGIN`        |                      | Sign every request in as this user, for development only                                                                                                              |
| `TASK_UI_NOW`              |                      | Pin the clock the hours read (RFC 3339), for the screenshots only                                                                                                     |

The flags `-l`, `--workspace` and `--data` override the matching variables; `-v` prints the build. `TASK_UI_ROOT` and `TASK_UI_HTPASSWD` set stop task-ui at start, with where the folders and the password file go. `/healthz` answers `ok` to anyone, for a proxy or a compose health check.

`task-ui migrate-layout` brings a workspace folder laid out as before, the projects right under the root and under each workspace, to the `home` and `src` layout: it reads the configuration and the database, moves each workspace's home project to `home/`, every other directory of the workspace to `src/`, and makes `home/` and `src/` where they are not, printing each move; `--dry-run` prints them and moves nothing. Run it with the server stopped. A folder laid out already prints nothing to move, and the command runs again without harm.

## Users and roles

Login is mandatory, through the sign in screen; there is no HTTP basic auth. Who may sign in is the password file, `.htpasswd` in the data directory: one `name:hash` per line, the htpasswd format, bcrypt hashes only. The file is the truth. task-ui reads it on startup and again whenever it changes. Adding, resetting or removing a user needs no restart: a user not in the file cannot sign in and the instance removes them. A file that is not there yet is a warning: the instance reads it once `task-ui passwd` makes it.

A sign in sets a session cookie that holds only a random token. The database keeps the hash of the token, the user and the expiry. A session survives a restart of the server and ends when it expires, on sign out, when the file removes the user, or when the user gets a new password.

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
task-ui forgejo-sync --dry-run               # the Forgejo integration's plan, workspaces, homes, projects and hooks; without the flag, applies it; --no-color
task-ui get-secret GITHUB_TOKEN              # in a run: prints a secret of the run's folder, for a script that reads it where it needs it
task-ui -v                                   # the version and the commit
```

`task-ui agent note|prompt|shell [args]` are the host commands of the Claude built-ins, `@claude-note`, `@claude-prompt` and `@claude`: the server runs them in a run's terminal, and they read the run's environment (see [Notes](guide/notes.md#the-claude-jobs)).

task-ui keeps what a user signs in with apart from the user row, and no reader of a user, the API included, holds a credential: `user_auth_password` holds the password hashes, `user_email` the addresses and `user_auth_token` the token hashes. The `email`, `backup` and `token` commands work on the database of a user the server has read from the password file. To remove a user, delete the line, or **Delete** the user on Settings, **Users**, or call `DELETE /api/users/{id}`; both take Allow Deletes on Users, `user/delete`, and remove the line and the user's sessions, tokens and policy at once.

The database holds what each user is: the admin flag, **Administrative privileges (all)**, which allows everything, and the role, runner or viewer, whose policy decides the rest; what each role may do is in [Settings, Permissions](guide/settings.md#permissions). The password file brings users in:

- The first read into an empty database makes every user in the file an admin.
- A file of one user makes that user an admin while no user is one.
- A name new to a database with users comes in as a viewer, the lowest role, with no flags.

After that an admin sets the flags on the user's profile, or with `task-ui user`, which works on the database without the server and keeps the flag on the last admin unless `--force`. Two more flags make a user an agent: **MCP agent user** lets the user's token reach the MCP endpoint, and **Run instance Claude when mentioned** has a note naming the user start the Notes agent as them (see [Notes](guide/notes.md)).

An instance set up before the database held the roles had them in `TASK_UI_ADMINS`, `TASK_UI_VIEWERS` and `TASK_UI_AGENT`. The first start over its database applies them once: the admins get the flag, the viewers the role, and the agent user both agent flags. It logs the import and records it, and reads the variables no more; remove them from the environment after. `TASK_UI_AGENT_JOB` is no setting: a note naming an agent user runs the Notes agent, `@claude-note`.

---

[Contents](README.md) · [API](api.md) · [Development](development.md)
