# project-table

Input scope: `data_projects`, from `Task\Project::table()`.

| Field   | Type | What it is                                                  |
|---------|------|-------------------------------------------------------------|
| `sized` | bool | Whether the page measured the directories, which draws Size |
| `rows`  | list | One project per row, in the order the host walked them      |

Each row:

| Field      | Type   | What it is                                                 |
|------------|--------|------------------------------------------------------------|
| `id`       | string | The folder ID, which the cell links                        |
| `href`     | string | The page of the folder, mounted                            |
| `bad`      | bool   | Whether a source of it does not parse                      |
| `sources`  | string | The files its conventions were read from, comma separated  |
| `jobs`     | int    | How many jobs it declares                                  |
| `checkout` | bool   | Whether the directory is a git checkout                    |
| `remote`   | string | The remote it was cloned from                              |
| `branch`   | string | The branch it is on                                        |
| `commit`   | string | The last commit, short                                     |
| `subject`  | string | That commit's subject, the cell's hover                    |
| `ago`      | string | How long ago it was made, "" for a checkout with no commit |
| `space`    | scope  | `components/space-mini.tpl`, empty where nothing measured  |
| `menu`     | scope  | `components/menu.tpl`, the row's actions                   |

The table of the Projects page and of a workspace's page, the same rows on both: a block draws it under its head, and a block with no project draws the head's empty line instead, so the component is never asked for an empty table.

The Size column is the page's own decision and not the row's: `sized` draws the header, and a row with no measure behind it holds an empty `space` and draws an empty cell under it. An instance with no meter measures nothing, and then the column is not there at all.
