# run-buttons

Input scope: `data_run`, from `Task\Control::runButtons()`.

| Field       | Type   | What it is                                           |
|-------------|--------|------------------------------------------------------|
| `folder`    | string | The folder the job runs in                            |
| `dir`       | string | The directory inside it; empty for the folder's own job |
| `job`       | string | The job's ID, which the run names                     |
| `name`      | string | The job's name, which the caret's hover names         |
| `what`      | string | `folder[/dir] / name`, the title the run is listed under |
| `place`     | string | `folder[/dir]`, which is the title Run force gives the run |
| `args`      | string | The command the arguments follow; empty for a job that takes none |
| `takesArgs` | bool   | Whether Run ends in `...`, for a job that opens a field first |
| `summarize` | bool   | Whether the caret offers Run and summarize            |
| `force`     | bool   | Whether it offers Run force, for a job in a workspace with hours |
| `menu`      | bool   | Whether there is a caret at all, which is either of the two |

The buttons carry `data-run` and the rest of the attributes `task-ui.js` reads:
it starts the run over the API, opens a field for the arguments first where
`data-args` says so, and asks for a reason where `data-force` does.

`place` is the title Run force gives the run and it leaves out the job's name,
which is what the Go page writes.
