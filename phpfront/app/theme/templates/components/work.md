# work

Input scope: `data_work`, one run of a thread, from
`Task\Threads\Work::byQuestion()`.

| Field | Meaning |
|-------|---------|
| `id`, `short` | the run, and the start of its ID |
| `stats` | what the run read and wrote, its cost and how many tools it called |

The component draws one faint line under a finished answer. The event cards a
run writes are not drawn as a log; the pinned status of a working thread carries
the state, and `status.md` carries its subagents.
