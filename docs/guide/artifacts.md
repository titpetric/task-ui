# Artifacts

The files runs leave behind: build output, reports, coverage, images,
anything a job wants kept. task-ui gathers them into one tree with every
run's files side by side, and keeps the versions a later run replaced.

![Artifacts](../assets/artifacts-dark.png)

## Leaving one

Every run gets an empty directory of its own in `$TASK_UI_ARTIFACTS`. A
job that writes files or directories there leaves artifacts:

```sh
go build -o "$TASK_UI_ARTIFACTS/app" .
mkdir -p "$TASK_UI_ARTIFACTS/coverage" && cp coverage.html "$TASK_UI_ARTIFACTS/coverage/"
```

When the run ends, task-ui moves the files into the artifact tree, at the
same paths. A job that writes `job1.txt` and another that writes
`job2.txt` leave both at the top of the tree. A later run that writes
`job1.txt` again replaces it there, and the history keeps the version
before, so a report a job writes on every run stays current. The run's
own directory goes, and nothing carries over into the next run's.

## The page

- **Space**: the bar shows what the tree takes: the latest files, and the
  history of earlier versions.
- **Files**: the latest version of every file, a directory at a time, as
  a table of names, sizes and times. A directory opens; a file opens in
  the browser, and an image or an svg opens in the image window, scaled to
  fit. Each row downloads its file, or its directory as a `.tgz`, and
  copies the link to it. The button over the table downloads the open
  directory, or everything, as a `.tgz`, and its link copies the same way:
  `/artifacts/<directory>.tgz`, and `/artifacts.tgz` for the whole tree.
- **History**: every version a run left, newest first, with the folder,
  job and run it came from, filtered by folder and job. A name downloads
  that version. A run's page lists its own.

You see the files of the folders you may see, in the table and in a
`.tgz`. Admins delete a record, and **Delete all**, at the top right,
deletes every file, its history and every record.

## History

The tree is a git repository: each run that leaves files is a commit.
Once the history holds more commits than `TASK_UI_ARTIFACT_HISTORY` (50
unless set), task-ui squashes it to one commit of the latest files, and
the earlier versions and their records go. See
[Running the server](../running.md).

---

Previous: [Runs](runs.md) · Next: [Scheduled](scheduled.md) · [Contents](../README.md)
