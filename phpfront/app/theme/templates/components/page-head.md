# page-head

Input scope: `data_head`.

| Field     | Type   | What it is                                           |
|-----------|--------|------------------------------------------------------|
| `title`   | string | The heading, which repeats the sidebar row            |
| `lead`    | string | The paragraph under it, as HTML; empty draws none     |
| `actions` | object | The head's own `[...]`, from `Task\Menu::data()`; empty draws none |

Every page opens with one. `lead` is drawn raw so it can hold a `<code>` or a
`<span class="mention">`, and the module writes it, so nothing a reader typed
reaches it.
