# work

Input scope: `data_work`, one run of a thread, from
`Task\Threads\Work::byQuestion()`.

| Field | Meaning |
|-------|---------|
| `id`, `short` | the run and the short name |
| `stats` | what the run read and wrote, or empty |
| `status` | the last state reported: `body`, `state`, `tool` |
| `events` | the run's cards, each with `seq`, `kind`, `body`, `prose`, `tool`, `input`, `error` |

Statuses are collapsed into `status`, and results are dropped because the
answer says them: the card flow holds `text`, `thinking`, `tool_use`,
`tool_result`, `system` and `error`. Every item writes `data-seq`, which the
script matches when the pane is read again.
