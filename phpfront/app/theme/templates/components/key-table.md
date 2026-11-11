# key-table

Input scope: `data_keys`, from `Task\Key::rows()`.

A list of [key-row](key-row.md) scopes, in the order the key folder answered
them. No keys draws nothing at all, not an empty table: the page above says
there are none, in its block head.

This is the table of the pages that list keys and nothing else — the Keys page,
Your SSH keys, a profile's Keys block and a workspace's keys. The Instance page
draws its own `table`, because it lists two named keys and puts a placeholder
row with a Generate or a Choose button where one of them is missing; it
includes [key-head](key-head.md) and [key-row](key-row.md) directly.
