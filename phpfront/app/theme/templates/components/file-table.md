# file-table

Input scope: `data_files`, from `Task\Files::panel()`, `::home()` or `::root()`.

| Field            | Type   | What it is                                           |
|------------------|--------|------------------------------------------------------|
| `up`             | string | The href of the row that goes a level up; empty draws none |
| `rows`           | list   | What the directory holds: its directories, then its files |
| `rows[].name`    | string | The name of the entry                                |
| `rows[].dir`     | bool   | Whether it is a directory                            |
| `rows[].sealed`  | bool   | Whether task-ui lists it without opening it          |
| `rows[].exec`    | bool   | Whether the file runs, by its mode or its shebang    |
| `rows[].href`    | string | What opens it; empty leaves the name plain           |
| `rows[].path`    | string | The path the preview window reads, under the folder  |
| `rows[].size`    | string | How big the file is, worded; empty for a directory   |
| `rows[].stamp`   | string | When it last changed, read                           |
| `rows[].ago`     | string | How long since it changed                            |
| `rows[].menu`    | object | The file's `[...]`, from `Task\Menu::data()`; empty draws none |

The Files panel of a folder's page, of a directory inside one, and of a
workspace's home: `..` up a level, the directories first and linked into, then
the files, each opening in the preview window.

A directory that is a folder of the catalogue links to that folder's page
instead of to the path, which is what `Task\Files` decides: the component
draws the href it was given and works nothing out.

A sealed entry draws plain and locked and carries no link at all, and neither
does an entry the folder refuses a path for, which is a symlink leading out of
it.
