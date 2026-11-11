# tri

Input scope: `data_cell`, from `Task\Matrix::cell()`.

| Field    | Type   | What it is                                                   |
|----------|--------|--------------------------------------------------------------|
| `path`   | string | The permission path, which the hidden field is named by       |
| `title`  | string | What the path is called: "Workspaces, Allow Edits"            |
| `effect` | string | `allow`, `deny`, or empty to inherit                          |
| `mark`   | string | The glyph on the button; empty for inherit                    |
| `label`  | string | The hover and the screen reader's text: the title and the effect |
| `locked` | bool   | Whether the button is off, which a matrix nobody may edit sets |
| `field`  | bool   | Whether the hidden field that Save reads is there              |
| `shown`  | bool   | Whether the result mark after the button is drawn              |
| `tone`   | string | `ok` or `bad`, the colour of the result mark                   |
| `says`   | string | `allowed` or `denied`, which the result's hover says           |
| `result` | string | The glyph of the result                                        |

One cell of a [matrix](matrix.md). The button is what a reader clicks, and
task-ui.js cycles it through inherit, allow and deny, turning the hidden field
on as it goes: a Save puts the fields that changed and nothing else.

The mark on the button is the subject's own entry; the mark after it is what
the role's rules and their own decide together. The two differ on every cell
that inherits, which is the whole reason both are drawn.

An `all` rule has no result of its own — it stands for the rules beside it
rather than for an action anyone takes — so `shown` is false on one however
much of the table shows results.
