# table

Input scope: `data_table`, from `Task\Table::data()`.

| Field              | Type   | What it is                                        |
|--------------------|--------|---------------------------------------------------|
| `columns`          | list   | The heading; empty draws no `thead`                |
| `columns[].title`  | string | What the heading cell says                         |
| `columns[].class`  | string | The `colgroup` class that sizes it; empty for none |
| `columns[].numeric`| bool   | Whether the heading is right aligned               |
| `rows`             | list   | The rows, each read by `table-row`                 |
| `width`            | int    | How many columns the empty row spans               |
| `empty`            | string | What an empty list says                            |

This is the plain list: a heading, rows, and a line when there are none. A
page whose table carries more than cells — a tree, a fold, a group heading —
writes its own markup in the module and does not bend this one.
