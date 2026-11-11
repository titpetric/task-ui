# integration-switch

Input scope: `data_switch`, from `mod_integrations::flag()`.

| Field  | Type   | What it is                                            |
|--------|--------|-------------------------------------------------------|
| `name` | string | What the box posts under, `values.<name>.enabled`     |
| `env`  | string | The variable it sends, which the title and label name |
| `on`   | bool   | Whether the switch is on                              |

The switch of one setting: on sends its variable in a run, off leaves it out and keeps the value for the next switch on. Its stored row is the setting's name with `_enabled` after it, which the API writes.
