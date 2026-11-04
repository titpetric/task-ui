# instance-foot

Input scope: none.

Closes what `instance-head.tpl` opens: the body of rows, the table and the wrap.

The pair is one table split in two because a section writes its own `<tr>` per setting, and the four sections of the Instance page — General, Runs, Paths and Integrations — close it the same way. Written out at each of them it was four copies of three tags, and a column added to the head is a column the foot never sees.
