# Rclone from dropbox example

task: Available tasks for this project:

* copy:       Copy Camera Uploads from Dropbox
* move:       Move Camera Uploads from Dropbox

```yaml
---
version: "3"

tasks:
  copy:
    desc: "Copy Camera Uploads from Dropbox"
    cmds:
      - rclone copy -P "Dropbox:Camera Uploads" ./Camera

  move:
    desc: "Move Camera Uploads from Dropbox"
    cmds:
      - rclone move -P "Dropbox:Camera Uploads" ./Camera
```