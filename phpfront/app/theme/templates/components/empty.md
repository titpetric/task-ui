# empty

Input scope: `data_empty`, built by the page.

| Field | Meaning |
|-------|---------|
| `text` | what the place says when it holds nothing |
| `tag` | `li` for a row inside an `ol`, `p` otherwise, the default |
| `href` | an optional place to go, drawn as a link |
| `label` | the link's text, with `href` |

Renders `.empty-state` with `role="status"`, so a reader hears the sentence
once the region it belongs to is reached. A list uses `tag: li` because a
`p` is not a child an `ol` accepts.
