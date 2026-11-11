# key-head

Input scope: none.

The columns of a keys table, the same wherever keys list: the Keys page, the
Instance page, a workspace's page, Your SSH keys and a profile. Name takes the
width of the names, Algorithm and Size as little as they need, Public key the
rest with its copy button, and the row's `[...]` last.

Algorithm and Size carry `hide-narrow`: on a phone the table draws as cards and
those two go, because the name already ends in both.

It is the head alone. The `table` and the `table-wrap` around it belong to the
page, because the Instance page lists two keys of five columns and adds a
placeholder row where one of them is missing, which no table of rows can draw.
[key-table](key-table.md) is the table for the pages that list keys and nothing
else.
