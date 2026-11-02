# state

Input scope: `data_state`, from `Task\Run::label()`.

| Field    | Type   | What it is                                               |
|----------|--------|----------------------------------------------------------|
| `id`     | string | The run to follow, a 26 character identifier; empty draws nothing |
| `state`  | string | The state, which colours the dot and names the label     |
| `ended`  | bool   | Whether the run is over                                  |
| `code`   | string | The exit code of a run that ended badly; empty draws none |

The contract is [the shared markup](../../../../../docs/framework.md): the
script fills a `span[data-run-id]` and keeps it current over the event stream
every page shares, and `data-finished` on a run that was already over means it
is never hydrated and never subscribed.

**Write `ended` when you know.** A page listing a hundred runs that says so
for the ninety-nine that finished asks the server once instead of a hundred
times.
