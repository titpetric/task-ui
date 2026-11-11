# TODO

What is open as of 2026-11-11, written down because the packages are about to move and every path below will need correcting against the new layout.

## Asked for and not started

### Threads index as a messenger view

`/threads` is a list. It should be a 20/80 split: the previous threads down the left, the selected thread as a chat on the right, switching between them without leaving the page. No thread selected is a new thread.

The button under the box reads `Post note` and should read `Send`. Review every other submit button against the same question: a box that sends a message takes `Send` and nothing more specific.

Files: `modules/threads/templates/index.tpl`, `modules/threads/mod_threads.php`, `modules/threads/assets/threads.css`.

### A module is enabled or disabled

Nearly every module under `phpfront/app/modules/` ships no `modinfo.php`. There is no page that lists them and no way to turn one off.

| What                     | Where                                           |
|--------------------------|-------------------------------------------------|
| `modinfo.php` per module | `phpfront/app/modules/<name>/`                  |
| `/settings/modules`      | a new `modules/settings/mod_modules.php`        |
| the enabled flag         | the `module` table, which already holds the row |
| MCP respects it          | `mcp/` module list                              |

The MCP list keeps every module in its answer and marks a disabled one unavailable, so a caller sees the name and the state rather than a gap.

## Needs a decision or a permission

### The stale plugin tree on prod

`/data/modules/titpetric/refinement` is what an older `phpscripts.Extract` wrote. It fails to register on every start:

```text
WARN a module did not register err="/data/modules/titpetric/refinement/modinfo.php: migrate.php: call to undefined method Task::migrate()"
```

`/m/titpetric/refinement` answers 404 and `/refinement` answers 200, so nothing is broken by it. Moving or deleting that directory is a write to the live container and was refused by the sandbox classifier; it needs naming.

### Deploy 112d721c

`fix(phpfront): a file the build drops leaves the extracted tree` is pushed and not deployed. Until it is, `/data/phpfront/modules/notes/` keeps registering `GET /notes` and the router refuses the threads module's route as a duplicate:

```text
WARN duplicate route "GET /notes" in modules/threads/mod_notes.php; previously registered by modules/notes/mod_notes.php
```

### gopkg.in/yaml.v3

splint's `blacklist` linter reports four files. `app/telemetry.go` is a drop-in swap for `github.com/goccy/go-yaml`; the other three use the `yaml.Node` AST or the strict decoder, which goccy spells differently.

| File                    | What it uses                              |
|-------------------------|-------------------------------------------|
| `app/telemetry.go`      | `Marshal`, `Unmarshal`                    |
| `form/form.go`          | `NewDecoder` with `KnownFields`           |
| `config/permissions.go` | `yaml.Node`, `MappingNode`, `ScalarNode`  |
| `detector/pipeline.go`  | `yaml.Node`, `ScalarNode`, `SequenceNode` |

`go:lint` is `splint ./... || true`, so none of this gates a commit.

## Carried from the port

### frontend/ cannot be deleted yet

| Fact                               | Count |
|------------------------------------|-------|
| Go files under `frontend/`         | 43    |
| files under `frontend/view/`       | 159   |
| `html/template` templates          | 69    |
| GET pages under `/legacy`          | 50    |
| root handlers belonging to neither | 8     |

The 69 templates are `html/template` and not MiniTPL, so they cannot move to `app/theme/templates`. The eight handlers are `POST /login`, `POST /logout`, `/healthz`, `/static/*`, `/artifacts/files/*`, `/files/*`, `/artifacts.tgz` and `/artifacts/*`, plus the `/m` plugin mount. They move to a package of their own first, then the 50 pages go.

`frontend/view/templates/_user.tpl` still draws the user menu that `theme/templates/components/user.tpl` dropped. Leave it: the page it serves is going.

### The .phpt fixtures

`app/tests/` holds fixtures for modules and for the theme together. Module fixtures go to `modules/<name>/tests/`, the rest to `theme/tests/`, and `app/tests/` goes. `bootstrap.php`, `setup.php` and `schema.sql` need a home and every `phpscript.yml` include path needs updating.

## Tooling

### plantuml writes the JVM version into every diagram

Every `schema.svg` carries a footer naming the local Java build, so a JDK patch rewrites all 14 with no change to the diagram:

```diff
-Java Version: 25.0.4.1+1-1-26.04.4-Ubuntu
+Java Version: 25.0.5.0.3+1-1-26.04.1-Ubuntu
```

Add `-nometadata` to the plantuml step of `phpfront/app/atkins.yml`.

### mdox

Every pipeline in the workspace is on `splint fix --markdown`; the shared `~/.atkins/skills/mdox.yml` installs splint and calls it. The `mdox` binary is still at `/usr/local/bin/mdox` and nothing reinstalls it. Delete it when convenient.

`modules/refinement/action/refinement.md` is the one file the formatter leaves alone, reporting that formatting would change the content of a code span.

### splint findings now block the edit hook

The prose and godoc linters that landed with `feat(splint): the prose of a tree is linted and its markdown formatted` report across whole packages, and the `PostToolUse` hook refuses an edit to a file in one. Known: `phpscript` (`"so a"` at 188x the baseline), `model`, `importfmt`. Each needs a comment sweep of the kind `chore(splint): the comments state behaviour directly` was.

### A flaky test

`TestSessions_ResolveCeiling` in `access` failed once under the sharded integration runner and passes alone and on a re-run. It reads a clock.

## Reported, not fixed

From the week-in-review of 2026-11-10:

- the upstream phpscript `lint --flatstack` constructor blind spot
- the four phpscript workarounds the runtime has since made unnecessary
- the godoc gate and the column-width cap
- `orbs.js` doing its work at load time
- `pending` never answered
- a dropped hub host returning on the next page load
- 32 findings marked plausible and never verified
