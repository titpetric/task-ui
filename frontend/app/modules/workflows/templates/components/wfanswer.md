# wfanswer

Input scope: `data_wfanswer`, the row under a workflow run holding its answer.

| Field  | Type   | What it is                                                      |
|--------|--------|-----------------------------------------------------------------|
| `id`   | string | The workflow run, which the chevron's `aria-controls` names     |
| `body` | string | Claude's answer as rendered HTML, drawn raw; empty draws no row |
| `cols` | int    | How many columns the row spans, which is the table's width      |
| `open` | bool   | Whether the row starts open; false writes `hidden`              |

The answer goes through `Task\Chrome::message()`, which renders the markdown and links the mentions the reader may open, so `body` is already HTML and the component draws it raw.

The row pairs with `wffold`: the same `id` ties the chevron to the row and both read the same `open`.
