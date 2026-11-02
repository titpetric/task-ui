# user

Input scope: `data_user`, from `Task\Who::of()`.

| Field      | Type   | What it is                                       |
|------------|--------|--------------------------------------------------|
| `name`     | string | The user's name                                  |
| `role`     | string | The role they act with, which colours the avatar |
| `initials` | string | The one or two letters the avatar draws          |
| `menu`     | object | The `[...]` beside them, from `Task\Menu::data()` |

The scope is `data_user` so a row drawing a user keeps its own `data_row`.

`userHTML` in task-ui.js draws the same markup for the rows a script adds, so
a change here changes the script too.
