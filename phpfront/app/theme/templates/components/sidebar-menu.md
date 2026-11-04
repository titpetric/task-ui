# sidebar-menu

Input scope: `data`, a list of rows in the order the sidebar table holds them,
less the hidden ones and the ones whose rule the reader lacks.

| Field    | Type   | What it is                                                |
|----------|--------|-----------------------------------------------------------|
| `title`  | string | What the row says                                         |
| `href`   | string | The whole link, the base in it where the base applies     |
| `icon`   | string | A sprite symbol, without the `i-` prefix                  |
| `active` | bool   | Whether the row is the page being read                    |
| `flows`  | bool   | Whether the instance's workflows list under the row       |
| `pulse`  | bool   | Whether the row carries the dot that pulses while running |

`Task\Sidebar::items()` builds it. The two flags are what the component keys
on, so renaming a page does not quietly empty either.

`href` is the one link in the theme the component does not prefix itself: a
module's row points at `/m/<vendor>/<name>`, which is another application on
the instance and sits outside this one's mount, so `Task\Sidebar::link()`
decides per row and the component writes what it got.

It also reads `chrome.running`, whether any run is in progress, which lights
the dot of the row that carries one, and `chrome.workflows`, which it hands to
`sidebar-workflows` under the row whose `flows` is set. Both are the page's
facts and not the row's, which is why neither is in the scope above.

A reader with no workflow at all gets no block, which is why the list is tested
here and not inside the component.
