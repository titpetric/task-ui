# integration-card

Input scope: `data_card`, one service as `mod_integrations::card()` draws it.

| Field        | Type   | What it is                                                   |
|--------------|--------|--------------------------------------------------------------|
| `id`         | string | The service's ID, which keys the card and its API path       |
| `number`     | int    | Its place on the rail, from 1                                |
| `title`      | string | What the service is called                                   |
| `configured` | bool   | Whether its settings are complete, which draws the badge     |
| `menu`       | scope  | `components/menu.tpl`, the Clear that empties every setting  |
| `desc`       | html   | What the service is for, rendered from its markdown          |
| `api`        | string | The endpoint Save puts to                                    |
| `next`       | string | Where a save goes, with the notice it carries                |
| `fields`     | list   | `components/integration-field.tpl`, one setting each         |
| `env`        | list   | `components/integration-env.tpl`, empty for every other card |
| `credential` | string | What the steps lead to, which the details name               |
| `steps`      | list   | How to get it, each rendered from its markdown               |
| `hook`       | html   | The push hook a sync installs, "" for a card with no sync    |
| `runs`       | bool   | Whether the two sync buttons are there                       |
| `sync`       | bool   | Whether the card ends with the line on the last sync         |
| `last`       | scope  | That run: `state`, `ago`, `who` and `href`; empty for none   |

One service of the Integrations page: a step of the rail, foldable, holding what the service is for, the form of its settings, the steps to its credential and the actions on it.

Nothing here is a row of the database. The registry is Go data and the settings are secrets the instance key encrypts, so the whole card arrives from `Task\Host::integrations()`, and a secret arrives as set or empty.

The two cards that differ say so as fields and not as their ID: the gateway holds its variables in `env` and one field, and the git host holds a `hook`, `runs` and `sync`. A card draws what it holds.
