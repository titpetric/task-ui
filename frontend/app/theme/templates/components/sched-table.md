# sched-table

Input scope: `data_sched`, from `Task\Schedule::table()`.

| Field                | Type   | What it is                                      |
|----------------------|--------|-------------------------------------------------|
| `can`                | bool   | Whether the reader may change a schedule, which draws the `[...]` |
| `rows`               | list   | The schedules, in the order the query read them |
| `rows[].expression`  | string | The cron expression, in a `code`                |
| `rows[].on`          | bool   | Whether the schedule runs                       |
| `rows[].next`        | list   | When it runs next, the day and the time apart   |
| `rows[].last`        | string | How long ago it last ran; empty says never      |
| `rows[].pick`        | bool   | Whether the cell picks another user, which an admin does |
| `rows[].user`        | object | Who it runs as, from `Task\Who::of()`           |
| `rows[].options`     | object | The users to pick from, read by `user-options`  |
| `rows[].api`         | string | Where the pick is saved, at the root            |
| `rows[].triage`      | bool   | Whether Claude triages a failed run of it       |
| `rows[].menu`        | object | The row's `[...]`, from `Task\Menu::data()`     |

The schedules of one job, as its own page lists them: the expression, when it
runs, who it runs as and whether it is on. No schedules draws nothing at all,
not an empty table.

It is not the table of the schedules page, which lists every schedule of every
folder and heads two columns more with the folder and the job; the expression
there wraps in a narrower column and carries `nowrap`. The two tables answer
different questions and are two components, as they are two templates in the Go
front end.

`can` is the reader's permission over the whole table and not a row's: a
schedule of a job is changed by whoever may run that job, so the one answer
holds for every row. A reader without it still sees the cell, empty, because
the columns line up across readers.

`api` carries no mount: the pick is a `PUT` of the API at the root, as every
`data-*` endpoint of a page is.
