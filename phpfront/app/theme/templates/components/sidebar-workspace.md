# sidebar-workspace

Input scope: `data_workspace`, one named workspace of the folder tree: its row
folds the projects under it, each a project with a fold of its own.

| Field      | Type   | What it is                                               |
|------------|--------|----------------------------------------------------------|
| `top`      | bool   | Whether the workspace's own folder is listed              |
| `id`       | string | The folder's ID, which the script reads                  |
| `name`     | string | The last segment of the ID                               |
| `href`     | string | The workspace's page, without the base                   |
| `title`    | string | What the summary says on a mouse-over                    |
| `active`   | bool   | Whether the page is about the workspace or something in it |
| `here`     | bool   | Whether the page is about the workspace itself            |
| `bad`      | bool   | Whether the folder's pipeline would not read             |
| `badge`    | string | What the badge of a bad folder says                      |
| `badTitle` | string | What that badge says on a mouse-over                     |
| `count`    | int    | How many projects it lists                               |
| `jobs`     | list   | The rows under it, for `sidebar-jobs`                    |
| `empty`    | string | What the line says where it lists none, for `sidebar-jobs` |
| `projects` | list   | The projects, each a `sidebar-project` scope             |

A workspace's row draws no icon: the chevron stands in front of the name and
the projects below it carry the icons.

Without `top` the projects stand on their own, with no row above them, and the
row's own fields go nowhere.
