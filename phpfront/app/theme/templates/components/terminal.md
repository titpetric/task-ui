# terminal

Input scope: `data_terminal`, the live screen of a run.

| Field      | Type   | What it is                                             |
|------------|--------|--------------------------------------------------------|
| `id`       | string | The run to attach to, which the frame follows          |
| `name`     | string | The run's title, for the image's label                 |
| `watching` | int    | How many windows are on the session; 0 says nobody is  |
| `frame`    | string | The last screen as an SVG, from `Task\Host::frame()`   |

The block belongs to a run the runner still holds: `Task\Host::session()` says
`live` and the state is neither of the ended ones. `frame` is drawn raw, which
is the one place a page embeds markup from the host; the renderer writes the
SVG of a terminal screen and nothing else.

`data-live` and `data-frame` are the contract with `task-ui.js`: it replaces
the frame as the run prints, and a click anywhere on the tile attaches.
