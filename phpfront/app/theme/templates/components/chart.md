# chart

Input scope: `data_chart`, from `Task\Activity::data()`.

| Field       | Type   | What it is                                                     |
|-------------|--------|----------------------------------------------------------------|
| `sum`       | string | The runs of the window and how many days it covers             |
| `failed`    | string | The failed runs of the window; empty draws none                |
| `median`    | string | The middle duration, already worded; empty draws none          |
| `charts`    | list   | A count chart each, in the order the page lays them out        |
| `durations` | object | The duration line of the latest runs; empty draws none         |

An empty `charts` draws nothing at all, which is the window with no run in it.

One chart of `charts`:

| Field   | Type   | What it is                                                       |
|---------|--------|------------------------------------------------------------------|
| `class` | string | The figure's classes, the span's own among them                  |
| `title` | string | The caption, "Last 24 hours"                                     |
| `label` | string | What the bars say to a reader who hears them                     |
| `n`     | string | How many columns the chart has, which the stylesheet reads as `--n` |
| `stats` | object | `text`, the finished runs in one line, and `title`, its hover; empty text draws none |
| `bars`  | list   | A column each                                                    |
| `lines` | object | `max`, `min` and `avg` as SVG paths, with `label`; empty draws no line |
| `ticks` | list   | A time marker each: `at`, where it sits in percent, and `label`, which may be empty |

One bar of `bars`:

| Field    | Type   | What it is                                                      |
|----------|--------|-----------------------------------------------------------------|
| `class`  | string | The hover's classes, which anchor it inside the chart           |
| `tip`    | string | What the column holds and when                                  |
| `hit`    | string | The same column's durations, for the hover under the lines      |
| `ok`     | string | The height of the part that ended ok; empty draws none          |
| `failed` | string | The height of the part that failed or ended killed              |
| `other`  | string | The height of the part still running or stopped                 |

The duration line of `durations`:

| Field     | Type   | What it is                                                     |
|-----------|--------|----------------------------------------------------------------|
| `label`   | string | What the line says to a reader who hears it                    |
| `caption` | string | The caption under it                                           |
| `line`    | string | The durations as an SVG path                                   |
| `area`    | string | The plot under the line, as the wash                           |
| `longest` | string | The slowest run of the line, already worded                    |
| `points`  | list   | A run each: `class`, `style`, `at`, `title` and `failed`       |

Every number is already a string in the shape the markup takes, a height or a
coordinate of the 100 by 100 plot, so no branch sits inside an attribute: a
stray space in a `style` breaks the bar it draws.

The hover is the shared `tip` markup, so a bar needs no title attribute of its
own.
