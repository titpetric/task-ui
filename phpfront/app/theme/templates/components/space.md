# space

Input scope: `data_space`, the space a directory takes.

| Field            | Type   | What it is                                           |
|------------------|--------|------------------------------------------------------|
| `total`          | string | The size of the files and the history, worded        |
| `files`          | string | The size of the files, worded                        |
| `history`        | string | The size of the history, worded                      |
| `ignored`        | string | The size git ignores, worded, empty where it ignores nothing |
| `filesPercent`   | string | The share of the bar the files take, to two decimals  |
| `historyPercent` | string | The share the history takes, to two decimals         |

One bar of the files of a directory and the git history beside them, with a
legend naming each. The total leaves out what git ignores and the last line of
the legend names it apart, because an ignored file is on the disk and not in
the directory's record.

Every size arrives worded, from `Task\Size::bytes()`, and the two shares arrive
as the numbers a width is written with: the component writes no arithmetic and
no format. A page with no meter behind it draws no bar at all and holds this
scope empty.
