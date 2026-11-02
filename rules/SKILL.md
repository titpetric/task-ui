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

## phpfront.yml, at zero

The [front end](../docs/phpfront.md) being ported. Its conventions are the module's with two places for a statement rather than one, and four rules that have nothing to do with SQL.

| Rule                                | What it refuses                                                    |
|-------------------------------------|--------------------------------------------------------------------|
| `front.sql-outside-storage`         | A statement outside `code/Task/` or a module's `storage/`          |
| `front.sql-literal-outside-storage` | The same, as a text match                                          |
| `front.prefix-in-php`               | `/development` in PHP or a template; the prefix is `app.FrontBase` |
| `front.route-names-a-file`          | An `@route` whose path ends in `.css`, `.js` and the rest          |
| `front.route-in-the-theme`          | An `@route` under `theme/`; the theme draws and answers nothing    |
| `front.model-edited`                | A hand edit to `code/Task/Model`, which `atkins models` writes     |

The last three are not tidiness either. A routed static file is a page the native router should have answered, with none of its caching. A route in the theme is a view in the wrong place. A hand-edited model drifts from the Go types the same `schema.yml` generates, and nothing says so until a column reads empty.

Why SQL in one place, and not just tidiness: `storage` is the only package this repository keeps integration tests for, so a query written anywhere else is a query no test covers. The storage tests are at 99.4% of statements over a throwaway database; a `db.QueryContext` in a handler has nothing behind it.

The same holds for a module, which is PHP: `docs/modules.md` says every statement a module runs lives in its one `<Module>Store` class, and `RefinementStore.php` repeats it in its header.

### Writing one

The protocol is the shared skill's, `/dev:semgrep`: write the bad shape and its legal neighbour into a probe tree, prove the rule fires on the first and not the second, then scan this tree. Two neighbours this file has already paid for:

- `model/types.mig.go` is mig's generated query builder and holds every SQL keyword. It is excluded by `*.mig.go`, and it says DO NOT EDIT.
- `$tpl->get()` renders a MiniTPL template and `$this->db->get()` reads a row. The PHP rule names the receiver for that reason; a rule on the method alone flagged two template renders in `RefinementPage.php`.
