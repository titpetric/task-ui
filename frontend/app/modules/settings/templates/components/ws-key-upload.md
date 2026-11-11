# ws-key-upload

Input scope: `data_wsup`, from `Task\Key::wsUpload()`.

| Field     | Type   | What it is                                            |
|-----------|--------|-------------------------------------------------------|
| `api`     | string | The endpoint the form posts to, at the API's own root |
| `next`    | string | Where the script goes once the key is added           |
| `dialog`  | bool   | The form is in a dialog, which carries its heading    |
| `folder`  | string | The workspace's folder, in the sentence               |
| `dir`     | string | The key folder the key is written to                  |
| `rows`    | int    | How tall the paste box is: shorter inside a dialog    |
| `comment` | string | What the Comment field opens with                     |
| `foot`    | object | The row of Cancel and the submit, for `ws-key-foot`   |

Laid out as `ws-key-generate`, with a file picker that reads into the paste box: the private half is read in the browser and posted as text, so the key arrives through the same endpoint whichever way it was given.
