# sidebar-workflows

Input scope: `data_workflows`, the instance's workflows in the order the
workflows page lists them, the ones the reader may not see already dropped.

| Field      | Type   | What it is                                               |
|------------|--------|----------------------------------------------------------|
| `id`       | string | The workflow's ID, which its page and the Run button use |
| `name`     | string | What the row says                                        |
| `hidden`   | bool   | Whether the reader hid the row from the sidebar          |
| `active`   | bool   | Whether the workflow's page is the one being read        |
| `running`  | bool   | Whether a run of it is in progress                       |
| `canRun`   | bool   | Whether the reader may start it                          |
| `prompted` | bool   | Whether a run of it takes input before it starts         |
| `title`    | string | The chain of its steps, the hover over its name          |

`Task\Workflow::sidebar()` builds it from `Task\Workflow::visible()`.

A hidden workflow stays in the list and the component leaves its row out, so
the block draws for a reader with any workflow at all.
