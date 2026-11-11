# menu

Input scope: `data_menu`, from `Task\Menu::data()`.

The scope is named `data_menu` and not `data` because a row draws a menu
inside its own scope: the row keeps `data_row` and hands the menu its own,
so neither overwrites the other.

| Field                | Type   | What it is                                      |
|----------------------|--------|-------------------------------------------------|
| `label`              | string | Names the button for a screen reader and a hover |
| `empty`              | bool   | Whether there is nothing to show, so no button  |
| `details`            | list   | The block on top: `label`, `value`, `mono`, `title` |
| `groups`             | list   | Lists of items, a line between them, danger last |
| `groups[][].icon`    | string | A sprite symbol, without the `i-` prefix        |
| `groups[][].label`   | string | What the entry says                             |
| `groups[][].note`    | string | The faint text after the label                  |
| `groups[][].href`    | string | A link; empty makes the entry a button          |
| `groups[][].download`| bool   | Whether the link downloads                      |
| `groups[][].disabled`| bool   | Whether the button is off                       |
| `groups[][].danger`  | bool   | Whether the entry is in the last group          |
| `groups[][].attrs`   | string | The `data-*`, `aria-*` and `title` attributes, built escaped |

`attrs` is drawn raw. `Task\Menu` is the only thing that writes it and it
escapes every value; a page that builds the string itself is a page that can
break out of the attribute.

task-ui.js opens the menu over the page, moves through it with the arrow keys
and closes it on Escape, a click outside or a pick; on a phone it opens as a
sheet.
