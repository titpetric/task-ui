# sidebar-folders

Input scope: `data`, the folder tree, which `Task\Chrome::folderTree()`
builds: the projects at the top of the root first, in the catalogue's order,
then each named workspace in its order with its own projects folded under it.

Each entry is a group, which is either a project or a workspace:

| Field       | Type | What it is                                                |
|-------------|------|-----------------------------------------------------------|
| `workspace` | bool | Whether the group is a named workspace rather than a project |

A workspace is a `sidebar-workspace` scope and a project a `sidebar-project`
one; both of those schemas are in their own files.

It also reads `chrome.can['project/edit']` for the empty line, which is the
one place this component reaches outside its own scope: the line it draws is a
link to a page, and whether to offer it is a permission.

The empty line is one line of markup, text and tags together, because
`scripts/compare.sh` compares a tag and the text beside it: a sentence split
across lines is a changed line even where the rendered page reads the same.
