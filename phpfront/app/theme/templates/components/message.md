# message

Input scope: `data_message`, one message of a thread, from
`Task\Threads\Talk::message()`.

| Field | Meaning |
|-------|---------|
| `id` | the note; the element is `#note-{id}` |
| `mine` | whether the reader said it |
| `agent` | whether the thread's agent wrote it: a run's answer, drawn on the agent's side with an `agent` chip |
| `author` | the user scope to draw |
| `body` | the rendered markdown |
| `audio` | the recording's URL, empty when the message has no recording |
| `state`, `working` | the note's state as a badge, and whether it is still being worked |
| `run`, `short` | the run behind it and the short name to show |
| `work` | runs to draw under the message, each a `data_work` |
| `ago`, `stamp` | the message's time, relative and full |

A run's work sits under the reply it filled, or under the question until one
exists. The component draws one message and nothing about the list it sits in;
the owner supplies the `ol.talk` and the `data-talk` marker.
