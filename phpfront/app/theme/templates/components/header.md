# header

Input scope: `data`.

| Field            | Type   | What it is                                      |
|------------------|--------|-------------------------------------------------|
| `base`           | string | The path the application is mounted at          |
| `user.name`      | string | The reader's name                               |
| `user.role`      | string | The role the reader acts with                   |
| `user.initials`  | string | The one or two letters the avatar draws         |
| `settings`       | bool   | Whether any step of the settings opens for them |

The logout form posts to `/logout`, which is the Guard's and not this
application's: the session is minted and read outside the base.

The prompt field, the panel the logo opens and the runs-as picker are the
`prompt/edit` part of the header. They arrive with the `dashboard` module,
which is what owns the prompt.
