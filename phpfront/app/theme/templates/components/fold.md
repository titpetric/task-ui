# fold

Input scope: `data_fold`.

| Field   | Type   | What it is                                      |
|---------|--------|-------------------------------------------------|
| `id`    | string | The body the toggle shows and hides             |
| `label` | string | The heading the toggle reads as                 |

The toggle of a folding block, which sits in the block's heading. The block
takes `data-fold` with the same id and its body takes `class="fold-body"` and
`id`, so the script folds the one from the other and the browser remembers a
folded block per reader.
