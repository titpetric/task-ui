# buttons

Input scope: `data_buttons`, from `Task\Buttons::data()`.

| Field             | Type   | What it is                                       |
|-------------------|--------|--------------------------------------------------|
| `small`           | bool   | Whether every button is small, as in a block head |
| `items`           | list   | The buttons, in the order the page added them    |
| `items[].icon`    | string | A sprite symbol before the label, without the `i-` prefix |
| `items[].label`   | string | What the button says                             |
| `items[].href`    | string | A link; empty makes the item a button            |
| `items[].download`| bool   | Whether the link downloads                       |
| `items[].disabled`| bool   | Whether the button is off                        |
| `items[].note`    | string | The hover of a button that is off                |
| `items[].primary` | bool   | Whether the button is the strong one of the row  |
| `items[].danger`  | bool   | Whether the button is in the bad colour          |
| `items[].attrs`   | string | The `data-*`, `aria-*` and `title` attributes, built escaped |
| `items[].bare`    | bool   | Whether the button is the icon alone, which draws as `btn-icon` |
| `items[].small`   | bool   | Whether this one button is small in a row that is not |
| `items[].faint`   | string | Why the reader gets no button here, drawn in place of one |

This is the buttons of a head and nothing around them: the head writes the
`<div class="actions">` that holds them, because one head holds both a row of
buttons and a `[...]` and they sit in one div together. A page head writes
them full size, `block-head` writes them small, and one button of a full size
row is small on its own where the page says so: a job's page draws its Pin
small beside an Attach that is not.

The items are the same shape as a menu's, so one place validates an attribute
name and escapes its value, and `attrs` is drawn raw.

A button with no `href` does nothing on its own: `task-ui.js` reads the
`data-*` attributes, as `data-attach` and `data-stop` on a run's page.

A `faint` item is a sentence where a button would be: the workflow page tells a
reader who may not run it why, in the place the Run button would stand, so the
row keeps its order and Edit stays to the right of it.

The size comes before the colour, `btn btn-sm btn-primary`, which is the order
the Go partial builds the class in and so the order `scripts/compare.sh`
demands: it forgives whitespace between tags and nothing inside an attribute.
A site that writes the class by hand rather than through the partial writes
`btn btn-primary btn-sm`, and the component that reproduces it writes the same
by hand: the Run button of a job.

A bare button is the icon with no label and none of the button's own chrome,
the one beside the `.tgz` link that copies it. It carries neither `small` nor a
colour, because there is nothing to colour: a `title` and an `aria-label` say
what it does.
