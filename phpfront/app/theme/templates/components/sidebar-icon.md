# sidebar-icon

Input scope: `data_icon`, the icon of a row of the folder tree. Any row scope
of `sidebar-folders` is one: the component reads two of its fields.

| Field     | Type   | What it is                                               |
|-----------|--------|----------------------------------------------------------|
| `icon`    | string | A sprite symbol, without the `i-` prefix                 |
| `dirty`   | int    | Changed and untracked files under the row, 0 for none     |
| `changed` | string | What the tooltip says, "" where `dirty` is 0              |

The component draws no newline of its own: a row writes it inside its link,
between the opening tag and the name.
