# block-head

Input scope: `data_block`, the head of a block.

| Field     | Type   | What it is                                               |
|-----------|--------|----------------------------------------------------------|
| `title`   | string | The block's name, in the `h3`                            |
| `note`    | string | The small text after the title; empty draws none         |
| `empty`   | string | The line of a block that lists nothing; empty draws none |
| `buttons` | object | The actions on the right, from `Task\Buttons::data()`    |

A block whose list is empty says so in the head, in the row between the title
and the buttons, and draws no table under it. The page decides that: `empty`
holds the sentence for a block with nothing in it and `""` for one with rows.

`buttons` is always set, with no items where the block has no actions. The
buttons are small here, the size `Task\Buttons` takes in its constructor.
