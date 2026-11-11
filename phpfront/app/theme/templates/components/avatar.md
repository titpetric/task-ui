# avatar

Input scope: `data_avatar`, from `Task\Who::avatar()`, and from
`Task\Who::of()`, which answers a superset of it.

| Field      | Type   | What it is                                       |
|------------|--------|--------------------------------------------------|
| `name`     | string | The user's name, which the title says            |
| `role`     | string | The role they act with, which colours the avatar |
| `initials` | string | The one or two letters the avatar draws          |

The letters in a circle that stand for a user. A name no row holds has no
role, and then the avatar draws without a `role-` class at all rather than
with an empty one.

The file ends without a newline, because the component is included in the
middle of a line: an `{include}` is a verbatim copy of the file and a trailing
newline would land inside the markup around it.

`avatar` in `_layout.tpl` is the Go side of this, and `userHTML` in
`task-ui.js` draws it for the rows a script adds, so a change here changes
both.
