# The convention rules

Semgrep rules that hold the conventions prose states, so a change that breaks one fails rather than being noticed in review a year later.

`atkins lint` runs them blocking. Nothing here is a backlog: every rule was written at zero and every rule is proven to fire, so a finding is a new violation and not a count to work down.

```sh
atkins lint                                                  # blocking, with the rest of the lint
semgrep scan --metrics=off --quiet --error --config rules .   # the same pass by hand
```

To count rather than block, which is what a rule written over an existing backlog needs:

```sh
mkdir -p cover
semgrep scan --metrics=off --quiet --config rules --json -o cover/semgrep.json .
jq -r '.results[].check_id' cover/semgrep.json | sort | uniq -c | sort -rn
```

## sql.yml, at zero

| Rule                             | What it refuses                                             |
|----------------------------------|-------------------------------------------------------------|
| `go-pdo-outside-storage`         | `github.com/titpetric/pdo` imported anywhere but `storage/` |
| `go-sql-outside-storage`         | `ExecContext`, `QueryContext`, `SelectContext` and the rest |
| `go-sql-literal-outside-storage` | A statement in a string literal outside `storage/`          |
| `php-sql-outside-store`          | `$this->db->get_all(...)` outside a module's `*Store.php`   |
| `php-sql-literal-outside-store`  | A statement in a literal outside a module's `*Store.php`    |

Both literal rules match `SELECT` as a bare keyword and carry `CREATE TABLE`, and both are case sensitive. The bare keyword is what catches `"SELECT " . COLUMNS . " FROM run"`, where the `FROM` is in a different literal and the narrower form read straight past it; the case policy is what keeps "Delete from the tree" a menu label and not a finding.

## phpfront.yml, at zero

The [front end](../docs/phpfront.md) being ported. Its conventions are the module's with two places for a statement rather than one, and four rules that have nothing to do with SQL.

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

Why SQL in one place, and not just tidiness: `storage` is the only package this repository keeps integration tests for, so a query written anywhere else is a query no test covers. The storage tests are at 99.4% of statements over a throwaway database; a `db.QueryContext` in a handler has nothing behind it.

The same holds for a module, which is PHP: `docs/modules.md` says every statement a module runs lives in its one `<Module>Store` class, and `RefinementStore.php` repeats it in its header.

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

Adopted from mobius's `rules/arrays.yml`. One rule, `comment-in-array`, and it guards a formatter defect rather than a taste: `phpscript fmt` relocates a comment written anywhere inside an expression to after the statement that holds it, stably, so a second pass changes nothing and the author who moved it sees no diff. [`docs/todo-correction.md`](../docs/todo-correction.md) section 23 has the probe.

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

### Writing one

The protocol is the shared skill's, `/dev:semgrep`: write the bad shape and its legal neighbour into a probe tree, prove the rule fires on the first and not the second, then scan this tree. Two neighbours this file has already paid for:

- `model/types.mig.go` is mig's generated query builder and holds every SQL keyword. It is excluded by `*.mig.go`, and it says DO NOT EDIT.
- `$tpl->get()` renders a MiniTPL template and `$this->db->get()` reads a row. The PHP rule names the receiver for that reason; a rule on the method alone flagged two template renders in `RefinementPage.php`.
