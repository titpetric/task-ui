# message

Input scope: `data_message`, one message of a thread, from
`Task\Threads\Talk::message()`.

| Field | Meaning |
|-------|---------|
| `id` | the note; the element is `#note-{id}` |
| `mine` | whether the reader said it |
| `author` | the user scope to draw |
| `body` | the rendered markdown |
| `audio` | the recording's URL, empty when the message has no recording |
| `state`, `working` | whether the message is still being worked |
| `run`, `short` | the run behind it and the short name to show |
| `work` | runs to draw under the message, each a `data_work` |
| `ago`, `stamp` | the message's time, relative and full |

The component draws one message and nothing about the list it sits in; the
owner supplies the `ol.talk` and the `data-talk` marker.
