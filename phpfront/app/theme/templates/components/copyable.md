# copyable

Input scope: `data_copy`.

| Field    | Type   | What it is                                                  |
|----------|--------|-------------------------------------------------------------|
| `value`  | string | What the button copies, the whole of it                     |
| `label`  | string | What it is, which names the button: "Copy the public key of git" |
| `shown`  | string | What the code draws; the value where there is nothing shorter to say |
| `fill`   | bool   | Whether it takes the width it has and ends in an ellipsis   |

A value to copy: the code with the copy button right after it. `shown` differs
from `value` where the cell has room for a word and not for the value — a
login beside a host, `ssh root@host` copied and `root@host` drawn — and the
whole value goes in the tooltip.

`fill` is the long value in a column that has to end somewhere: the public key
of a key, which the CSS ellipsises at the cell's width. The tooltip carries the
whole line either way, so the value is always readable without the clipboard.

`label` is the sentence after "Copy", lower case, with its article: the title
and the `aria-label` are the same string and a screen reader reads it as a
command.
