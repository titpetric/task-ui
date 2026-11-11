# crumbs

Input scope: `data`.

| Field          | Type   | What it is                                      |
|----------------|--------|-------------------------------------------------|
| `base`         | string | The path the application is mounted at          |
| `crumbs`       | list   | The trail, in order                             |
| `crumbs[].text` | string | What the step says                             |
| `crumbs[].href` | string | Where it goes, without the base; "" for the last |

The Dashboard is always the first step and is not in the list. An empty list
draws nothing, which is the dashboard itself.

`Task\Chrome::heading()` sets the title and the trail in one call.
