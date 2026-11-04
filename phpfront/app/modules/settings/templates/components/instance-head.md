# instance-head

Input scope: none.

The table four sections of the Instance page draw: the setting, the value it runs with, the environment variable that sets it and what it does. One width per column across the page, so General, Runs, Paths and Integrations line up with each other however long a value is.

It opens the table and leaves the rows to the section, which writes a `<tr>` per setting: a value is a path, a badge, a user, a link or a form, and no table of data draws those four from one shape.

`instance-foot.tpl` closes what this opens.

It is the Instance page's own and not the theme's: the settings of a server are listed on one page, and a second page listing them would be the same page.
