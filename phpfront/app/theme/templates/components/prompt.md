# prompt

Input scope: `data_prompt`, the prompt box of a page, from `Task\Prompt`.

| Field | Meaning |
|-------|---------|
| `title` | the heading, `Prompt` |
| `note` | the sentence after the heading |
| `open` | whether the box starts open |
| `hidden` | whether the box waits out of sight until a thread comes |
| `watch` | whether the reader only reads, so the box holds no form |
| `composer` | the box, as `components/composer.tpl` reads it |
| `threads` | the threads asked here, each `href`, `title` and `state` |
| `threads_href` | the Threads page filtered to this place |

A box with `watch` holds no form: a reader with `prompt/view` and without
`prompt/edit` reads the threads and writes none. A run's box is hidden while
the run goes and it has no thread yet; the script shows it once one arrives.

The page builds the composer with `Task\Composer::prompt()`, which names the
place, the agent and where Send goes. The box posts through `form[data-send]`,
and `data-next` opens the thread the API answers.
