# Testing

task-ui has two kinds of tests, and `go test ./...` runs only the first.

## The budget

**A test takes under a second.** One that takes longer is waiting on something it could be told about, and the run says which:

```bash
atkins test:scope      # the packages a change touched, inside the budget
atkins test:budget     # every package, inside the budget
```

`scripts/test.sh` is both. It scopes by asking git what changed and `go list` what imports it, so a change to one package runs that package and the packages built on it rather than the ninety that cannot have noticed. It then reads its own run through `gotestsum tool slowest` and fails on anything over the budget, which is measured on the machine that ran it rather than taken from a table.

```bash
scripts/test.sh --all                 # every package
scripts/test.sh --base HEAD~3         # changed against another ref
scripts/test.sh --budget 500ms        # a tighter one
scripts/test.sh --tag integration ./storage/...
```

### What a test must not wait on

- **No `time.Sleep`.** Sleeping for a thing to happen is slow when it works and flaky when it does not. Wait on the channel, the subscription or the `Done()` the code already offers; to prove something did *not* happen, read the channel with a `default` and assert it was empty.
- **No `sleep` in a fixture's script.** A job that has to still be running when the test looks writes a file and waits for another to appear; the test writes that file when it is ready. A gate is immediate, a sleep is a guess.
- **Nothing that calls out to Claude or to a real host.** A fake on the PATH answers in microseconds and asserts what was asked of it.

### What a fixture must not build twice

Two costs were measured and both are now paid once per test binary rather than once per test:

| What                 | Was  | Is     | How                                                                                  |
|----------------------|------|--------|--------------------------------------------------------------------------------------|
| a migrated database  | 64ms | 0.22ms | `newDB(t)` in `storage/storage_integration_test.go`, one in-memory database per test |
| a 4096 bit RSA key   | 1.5s | 0ms    | `internal/testkey.RSA(t)` reads one generated once into `testdata`                   |
| the parsed templates | 65ms | once   | `shared(t)` in the view tests                                                        |

Reach for the same shape when a fixture builds something expensive and unchanging: build it once behind a `sync.Once`, hand out a copy.

### Finding the waste

```bash
atkins test:duplicate            # which two tests cover the same lines
scripts/test-duplicate.sh ./runner/... --overlap 90
```

It runs each test alone with a coverage profile of its own and compares the line sets. A pair over the threshold is a candidate and not a verdict: two tests may cover one function and assert different things about it. The report is for finding the pairs worth reading.

## Unit tests

A unit test stays inside the test process. It calls the code of its package, puts fakes behind the interfaces that code consumes, serves a handler with `httptest.NewRecorder` and `ServeHTTP`, and reads and writes files under `t.TempDir()`. Unit tests are the default run:

```bash
go test -cover ./...
```

The whole run takes seconds. `task test` and the atkins job `test` run the same command.

### A template test asserts that it rendered

A template is read, and a copy of its output pasted into a test says only that nobody has changed the template since: it fails on every edit, it is rewritten to match whatever came out, and it has caught nothing. A test may assert two things. The render returned no error, which fails when a template reaches for a field the page model does not hold. And it wrote something.

```go
for _, name := range v.Names() {
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, name, page), name)
	assert.True(t, buf.Len() > 0, name)
}
```

Forbidden: `assert.Contains` over markup, and an exact count. `assert.Len(t, rows, 3)` and `assert.Equal(t, 1, strings.Count(body, …))` assert the fixture; `assert.True(t, len(rows) > 0)` is the form. A golden file a tool regenerates is a diff against the last state and not an assertion somebody typed, so the `.phpt` fixtures and `scripts/compare-site.sh` below stay as they are.

`frontend/view` holds the markup assertions, and they are being removed. Until then `shared(t)` skips the tests that hold them, and `atkins test:view` runs them, which is the only job that does:

```bash
atkins test:view     # TASK_UI_VIEW_TESTS=1, under the caps the file sets
```

A failure over a page says where to look, not what the page holds: `assert.Brief` returns the haystack's length and the text around the longest prefix of the needle it holds.

Run the suite through atkins, not as a bare `go test ./...`. The `env.vars` block at the top of `atkins.yml` holds `GOFLAGS=-p=1 GOMAXPROCS=2 GOMEMLIMIT=2GiB` for every job, which builds one package at a time; uncapped, `go test ./...` links a hundred of them four at a time, and that is the run the kernel killed.

| Measured                                  | Value             |
|-------------------------------------------|-------------------|
| one failing `Contains` over the dashboard | 26,297 bytes      |
| the same failure through `assert.Brief`   | 473 bytes         |
| free memory when the kernel killed it     | 787 MB, swap full |

### A test binary holds tens of megabytes

Measure one, and run it from the package's own directory: a test that names a relative path walks the wrong tree from anywhere else, and `../..` out of `/tmp` is the whole filesystem.

```bash
go test -c -o /tmp/pkg.test ./workspace
cd workspace && /usr/bin/time -f "%M KB peak" /tmp/pkg.test -test.count=1
```

| Binary                                            | Peak   |
|---------------------------------------------------|--------|
| the heaviest of the 60 unit binaries              | 34 MB  |
| `storage` under `-tags integration`, 194 tests    | 46 MB  |
| `MaxBundle` of padding held as one `bytes.Repeat` | 190 MB |

A binary that peaks in the hundreds holds an input it could read in pieces. `TestReadBundle_TooLarge` feeds `ReadBundle` the padding that trips `MaxBundle` through an `io.LimitReader` over a reader of zeros, which is the last row of the table it does not pay. The other rule is the one `view.New` follows: a parsed template set is built once per binary, never once per fixture.

Three tools look at a template. `atkins precompile` compiles every `.tpl` and reports a syntax error, `phpscript lint ./...` reports a name the interpreter cannot resolve, and a six-line script renders one template over data you wrote. The `templates` skill has the script and the one trap in it.

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

An integration test gets five seconds rather than one, because it starts processes; `TASK_UI_TEST_BUDGET` moves that, and the run names anything over it. What takes longer than five seconds is waiting.

One `go test` per package, and a package holding more tests than the batch size is split by test name into batches of that size. The names come from `go test -list`, so a test added to a split package runs without the script being edited. Three variables set the bounds:

| Variable             | Default | What it sets                                     |
|----------------------|---------|--------------------------------------------------|
| `TASK_UI_TEST_BATCH` | `20`    | Tests per `go test`, over which a package splits |
| `TASK_UI_TEST_MEM`   | `2GiB`  | `GOMEMLIMIT` of each `go test`                   |
| `TASK_UI_TEST_PROCS` | `2`     | `GOMAXPROCS` of each `go test`                   |

The split is what keeps the suite inside the memory cap. A cap binds a process, not a run: `go test -tags integration ./...` builds one binary per package, and `frontend` holds enough tests that running them together needs more than the cap, which ends the whole suite in an OOM that names no package. One `go test` per batch keeps a failure local and prints which package and which batch it is.

Nothing here runs a benchmark: the repository holds none, and `go test` runs one only under `-bench`. Coverage is its own target, `task cover`, because `-cover` instruments every package and the correctness run does not read the number.

## Where integration tests belong

Integration tests belong in the `storage` package, and it is the only one that holds them: 28 files over a throwaway database, testing the real implementation with no mocks. A package that consumes storage, such as `api`, `frontend`, `mcp` or `note`, tests its own code with unit tests over fakes.

A driver — `git`, `runner`, `sshprobe`, `agent`, the packages that wrap a process or a protocol — has none of its own. The instance the shots suite builds is what runs them: real runs through the real runner, over real checkouts, read back through the pages that draw them, with the counters of that server in the coverage the suite reports. A driver whose behaviour no page reaches needs a test of its own, and this is the package it goes in.

## The front end

The pages are tested by an instance, not by a handler test: `scripts/demo.sh` builds the server, serves a copy of `testdata/demo` from four instances over one database, sets the instance up through the API, starts real runs, and then reads every page of it twice.

```bash
scripts/demo.sh                              # the screenshots, the crawl and the comparison
TASK_UI_SHOT_FILTER='^$' scripts/demo.sh     # the crawl and the comparison alone
task test:coverage -- shots                  # the same, instrumented, as a suite
```

Five things are asserted over that instance:

- **Every link answers.** `scripts/crawl.sh` starts at `/` and fetches every `href` that stays on the site, once each, and names what did not answer 200. The page list is the site's own, so a page is crawled the run after it is linked.
- **Every page it reached is the same page on the front end being replaced.** The ported pages are what the instance serves, at `/`, and the Go pages answer under `app.GoBase`; `scripts/compare-site.sh` hands each crawled path to `scripts/compare.sh`, which fetches `/<path>` and `/legacy/<path>` with one session, normalises both and diffs them; the exit code is the count that differ. [The port](phpfront.md) is what this is for, and the demo data is what makes it worth anything — a page compared on data that reaches none of its branches is compared on nothing. Both front ends are crawled and the page lists are read together, because each crawl is blind to a page only the other links; the prefixed crawl's status is discarded, since a Go page links the root as well as itself and prefixing those paths fabricates a 404.
- **The same pages again as the viewer**, over the same database. A permission gate one front end has and the other does not is invisible to an admin's sweep, because both draw everything.
- **The sign in page**, on the instance that has no autologin, because it is the one page a reader with no session is given and nowhere else draws it.
- **One page of each layout through a browser**, in both palettes. `scripts/compare-shots.sh` shoots `/<path>` and `/legacy/<path>` with `scripts/chromium.sh` and counts the pixels that are not the same colour with `magick compare -metric AE`; the bar is zero. A diff of the markup cannot see a stylesheet rule that matches a class one front end stopped writing, a font that resolves on one path, or an SVG drawn outside its box. The set is the distinct layouts and not the distinct pages, because a shot is a browser. The run page is not in it: its output block is a terminal the browser writes out of the recording it fetches after the load event, so the shot catches the placeholder or the screen depending on the moment, in either front end: two shots of one of them differ as far as a shot of each. A block the browser fills itself is compared as markup and never as pixels.

A page that differs is read a second time with the two fetches the other way round, and only what both readings hold in common is reported. Two kinds of difference are not one: a value that moves while the sweep walks — an mtime a run is rewriting — is in one reading and not the other, and a stamp of 179 seconds ago that the second fetch draws as 181 is `2m ago` against `3m ago` in the same direction on every reload, which the swap puts in the other diff. What survives both is the page.

The server is built with `go build -cover` when `TASK_UI_COVER_DIR` is set, so what a browser and a crawl reach counts towards the coverage `scripts/test-coverage.sh` reports. That is what the suite is for: the pages are read by the thing that serves them.

Each suite's whole output is kept at `cover/logs/<suite>.log`, and a failure prints the last thirty lines of it and that path. The sweep names a page where it reached it, so the one of a hundred that differed is usually further up than the tail.

## A module's PHP

A [module](modules.md) is PHP, and PHP is tested with phpscript, not with Go. Two commands do it, both run from the module's root, the directory `modinfo.php` sits in:

```bash
phpscript test ./tests/...     # the .phpt fixtures
phpscript run precompile.php   # a script, end to end
```

`atkins test` in the module runs the first and `atkins precompile` the second, and `atkins` runs fmt, lint, test and precompile in that order.

Two trees hold fixtures and each roots them at its own `phpscript.yml`: a module's, and `phpfront/app`, [the front end being ported](phpfront.md), whose `atkins` runs the same jobs over its own. `scripts/test-coverage.sh` runs both as the `php` suite, from their own directories.

A fixture is a `.phpt` file: front matter with a `name` and a `description`, the PHP, and the output it must print, each part separated by a `---` line.

```text
name: RefinementManager::title
description: >
  The first line of the comment a thread opened on, for the list.
---
<?php

echo RefinementManager::title("Design the user tables\nand the logins."), "\n";
---
Design the user tables
```

Print with `echo`, not `var_dump`. `echo` writes the value and nothing else, which is the same in every runtime; `var_dump` writes the type and the length beside it, and those are the runtime's own and move when a binding changes what a value is made of.

### The suite root

`phpscript test` roots a fixture at the directory the fixture sits in, so a fixture in `tests/` sees `tests/` as `/` and nothing of the module. A `phpscript.yml` in the tree moves that root: the directory holding the file is the application root for every fixture below it. The module keeps one at its own root, so a fixture reads the tree `index.php` reads and `require "ext/autoload.php"` means the same path in both.

```yaml
test:
  include: tests/bootstrap.php
```

`include` runs ahead of every fixture. `tests/bootstrap.php` requires `ext/autoload.php` and nothing else, so a fixture names the class it is about and the classmap finds it.

### What runs and what cannot

phpscript is the runtime, not the instance. It carries the PHP standard library and nothing of task-ui: `Task\Module`, `Task::register`, `Task::dispatch`, `Task::markdown`, `Task::artifacts`, `Task::artifact` and `Database` are Go bindings the host registers on a runtime it builds itself, and a `phpscript test` or `phpscript run` has none of them.

So a fixture covers what stands on its own, which is most of what a module decides:

| Runs under phpscript                                                       | Does not                                             |
|----------------------------------------------------------------------------|------------------------------------------------------|
| a manager's static side: routing a path, a title, grouping rows for a page | a manager's instance methods, which reach `Task::`   |
| composing a prompt out of rows, over plain arrays                          | the store, a class over a `Database`                 |
| the stamp functions, over the runtime's own `DateTime`                     | `modinfo.php`, which is `Task::` from its first line |
| MiniTPL: compiling a template, rendering one with data                     |                                                      |

Naming a class is not constructing one, which is why the static side is static: `RefinementManager::title` and `RefinementPrompt::flatten` are called straight off the class, and the fixture never reaches the constructor that wants a `Database`.

For what cannot run, `stubs/` holds PHP declarations of the bindings. They are for `phpscript lint --include`, so the linter knows the names; nothing requires them at runtime and they do not make a binding callable.

### What the Go tests cover

The Go side of a module is packaging, not behaviour, and `phpscripts/phpscripts_test.go` is where it is pinned: that the embedded tree extracts, that a module installs its composer tree into `ext/` and not `vendor/`, and that every `require` a module writes names a file that shipped. Behaviour belongs in a `.phpt` fixture. Do not write a Go integration test to cover what a module's PHP does.

---

[Contents](README.md) · [Development](development.md) · [Modules](modules.md) · [Running the server](running.md)
