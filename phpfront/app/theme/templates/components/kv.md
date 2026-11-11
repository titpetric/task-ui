# kv

Input scope: `data_kv`, a table of a label and a value a row each.

| Field              | Type   | What it is                                        |
|--------------------|--------|---------------------------------------------------|
| `rows`             | list   | The rows, in the order the page lists them        |
| `rows[].label`     | string | The `th` on the left                              |
| `rows[].class`     | string | The class of the `td`; empty for none             |
| `rows[].who`       | object | Who started a run, read by `who`; empty draws none |
| `rows[].state`     | object | A run's state, read by `state`; empty draws none  |
| `rows[].link`      | object | A link: `href`, `icon` and `text`; an empty `href` draws none |
| `rows[].code`      | list   | Values in `code`, a space after each              |
| `rows[].text`      | string | The plain value                                   |
| `rows[].faint`     | string | A faint value, as the finish of a run in progress |
| `rows[].after`     | string | The dim text after the value; empty draws none    |
| `rows[].badge`     | object | A badge after the value: `tone` and `text`; empty `text` draws none |

Every key is set on every row. A cell is drawn in the order of the table
above, so a row that holds a state and `after` reads as the state and then why
it ended. Each piece is named on its own, so a page builds no HTML and the
escaping stays here.

A page whose detail table holds more than this writes its own markup in the
module, as the Artifacts table of a run's page does.
