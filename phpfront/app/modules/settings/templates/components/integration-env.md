# integration-env

Input scope: `data_env`, from `mod_integrations::env()`: one variable a row.

| Field    | Type   | What it is                                                 |
|----------|--------|------------------------------------------------------------|
| `env`    | string | The variable's name, the row's first cell                  |
| `help`   | string | What it does, under the name                               |
| `source` | bool   | Its value is the key the reader typed, so the cell says so |
| `empty`  | bool   | It is sent with no value, so the cell says so              |
| `id`     | string | The input's ID, for a variable the reader fills in         |
| `name`   | string | What the form posts it under, `values.<name>`              |
| `value`  | string | What the input holds, the default where it is empty        |

The variables of the gateway, which a run gets while its key is set. The key itself is the one field of the form above; these are what the key turns into, so they are a table and not a field each: two the key fills, one sent empty, and the rest lines of text with their defaults in them.
