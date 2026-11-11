# user

Input scope: `data_user`, from `Task\Who::of()`.

| Field      | Type   | What it is                                       |
|------------|--------|--------------------------------------------------|
| `name`     | string | The user's name                                  |
| `role`     | string | The role they act with, which colours the avatar |
| `initials` | string | The one or two letters the avatar draws          |
| `href`     | string | Their profile, and `""` for a name the reader may not open |

The scope is `data_user` so a row drawing a user keeps its own `data_row`.

A user reference is a link to the profile and nothing else. Every action on a
user is on that page: Edit as its primary button, the runs by them and the
delete in its `[...]`.

The avatar is `avatar.tpl`, which this scope is a superset of, so the place
that draws a user and the place that draws an avatar alone draw one thing.

`userHTML` in task-ui.js draws the same markup for the rows a script adds, so
a change here changes the script too.
