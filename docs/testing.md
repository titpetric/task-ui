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

---

[Contents](README.md) · [Development](development.md) · [Running the server](running.md)
