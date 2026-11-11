# The front end's convention rules

Semgrep rules that hold what the PHP of this tree may not do. The Go rules stay in [`rules/`](../../../rules/SKILL.md) at the repository root; these moved here because every one of them is about phpscript, and their path globs are written against this directory.

`atkins ruleslint` runs them blocking: `.ruleslint.conf` names the command in its `linters` block, after the structure checks. Nothing here is a backlog: every rule was written at zero and every rule is proven to fire, so a finding is a new violation and not a count to work down.

```sh
atkins ruleslint                                                                                   # blocking, after the structure checks
git ls-files -co --exclude-standard | xargs semgrep scan --metrics=off --quiet --error --config semgrep   # the same pass by hand
```

Given a directory, semgrep walks the files git tracks and says so in its own summary, so a module written this hour is scanned by nothing. The file list is git's own: what it tracks and what it does not, minus what it ignores.

To count rather than block, which is what a rule written over an existing backlog needs:

```sh
mkdir -p ../cover
semgrep scan --metrics=off --quiet --config semgrep --json -o ../cover/semgrep.json .
jq -r '.results[].check_id' ../cover/semgrep.json | sort | uniq -c | sort -rn
```

## phpfront.yml, at zero

The [front end](../../../docs/phpfront.md) being ported. Its conventions are the module's with two places for a statement rather than one, and four rules that have nothing to do with SQL.

| Rule                                | What it refuses                                                    |
|-------------------------------------|--------------------------------------------------------------------|
| `front.sql-outside-storage`         | A statement outside `code/Task/` or a module's `storage/`          |
| `front.sql-literal-outside-storage` | The same, as a text match                                          |
| `front.prefix-in-php`               | A mount in PHP or a template; `app.FrontBase` and `app.GoBase`     |
| `front.route-names-a-file`          | An `@route` whose path ends in `.css`, `.js` and the rest          |
| `front.route-in-the-theme`          | An `@route` under `theme/`; the theme draws and answers nothing    |
| `front.model-edited`                | A hand edit to `code/Task/Model`, which `atkins models` writes     |
| `front.allocation-in-expression`    | `new X` inside an argument, a ternary or an array value            |

The last three are not tidiness either. A routed static file is a page the native router should have answered, with none of its caching. A route in the theme is a view in the wrong place. A hand-edited model drifts from the Go types the same `schema.yml` generates, and nothing says so until a column reads empty.

A module is PHP: `docs/modules.md` says every statement a module runs lives in its one `<Module>Store` class, and `RefinementStore.php` repeats it in its header. The Go half of the same convention is `rules/sql.yml`, where `storage` is the only package the repository keeps integration tests for.

## transactions.yml, at zero

Adopted from mobius's `rules/transactions.yml`, a sibling repository on the same runtime, which found this first.

| Rule                           | What it refuses                                              |
|--------------------------------|--------------------------------------------------------------|
| `begin-without-defer-rollback` | `begin()` with no `defer($db->rollback);` on the next line   |
| `defer-closure`                | `defer(function () { ... })`, whose body captures nothing    |

A fault between `begin()` and `commit()` is not an `Exception` — a PHP `Error` and a fault the interpreter raises are neither — so `catch` does not cover it. The frame returns either way and a deferred callback runs on the way out. `commit()` makes the deferred rollback a no-op, so the line is free.

It cost one finding on adoption, `code/Task/Sidebar.php`, where the only transaction in the front end had no rollback on any path: the pooled connection kept a write transaction open until the render context timed out, which on SQLite is a lock every other writer waits on.

A comment between the two statements is legal — the pattern matches the AST, not the lines — which is where the reason for the `defer` belongs.

## arrays.yml, at zero

Adopted from mobius's `rules/arrays.yml`. One rule, `comment-in-array`, and it guards a formatter defect rather than a taste: `phpscript fmt` relocates a comment written anywhere inside an expression to after the statement that holds it, stably, so a second pass changes nothing and the author who moved it sees no diff. [`docs/todo-correction.md`](../../../docs/todo-correction.md) section 23 has the probe.

Generic, because semgrep's PHP patterns cannot see comments. A comment above a statement that merely contains an inline `array(...)` stays legal, which is where every comment in this tree already sits.

## suppression.yml, at zero

Adopted from mobius's `rules/suppression.yml`. One rule, `error-suppression`: no `@`. It hides the failure the author thought of and every other one with it, and this runtime ignores it anyway. Ask the question instead — `file_exists()` before `filemtime()`, `isset()` before a read.

mobius's rule carries no paths block. Ours excludes `ext/` and `tests/`, because the four findings it had here were all false and all outside our own code: two copies of composer's vendored `InstalledVersions.php`, and two `.phpt` expected-output blocks where a rendered `@mention` reads as code to a PHP parser. Neither exclude may be widened past those directories — a suppressed call in a page is the thing this refuses.

## phpscript.yml, at zero

Ours, not mobius's, and both rules retire the day the runtime is fixed. Each message names the `docs/todo-correction.md` section that holds its probe.

| Rule                     | What it refuses                                                  |
|--------------------------|------------------------------------------------------------------|
| `whole-float-literal`    | `100.0`, which `fmt` rewrites to `100` and so retypes            |
| `self-in-const-array`    | `const X = array(self::Y)`, which resolves against the reader    |

Both are the kind of defect a linter cannot see: `fmt` and `lint` are clean on either, and the page draws wrong or 500s. The float rule matches only a literal in a value position — after an operator, a paren, a comma, a `=>` or a `return` — because a generic match reads inside a string too, and `Clock::stamp()` carries the Go time layout `2006-01-02 15:04:05.000000`, whose fractional seconds are not a float.

## Writing one

The protocol is `rules/SKILL.md`, Writing one, and the shared skill it names, `/dev:semgrep`. A path glob here is written against this directory, not the repository root: `**/*.php` is every PHP file of the app, and `code/Task/**` is the one package that owns a statement.

`$tpl->get()` renders a MiniTPL template and `$this->db->get()` reads a row, which is why the PHP rule names the receiver; a rule on the method alone flagged two template renders in `RefinementPage.php`.
