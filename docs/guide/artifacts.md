# Artifacts

The files runs leave behind: build output, reports, coverage, images, anything a job wants kept. task-ui gathers them into one tree with every run's files side by side, and keeps the versions a later run replaced.

![Artifacts](https://cdn.si/titpetric/task-ui/docs/assets/artifacts-dark.png)

## Leaving one

Every run gets an empty directory of its own in `$TASK_UI_ARTIFACTS`. A job that writes files or directories there leaves artifacts:

```sh
go build -o "$TASK_UI_ARTIFACTS/app" .
mkdir -p "$TASK_UI_ARTIFACTS/coverage" && cp coverage.html "$TASK_UI_ARTIFACTS/coverage/"
```

When the run ends, task-ui moves the files into the artifact tree, at the same paths. A job that writes `job1.txt` and another that writes `job2.txt` leave both at the top of the tree. A later run that writes `job1.txt` again replaces it there, and the history keeps the version before. A report a job writes on every run stays current. The run's own directory goes, and nothing carries over into the next run's.

## The page

- **Space**: the bar shows what the tree takes: the latest files, and the history of earlier versions.
- **Files**: the latest version of every file, a directory at a time, as a table of names, sizes and times. A directory opens; a file opens in the preview window: an image scaled to fit, markdown rendered as a page, an `.html` file as the page it is, a sound file in a player, plain text, and for anything else its type and the download. A file whose type is `audio/*` plays in place, `.mp3`, `.wav`, `.ogg`, `.opus`, `.m4a`, `.aac` and `.flac` among them, and seeks by byte range without loading the whole file. An `.html` file renders in a frame of its own, sandboxed and without its scripts, because a job wrote it and the instance did not. The window picks among the versions the runs wrote, newest first, and downloads the one it shows. Ctrl and a click open the file on its own. Each row downloads its file, or its directory as a `.tgz`, and copies the link to it. The button over the table downloads the open directory, or everything, as a `.tgz`, and its link copies the same way: `/artifacts/<directory>.tgz`, and `/artifacts.tgz` for the whole tree.
- **History**: every version a run left, newest first, with the folder, job and run it came from, or the user who wrote it over MCP or the API, filtered by folder and job. A name opens that version in the preview window. A run's page lists its own.

You see the files of the folders you may see, in the table and in a `.tgz`. With Allow Edits on Artifacts, `artifact/edit`, you delete a file or a directory from the tree, with the button on its row: the history keeps its versions, and the preview window still shows them. With Allow Deletes on Artifacts, `artifact/delete`, you delete a record, and **Delete all**, in the **⋯** menu at the top right, deletes every file, its history and every record.

## What a field uploads

Every field somebody writes markdown in - a prompt, a note, a reply, a file of the tree being edited - takes a file by paste, by drop or from the picker, and puts it in the artifact tree. The markdown that draws it lands where the caret was, and the file is an artifact like any other: the history keeps the version it replaced, and a later upload of the same name replaces it again.

Where it lands is the page's own directory and not the top of the tree:

| Written from                       | Where it lands                  |
|------------------------------------|---------------------------------|
| A prompt, a note or a reply        | `prompts/<workspace>-<project>` |
| A page about no place              | `prompts/internal`              |
| A file open in the artifact window | the directory that file is in   |

So a picture pasted into a prompt about `malta/apps` is at `prompts/malta-apps/`, and one dropped into `reports/weekly.md` while it is being edited is at `reports/`. The tree reads as what wrote it rather than as one directory of everything, which is also what a job should do with `$TASK_UI_ARTIFACTS`: write into a directory of its own and the Files table stays readable.

An upload takes Allow Edits on Artifacts, `artifact/edit`, which is the same rule as writing one over the API; a reader without it has no upload tool on the bar.

## Editing a file

A markdown file of the tree opens in the preview window as a page. With Allow Edits on Artifacts, **Edit** draws the editor over it, and **Save** writes the markdown back to the same path, which the history keeps the version before of. Only the latest version opens for editing: an older one is there to be read, and writing it back would be a restore rather than an edit.

## Writing one

An agent outside task-ui writes into the tree too, over MCP with `write_artifact`, or with `PUT /api/artifacts/files/<path>` and the file as the body: a relative path, with no `..` and no dot file, within the size limit the [API](../api.md) names. The file replaces the version at that path, which the history keeps, as a run's file would. Its record has no run, folder or job: the History names the user who wrote it, and whoever sees artifacts sees it. `list_artifacts` lists a directory and `read_artifact` reads a file, the latest version or an older one. Writing takes Allow Edits on Artifacts, `artifact/edit`, which the runner role holds; with it a runner also deletes a file from the tree (see [MCP](mcp.md#tools)).

## History

The tree is a git repository: each run that leaves files is a commit. Once the history holds more commits than `TASK_UI_ARTIFACT_HISTORY` (50 unless set), task-ui squashes it to one commit of the latest files, and the earlier versions and their records go. See [Running the server](../running.md).

**Artifact history** on the [Data retention](settings.md#data-retention) settings page keeps the earlier versions for an age and within a size: `@cleanup` drops the oldest commits of the history, and the versions only they held, until both rules hold. The latest version of every file in the tree stays.

---

Previous: [Runs](runs.md) · Next: [Scheduled](scheduled.md) · [Contents](../README.md)
