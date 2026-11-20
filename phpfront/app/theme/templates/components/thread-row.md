# thread-row

Input scope: `data_row`, one row of a thread list, from
`Task\Threads\Timeline::row()` or `::chat()`.

| Field | Meaning |
|-------|---------|
| `id` | the root note |
| `href` | the thread page, `/threads/{id}` |
| `kind` | `note`, `prompt`, `voice` or `chat` |
| `said` | the opening line, cut for the row |
| `state` | `open`, `working`, `done`, `failed`, empty for a chat |
| `unread` | a chat's unread count, zero elsewhere |
| `place` | the directory the thread belongs to, empty at the root |
| `messages` | how many messages the thread holds, drawn when more than one, unless the row draws Stop |
| `heard` | whether any message has a recording |
| `can_end` | whether the reader may stop the working thread's run: the row draws a Stop button |
| `ago`, `stamp` | the newest activity, as a relative and a full stamp |
| `person` | the user scope the row draws on the trailing side |
| `selected` | whether the pane shows this thread |

One component serves the Chats and Threads lists; the list owner writes the
`ol` and its class.
