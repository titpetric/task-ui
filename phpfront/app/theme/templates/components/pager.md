# pager

Input scope: `data_pager`, from `Task\Pager::of()`.

| Field  | Type   | What it is                                                  |
|--------|--------|-------------------------------------------------------------|
| `prev` | string | The page of newer rows, without the mount; empty draws none |
| `next` | string | The page of older rows, without the mount; empty draws none |

The pages either side of a page of a list. Both links are written under
`{base}`, so a path is the one the page's own filter names and the mount is
spelled once.

Neither link draws the nav at all, which is a list that fits on one page.
