# integration-field

Input scope: `data_field`, one setting as `mod_integrations::field()` draws it.

| Field         | Type   | What it is                                                     |
|---------------|--------|----------------------------------------------------------------|
| `kind`        | string | `who`, `env`, `secret`, `select` or `text`: the branch drawn   |
| `id`          | string | The input's ID, the service and the setting                    |
| `name`        | string | What the form posts it under, `values.<name>`                  |
| `label`       | string | What the setting is called                                     |
| `wide`        | bool   | Whether the field takes the whole row                          |
| `badge`       | string | `set` or `empty` for a secret, "" for every other kind         |
| `tone`        | string | The badge's colour: `ok` for a secret with a value             |
| `help`        | string | The line under the input                                       |
| `spaced`      | bool   | Whether a space follows the help, where a variable follows it  |
| `env`         | string | The variable the setting sets in a run, "" for none            |
| `placeholder` | string | What an empty input says; a set secret says it keeps its value |
| `value`       | string | What the input holds, the default where the setting is empty   |
| `options`     | list   | `value` and `selected`, for the list to pick from              |
| `toggle`      | bool   | Whether the label shares its row with a switch                 |
| `switch`      | scope  | `components/integration-switch.tpl`, empty for no switch       |
| `who`         | scope  | `components/who-field.tpl`, empty for every other kind         |

One setting of a service's form. The kind picks the branch: the Who boxes, a switch alone for a variable whose value is not the reader's, a secret, a list to pick from, or a line of text.

A secret holds no value here. The badge says whether one is set and the placeholder says that an empty field keeps it, which is the whole of what the form knows: the value is encrypted with the instance key, which does not enter PHP.
