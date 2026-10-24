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
starts one; `#menu` opens the drawer; `#actions=<n>` opens the **⋯**
menu of row n.

Under 860px the sidebar is a drawer behind the menu button, with the
folders opening into their jobs, and the tables of jobs and runs fold
into cards.

## Row actions

A row of a table shows its main action as a button and the rest
behind its **⋯** button, the same menu on every page: users, keys,
secrets, projects, workspaces, schedules, workflows, runs, notes and
prompts, artifacts, a folder's Jobs, Files and env files, and the inventory's
hosts.

- **Visible**: the row's main purpose, at most one or two buttons:
  **Run** on a job or a workflow, **View** or **Attach** on a run, the
  copy button of a key's public key.
  Everything else goes in the menu.
- **Items**: an icon and a text label each, never an icon alone. A link
  navigates; a button does what its data attributes say, as a loose
  button does: `data-delete`, `data-api` with `data-method` and
  `data-body`, `data-copy`, `data-download`, `data-confirm`,
  `data-confirm-ok` and `data-next`.
- **Order**: the primary items first (open, edit, view), the secondary
  ones (copy, download, move) after a line, the danger group last, in
  the bad colour: Delete, Remove, Kill. The menu puts every danger item
  in the last group, wherever the template lists it.
- **Details**: the top of the menu holds what the row's columns leave
  out, as label and value: a key's fingerprint and dates, a run's ID and
  start, the version of an artifact.
- **Accessibility**: the button has `aria-haspopup="menu"`,
  `aria-expanded` and the label "Actions on <name>"; the items have
  `role="menuitem"`. A click opens the menu, never a hover; Enter, Space
  or the down arrow open it on the first item, the arrows, Home and End
  move, Escape closes it and returns the focus to the button. A click
  outside, a scroll or a pick close it, and one menu is open at a time.
  On a phone, under 640px, it opens as a sheet at the bottom.

A template draws one with the `menu` template and the functions of the
same names:

```
{{ template "menu" (menu (print "Actions on " .Name)
	(detail "Fingerprint" .Fingerprint "mono" "")
	(item "edit" "Edit links" "href" (print "/settings/keys/" .Name))
	(item "copy" "Copy public key" "data-copy" .PublicKey)
	sep
	(and ($.Can "key/delete") (danger "trash" "Remove" "data-delete" (print "/api/keys/" .Name) "data-confirm" "Remove it?" "data-next" "reload"))) }}
```

`item` and `danger` take an icon of the sprite, a label and pairs of an
attribute and its value: `href`, `download`, `note`, `disabled` (its value
the note) and any `data-*`, `aria-*` or `title`, the values escaped; an
unknown attribute fails the render. `detail` takes a label, a value and
`title` or `mono`. `sep` ends a group. `(and $cond (item ...))` leaves an
item out while the condition is empty, and a menu with nothing in it
draws no button. `menuHTML` in `task-ui.js` draws the same for the rows
a script adds. The CSS is in `css/menus.css`.

A keys table has the same columns on every page, the Keys page, the
Instance page, a workspace's page, Your SSH keys and a profile: Name,
Algorithm, Size in bits, Public key shortened with a copy button for
the whole line, Purpose, and the **⋯** menu with the fingerprint, the
dates and the links in its details.

## Lists and actions

[Sitemap](sitemap.md) lists every page with its lists, its head
actions, its row actions, its forms and its empty lines. The pages
follow one set of rules:

- **Create actions** sit in the page head, top right, for the page's
  main list: **Add project**, **Add workspace**, **Add user**, **Add
  secret**, **Add workflow**, each with the plus icon, and **Upload
  key** and **Generate key** for keys. One button per head is primary.
  A second list on a page holds its own in its block head, small and
  plain: a workspace's keys, a profile's keys, a workspace's secrets.
- **Page-level danger**, deleting what the page shows, goes in a **⋯**
  menu at the end of the page head, never a loose red button: a
  workspace, a project's checkout, a workflow, a user, every artifact.
- **Words**: "Add <thing>" creates. "Delete" ends a record or data for
  good: a user, a secret, a schedule, a workflow, a note, an artifact, a
  key's backup, a workspace with its keys and secrets. "Remove" lets go
  of what can come back: an SSH key, whose files stay in the key folder;
  "Remove checkout", which the remote can clone again; a step of a
  workflow. The confirm button repeats the item's label. A link to a
  settings page reads "<Kind> settings", Project settings, Workspace
  settings, and the page's title repeats it, with the folder as a
  badge.
- **Empty lists**: a list with no rows draws no table. Its block head
  holds one line, "No keys added.", between the heading and the buttons
  that add to it, from `blockhead`:

  ```
  {{ template "blockhead" (head "Keys" "SSH keys the workspace's runs get" (and (not .Keys) "No keys added.")
  	(item "upload" "Upload key" "href" "/settings/workspaces/acme/keys/upload")
  	(item "key" "Generate key" "href" "/settings/workspaces/acme/keys/generate")) }}
  {{ with .Keys }}...the table...{{ end }}
  ```

  `head` takes the title, a note, the empty line or a false value, and
  buttons made with `item`, `primary` or `danger`, a false value left
  out. A list a filter empties, runs or artifacts, keeps its table and
  an empty row.
- **Copy**: a value to copy is the `copyable` template, `(list value
  label [shown])`: the value in code and the copy button right after it,
  for a public key, an ssh login, a link or a command. The same value
  may also have Copy in the row's menu.
- **One template per entity**: keys list through `keyhead` and
  `keyrow`, projects through `projecttable`, a user through `user`, a
  row's menu through `menu`. A page that lists them again uses those.
- **Forms**: a label above each field, its help in faint text beside
  the label, Cancel and the primary submit at the bottom right in
  `form-foot`. A page of more than one form or list puts each in a block
  with a heading: a workspace's Settings, Projects and Keys.

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
| `layout`    | where a folder ID lives on disk and back: `<ws>/home` and `<ws>/src/<project>` under the root, `home` and `src/<project>` for `_`; the only place that maps IDs to directories, and the layout migration |
| `folder`    | the catalogue: a `Folder` with its `Job`s, discovery under the root, the project configuration |
| `runner`    | sessions: a `Process` on a pty, the terminal `Frame`, subscribers, artifacts |
| `scheduler` | cron `Expression`s and the `Loop` that starts due schedules                  |
| `hours`     | the office hours `Rules` a folder or job may run in                          |
| `recording` | the asciicast files                                                          |
| `access`    | who may do what: the password file, `Sessions`, tokens, the `Guard`          |
| `secret`    | the secrets, encrypted at rest with the instance's encryption key through its `Vault`, their scopes and their read log |
| `keystore`  | every key of the instance in one flat folder, what each is for and whose runs get it |
| `envfile`   | the env files users fill for a folder, written in it before each run          |
| `sshkey`, `git`, `dotenv` | the key pairs, the checkouts, the dotenv lines it writes      |
| `permission` | what a user may do, from `config/permissions.yml` and the policies of the user and the role |
| `config`    | the embedded configuration: `permissions.yml` and `context.md`               |
| `note`      | the notes board: messages, replies and the notes that hand someone an instruction |
| `workflow`  | workflows: named chains of jobs, started one after another as ordinary runs  |
| `workspace` | what a workspace holds beside its checkout: its keys, and its export bundle  |
| `agent`     | the host commands of the Claude built-ins, `task-ui agent note\|prompt\|shell` |
| `claude`    | Claude Code for task-ui's own uses, such as the summary of a run's output   |
| `integration` | the settings of the services task-ui talks to, encrypted with the instance key |
| `beszel`    | the client of a Beszel hub, for the host inventory                           |
| `artifactstore` | the files runs leave, as one tree with its history in git                    |
| `retention` | the Data retention rules and `task-ui cleanup`, which applies them           |
| `channel`   | `task-ui channel`, the Claude Code channel of the notes that name a user     |
| `event`     | the in-memory bus the pages follow, by topic                                 |
| `sshagent`  | an ssh agent on a socket of its own, serving keys from memory to a run       |
| `sshprobe`  | who task-ui signs in to a host as over ssh                                   |
| `hostrun`   | the runs of `@host:<host>`, a command on a host of the inventory             |
| `hostinfo`  | a host's filesystems and its containers' sizes over its ssh login, for the expanded row of the inventory |
| `hostpath`  | the paths of the container task-ui runs in as its docker host sees them, for a run's `docker run -v` |
| `skills`    | asks atkins what runs in a directory: its pipeline and the skills that apply |
| `tooling`   | asks atkins what the tree has been running, every command with its time, the `Report` behind the panel on the runs page |
| `markdown`  | markdown rendered as HTML                                                    |
| `highlight` | the coloured HTML of the files a pipeline reads                              |
| `diskusage` | how much space a directory takes, files and git history apart                |
| `argv`      | the arguments a user types for a job, split as a shell splits words          |
| `internal/rulelint` | matches the permission checks of the code against `config/permissions.yml`   |
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
parallel, and share nothing. `go test ./...` runs the unit tests alone;
the tests that run processes, open a database or use the network sit
behind the `integration` build tag, see [Testing](testing.md).
`go run ./internal/cmd/rulelint` (`task
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

---

[Contents](README.md) · [Running the server](running.md) · [API](api.md)
