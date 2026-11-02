# icons

Input scope: none.

The sprite, once per page. Every icon on the page is a `<use href="#i-NAME"/>`
into it, and a sidebar row's icon column names one without the `i-` prefix.

The symbols: `project`, `folder`, `folder-open`, `grid`, `list`, `box`, `flow`,
`prompt`, `summary`, `grip`, `plus`, `server`, `plug`, `note`, `clock`,
`pulse`, `users`, `settings`, `logout`, `copy`, `rotate`, `trash`, `edit`,
`key`, `secret`, `file`, `script`, `terminal`, `play`, `stop`, `pin`, `eye`,
`menu`, `expand`, `collapse`, `more`, `chev`, `up`, `down`, `download`,
`upload`.

`Task\Sidebar::icon()` falls back to `box` for a name the sprite would not
accept, so a value out of the database cannot reach the markup as anything
else.
