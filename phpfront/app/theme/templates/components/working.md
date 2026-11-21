# working

Input scope: `data_working`, the live answer of a run still going, from
`Task\Threads\Talk::live()`.

| Field | Meaning |
|-------|---------|
| `label` | what Claude reports doing, such as `Running Bash` |
| `state` | `starting`, `thinking`, `writing`, `tool`, `result`, `done` |
| `started` | when the run started, in milliseconds, for the elapsed time |
| `run` | the run to stop or kill |
| `canEnd` | whether the reader may stop or kill it |
| `agents` | the subagents the run started, each with `id`, `task`, `kind`, `prompt`, `state`, `result`, `className` and `badge` |
| `todos` | the checklist the run last wrote, each with `text`, `state` and the `className` of its dot |
| `progress` | the last text Claude printed, drawn in place of the one before |
| `progressHtml` | that text as a markdown block quote, which the template draws |
| `progressClass` | the class of the text block, faint while it is empty |
| `stats` | the context line: tools, tokens and cost |
| `tools` | how many tools the run has called |

The block draws inside the answer it is writing, not above the conversation.
The script replaces `[data-status-text]` as events arrive, counts
`[data-status-elapsed]` from `data-started`, replaces `[data-work-progress]`
with the newest text, replaces `[data-status-todos]` from each TodoWrite call,
adds a `[data-agent]` line for each Task call, and marks its result.
