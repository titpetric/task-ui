# The convention rules

Semgrep rules that hold the conventions prose states, so a change that breaks one fails rather than being noticed in review a year later.

These are the Go rules. The PHP rules of the front end live beside the tree they read, in [`phpfront/app/semgrep/`](../phpfront/app/semgrep/SKILL.md), and `atkins ruleslint` runs them from there.

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

The PHP half of the same convention is `front.sql-outside-storage` and `front.sql-literal-outside-storage`, under `phpfront/app/semgrep/`.

The literal rules match `SELECT` as a bare keyword and carry `CREATE TABLE`, and both are case sensitive. The bare keyword is what catches `"SELECT " . COLUMNS . " FROM run"`, where the `FROM` is in a different literal and the narrower form read straight past it; the case policy is what keeps "Delete from the tree" a menu label and not a finding.

Why SQL in one place, and not just tidiness: `storage` is the only package this repository keeps integration tests for, so a query written anywhere else is a query no test covers. The storage tests are at 99.4% of statements over a throwaway database; a `db.QueryContext` in a handler has nothing behind it.

## Writing one

The protocol is the shared skill's, `/dev:semgrep`: write the bad shape and its legal neighbour into a probe tree, prove the rule fires on the first and not the second, then scan this tree. One neighbour this file has already paid for: `model/types.mig.go` is mig's generated query builder and holds every SQL keyword, excluded by `*.mig.go`, and it says DO NOT EDIT.
