# sidebar-project

Input scope: `data_project`, one project of the folder tree: its row, the jobs
it lists and the folders inside it.

| Field      | Type   | What it is                                               |
|------------|--------|----------------------------------------------------------|
| `top`      | bool   | Whether the project's own folder is listed                |
| `id`       | string | The folder's ID, which the script reads                  |
| `name`     | string | The last segment of the ID                               |
| `href`     | string | The project's page, without the base                     |
| `title`    | string | What the summary says on a mouse-over                    |
| `active`   | bool   | Whether the page is about the project or something in it  |
| `here`     | bool   | Whether the page is about the project itself              |
| `icon`     | string | A sprite symbol, without the `i-` prefix                 |
| `dirty`    | int    | Changed and untracked files under it, 0 for none          |
| `changed`  | string | What the icon says on a mouse-over, "" where `dirty` is 0 |
| `bad`      | bool   | Whether the folder's pipeline would not read             |
| `badge`    | string | What the badge of a bad folder says                      |
| `badTitle` | string | What that badge says on a mouse-over                     |
| `count`    | int    | How many jobs the project's own pipeline declares         |
| `counted`  | bool   | Whether to draw the count at all                         |
| `jobs`     | list   | The rows under it, for `sidebar-jobs`                    |
| `empty`    | string | What the line says where it lists none, for `sidebar-jobs` |
| `nested`   | list   | The folders inside it, each a `sidebar-folder` scope      |

Without `top` the folders inside stand on their own, with no row above them,
and the row's own fields go nowhere.

The jobs block is left out where the project lists none, so the empty line of
`sidebar-jobs` belongs to a folder inside one and never to a project.
