# key-row

Input scope: `data_key`, from `Task\Key::row()`.

| Field      | Type   | What it is                                                |
|------------|--------|-----------------------------------------------------------|
| `name`     | string | The key's name, which ends in its algorithm and size      |
| `href`     | string | The page that sets whose runs get it; empty draws the name plain |
| `encrypts` | bool   | Whether it is the encryption key, which the badge says    |
| `backup`   | string | The backup a rotation left serving beside it; empty draws none |
| `comment`  | string | The comment in the key file, under the name               |
| `algo`     | string | The algorithm                                             |
| `size`     | int    | The size in bits                                          |
| `copy`     | object | The public key, read by [copyable](copyable.md)           |
| `menu`     | object | The row's `[...]`, from `Task\Menu::data()`               |

One key of a keys table, five cells wide, under [key-head](key-head.md).

`href` is empty on three keys and not one: the encryption key, which is handed
to nobody and so links nobody; a key of a workspace, which every run of that
workspace gets whatever the link rows say; and every key at all to a reader
without `key/edit`. The name is the same either way, in `mono`, so the column
reads alike across readers.

BACKUP ACTIVE is a rotation that kept the previous material. It is a state and
not an action: the old key still opens what the new one has not re-encrypted
and still signs in to hosts that know only its public half, which is why the
`[...]` has Rotate off and Delete backup in its place.

The `id` is `key-<name>`, which is where a rotation's redirect lands.
