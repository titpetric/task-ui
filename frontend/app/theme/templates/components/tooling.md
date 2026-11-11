# tooling

Input scope: `data_tooling`, from `Task\Tooling::data()`.

| Field    | Type | What it is                                         |
|----------|------|----------------------------------------------------|
| `scopes` | list | A database atkins records to each, the project's first |

An empty scope draws nothing at all, which is a tree nobody has run a job in
through atkins.

One scope of `scopes`:

| Field        | Type   | What it is                                                     |
|--------------|--------|----------------------------------------------------------------|
| `database`   | string | The file the rows came from, which the line names and titles    |
| `name`       | string | What the scope is called, project or workspace                  |
| `repository` | string | The checkout a project's rows came from; empty draws none       |
| `sum`        | string | The commands of the window and what they took, already worded   |
| `charts`     | list   | One chart per window, in the order the page lays them out       |
| `latest`     | list   | The newest commands, newest first; empty draws no list          |
| `toggle`     | string | What the details summary says                                   |
| `commands`   | list   | A command key each; empty draws no details at all               |

One chart of `charts`:

| Field    | Type   | What it is                                                      |
|----------|--------|-----------------------------------------------------------------|
| `class`  | string | The figure's classes, the span's width among them                |
| `title`  | string | The caption, "Last 24 hours"                                     |
| `stats`  | string | The window's commands, its time and its busiest slot, in one line |
| `bars`   | object | The group of bars                                                |
| `ticks`  | list   | A time marker each: `at`, where it sits in percent, and `label`  |
| `legend` | list   | A name each, in the order the bars of a slot draw them; empty draws one line saying so |

The group of `bars`:

| Field   | Type   | What it is                                                       |
|---------|--------|------------------------------------------------------------------|
| `class` | string | The group's classes, its density among them                      |
| `style` | string | `--n`, how many slots it has, and `--k`, how many bars a slot has |
| `label` | string | What the group says to a reader who hears it                     |
| `slots` | list   | A column each                                                    |

One slot of `slots`:

| Field      | Type   | What it is                                                   |
|------------|--------|--------------------------------------------------------------|
| `class`    | string | The hover's classes, which anchor it inside the chart        |
| `title`    | string | What ran in the column and when                              |
| `segments` | list   | A bar per name of the legend, in its order; empty draws the column alone |

One segment of `segments`:

| Field   | Type   | What it is                                                       |
|---------|--------|------------------------------------------------------------------|
| `class` | string | The bar's classes, its anchor and its colour                     |
| `style` | string | Its height, in percent of the window's tallest                   |
| `title` | string | The name, its time and its runs in the slot; empty draws a bar that says nothing |

One name of `legend`:

| Field   | Type   | What it is                                      |
|---------|--------|-------------------------------------------------|
| `color` | string | Its swatch's colour class                       |
| `name`  | string | The command, or the tool, or `other`            |
| `total` | string | What it took in the window, already worded      |
| `runs`  | string | How many times it ran, already worded           |

One command of `latest`:

| Field     | Type   | What it is                                                 |
|-----------|--------|------------------------------------------------------------|
| `stamp`   | string | When it ran, as every timestamp column reads               |
| `when`    | string | Its clock time, with its weekday before on another day      |
| `color`   | string | The colour its command holds on the charts                  |
| `title`   | string | The command and the job it ran under, for the hover         |
| `command` | string | The command key                                             |
| `job`     | string | The job it ran under; empty draws none                      |
| `took`    | string | How long it took, already worded                            |
| `failed`  | bool   | Whether it ended badly, which draws failed instead of ok    |
| `exit`    | int    | The code it ended on, which the hover names                 |

One command of `commands`:

| Field      | Type   | What it is                                           |
|------------|--------|------------------------------------------------------|
| `command`  | string | The command key                                      |
| `tool`     | string | The tool it runs through                             |
| `runs`     | int    | How many times it ran in the window                  |
| `total`    | string | What it took between them, already worded            |
| `average`  | string | What one run took on average                         |
| `max`      | string | What the slowest run took                            |
| `failures` | int    | How many ended badly; zero draws a plain zero        |
| `last`     | string | How long since it last ran, or `-`                   |

Every number a style or a sentence holds is already a string in the shape the
markup takes, so no branch sits inside an attribute: a stray space in a `style`
breaks the bar it draws.

The hover is the shared `tip` markup, so a bar needs no title attribute of its
own.
