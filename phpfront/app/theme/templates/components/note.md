# note

Input scope: `data_note`, one note with its replies.

| Field         | Type   | What it is                                              |
|---------------|--------|---------------------------------------------------------|
| `id`          | string | The note, which the anchor and the reply form name      |
| `body`        | string | The note as HTML, from `Task\Chrome::message()`          |
| `raw`         | string | The note as it was written, for the field an edit opens  |
| `author`      | object | Who wrote it, from `Task\Who::of()`                     |
| `ago`         | string | How long since it was written                            |
| `stamp`       | string | When it was written, which the hover says               |
| `edited`      | bool   | Whether it has been rewritten                            |
| `editedStamp` | string | When it was rewritten, which that hover says            |
| `edits`       | bool   | Whether this reader may rewrite it, by clicking the text |
| `state`       | string | An instruction's state; empty for a plain note          |
| `working`     | bool   | Whether that state is working, which a reply draws      |
| `run`         | string | The run the agent worked in; empty for none             |
| `short`       | string | The first eight characters of the run                   |
| `menu`        | object | The row's `[...]` menu, from `Task\Menu::data()`         |
| `replies`     | list   | The replies, oldest first, each the same less `replies`  |
| `editor`      | object | The Reply field, read by `editor`                        |

`Task\Note::card()` builds every field of it and `Task\Note::thread()` adds
`replies`, so the board and a run's prompt box draw one markup over one scope.

Every note carries the Reply field and no permission hides it, which is what
`_thread.tpl` does in Go: `note/edit` is the API's and the Edit item's, so a
reader without it is refused on the post and not before it.

`body` is drawn raw: it is markdown through the renderer that owns the content
security policy, which skips raw HTML and keeps a link to a scheme it knows.
`Task\Chrome::message()` is what marks a `@name` and links the ones whose
profile the reader may open.

The reply form carries `data-note-form` and `data-parent`, and no action: the
script posts it to `/api/notes` and refreshes the list in place. `edits` is
what opens the same form over the text, which is why `raw` is in the scope
beside `body`.

The scopes are `data_note`, `data_user` for the author, `data_menu` for a menu
and `data_markdown` for the Reply field, so a note, the user it names, either
menu and the field do not overwrite each other.
