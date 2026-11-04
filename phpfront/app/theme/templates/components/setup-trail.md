# setup-trail

Input scope: `data_trail`, from `Task\Setup::trail()`.

| Field              | Type   | What it is                                           |
|--------------------|--------|------------------------------------------------------|
| `previous`         | object | The step before this one; empty draws no link        |
| `next`             | object | The step after it; empty draws no link               |
| `steps`            | list   | Every step the reader may open, in order             |
| `steps[].name`     | string | What the step is called                              |
| `steps[].href`     | string | The page it opens, the prefix in it                  |
| `steps[].current`  | bool   | Whether it is the step being read, which is not a link |

The foot of a settings page: the step before it, the whole trail, and the step
after. It sits at the end of the page's body, where the Go layout draws it, so
each settings template includes it last.

The page decides whether there is a trail at all and includes this only then.
`Task\Setup::trail()` answers an empty array for a page that names no step,
for a step the reader may not open, and for a reader with one step alone:
there is nothing to walk.
