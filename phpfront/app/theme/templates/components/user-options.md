# user-options

Input scope: `data_options`.

| Field          | Type   | What it is                                            |
|----------------|--------|-------------------------------------------------------|
| `users`        | list   | One option per user, in the order they are offered    |
| `users[].name` | string | The name, which is the value and the first word       |
| `users[].role` | string | The role, in brackets, where the long form writes it   |
| `users[].keys` | string | The instance SSH key names, space apart, for a script  |
| `users[].said` | string | `keys: a, b`, or `no keys`                             |
| `selected`     | string | The name the select opens on                           |
| `short`        | bool   | Whether the option leaves the role out                 |

The options of a select that picks who a run is as, from
`Task\User::options()`. The `<select>` around them is the page's: each carries
its own class, id, name and label.
