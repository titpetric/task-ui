# Testing

task-ui has two kinds of tests, and `go test ./...` runs only the first.

## Unit tests

A unit test stays inside the test process. It calls the code of its
package, puts fakes behind the interfaces that code consumes, serves a
handler with `httptest.NewRecorder` and `ServeHTTP`, and reads and
writes files under `t.TempDir()`. Unit tests are the default run:

```bash
go test -cover ./...
```

The whole run takes seconds. `task test` and the atkins job `test` run
the same command.

## Integration tests

An integration test crosses a process boundary. It does at least one of
these:

- runs another process: `sh`, `git`, `ssh`, `docker`, a fake `claude`
  script, or the test binary itself;
- queries a database, sqlite included, even `:memory:` or a shared
  in-memory database;
- opens a network listener or dials one: `httptest.NewServer`,
  `net.Listen`, a real HTTP request or an ssh connection.

An integration test lives in a file named `<name>_integration_test.go`
beside the code, and the file opens with the build tag:

```go
//go:build integration
```

A file holds tests of one kind. Helpers that only integration tests
call go in the tagged file with them; a helper that both kinds call
stays in the untagged file and crosses no boundary itself.

Integration tests take longer and need `git`, `sh` and the other tools
on the machine. They run with the tag alone:

```bash
go test -tags integration -count=1 ./...
```

`task test:integration` and the atkins job `test:integration` run the
same command.

## Where integration tests belong

Integration tests belong in the `storage` package and in the drivers,
the packages that wrap a process or a protocol: `git`, `runner`,
`sshprobe`, `agent` and the like. They test the real implementation,
with no mocks. A package that consumes storage or a driver, such as
`api`, `frontend`, `mcp` or `note`, tests its own code with unit tests
over fakes.

---

[Contents](README.md) · [Development](development.md) · [Running the server](running.md)
