# space-mini

Input scope: `data_mini`, from `Task\Size::mini()`.

| Field            | Type   | What it is                                           |
|------------------|--------|------------------------------------------------------|
| `title`          | string | The whole measure as one sentence, the cell's hover   |
| `total`          | string | The size of the files and the history, worded        |
| `filesPercent`   | string | The share of the bar the files take, to two decimals  |
| `historyPercent` | string | The share the history takes, to two decimals         |

The space a directory takes in a table cell: a short bar of the files and the
git history behind them, and the total beside it. It is `space` with no room
for a legend, so what the legend would name goes into the title instead.

Every size arrives worded and the two shares arrive as the numbers a width is
written with: the component writes no arithmetic and no format. A row with no
measure behind it holds this scope empty and the cell draws nothing.
