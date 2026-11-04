# git-line

Input scope: `data_git`, from `Task\Place::git()`.

| Field              | Type   | What it is                                           |
|--------------------|--------|------------------------------------------------------|
| `title`            | string | What the whole line says on a mouse-over: the checkout, and what it tracks |
| `branch`           | string | The branch the checkout is on; empty means detached  |
| `detached`         | string | The commit it is detached at, for a checkout on no branch |
| `upstream`         | string | The branch it tracks; empty means it tracks none     |
| `ahead`            | int    | How many commits there are to push                   |
| `behind`           | int    | How many commits there are to pull                   |
| `dirty`            | int    | How many files are changed or untracked              |
| `files`            | string | Those files worded, which the button says            |
| `clean`            | bool   | Whether it tracks a branch and has nothing to push, pull or commit |
| `where`            | string | The checkout under the root, or the root itself      |
| `changes`          | list   | The changed files, drawn under the button            |
| `changes[].code`   | string | The two letters git reports the change as            |
| `changes[].state`  | string | Those letters worded                                 |
| `changes[].path`   | string | The file, under the checkout                         |
| `changes[].from`   | string | Where a rename came from; empty for every other change |

The git line of a folder's page and of a directory browsed inside one: the
checkout the place is in, its branch, what it has to push and pull and what
changed. The count of the changed files is a button that shows and hides the
table of them under it.

A place in no checkout draws no line at all, which is the null
`Task\Place::git()` answers: the page tests it and includes nothing.
