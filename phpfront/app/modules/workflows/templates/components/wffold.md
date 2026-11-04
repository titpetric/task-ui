# wffold

Input scope: `data_wffold`, the chevron in front of a workflow run's name.

| Field    | Type   | What it is                                                        |
|----------|--------|-------------------------------------------------------------------|
| `id`     | string | The workflow run, which the chevron's `aria-controls` names       |
| `answer` | bool   | Whether the run has an answer at all; false draws the empty place |
| `open`   | bool   | Whether the answer row under it starts open                       |

A run without an answer keeps the place empty rather than drawing nothing, so the names of the rows line up; its state is on the dot beside it.

`data-answer-toggle` is the contract with `task-ui.js`: it folds the row `aria-controls` names and turns `aria-expanded` over.
