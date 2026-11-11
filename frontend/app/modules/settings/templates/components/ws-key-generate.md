# ws-key-generate

Input scope: `data_wsgen`, from `Task\Key::wsGenerate()`.

| Field     | Type   | What it is                                            |
|-----------|--------|-------------------------------------------------------|
| `api`     | string | The endpoint the form posts to, at the API's own root |
| `next`    | string | Where the script goes once the key is added           |
| `dialog`  | bool   | The form is in a dialog, which carries its heading    |
| `folder`  | string | The workspace's folder, in the sentence and the name  |
| `dir`     | string | The key folder the key is written to                  |
| `comment` | string | What the Comment field opens with                     |
| `foot`    | object | The row of Cancel and the submit, for `ws-key-foot`   |

Whose runs get the key is no field: it is a key of this workspace, so every run of this workspace gets it and there is nothing left to choose.

The heading and the sentence are the dialog's. A page of its own writes them in its own head, so they are not repeated here.
