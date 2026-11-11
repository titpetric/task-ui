# ws-key-dialogs

Input scope: `data_wsdialogs`, from `Task\Key::wsDialogs()`.

| Field      | Type   | What it is                               |
|------------|--------|------------------------------------------|
| `generate` | object | The generate form, for `ws-key-generate` |
| `upload`   | object | The upload form, for `ws-key-upload`     |

The two dialogs the Upload key and Generate key buttons of a workspace's keys open. Both pages that list a workspace's keys open the same two and return to the page they were opened from, so the pair is one component.

A page draws it only where the reader may edit the workspace: the buttons that open the dialogs are gone otherwise, and a dialog nothing opens is markup nobody reaches.
