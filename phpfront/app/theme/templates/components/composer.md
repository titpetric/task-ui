# composer

Input scope: `data_composer`, from `Task\Composer`.

| Field | Meaning |
|-------|---------|
| `mode` | what Send posts: `ask`, `note` or `chat` |
| `modes` | the modes a reader may switch to, each `key` and `label` |
| `editor` | the markdown field, as `components/editor.tpl` reads it |
| `agents`, `agent` | who may answer, and who does |
| `show_place` | whether the box has a folder to ask in |
| `parent` | the thread a message replies in |
| `channel` | a chat's channel |
| `folder`, `dir`, `run`, `area`, `origin` | where the message belongs |
| `mic` | whether the box has a record button |
| `pane` | whether the box sits in the Threads pane |
| `url` | where an `ask` posts, the prompt endpoint unless the page names one |
| `next` | where a page goes once the message is sent |

`form[data-send]` is the one path a message takes. `task-ui.js` reads the
mode, posts `/api/prompts`, `/api/notes` or `/api/threads`, and goes to
`data-next` with the note the API answers.
