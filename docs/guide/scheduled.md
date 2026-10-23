# Scheduled

Jobs that run on their own, on a cron expression.

![Scheduled](../assets/schedules-dark.png)

## Scheduling a job

A job's page (open it from the sidebar) has a **Schedules** section:
enter an expression and **Add schedule**. The Scheduled page lists every
schedule with its next and last run, and its **Add schedule** form takes
a folder and job too.

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

**Args** on the Scheduled form go to a job that takes arguments, after
`--`, on every run of the schedule; a job that takes none refuses them.

A schedule runs a command on hosts of the [inventory](inventory.md)
too: `@host:<name>` of the root, `_`, on the host, or `@host:*` on every
host with an ssh login, with the command as its args. With the `hosts`
permission, the job picker of the Scheduled form lists them under
**Hosts**, beside the root `_` in the folder picker, and **Args** takes
the command whole. Through `POST /api/schedules`:

```json
{"folder": "_", "job": "@host:*", "args": ["df -h /"], "expression": "0 6 * * *"}
```

Its user needs the `hosts` permission; `@host:*` starts a batch, a run
a host. The bare `@host` names no host, and the API refuses it.

## Cleanup

`@cleanup` of the root, `_`, is `task-ui cleanup`: it applies the
[Data retention](settings.md#data-retention) rules. It removes the
recording files of finished runs, and only those, the earlier versions
of the artifact tree, never the latest of a file in the tree, and the
reads of the secret read log, the oldest first, and prints, for each
kind, what it removed and what stays. After it removed rows it
vacuums the database and prints its size before and after. With
`--dry-run` as its arguments it prints what would go and changes
nothing.

When task-ui first starts it schedules `@cleanup` `@daily` as the first
admin, once: the schedule is yours to change, disable or delete, and
task-ui adds no other. Admins alone run `@cleanup`, by hand, on a
schedule, or as a workflow step.

## How a scheduled run runs

It starts in the background as the schedule's user, with their
permissions, and shows up like any other run under their name: on the
dashboard, in Runs and in Recents. When the user may not run the job
at that time, because a permission, a folder grant or an allow list
changed, the scheduler skips the run and logs why. When its time comes while the job still runs,
the schedule skips that time. Office hours do not gate a schedule: the schedule
is the decision.

**Disable**, in a schedule's **⋯** menu, stops it without losing it,
**Enable** starts it again. The schedule's user changes or deletes it, as does anyone else
who may run its job: changing takes Allow Edits on Schedules,
`schedule/edit`, deleting Allow Deletes, `schedule/delete`. When its job leaves the folder, task-ui disables the schedule.

---

Previous: [Artifacts](artifacts.md) · Next: [Notes](notes.md) · [Contents](../README.md)
