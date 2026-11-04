# table-row

Input scope: `data_row`, one row of a table.

| Field           | Type   | What it is                                      |
|-----------------|--------|-------------------------------------------------|
| `id`            | string | The row's anchor; empty writes no attribute     |
| `run`           | string | The run the row is of, which the script keys on |
| `class`         | string | The row's class; empty writes none              |
| `cells`         | list   | The cells, in order                             |
| `cells[].html`  | string | The cell, as HTML, drawn raw                    |
| `cells[].class` | string | The cell's class; empty writes none             |
| `cells[].span`  | int    | `colspan`; 0 writes none                        |

`data-run-id` on the row claims `tr[data-run-id]`, where the state label claims
`span[data-run-id]`.
