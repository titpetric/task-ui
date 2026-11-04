# confirm

Input scope: none.

The two windows that ask before acting, written once in the shell because any
page may ask.

| Window                   | What opens it         | What it asks                               |
|--------------------------|-----------------------|--------------------------------------------|
| `workflow-input-dialog`  | `data-workflow-input` | Optional input for a workflow's prompt     |
| `confirm-dialog`         | `data-confirm`        | Whether to go ahead with what cannot be undone |

The confirm window's text is the opener's, so a row says what it is about to
delete rather than leaving the window to guess. Both are empty markup the
script fills; the contract is docs/framework.md.
