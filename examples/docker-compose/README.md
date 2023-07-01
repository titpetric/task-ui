# Docker compose

task: Available tasks for this project:

* build:        Build docker image
* down:         Shut down env
* logs:         Tail container logs (live)
* pull:         Pull docker images
* status:       Environment status      (aliases: ps)
* up:           Bring up env

```yaml
---
version: "3"

tasks:
  build:
    desc: "Build docker images"
    cmds:
      - docker compose build

  pull:
    desc: "Pull docker images"
    cmds:
      - docker compose pull

  status:
    aliases: [ps]
    desc: "Environment status"
    cmds:
      - docker compose ps

  up:
    desc: "Bring environment up"
    cmds:
      - docker compose up -d --remove-orphans

  down:
    desc: "Bring environment down"
    cmds:
      - docker compose down

  logs:
    desc: "Tail container logs (live)"
    cmds:
      - docker compose logs --tail=10 -f
```
