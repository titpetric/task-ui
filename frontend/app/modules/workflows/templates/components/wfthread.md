# wfthread

Input scope: `data_thread`, from `Task\Pipeline::thread()`.

| Field        | Type   | What it is                                                        |
|--------------|--------|-------------------------------------------------------------------|
| `id`         | string | The workflow run, which the script puts the thread in again by    |
| `workflowId` | string | The workflow it is a run of                                       |
| `page`       | string | The run's own page, without the mount, polled while it goes       |
| `running`    | bool   | Whether the run has not ended                                     |
| `noteState`  | string | `working`, `done` or `failed`, which colours the card             |
| `name`       | string | The workflow's name, which is the author of a reply with no agent |
| `user`       | object | Who started it, read by `user`                                    |
| `ago`        | string | How long since it started                                         |
| `stamp`      | string | When it started, as the hover reads it                            |
| `doneAgo`    | string | How long since it ended                                           |
| `doneStamp`  | string | When it ended, as the hover reads it                              |
| `state`      | string | How it ended, which the dot carries                               |
| `agent`      | object | The agent that answered, read by `user`; empty names the workflow |
| `question`   | string | What it was asked, already markdown; empty says it had no input   |
| `answer`     | string | What came back, already markdown; empty draws the state instead   |
| `waiting`    | string | What the reply says while the run goes                            |
| `ended`      | string | What the reply says for a run that ended with no answer           |

One run of a workflow as a thread: what it was given over what came back.

`question` and `answer` are already HTML and are written raw; everything else is escaped where it is drawn. `data-follow` and `data-page` are the contract with `task-ui.js`, which puts the thread in again off the page once the workflow's topic says the run ended.
