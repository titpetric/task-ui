# Digital Ocean (doctl)

task: Available tasks for this project:

* create:         Create a new droplet
* delete:         Delete a droplet
* init:           Initialize doctl
* list:           List all droplets
* ssh:            SSH into a droplet
* ssh-keys:       List all ssh keys

```yml
---
version: "3"

tasks:
  list:
    desc: "List all droplets"
    cmds:
      - doctl compute droplet list

  create:
    desc: "Create a new droplet"
    cmds:
      - |
        echo "Droplet name:"
        read DROPLET
        doctl compute droplet create ${DROPLET} --image "ubuntu-20-04-x64" --size "s-1vcpu-1gb" --region "nyc1" --ssh-keys {{ .sshKeyFingerprint }}
    vars:
      sshKeyFingerprint:
        sh: task ssh-keys | head -n 1

  delete:
    desc: "Delete a droplet"
    cmds:
      - |
        echo "Droplet name:"
        read DROPLET
        doctl compute droplet delete ${DROPLET}

  ssh:
    desc: "SSH into a droplet"
    interactive: true
    cmds:
      - |
        echo "Droplet name:"
        read DROPLET
        doctl compute ssh ${DROPLET}

  ssh-keys:
    desc: "List all ssh keys"
    cmds:
      - doctl compute ssh-key list --format=FingerPrint --no-header

  init:
    desc: "Initialize doctl"
    interactive: true
    cmds:
      - doctl auth init
```