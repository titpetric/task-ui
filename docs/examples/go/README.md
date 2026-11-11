# Go development

task: Available tasks for this project:

* benchmark:       Run benchmarks for the Go project
* build:           Build the Go project
* test:            Run tests for the Go project

```yaml
---
version: "3"

tasks:
  build:
    desc: "Build the Go project"
    cmds:
      - go build ./...

  test:
    desc: "Run tests for the Go project"
    cmds:
      - go test -race -cover ./...

  benchmark:
    desc: "Run benchmarks for the Go project"
    cmds:
      - go test -race -cover -benchmem -run=^$ -bench=".*" ./...
```