# icons

Input scope: none.

There is no icon component. `theme/assets/css/icons.css` draws an icon from
`theme/assets/icons/<name>.svg` as a CSS mask, coloured by `currentColor`, and
the page writes `<i class="ic ic-NAME" aria-hidden="true"></i>`. A name outside
that list draws nothing, so a value out of the database cannot reach the markup
as anything else.

The names: `back`, `bell`, `box`, `chev`, `clock`, `collapse`, `copy`, `down`,
`download`, `edit`, `expand`, `eye`, `file`, `flow`, `folder`, `folder-open`,
`grid`, `grip`, `key`, `list`, `logout`, `magnifier`, `menu`, `mic`, `more`,
`note`, `pin`, `play`, `plug`, `plus`, `project`, `prompt`, `pulse`, `rotate`,
`script`, `secret`, `send`, `server`, `settings`, `stop`, `summary`, `terminal`,
`trash`, `up`, `upload`, `users`, `x`.
