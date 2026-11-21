# work

Input scope: `data_work`, one run of a thread, from
`Task\Threads\Work::byQuestion()`.

| Field | Meaning |
|-------|---------|
| `id`, `short` | the run, and the start of its ID |
| `trace` | how the run spent its time: `ms`, `count`, a state each with `key`, `label`, `ms`, `count`, `share` and `span`, and the `segments` between events |
| `stats` | what the run read and wrote, its cost and how many tools it called |

The component draws one faint line under a finished answer: a stacked bar of
thinking, tool calls and waiting, and the run's statistics under it. The bar
is one line and a button; `threads.js` cycles it between duration and count,
which group the segments by state, and time, which draws each segment with its
own width. Hovering a segment names it and its value.

The event cards a run writes are not drawn as a log; a working answer carries
its state, and `working.md` carries its subagents.
