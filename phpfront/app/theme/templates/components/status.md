# status

Input scope: `data_status`, the pinned line of a working conversation, from
`Task\Threads\Conversation::status()`.

| Field | Meaning |
|-------|---------|
| `working` | whether the thread has a run going |
| `label` | what Claude reports doing, such as `Running Bash` |
| `state` | `starting`, `thinking`, `writing`, `tool`, `result`, `done` |
| `started` | when the run started, in milliseconds, for the elapsed time |
| `run` | the run to stop or kill |
| `can_end` | whether the reader may stop or kill it |

The script replaces `[data-status-text]` as events arrive and counts
`[data-status-elapsed]` from `data-started`. Nothing draws when the thread is
not working.
