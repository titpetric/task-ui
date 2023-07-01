# Terraform

task: Available tasks for this project:

* apply:       Apply Terraform configuration
* lint:        Lint Terraform files

```yaml
---
version: "3"

tasks:
  lint:
    desc: "Lint Terraform files"
    cmds:
      - terraform fmt -check=true -recursive

  apply:
    desc: "Apply Terraform configuration"
    cmds:
      - terraform init
      - terraform validate
      - terraform plan
      - terraform apply
```