# Scheduled

Jobs that run on their own, on a cron expression.

![Scheduled](../assets/schedules-dark.png)

## Scheduling a job

A job's page (open it from the sidebar) has a **Scheduled** section:
enter an expression and **Schedule**. The Scheduled page lists every
schedule with its next and last run, and the form there takes a folder
and job too.

An admin picks who a schedule runs as: **Run as** on both forms lists
the users, with the admin first selected. task-ui checks that the picked
user may run the job: they see its folder, hold Run Jobs, and the
job's allow lists name them or their role. If they may not, the form
says so and saves nothing. On the lists, the **Runs as** column shows
each schedule's user with the **⋯** menu; an admin picks another user
there, and the schedule saves on the change. Everyone else schedules
jobs as themselves.

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

It starts in the background as the schedule's user, with their
permissions, and shows up like any other run under their name: on the
dashboard, in Runs and in Recents. When the user may not run the job
at that time, because a permission, a folder grant or an allow list
changed, the scheduler skips the run and logs why. When its time comes while the job still runs,
the schedule skips that time. Office hours do not gate a schedule: the schedule
is the decision.

**Disable** stops a schedule without losing it, **Enable** starts it
again. The schedule's user or an admin changes or deletes it, as does
anyone else who may run its job. When its job leaves the folder, task-ui disables the schedule.

---

Previous: [Artifacts](artifacts.md) · Next: [Notes](notes.md) · [Contents](../README.md)
