# conversation-head

Input scope: `data_head`, the title line of one conversation, from
`Task\Threads\Conversation::view()`.

| Field | Meaning |
|-------|---------|
| `id` | the root note, which the agent picker names |
| `title` | the thread's opening line |
| `meta` | what it is, where it lives and its state |
| `agents` | the names a reply may be posted as |
| `agent` | the one the thread uses |
| `can_pick` | whether the reader may change it now |
| `menu` | the thread's actions, as `Task\Menu::data()` |

This is the page's only `h2` when a conversation is open; the Threads page
draws its own heading in the blank pane while nothing is selected.
