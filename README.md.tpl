# Task UI

Run your Taskfile.yml from the browser.

![](https://raw.githubusercontent.com/titpetric/task-ui/main/.github/assets/task-ui.gif)

Task UI is meant for Docker environments. A generic Dockerfile exists, which
bundles typical dependencies like `task`, `ttyrec`, `docker`, `docker compose`.

To use, start by navigating to the `docker/` subfolder. It contains a Taskfile,
with the typical commands to build and run task-ui from a docker image. For
examples with Taskfiles you could run, look into the folder.

# Running

To set up your project to run with Task UI, it's recommended you use the
example docker compose setup here:

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
- If you don't want to use docker, remove the volume for `docker.sock`.

The image provides an `id_ecdsa` key to use for ssh hops. The recommended
way to deploy is to provide your own `docker/root/.ssh` folder with
the ssh keys. You can regenerate the ssh key with `task docker:gen`.

