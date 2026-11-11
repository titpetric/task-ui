# line-chart

Input scope: `data_line_chart`, from `Task\Chart::data()`.

| Field    | Type   | What it is                                                      |
|----------|--------|-----------------------------------------------------------------|
| `title`  | string | The caption over the drawing                                    |
| `svg`    | string | The drawing, written raw                                        |
| `legend` | list   | A series each, in the order the chart was given them            |
| `total`  | object | The stack's own row, under the series; empty draws none         |
| `empty`  | bool   | No series has a record in the span, and the line says so        |

One row of `legend`, and `total` itself:

| Field   | Type   | What it is                                                       |
|---------|--------|------------------------------------------------------------------|
| `name`  | string | The series' name                                                 |
| `color` | string | Which of the six colours draws it, as the stylesheet reads it    |
| `min`   | string | Its least, already worded in the chart's unit                    |
| `avg`   | string | Its mean                                                         |
| `max`   | string | Its most                                                         |

The drawing is a whole `<svg>` element built in PHP, so the component writes it
raw and nothing in the markup knows its geometry. A legend's numbers are already
strings in the chart's unit, so no branch sits inside them.

An empty chart draws its caption and one line of prose: the span is the page's,
not the chart's, so a chart with no record still says which span it covers.
