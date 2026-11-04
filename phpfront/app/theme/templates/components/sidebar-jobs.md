# sidebar-jobs

Input scope: `data_jobs`, the jobs under one row of the folder tree. Any row
scope of `sidebar-folders` is one: the component reads two of its fields, and
every row of the tree carries both.

| Field   | Type   | What it is                                                 |
|---------|--------|------------------------------------------------------------|
| `jobs`  | list   | The folder's own jobs, then the pins its row lists besides them |
| `empty` | string | What the line says where the folder lists neither           |

Each entry of `jobs`:

| Field     | Type   | What it is                                               |
|-----------|--------|----------------------------------------------------------|
| `key`     | string | What the job's run is listed under, which the script reads |
| `pinned`  | bool   | Whether the row is a pin rather than a job of the pipeline |
| `active`  | bool   | Whether the job is the one the page is about              |
| `running` | bool   | Whether a run of it is in progress                        |
| `attach`  | string | The run to attach to, "" where none is in progress        |
| `href`    | string | The job's page, without the base                          |
| `hint`    | string | What the link says on a mouse-over, "" for a job of the pipeline |
| `kind`    | string | The convention in front of the job's name, "" where its name says it |
| `label`   | string | What the link says                                        |
| `can`     | bool   | Whether the reader may start it                           |
| `folder`  | string | The folder it runs in, which the run button carries       |
| `dir`     | string | The directory inside the folder, "" for the folder itself  |
| `job`     | string | The job the run button names                              |
| `args`    | string | The arguments it runs with, as JSON, "" for none           |
| `said`    | string | What the run button says it runs                          |
| `name`    | string | The job's own name, which the attach button says           |

A row draws a button at most: the one that attaches where a run is in
progress, the one that starts where the reader may, and neither otherwise.

The empty line is drawn where `jobs` is empty, which is a folder with no job
and no pin of its own. A project's row leaves the whole block out instead, so
the line belongs to a folder inside one.
