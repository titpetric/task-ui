# table

Input scope: `data_table`, from `Task\Table::data()`.

| Field    | Type   | What it is                                                 |
|----------|--------|------------------------------------------------------------|
| `class`  | string | The table's own classes, after `table`; empty for none     |
| `wrap`   | string | The wrap's own classes, after `table-wrap`; empty for none |
| `header` | list   | Rows of `th`; empty draws no `thead`                       |
| `body`   | list   | Rows of `td`                                               |
| `footer` | list   | Rows of `td`; empty draws no `tfoot`                        |

Every row is read by `table-row`.
