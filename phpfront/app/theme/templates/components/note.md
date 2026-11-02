# note

Input scope: `data_note`, one note with its replies.

| Field                | Type   | What it is                                      |
|----------------------|--------|-------------------------------------------------|
| `id`                 | string | The note, which the anchor and the reply form name |
| `body`               | string | The note as HTML, from `Task\Chrome::message()` |
| `author`             | object | Who wrote it, from `Task\Who::of()`             |
| `ago`                | string | How long since it was written                   |
| `edited`             | bool   | Whether it has been rewritten                   |
| `state`              | string | An instruction's state; empty for a plain note  |
| `run`                | string | The run the agent worked in; empty for none     |
| `short`              | string | The first eight characters of the run           |
| `replyable`          | bool   | Whether the reader may reply                    |
| `menu`               | object | The row's `[...]` menu, from `Task\Menu::data()` |
| `replies`            | list   | The replies, oldest first, each the same less `replies` |

`body` is drawn raw: it is markdown through the renderer that owns the content
security policy, which skips raw HTML and keeps a link to a scheme it knows.
`Task\Chrome::message()` is what marks a `@name` and links the ones whose
profile the reader may open.

The scopes are `data_note`, `data_user` for the author and `data_menu` for a
menu, so a note, the user it names and either menu do not overwrite each
other.
