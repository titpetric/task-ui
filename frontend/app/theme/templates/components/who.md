# who

Input scope: `data_who`, from `Task\Who::ofRun()`: who started a run and how.

| Field       | Type   | What it is                                             |
|-------------|--------|--------------------------------------------------------|
| `user`      | object | The user, from `Task\Who::of()`, read by `user`        |
| `scheduled` | bool   | Whether the scheduler started the run                  |
| `reason`    | string | Why the run went outside its window; empty draws none  |
| `forced`    | bool   | Whether the run killed the one before it               |

The badges follow the name in one span, so a table cell and a detail row word
a run's start the same way. A run with a `reason` draws the out of hours badge
and no forced badge: the window is the stronger fact, and the scheduler's own
badge is beside either.
