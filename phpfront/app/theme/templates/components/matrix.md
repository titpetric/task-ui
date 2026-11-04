# matrix

Input scope: `data_matrix`, from `Task\Matrix::data()`.

| Field         | Type   | What it is                                                |
|---------------|--------|-----------------------------------------------------------|
| `api`         | string | Where Save puts the cells that changed                     |
| `fixed`       | bool   | An admin's policy: every cell allows and none changes      |
| `editable`    | bool   | Whether the cells change and the Save is drawn             |
| `effective`   | bool   | Whether the result mark after each cell is drawn           |
| `rules`       | list   | The column heads: `key` and `title`                        |
| `sections`    | list   | A row each: `title`, and `cells` read by [tri](tri.md)     |
| `permissions` | list   | A row each under Application, a cell read by [tri](tri.md) |
| `columns`     | int    | The whole width: the section column and one per rule       |
| `span`        | int    | The width an application row's one cell covers             |

One subject's policy: the sections of `config/permissions.yml` down the side,
the rules of every section across the top, then the permissions that belong to
no section under an Application heading.

A matrix is drawn for a role, on the permissions page, and for one user, on the
page that edits them. The two differ in where Save puts and whether the result
marks are there, and both are fields of the scope, which is why this is a
component and not markup of either page.

task-ui.js finds the form by `data-matrix` and drives the cells; the Save puts
only the paths whose cell moved. A matrix with no `editable` draws no form foot
and every button in it is off, so a reader who may look and not change sees the
same table without a control that would refuse them.
