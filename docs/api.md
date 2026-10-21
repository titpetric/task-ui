# API

Everything the pages do calls `/api`, JSON over the session cookie,
or over a token in an `Authorization: Bearer` header (`task-ui token
<user> <name>` prints one). Writes need the same origin. The browser
pages are the reference client; `scripts/run.sh` is the curl one.

The surface below, and the MCP tools on it, stay frozen: a change adds a
route or a tool only for a need these cannot serve, and states the reason.

| Added | For |
|-------|-----|
| `DELETE /api/artifacts` | **Delete all** on the Artifacts page: deleting the records one by one leaves the files and their history |
| `/api/runs/{id}/summary` | **Summary** on the Runs page: Claude's summary of a run, written once and kept |
| `/api/workflow-runs/{id}/summary` | **Summary** on a workflow's runs: Claude's summary of every step, written once and kept |
| `POST /api/users` | **Create user** on the Users page: the password file had no writer but `task-ui passwd` on the server |
| `/api/artifacts/files/{path}` | The preview window: a file of the tree, rendered, with its versions, and the delete of one file, which the records cannot reach |
| `/api/notes`, `list_notes`, `note_comment` | Notes, and instructions to the agent, which reads and reports on them over MCP |
| `/api/integrations`, `/api/inventory` | Settings, Integrations, which keeps the services' tokens encrypted, the inventory from Beszel, and commands run on its hosts |
| `/api/workspaces`, `workspace` on `POST /api/projects` | Workspaces: a checkout whose folder is the namespace of the projects cloned inside it, a resource the project routes cannot express. Any number of workspaces may share a repository, and deleting one takes its projects with it |
| `POST /api/workspaces/{folder}/prompt` | **Prompt** on a workspace's folder page: Claude Code on a prompt in the workspace's checkout, a run of `internal/claude-prompt` whose environment, `TASK_UI_PROMPT` and the workspace, `POST /api/runs` cannot carry |
| `sort` on `PUT /api/workspaces/{folder}` | **Move up** and **Move down** on the Workspaces page: the order of the workspaces in the sidebar, which no field held |
| `/api/workspaces/{folder}/keys`, `/keys/generate`, `/keys/upload`, `/keys/{name}` | **Keys** on a workspace's configure page: a workspace's own encryption key and SSH keys, and the grants that narrow an SSH key to roles and users. `/api/keys` holds the instance's keys alone, with no notion of a workspace |
| `workspace` on `PUT /api/projects/{folder}` | **Workspace** on a project's configure page: moves the project's checkout, row and the users' folder grants to another workspace. No route changed a folder's ID. `left` lists the schedules, secrets, env files and workflow steps that still name the old one |
| `workspace:<name>` scope on `/api/secrets`, `rekeyed` on `POST /api/workspaces/{folder}/keys/generate` and `/upload` | Secrets per workspace: a workspace's shared secrets, encrypted with its key and kept from every other workspace. `all` or no scope is the default workspace's, `workspace:_`; a rotation of a workspace's key writes its secrets again |
| `POST /api/workspaces/{folder}/export`, `POST /api/workspaces/import` | **Export** on a workspace's configure page and **Import** on the Workspaces page: a workspace moves between instances with its keys, its grants and its secrets' ciphertext, which no route hands out. A POST keeps the passphrase out of the URL |
| `GET /api/files/{folder}/{path}` | The preview window on a folder page's **Files**: a file of the checkout, rendered as `/api/artifacts/files/{path}` renders an artifact. `/api/sources` serves only the files a pipeline reads, and no route served the rest |
| `./<path>` as `job` on `POST /api/runs` | **Run** beside a file that runs on a folder page's **Files** and in its window: the file as a job of its folder, under the folder's allow rules and hours, with the env, keys and pull of any job. A file outside the scripts folder had no job ID. `run_job`, schedules and workflow steps resolve a job the same way and take it too |
| `POST /api/keys/encrypt/rotate`, `encrypt` and `overlaps` on `GET /api/keys` | The instance row of Settings, Keys: the instance's encryption key, kept apart from the SSH keys under `TASK_UI_DATA/instance/encrypt`, has no name among them for `/api/keys/generate` to rotate. A key either encrypts or authenticates, so `overlaps` names a key held for both, and an upload of key material held for the other use, here or in any workspace, answers `409` |
| `dir` and `args` on `POST /api/runs`, `dir` and `args` on run rows, `takes_args` on jobs | The job cards of a directory inside a folder, such as `malta/dockerfiles` in the checkout of `malta`: that directory's own targets, run there under the folder's rules, which no folder or job ID named, and arguments for an atkins job that declares `args: true`, which reach it after `--` as argv. Run again repeats both. Schema `run_dir_args` adds the two columns |
| `GET /api/folders/{id}/actions?dir=`, `@<skill>:<job>` as `job`, `dir` on `GET /api/runs` | The actions bar of a folder page and of a directory inside a folder: the git line of the checkout and the jobs of the atkins skills that apply in that directory. `/api/folders` lists the catalogued jobs only, and no route read `git status` or asked atkins for skills; a skill job had no job ID. A skill job runs `atkins <skill>:<job>` under the folder's rules; `run_job`, schedules and workflow steps take it too. `dir` with `job` on the runs listing keeps one directory's job apart from the folder's job of the same name. Read only, for whoever sees the folder |
| `summarize` on `POST /api/inventory/run` | **Run and summarize** on the Inventory page: Claude reads every host's output as one report, after the run, in the same request |

| Method | Path | Does |
|--------|------|------|
| `GET` | `/api/folders?folder=` | the folders you can see and their jobs, with `can_run` and `last_run` |
| `GET` | `/api/folders/{id}/actions?dir=` | the actions bar of a folder, or of `dir` inside it: `git` (`checkout`, `branch` or `detached`, `upstream`, `ahead`, `behind`, `dirty`), `skills` with their `dir`, `when` and `jobs`, and the place's own `jobs`, each job with `can_run` and `last_run`; `404` for a directory the folder's Files would not open |
| `GET` | `/api/runs?folder=&job=&dir=&state=&limit=` | the runs, newest first, with `live`, `attached` and `can_run`; with `job`, `dir` names the directory of the job, empty for the folder's own |
| `POST` | `/api/runs` `{folder, dir, job, args, force, reason, cols, rows}` | starts a run; `job` is a job's ID, `./<path>` for a file of the folder that runs, executable or with a shebang, and not a dot file, or `@<skill>:<job>` for a job of an atkins skill that applies in `dir`, which runs where the skill matched; `dir`, a slash path inside the folder, makes `job` one of that directory's own targets, run there, `404` for a directory the folder's Files would not open; `args`, a line split as a shell splits words with nothing expanded, or a list, at most 32 of at most 512 bytes, follow the job's argv after `--`, `400` for a job without `takes_args`; `409` with the run in the way unless forced; `403` out of hours unless forced with a reason |
| `GET` | `/api/runs/{id}` | one run |
| `POST` | `/api/runs/{id}/stop`, `/kill` | ends a run |
| `GET` | `/api/runs/{id}/frame.svg`, `/frame.txt`, `/cast` | the last frame, the recording |
| `WS` | `/api/runs/{id}/attach` | binary frames are output; text frames are `{"type":"input"\|"resize"\|"stop"\|"kill"}` in and `{"type":"status"}` out |
| `GET` `DELETE` | `/api/artifacts`, `/api/artifacts/{id}` | list, download a version, delete a record; `DELETE /api/artifacts` empties the tree, its history and the records (admins) |
| `GET` `POST` | `/api/runs/{id}/summary` | a finished run's summary: `GET` the kept one, `404` without; `POST` returns the kept one or has Claude write it from the end of the output and keeps it; `summary` as markdown, `html`, `kept` (Claude Code integration) |
| `GET` `POST` | `/api/workflow-runs/{id}/summary` | a finished workflow run's summary, as for a run: Claude reads each step's job, state and the end of its output; `404` for a run with a step you may not see |
| `GET` `DELETE` | `/api/artifacts/files/{path}?version=` | a file of the tree: `kind` (`image`, `markdown`, `text` or `file`), markdown as `html`, text as `text`, and its `versions`; `DELETE` takes a file or a directory out of the tree, the history keeps it (admins) |
| `GET` `POST` `PUT` `DELETE` | `/api/schedules`, `/api/schedules/{id}` | `{folder, job, expression, enabled}` |
| `GET` `POST` `PUT` `DELETE` | `/api/workflows`, `/api/workflows/{id}` | `{name, steps: [{folder, job}]}`; owner or admin changes |
| `GET` `POST` | `/api/workflows/{id}/runs` | runs, newest first, with their steps; `POST` starts one, 409 while one runs |
| `POST` | `/api/workflow-runs/{id}/stop` | stops the step running, skips the rest |
| `GET` `POST` | `/api/notes?mention=&state=&limit=` | the threads, newest first, with their replies; `POST {parent_id, body}` leaves a note or a reply, and one naming the agent starts it |
| `POST` | `/api/notes/{id}/comments` | `{body, state}`: a reply, and the instruction's state: working, done or failed |
| `DELETE` | `/api/notes/{id}` | a note and its replies; its author or an admin |
| `GET` `PUT` `DELETE` | `/api/integrations`, `/api/integrations/{id}` | the integrations, `claude` and `beszel`: `{values}` by setting name; secrets come back only as `set`, and an empty one keeps its value (admins) |
| `GET` | `/api/inventory` | the hosts of the Beszel integration, largest first; 404 without it |
| `POST` | `/api/inventory/run` | `{command, hosts, summarize}`: runs the command on the hosts over their ssh logins at once, every host with a login when `hosts` is empty; each host's `output`, `exit_code` and `error` (admins). With `summarize`, also the markdown `report` of every output and Claude's `summary` of it, or `summary_error` |
| `GET` `POST` `PUT` `DELETE` | `/api/projects`, `/api/projects/{folder}` | clone `{url, id, branch, workspace}`, into `<workspace>/<id>` when `workspace` names one, at the root for `_` or empty; `409` at the root for a workspace's name or a directory there; configure `{title, description, allow, sources, commands, job_allow, hours, job_hours, workspace}`; a `workspace` other than the project's moves it, answering the row with `left`, and `409` for a workspace, a missing target, a taken ID or a run going |
| `GET` `POST` `PUT` `DELETE` | `/api/workspaces`, `/api/workspaces/{folder}` | clone `{url, folder, title, description, branch}` into `<root>/<folder>`, with its key directories and no encryption key; `409` for a folder a workspace, a project or a directory holds; `PUT {title, description, sort}`, the lowest sort first, left out it stays; `DELETE` removes the checkout, the projects inside it and their rows, their managed env files, its secrets and its keys; the default workspace `_`, the root, lists first, keeps its sort, and `DELETE` answers `409` (admins) |
| `POST` | `/api/workspaces/{folder}/prompt` `{prompt, cols, rows}` | starts `internal/claude-prompt` as you, with `TASK_UI_PROMPT` and the workspace in `TASK_UI_WORKSPACE_DIR`, and answers `201` with the run; `400` for an empty prompt, `409` while a prompt runs (admins) |
| `POST` | `/api/workspaces/{folder}/export` `{passphrase}`, `/api/workspaces/import` (multipart: `bundle`, `passphrase`, `folder`) | export answers the workspace as `<folder>.taskui-ws`, encrypted to the passphrase: its row, projects, env files, secrets as ciphertext, key files and grants; import adds it under `folder`, the bundle's own when left out and required for `_`, clones its projects and answers `201 {workspace, projects: [{folder, cloned, output}], unknown_users}`; `400` for a wrong passphrase or a bundle it refuses, `409` for a folder held and for bundled keys held here for the other use or an encryption key another workspace holds, at most 64 MiB (admins) |
| `GET` `POST` `PUT` `DELETE` | `/api/workspaces/{folder}/keys`, `/keys/generate`, `/keys/upload`, `/keys/{name}` | a workspace's keys, public halves and fingerprints only: `encrypt`, `keys` with their backups, `policies` and `users` by key name; generate `{type, name, comment, policies, users}`, upload `{content, name, comment, policies, users}` as `id_<type>_<name>`; a key without a grant goes to every run in the workspace; `PUT` sets `{policies, users}`, `DELETE` removes a key and its grants; `encrypt: true` on generate or upload adds the ed25519 encryption key a workspace starts without, or rotates or replaces it, which no run gets and no route deletes, and writes the workspace's secrets again, `rekeyed` their count, `rekey_error` when that fails and the backup still opens them; upload answers `409` for key material held for the other use by the instance or any workspace, and for an encryption key another workspace or the instance holds (admins) |
| `GET` `POST` `PUT` | `/api/users`, `/api/users/{id}` | `POST {name, password, folders}` adds a user to the password file (admins; `409` for a name it holds); `PUT {folders}`; roles come from the role lists, and removing a user is removing its line |
| `GET` `POST` `DELETE` | `/api/secrets`, `/api/secrets/{id}` | `{scope, name, content}`; `scope` is `workspace:<name>`, shared within the workspace, or a folder; `all` or empty is `workspace:_`; `400` for an unknown workspace or folder; content is never returned; `409` for a scope whose workspace has no encryption key yet |
| `POST` | `/api/secrets/rekey` | writes every secret again for its scope's current key, a workspace's or, for the integrations, the instance's encryption key |
| `GET` `POST` `PUT` `DELETE` | `/api/keys`, `/api/keys/generate`, `/api/keys/upload`, `/api/keys/{name}` | the SSH keys of `TASK_UI_SSH_DIR`, none that encrypts: `{type, name, comment, policies, users}`, `{content, name, comment, policies, users}`; `name` makes `id_<type>_<name>` for someone; `PUT` sets a key's `{policies, users}`, the roles and users whose runs get it; no secret changes with them; `GET` adds `encrypt`, the instance's encryption key, public half and fingerprint, and `overlaps`, `[{fingerprint, encrypt, access}]`, the keys held for both uses by their paths; upload answers `409` for an encryption key, private or public half (admins) |
| `POST` | `/api/keys/encrypt/rotate` | generates a new instance encryption key, keeps the old one as a backup and writes the integrations' settings again, `201 {key, rekeyed}`, `rekey_error` when that fails and the backup still opens them (admins) |
| `GET` `PUT` | `/api/env/{folder}`, `/api/env/{folder}/{path}` | the env files; `{content}` saves and generates |
| `GET` | `/api/files/{folder}/{path}` | a file of a folder's checkout, for whoever sees the folder: `kind` (`image`, `markdown`, `text` or `file`), markdown as `html`, text as `text` and coloured as `html`, `truncated` past 1 MiB; `404` for a directory, a path with `..`, a symlink out of the checkout, `.git` or a path through it, and any dot file, such as `.env` or `.claude/.secret`, or a symlink to one; a dot directory such as `.claude/` opens. `/files/{folder}/{path}` is the file itself, `?download` to save it |

## MCP

`/api/mcp` is an MCP server over the streamable HTTP transport: nine
tools to list jobs and workflows, run them, read the runs, and read and
answer the notes, as the token's user. Its use is in [the MCP chapter](guide/mcp.md).

---

[Contents](README.md) · [Running the server](running.md) · [Development](development.md)
