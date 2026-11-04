# sidebar-folder

Input scope: `data_folder`, one folder inside a project of the folder tree: a
disclosure of its jobs, which names its path below the project.

| Field      | Type   | What it is                                               |
|------------|--------|----------------------------------------------------------|
| `id`       | string | The folder's ID, which the script reads                  |
| `name`     | string | The last segment of the ID                               |
| `dir`      | string | The path between the project and the folder, "" for one right inside it |
| `href`     | string | The folder's page, without the base                      |
| `title`    | string | What the summary says on a mouse-over                    |
| `active`   | bool   | Whether the folder is the one the page is about          |
| `open`     | bool   | Whether the row starts unfolded                          |
| `icon`     | string | A sprite symbol, without the `i-` prefix                 |
| `dirty`    | int    | Changed and untracked files under it, 0 for none          |
| `changed`  | string | What the icon says on a mouse-over, "" where `dirty` is 0 |
| `bad`      | bool   | Whether the folder's pipeline would not read             |
| `badge`    | string | What the badge of a bad folder says                      |
| `badTitle` | string | What that badge says on a mouse-over                     |
| `count`    | int    | How many rows the folder lists                           |
| `jobs`     | list   | The rows, for `sidebar-jobs`                             |
| `empty`    | string | What the line says where it lists none, for `sidebar-jobs` |

It passes its own scope on to `sidebar-icon` and to `sidebar-jobs`, both of
which read a row of the tree.
