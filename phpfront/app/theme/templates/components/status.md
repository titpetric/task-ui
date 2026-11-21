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
| `agents` | the subagents the run started, each with `id`, `task`, `kind`, `prompt`, `state` and `result` |
| `can_end` | whether the reader may stop or kill it |

The script replaces `[data-status-text]` as events arrive, counts
`[data-status-elapsed]` from `data-started`, appends a `[data-agent]` line for
each Task call, and marks its result. A subagent's brief and result sit behind
its summary. Nothing draws when the thread is not working.
