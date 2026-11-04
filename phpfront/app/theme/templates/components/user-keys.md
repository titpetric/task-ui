# user-keys

Input scope: `data_user_keys`, from `Task\Key::ofUser()`.

| Field   | Type   | What it is                                                 |
|---------|--------|------------------------------------------------------------|
| `block` | object | The head, read by [block-head](block-head.md)              |
| `keys`  | list   | The rows, read by [key-table](key-table.md)                |

The Keys block of one user: the instance SSH keys linked to them by name, with
Generate key and Upload key for a reader with Allow Edits on Instance Keys. A
profile draws it and so does the page that edits a user, which is why it is a
component and not markup of either: the two differ only in where a change
returns to, and that is a field of the rows.

The keys of the user's role are not in it. A role's key comes with the role,
and the head says so where the table is empty: unlinking a key from a user
takes away nothing the role grants.

The `id="keys"` is on the section, so `#keys` on either page opens at the
block. One page draws the block once; a second would repeat the id.
