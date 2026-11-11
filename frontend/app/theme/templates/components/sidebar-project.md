# sidebar-project

Input scope: `data_project`, one project of the folder tree: a row, a link and
nothing under it.

| Field      | Type   | What it is                                                |
|------------|--------|-----------------------------------------------------------|
| `id`       | string | The folder's ID, which the script reads                   |
| `name`     | string | The last segment of the ID                                |
| `href`     | string | The project's page, without the base                      |
| `title`    | string | What the link says on a mouse-over                        |
| `active`   | bool   | Whether the page is about the project or something in it  |
| `here`     | bool   | Whether the page is about the project itself              |
| `icon`     | string | A sprite symbol, without the `i-` prefix                  |
| `dirty`    | int    | Changed and untracked files under it, 0 for none          |
| `changed`  | string | What the icon says on a mouse-over, "" where `dirty` is 0 |
| `bad`      | bool   | Whether the folder's pipeline would not read              |
| `badge`    | string | What the badge of a bad folder says                       |
| `badTitle` | string | What that badge says on a mouse-over                      |
| `count`    | int    | How many jobs the project's own pipeline declares         |
| `counted`  | bool   | Whether to draw the count at all                          |

A project whose directory declares no job is a row all the same, drawn from
its ID: the sidebar lists the workspaces and their projects, and the jobs of a
project are on its page.
