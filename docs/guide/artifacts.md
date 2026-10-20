# Artifacts

Files the runs left behind: build output, reports, logs, anything a job
wants kept.

![Artifacts](../assets/artifacts-dark.png)

## Leaving one

Every run gets a directory of its own in `$TASK_UI_ARTIFACTS`. A job that
writes files there leaves artifacts:

```sh
go build -o "$TASK_UI_ARTIFACTS/app" .
cp report.html "$TASK_UI_ARTIFACTS/"
```

When the run ends, task-ui records each file with its size and type.

## Finding one

The Artifacts page lists them all, newest first, with the folder, job and
run they came from, filtered by folder and job. A run's page lists its
own. A name downloads the file. Admins delete them.

---

Previous: [Runs](runs.md) · Next: [Scheduled](scheduled.md) · [Contents](../README.md)
