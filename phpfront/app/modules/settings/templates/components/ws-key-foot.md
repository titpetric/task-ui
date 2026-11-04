# ws-key-foot

Input scope: `data_wsfoot`, the foot of a workspace key's form, from `Task\Key::wsFoot()`.

| Field    | Type   | What it is                                             |
|----------|--------|--------------------------------------------------------|
| `dialog` | bool   | The form sits in a dialog, so Cancel closes it         |
| `cancel` | string | Where Cancel goes on a page of its own, with the mount |
| `label`  | string | What the submit button says                            |

The two forms that add a key to a workspace end in the same row, and each is drawn twice: once on a page of its own and once in the dialog the keys page opens. In the dialog Cancel closes it and goes nowhere, so the link is a button there and an `a` on a page.
