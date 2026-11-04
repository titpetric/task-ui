# wfstep.tpl

One step card of the workflow form: its node on the rail, its head with the grip, its kind and the step in one line, the `[...]` menu, and the fields of its kind under it.

Input scope: `data_step`. It also sets `data_menu`, `data_space`, `data_hostjobs` and `extra` for the components it includes.

| Key            | Type   | What                                            |
|----------------|--------|-------------------------------------------------|
| `kind`         | string | job, workflow, script, prompt or summary        |
| `disabled`     | bool   | runs skip the step                              |
| `open`         | bool   | the card is unfolded                            |
| `icon`         | string | `#i-<icon>` of the sprite                       |
| `kindLabel`    | string | the kind as the card names it                   |
| `line`         | string | the step in one line, the folded card's heading |
| `menu`         | map    | the `[...]` menu, the scope of menu.tpl         |
| `missing`      | string | what the step's place lacks, "" for nothing     |
| `runs`         | bool   | the step runs a job or a workflow               |
| `script`       | bool   | the step runs a command of its own              |
| `asks`         | bool   | the step is a prompt                            |
| `sums`         | bool   | the step is a summary                           |
| `folderLabel`  | string | Folder, or Workflow for a step that runs one    |
| `pickNone`     | bool   | neither a folder nor a workflow is picked       |
| `places`       | list   | the folders the step may run in                 |
| `hosts`        | bool   | offer the root's host jobs                      |
| `rootSelected` | bool   | the root is the folder picked                   |
| `flows`        | list   | the workflows the step may run                  |
| `jobs`         | list   | every job of every folder                       |
| `hostJobs`     | list   | the host jobs, the scope of host-jobs.tpl       |
| `args`         | string | the arguments on one line                       |
| `command`      | string | the command of a script step                    |
| `prompt`       | string | the prompt's text                               |
| `none`         | bool   | offer no folder, the workspace root             |
| `noneSelected` | bool   | no folder is picked                             |
| `extra`        | map    | the place the reader has no folder for, or null |
| `dir`          | string | the directory inside the folder                 |
| `question`     | string | what a summary asks Claude                      |
| `instruction`  | string | the instruction beside it                       |
| `space`        | map    | the Workspace field, the scope of stepspace.tpl |

Each item of `places`, and `extra`:

| Key         | Type   | What                                  |
|-------------|--------|---------------------------------------|
| `id`        | string | the folder                            |
| `workspace` | string | the workspace it is in                |
| `rel`       | string | the folder relative to that workspace |
| `label`     | string | the folder as a reader reads it       |
| `selected`  | bool   | it is the one picked, `places` only   |

Each item of `flows`:

| Key        | Type   | What                 |
|------------|--------|----------------------|
| `id`       | string | the workflow         |
| `name`     | string | its name             |
| `selected` | bool   | it is the one picked |

Each item of `jobs`:

| Key        | Type   | What                     |
|------------|--------|--------------------------|
| `id`       | string | the job                  |
| `folder`   | string | the folder it belongs to |
| `name`     | string | its name                 |
| `selected` | bool   | it is the one picked     |
