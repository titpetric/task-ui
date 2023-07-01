# Remote shell taskfile

task: Available tasks for this project:

* bash:       Drop into a bash shell

```yaml
---
version: "3"

tasks:
  bash:
    desc: "Drop into a bash shell"
    interactive: true
    cmds:
      - bash
```
