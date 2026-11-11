# pipenode

Input scope: `data_node`, from `Task\Pipeline::nodes()`.

| Field        | Type   | What it is                                                   |
|--------------|--------|--------------------------------------------------------------|
| `pos`        | int    | Which step it is, counted from 0, which the script reads     |
| `then`       | bool   | Whether a step comes before it; read by a chain and not here |
| `state`      | string | How the step ended; empty for a step that has not run        |
| `disabled`   | bool   | Whether runs skip the step                                   |
| `href`       | string | Where the name links, without the mount                      |
| `runId`      | string | The run the step made of a job; empty draws no screen        |
| `workflow`   | bool   | Whether the step runs a workflow, which the icon says        |
| `summary`    | bool   | Whether the step asks for the summary                        |
| `check`      | bool   | Whether the step is the check that sends the steps round     |
| `prompt`     | string | What the step asked Claude; empty draws no paragraph         |
| `text`       | string | The step's name                                              |
| `small`      | string | The line under the name: the state, or which step it is      |
| `expand`     | bool   | Whether the whole output is longer than the screen           |
| `lines`      | int    | The lines the whole output has; 0 names none on the button   |
| `frame`      | string | The screen the run ended on, as SVG, written raw             |
| `live`       | bool   | Whether the screen is polled while the run goes              |
| `subId`      | string | The workflow run the step made; empty draws no inner list    |
| `workflowId` | string | The workflow that run is of                                  |
| `sub`        | list   | The steps of that run, each read by `pipesub`                |

One step of a pipeline. `sub` is the chain of the workflow run the step made, and it is read by `pipesub` and not by this component: MiniTPL's `{include}` is a verbatim replace that cannot nest into itself, so the one depth a pipeline draws is unrolled into the two components.
