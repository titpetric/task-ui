# meter

Input scope: `data_meter`, from `Task\Meter::of()`.

| Field   | Type   | What it is                                                     |
|---------|--------|----------------------------------------------------------------|
| `share` | float  | The share itself, 0 to 100, which decides the tone             |
| `width` | string | The same share to a tenth, the width of the bar in percent     |
| `pct`   | string | The same share whole, the number beside the bar                |

The three are one number three ways: the tone is a comparison, the width is a
style and the number is what a reader reads, and each wants its own rounding.

A share of 90 or more draws `high`, one of 75 or more `warm`, and the rest the
plain tone the stylesheet gives the bar.
