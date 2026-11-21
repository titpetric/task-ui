# work

Input scope: `data_work`, one run of a thread, from
`Task\Threads\Work::byQuestion()`.

| Field | Meaning |
|-------|---------|
| `id`, `short` | the run, and the start of its ID |
| `stats` | what the run read and wrote, its cost and how many tools it called |

The component draws one faint line under a finished answer. The event cards a
run writes are not drawn as a log; a working answer carries its state, and
`working.md` carries its subagents.
