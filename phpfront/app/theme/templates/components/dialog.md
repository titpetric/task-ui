# dialog

Input scope: none.

The five windows a page opens over itself, written once in the shell because
any page may open any of them. Each is empty markup: the script fills it from
the API and shows it, and a reader with no JavaScript never sees one.

| Window            | What opens it         | What it shows                                      |
|-------------------|-----------------------|----------------------------------------------------|
| `run-dialog`      | `data-run-id`         | One run: its terminal, its state and what may act on it |
| `preview-dialog`  | `data-file`           | An image, an svg, an HTML page or a document of the artifacts |
| `pipeline-dialog` | `data-pipeline`       | One workflow run's pipeline, fetched from its page |
| `summary-dialog`  | `data-summary`        | Claude's summary of a run or a workflow run        |
| `source-dialog`   | `data-source`         | A file a folder's pipeline reads, coloured         |

The preview window's Edit and Save are hidden until a markdown file of the
artifact tree is open at its latest version and the body says the reader may
write one, `data-artifact-edit`. Edit draws the editor over the rendered
document and Save puts the markdown back at the same path.

A file of a folder's checkout opens the same way where the body says
`data-file-edit`: Edit draws the editor over the file, Cancel puts the
preview back, and Save writes the file and commits it with the checkout's
git user. Markdown keeps the rich editor; every other kind opens as source,
in the code face and coloured by the extension. The saved window offers Push,
which sends the checkout to its remote with the instance key.

The window offers no Download. A file is read in task-ui, which is the
environment for it, and the transfers that remain are a directory's `.tgz` on
the artifacts page, a workspace's export, and a key's public half.

The run window's foot is every action a run takes, each hidden until the state
says it applies: Attach, Run again, Take input, Stop, Kill, Pause, Restart and
Run force. `run-reason` asks why a job runs outside its hours, and the instance
saves the answer with the run. The contract between an attribute and the script
is docs/framework.md.
