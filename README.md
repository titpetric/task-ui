# Task UI

Run your `Taskfile.yml` from the browser.

![](https://raw.githubusercontent.com/titpetric/task-ui/main/.github/assets/task-ui.png)

Start the Docker image with `task docker:run`.

Task UI is meant for Docker environments. A generic Dockerfile exists, which
bundles typical dependencies like `task`, `ttyrec`, `docker`, `docker compose`.

To use, start by navigating to the
[docker](https://github.com/titpetric/task-ui/tree/main/docker)
subfolder. It contains a Taskfile, with the typical commands to build and
run task-ui from a Docker image. For examples with Taskfiles you could
run, look into the folder
[examples](https://github.com/titpetric/task-ui/tree/main/examples).

The layout is somewhat responsive, supporting mobile.

# Running

To set up your project to run with Task UI, it's recommended you use the
example Docker Compose setup here:

```yaml
services:
  runner:
    image: titpetric/task-ui
    restart: always
    build: .
    command:
      - --history-enable
    ports:
    - 3000:3000
    volumes:
    - $PWD/app:/app
    - /var/run/docker.sock:/var/run/docker.sock:ro
```

In particular, you should mount your `/app` folder which contains your
`Taskfile.yml`, `docker-compose.yml` and other files. Task UI will run
with what you provide it with.

- If you don't want history, remove the `command` flags.
- If you don't want to use Docker, remove the volume for `docker.sock`.

The image provides an `id_ecdsa` key to use for SSH hops. The recommended
way to deploy is to provide your own `docker/root/.ssh` folder with
the SSH keys. You can regenerate the SSH key with `task docker:gen`.

# Development

task: Available tasks for this project:

* default:            Run everything
* fix:                Fix code
* install:            Install task-ui
* run:                Run task-ui
* test:               Test task-ui
* docker:build:       Build task-ui docker image
* docker:gen:         Generate ssh key for docker image
* docker:push:        Push task-ui to registry
* docker:run:         Run task-ui in docker env

## task: default

Run everything

commands:
 - Task: fix
 - Task: install
 - Task: test

## task: test

Test task-ui

commands:
 - CGO_ENABLED=1 go test -race -count=1 -cover ./...
 - CGO_ENABLED=0 go test -count=1 -cover ./...

## task: run

Run task-ui

commands:
 - task-ui --history-enable

## task: install

Install task-ui

commands:
 - CGO_ENABLED=0 go install .

## task: fix

Fix code

dependencies:
 - deps:goimports

commands:
 - goimports -w .
 - go fmt ./...
 - go vet .
 - go mod tidy
 - ./README.md.sh > README.md

