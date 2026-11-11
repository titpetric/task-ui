# prompt

Input scope: `data_prompt`, the prompt box of a page.

| Field           | Type   | What it is                                           |
|-----------------|--------|------------------------------------------------------|
| `title`         | string | The heading, `Prompt`, or `Prompts` for a box that only lists |
| `note`          | string | The small sentence after the heading                 |
| `open`          | bool   | Whether the box starts open                          |
| `hidden`        | bool   | Whether the box waits out of sight until a thread comes |
| `watch`         | bool   | Whether the reader only reads, so there is no form   |
| `api`           | string | The route the form posts to, under `/api/`           |
| `fields`        | list   | The hidden fields: `name` and `value` each           |
| `editor`        | object | The field the question is written in, read by `editor` |
| `lead`          | string | The foot up to who the prompt runs as                |
| `tail`          | string | The foot after it                                    |
| `picks`         | list   | The folders the box offers to run in: `value` and `text` each; empty draws no select |
| `after`         | string | The foot after the folder select                     |
| `options`       | object | The users an admin picks from, read by `user-options`; no `users` draws `you` |
| `place`         | string | The folder the prompts run in, as their notes hold it |
| `area`          | string | The page the threads belong to rather than a place; empty draws no `data-area` |
| `threads`       | list   | The threads, newest first, each read by `note`       |

A box with `watch` holds no form: a reader with `prompt/view` and without
`prompt/edit` reads the threads and writes none. A run's box is hidden while
the run goes and it has no thread yet, and the script shows it once one
arrives.

`lead`, `tail` and `after` are the foot around the two picks, so the sentence
of a page reads as one line: the run page's says what Claude gets and what it
does with a fix, and the Workflows page's names the folder Claude runs in
between `tail` and `after`.

A box of an area belongs to a page and not to a place: `area` goes on the
threads div, which is what `task-ui.js` polls the area's threads from, and the
place is whatever the folder select picked.

`data-api`, `data-refresh-notes` and `data-notes` are the contract with
`task-ui.js`: it posts the form over the API, draws the new thread and polls
the threads while a prompt runs.
