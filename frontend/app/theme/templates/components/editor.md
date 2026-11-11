# editor

Input scope: `data_markdown`, a field somebody writes markdown in. The scope
is not `data_editor`: that one is the workflow step editor's, and a page may
draw both.

| Field         | Type   | What it is                                                  |
|---------------|--------|-------------------------------------------------------------|
| `name`        | string | The name the form posts it under                            |
| `rows`        | int    | How tall the field is before anything is typed              |
| `label`       | string | The aria-label, which says what is being written            |
| `placeholder` | string | What the empty field says; empty draws no placeholder       |
| `mentions`    | bool   | Whether an `@` opens the list of users                      |
| `toolbar`     | string | `full`, every tool, or `inline`, the five of a short field  |

`data-markdown` is the contract with `task-ui.js`: it mounts mdedit over the
field, and the textarea stays under it holding what the form posts. A reader
whose browser ran no script writes markdown into the textarea itself, which is
what the field was before the editor.

The field is not `required`. A field the editor drew over is hidden, and a
browser refuses to post a form holding a required control nobody can focus;
the script is what refuses an empty one, as it did before.

Where the file a writer drops into the field lands is not here: the body
carries `data-upload`, the directory of the artifact tree the page writes
into, and every field on the page reads it off the body. `Task\Upload` is what
answers it.
