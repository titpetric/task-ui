# control-table

Input scope: `data_control`, a list of rows from `Task\Control::of()`.

| Field   | Type   | What it is                                           |
|---------|--------|------------------------------------------------------|
| `label` | string | The badge the row is named by: the pipeline file, or the group's kind |
| `title` | string | What the badge says on a hover; empty draws no attribute |
| `href`  | string | The place the row's targets run in, which the badge links to; empty draws a plain badge |
| `empty` | string | What a row of a pipeline file with no targets says instead of its buttons |
| `jobs`  | list   | The row's buttons, each read by `control-job`         |
| `menu`  | object | The row's `[...]`, from `Task\Menu::data()`           |

Every row is a named row: its buttons carry their labels beside the badge, so
the table has no head and no column, and the `colspan` is the one column the
buttons sit in. A row of a pipeline file is the one kind whose badge is a link
and the one kind that says anything when it holds no job.

The dashboard's control centre is the same jobs in columns, a column per name
that repeats, with a head over them. That is a second table and not this one:
nothing here draws a column.
