# Development

## Screenshots

Every view, in both themes and on a phone, is in
[docs/screenshots.md](screenshots.md); the pictures are under
`docs/assets/`.

`scripts/demo.sh` regenerates them: it serves a copy of `testdata/demo`,
sets the instance up through the API with curl, starts real runs with
`scripts/run.sh`, walks every link of every page with `scripts/crawl.sh`
and captures each view with headless chromium through
`scripts/chromium.sh`, the phone views included. A run opens in its
window from a link: `/runs/<id>#run=<id>`; `#start=<folder>/<job>`
starts one; `#menu` opens the drawer.

Under 860px the sidebar is a drawer behind the menu button, with the
folders opening into their jobs, and the tables of jobs and runs fold
into cards.

## Layout

Packages are the bounded contexts, one per domain, singular. The
dependency runs one way: `app` to everything, `api` and `frontend` to the
domains, the domains to `model` and `storage`, `model` to nothing of ours.

| Package     | Holds                                                                        |
|-------------|------------------------------------------------------------------------------|
| `schema`    | the SQL migrations, the source of truth of the data model (mig)              |
| `model`     | the mig generated types and the rules on them                                |
| `storage`   | CRUD per table over pdo, one in-memory database per test                     |
| `detector`  | the conventions: a `Detector` per file kind, returning a `Source` of `Target`s |
| `folder`    | the catalogue: a `Folder` with its `Job`s, discovery under the root, the project configuration |
| `runner`    | sessions: a `Process` on a pty, the terminal `Frame`, subscribers, artifacts |
| `scheduler` | cron `Expression`s and the `Loop` that starts due schedules                  |
| `hours`     | the office hours `Rules` a folder or job may run in                          |
| `recording` | the asciicast files                                                          |
| `access`    | who may do what: the password file, `Sessions`, tokens, the `Guard`          |
| `secret`    | the secrets, encrypted at rest per workspace, through each one's `Vault`     |
| `envfile`   | the dotenv files a folder reads, as task-ui manages and generates them       |
| `sshkey`, `git`, `dotenv` | the key pairs, the checkouts, the file format                 |
| `api`       | the JSON endpoints and the attach socket                                     |
| `mcp`       | the MCP server: JSON-RPC tools that call the API as the token's user         |
| `frontend`  | the pages; `frontend/view` renders them from embedded templates and assets   |
| `app`       | the composition root: configuration, wiring, serving                         |

A convention is a `Detector` in the `detector` package: one type that
reads a folder and returns a `Source`, added to the `Detectors` list. The
rest of task-ui runs on sources and jobs and never asks which kind they
are.

`splint ./...` (`task lint`) is the bar: every file has a test beside it,
every exported symbol a test that reaches it, every symbol a file named
for it, and every dependency one import per package. Tests are black box,
parallel, and share nothing. `go run ./internal/cmd/rulelint` (`task
lint` too) matches the permission checks against
`config/permissions.yml`: every rule a check names is in the file, and
every rule the file holds has a check. A check names its rule as a
string literal, `h.require(w, r, "secret/edit")` in Go and
`{{ if .Can "secret/edit" }}` in a template, so the lint finds it.
`config/context.md` tells an agent what task-ui is, area by area, for
the MCP instructions and every Claude run; keep it in step with the
features. Its test fails when a section or permission of
`config/permissions.yml`, an MCP tool or a built-in job has no line
there.
`scripts/smoke.sh` (`task smoke`) then
exercises a fresh instance end to end with curl and headless chromium:
runs, hours, secrets and keys, env files, schedules, users, projects,
artifacts, every page, and the browser's console.

---

[Contents](README.md) · [Running the server](running.md) · [API](api.md)
