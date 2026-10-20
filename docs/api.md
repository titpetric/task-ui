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
| `/api/artifacts/files/{path}` | The preview window: a file of the tree, rendered, with its versions, and the delete of one file, which the records cannot reach |
| `/api/notes`, `list_notes`, `note_comment` | Notes, and instructions to the agent, which reads and reports on them over MCP |
| `/api/integrations`, `/api/inventory` | Settings, Integrations, which keeps the services' tokens encrypted, the inventory from Beszel, and commands run on its hosts |
| `summarize` on `POST /api/inventory/run` | **Run and summarize** on the Inventory page: Claude reads every host's output as one report, after the run, in the same request |

| Method | Path | Does |
|--------|------|------|
| `GET` | `/api/folders?folder=` | the folders you can see and their jobs, with `can_run` and `last_run` |
| `GET` | `/api/runs?folder=&job=&state=&limit=` | the runs, newest first, with `live`, `attached` and `can_run` |
| `POST` | `/api/runs` `{folder, job, force, reason, cols, rows}` | starts a run; `409` with the run in the way unless forced; `403` out of hours unless forced with a reason |
| `GET` | `/api/runs/{id}` | one run |
| `POST` | `/api/runs/{id}/stop`, `/kill` | ends a run |
| `GET` | `/api/runs/{id}/frame.svg`, `/frame.txt`, `/cast` | the last frame, the recording |
| `WS` | `/api/runs/{id}/attach` | binary frames are output; text frames are `{"type":"input"\|"resize"\|"stop"\|"kill"}` in and `{"type":"status"}` out |
| `GET` `DELETE` | `/api/artifacts`, `/api/artifacts/{id}` | list, download a version, delete a record; `DELETE /api/artifacts` empties the tree, its history and the records (admins) |
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
| `GET` `POST` `PUT` `DELETE` | `/api/projects`, `/api/projects/{folder}` | clone `{url, id, branch}`; configure `{title, description, allow, sources, commands, job_allow, hours, job_hours}` |
| `GET` `PUT` | `/api/users`, `/api/users/{id}` | `{folders}`; names, passwords and roles come from the password file and the role lists |
| `GET` `POST` `DELETE` | `/api/secrets`, `/api/secrets/{id}` | `{scope, name, content}`; content is never returned; `409` without a key to encrypt with |
| `POST` | `/api/secrets/rekey` | writes every secret again for the current keys |
| `GET` `POST` `PUT` `DELETE` | `/api/keys`, `/api/keys/generate`, `/api/keys/upload`, `/api/keys/{name}` | `{type, name, comment, policies, users}`, `{content, name, comment, policies, users}`; `name` makes `id_<type>_<name>` for someone; `PUT` sets a key's `{policies, users}`, the roles and users whose runs get it; a new key rekeys the secrets, the last key that opens them cannot go |
| `GET` `PUT` | `/api/env/{folder}`, `/api/env/{folder}/{path}` | the env files; `{content}` saves and generates |

## MCP

`/api/mcp` is an MCP server over the streamable HTTP transport: nine
tools to list jobs and workflows, run them, read the runs, and read and
answer the notes, as the token's user. Its use is in [the MCP chapter](guide/mcp.md).

---

[Contents](README.md) · [Running the server](running.md) · [Development](development.md)
