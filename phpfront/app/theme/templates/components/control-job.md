# control-job

Input scope: `data_cc`, from `Task\Control::button()`.

| Field      | Type   | What it is                                           |
|------------|--------|------------------------------------------------------|
| `key`      | string | What the job's run is listed under, which the script redraws the slot by |
| `folder`   | string | The folder the job runs as                            |
| `job`      | string | The job's ID, which a run names                       |
| `name`     | string | The job's name, which a confirmation names            |
| `label`    | string | What the button says: the name, or the ID where the two differ |
| `last`     | string | The colour the last run gives it: `never`, `ok` or `bad` |
| `can`      | bool   | Whether the reader may run it                         |
| `dir`      | string | The directory inside the folder it runs in; empty for the folder's own |
| `desc`     | string | What it does, where that says more than its name; empty draws none |
| `args`     | string | The command the arguments follow; empty for a job that takes none |
| `confirm`  | string | What it asks before it runs; empty asks nothing       |
| `path`     | string | The file a job of a file runs; empty for every other kind |
| `terminal` | bool   | Whether it takes a terminal, which the glyph and the hover say |
| `what`     | string | `folder[/dir] / name`, the title the run is listed under |
| `run`      | string | The run in progress; empty draws the Run instead      |
| `href`     | string | The job's own page, where a reader who may not run it goes |
| `title`    | string | The hover of whichever of the three the button is     |

The button is one of three and never two: an eye on the run in progress, a Run,
or a `#!` to the job's page for a reader without the permission. `title` is the
hover of the one drawn, which `Task\Control::says()` writes, because each of
the three says a different sentence out of the same values.

The scope is `data_cc` and not `data_job`: the slot is `data-cc`, and a page
that draws this table names a job of its own.

Every row of the Jobs table is a named row, so the button always carries its
label. The control centre of the dashboard draws the same job as a bare button
in a column, which is a second shape and not this one.
