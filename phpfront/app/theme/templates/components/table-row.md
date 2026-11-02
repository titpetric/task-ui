# table-row

Input scope: `data_row`, one row of a table.

| Field          | Type   | What it is                                           |
|----------------|--------|------------------------------------------------------|
| `id`           | string | The row's anchor; empty writes no attribute           |
| `run`          | string | The run the row is of, which the script keys on       |
| `class`        | string | The row's class; empty writes none                    |
| `cells`        | list   | The cells, in column order                            |
| `cells[].html` | string | The cell, as HTML                                     |
| `cells[].class`| string | The cell's class; empty writes none                   |

`html` is drawn raw, so the module builds it: a cell holding a value a reader
typed escapes it, and `Task\Table::cell()` is what does that.

`data-run-id` on the row is the runs table's contract, which updates a row
whole from the event stream. It claims `tr[data-run-id]`, where the state
label claims `span[data-run-id]`, so the two never meet.
