# sidebar-folders

Input scope: `data`, the folder tree: the projects at the top of the root
first, then each named workspace in its order with its projects folded under
it.

| Field       | Type   | What it is                                                |
|-------------|--------|-----------------------------------------------------------|
| `title`     | string | The folder's title                                        |
| `href`      | string | The link, without the base                                |
| `icon`      | string | A sprite symbol, without the `i-` prefix                  |
| `workspace` | bool   | Whether the group is a workspace rather than a project    |
| `dirty`     | int    | Changed and untracked files under the checkout, 0 for none |
| `jobs`      | list   | `name` and `href` per job the folder declares             |

It also reads `chrome.can['project/edit']` for the empty line, which is the one
component that reaches outside its own scope: the line it draws is a link to a
page, and whether to offer it is a permission.

The empty line is written out once per branch. MiniTPL emits a newline after
every compiled tag, so an `{if}` directly before punctuation renders a space
in front of the comma; the whole sentence sits inside each branch instead.

The tree, its workspace order, each workspace's sidebar limit and the reader's
pins arrive with the `folders` module, which owns the catalogue. Until then
`data` is empty and the component draws the empty line.
