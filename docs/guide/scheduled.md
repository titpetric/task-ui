# Scheduled

Jobs that run on their own, on a cron expression.

![Scheduled](../assets/schedules-dark.png)

## Scheduling a job

A job's page (open it from the sidebar) has a **Scheduled** section:
enter an expression and **Schedule**. The Scheduled page lists every
schedule with its next and last run, and the form there takes a folder
and job too.

An expression has five fields, minute, hour, day of month, month and
day of week, in the server's time zone:

| Expression       | Runs                              |
|------------------|-----------------------------------|
| `0 3 * * *`      | every day at three in the morning |
| `*/15 * * * 1-5` | every fifteen minutes on weekdays |
| `30 6 1 * *`     | at 6:30 on the first of the month |
| `@hourly`, `@daily`, `@weekly`, `@monthly` | as they say |

Month and day names work too (`mon-fri`, `jan`).

## How a scheduled run runs

It starts in the background under the name of whoever scheduled it, with
their permissions, and shows up like any other run: on the dashboard, in
Runs and in Recents. When its time comes while the job still runs,
the schedule skips that time. Office hours do not gate a schedule: the schedule
is the decision.

**Disable** stops a schedule without losing it, **Enable** starts it
again. When its job leaves the folder, task-ui disables the schedule.

---

Previous: [Artifacts](artifacts.md) · Next: [Notes and Chat](notes.md) · [Contents](../README.md)
