# sidebar-menu

Input scope: `data`, a list of rows in the order the sidebar table holds them,
less the hidden ones and the ones whose rule the reader lacks.

| Field    | Type   | What it is                                                |
|----------|--------|-----------------------------------------------------------|
| `title`  | string | What the row says                                         |
| `href`   | string | The link, without the base                                |
| `icon`   | string | A sprite symbol, without the `i-` prefix                  |
| `active` | bool   | Whether the row is the page being read                    |
| `flows`  | bool   | Whether the instance's workflows list under the row       |
| `pulse`  | bool   | Whether the row carries the dot that pulses while running |

`Task\Sidebar::items()` builds it. The two flags are what the component keys
on, so renaming a page does not quietly empty either.

`flows` is read by `sidebar-workflows`, which arrives with the `workflows`
module.
