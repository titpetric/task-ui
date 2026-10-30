# Testing

task-ui has two kinds of tests, and `go test ./...` runs only the first.

## Unit tests

A unit test stays inside the test process. It calls the code of its package, puts fakes behind the interfaces that code consumes, serves a handler with `httptest.NewRecorder` and `ServeHTTP`, and reads and writes files under `t.TempDir()`. Unit tests are the default run:

```bash
go test -cover ./...
```

The whole run takes seconds. `task test` and the atkins job `test` run the same command.

## Integration tests

An integration test crosses a process boundary. It does at least one of these:

- runs another process: `sh`, `git`, `ssh`, `docker`, a fake `claude` script, or the test binary itself;
- queries a database, sqlite included, even `:memory:` or a shared in-memory database;
- opens a network listener or dials one: `httptest.NewServer`, `net.Listen`, a real HTTP request or an ssh connection.

An integration test lives in a file named `<name>_integration_test.go` beside the code, and the file opens with the build tag:

```go
integration

package p
```

A file holds tests of one kind. Helpers that only integration tests call go in the tagged file with them; a helper that both kinds call stays in the untagged file and crosses no boundary itself.

Integration tests take longer and need `git`, `sh` and the other tools on the machine. `scripts/test-integration.sh` runs them, and `task test:integration` and the atkins job `test:integration` call it:

```bash
scripts/test-integration.sh                  # every package
scripts/test-integration.sh ./storage/...    # one package or pattern
task test:integration -- ./frontend/...      # the same, through task
```

One `go test` per package, and a package holding more tests than the batch size is split by test name into batches of that size. The names come from `go test -list`, so a test added to a split package runs without the script being edited. Three variables set the bounds:

| Variable             | Default | What it sets                                     |
|----------------------|---------|--------------------------------------------------|
| `TASK_UI_TEST_BATCH` | `20`    | Tests per `go test`, over which a package splits |
| `TASK_UI_TEST_MEM`   | `2GiB`  | `GOMEMLIMIT` of each `go test`                   |
| `TASK_UI_TEST_PROCS` | `2`     | `GOMAXPROCS` of each `go test`                   |

The split is what keeps the suite inside the memory cap. A cap binds a process, not a run: `go test -tags integration ./...` builds one binary per package, and `frontend` holds enough tests that running them together needs more than the cap, which ends the whole suite in an OOM that names no package. One `go test` per batch keeps a failure local and prints which package and which batch it is.

Nothing here runs a benchmark: the repository holds none, and `go test` runs one only under `-bench`. Coverage is its own target, `task cover`, because `-cover` instruments every package and the correctness run does not read the number.

## Where integration tests belong

Integration tests belong in the `storage` package and in the drivers, the packages that wrap a process or a protocol: `git`, `runner`, `sshprobe`, `agent` and the like. They test the real implementation, with no mocks. A package that consumes storage or a driver, such as `api`, `frontend`, `mcp` or `note`, tests its own code with unit tests over fakes.

## A module's PHP

A [module](modules.md) is PHP, and PHP is tested with phpscript, not with Go. Two commands do it, both run from the module's root, the directory `modinfo.php` sits in:

```bash
phpscript test ./tests/...     # the .phpt fixtures
phpscript run precompile.php   # a script, end to end
```

`atkins test` in the module runs the first and `atkins precompile` the second, and `atkins` runs fmt, lint, test and precompile in that order.

A fixture is a `.phpt` file: front matter with a `name` and a `description`, the PHP, and the output it must print, each part separated by a `---` line.

```text
name: refinement_kind
description: >
  Whether the thread page embeds an artifact or links it.
---
<?php

echo refinement_kind("diagrams/schema.svg"), "\n";
---
image
```

Print with `echo`, not `var_dump`. `echo` writes the value and nothing else, which is the same in every runtime; `var_dump` writes the type and the length beside it, and those are the runtime's own and move when a binding changes what a value is made of.

### The suite root

`phpscript test` roots a fixture at the directory the fixture sits in, so a fixture in `tests/` sees `tests/` as `/` and nothing of the module. A `phpscript.yml` in the tree moves that root: the directory holding the file is the application root for every fixture below it. The module keeps one at its own root, so a fixture reads the tree `index.php` reads and `require "lib/render.php"` means the same path in both.

```yaml
test:
  include: tests/bootstrap.php
```

`include` runs ahead of every fixture. `tests/bootstrap.php` requires `ext/autoload.php` and the module's own declarations, so a fixture names the function it is about and nothing else.

### What runs and what cannot

phpscript is the runtime, not the instance. It carries the PHP standard library and nothing of task-ui: `Task\Module`, `Task::register`, `Task::dispatch`, `Task::migrate`, `Task::markdown`, `Task::artifacts`, `Task::artifact` and `Database` are Go bindings the host registers on a runtime it builds itself, and a `phpscript test` or `phpscript run` has none of them.

So a fixture covers what stands on its own, which is most of what a module decides:

| Runs under phpscript | Does not |
|----------------------|----------|
| the pure functions of `lib/`: routing a path, a title, an artifact's kind | anything calling `Task::` |
| the stamp functions, over the runtime's own `DateTime` | `storage/`, a class over a `Database` |
| MiniTPL: compiling a template, rendering one with data | `modinfo.php` and `migrate.php`, which are `Task::` from their first line |

Loading a file is not calling it: `lib/thread.php` declares functions that reach `Task::` and the repository, and the declaration costs nothing, so the bootstrap loads the file and a fixture calls the functions that stand alone.

For what cannot run, `stubs/` holds PHP declarations of the bindings. They are for `phpscript lint --include`, so the linter knows the names; nothing requires them at runtime and they do not make a binding callable.

### What the Go tests cover

The Go side of a module is packaging, not behaviour, and `phpscripts/phpscripts_test.go` is where it is pinned: that the embedded tree extracts, that a module installs its composer tree into `ext/` and not `vendor/`, and that every `require` a module writes names a file that shipped. Behaviour belongs in a `.phpt` fixture. Do not write a Go integration test to cover what a module's PHP does.

---

[Contents](README.md) · [Development](development.md) · [Modules](modules.md) · [Running the server](running.md)
