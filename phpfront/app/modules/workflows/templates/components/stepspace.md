# stepspace.tpl

The Workspace field of a step card of the workflow form.

Input scope: `data_space`.

| Key             | Type | What                                      |
|-----------------|------|-------------------------------------------|
| `none`          | bool | offer no workspace, the workspaces root   |
| `noneSelected`  | bool | that option is the one picked             |
| `spaces`        | list | the workspaces                            |
| `flows`         | bool | offer Workflows, for a step that runs one |
| `flowsSelected` | bool | that option is the one picked             |

Each item of `spaces`:

| Key        | Type   | What                 |
|------------|--------|----------------------|
| `id`       | string | the workspace        |
| `title`    | string | its name             |
| `selected` | bool   | it is the one picked |
