# Inventory

The hosts Beszel monitors, inline in task-ui, the largest first by cores,
then RAM: whether each is up, its icon green when up and red otherwise, who task-ui signs in to it as over ssh, its
cores, how busy its CPU is, and how full its memory and disk are, with what
it uses of each and the size, in GB. A share turns amber past 75% and red
past 90%. A host whose agent sends no core count shows its threads.

The SSH column is the first of `debian@` and `root@` that signs in to the
host with the instance keys, without a password; the button beside it
copies the `ssh` command. task-ui tries every ten minutes, and a host it
reaches slowly fills in on the next view.

## Running a command on the hosts

A user with the `hosts` permission, Run on Hosts, runs a shell command
on the hosts at once, like pssh: pick the
hosts, by default every one with an ssh login (**All**, **None** and
**Invert** set them at once), write the command or take a
quick one (`uptime`, `df -h /`, `free -h`, the containers, the kernel, the
OS), and **Run**, or Ctrl+Enter.

Each host gets a run of its own, a run of `@host`, the root's built-in
job beside `@shell`: it runs the command over the host's login, on a
terminal, for at most two minutes; task-ui kills it past that. Pagers are
off (`PAGER=cat`), so `systemctl status` and the like end on their own.
A run of `@host` is a run like any other: it has a row, a recording, live
output, a place in the recents and on [Runs](runs.md), where
`job=@host` keeps them, and its page, which names it `@host · <host>`.
Its target is the host, its args the command, so **Run again** runs the
same command on the same host. A host without an ssh login gets a run
that fails saying so; task-ui skips a host still running a command of
`@host`. Who ran what on which host is the run history.

The runs started together are a batch. The panel shows your last batch
under the form, a row a host: the host, linked to its run's page, how
the run ended and how long it took, **View** or **Attach**, and what it
printed, both streams, in the colours a program prints on a terminal,
scrolling inside the row. While a run of the batch runs, the panel
follows it on its own; after a reload it shows the batch again. **In
Runs** lists the batch at `/runs?batch=<id>`.

**Run and summarize** runs the command the same way, and after the last
host finishes asks Claude about the batch, as you: a prompt thread about
the batch's first run, "Summarize the runs on these hosts.", which
Claude answers with every host's run and output in one or two short
paragraphs, leading with what needs attention, and a table with a row a
host. The thread shows under the panel and on the page of every run of
the batch, where you can ask about the batch again. The button shows
when the Claude Code integration has a token for you (see
[Settings](settings.md#integrations)).

A [schedule](scheduled.md) or a [workflow](workflows.md) step runs
`@host` too, with a target, a host or `*` for every host with an ssh
login, and the command as its args; `*` starts a batch. A workflow step
on `*` ends ok when every host's run did. Both take the `hosts`
permission of the user they run as.

The logins sign in as `debian` or `root`, so the panel shows only with
`hosts`, which no role but admin holds until
[Settings, Permissions](settings.md#permissions) grants it. The page
itself takes Allow View on Inventory, `inventory/view`, which the runner
and viewer roles hold.

## Prompts on the hosts

A [prompt](projects.md#prompt) on a folder page sees the inventory. Its
run gets the hosts in `TASK_UI_HOSTS`, a line a host in the order of the
table: the name, the host, the user of its SSH column (empty for a host
task-ui does not reach) and `up` or `down`, separated by tabs. Claude
reads them with how to sign in,
`ssh -o BatchMode=yes -o ConnectTimeout=10 -o StrictHostKeyChecking=accept-new user@host 'command'`,
and the commands above as examples. It runs a question about every host
on each in parallel, prefers read-only commands, changes nothing on a
host unless the question calls for a change, and names the hosts it
could not reach. A question for instance:

> list running containers on all hosts, check for any usage of php or
> frankenphp images

The ssh agent of the run holds the keys that sign in: the instance keys
[Settings, Keys](settings.md#keys) grants your role or your name, and the
SSH keys of the workspace its grants let through. The SSH column comes
from every instance key, so a host there may still refuse a prompt whose
run lacks the key that reaches it; grant that key to the admin role, or
to you, in Settings, Keys. When Beszel does not answer, the prompt runs
all the same, and Claude reads that task-ui got no inventory. The
notes agent, `@claude` on the Notes board, gets no inventory.

task-ui reads the hub at most once a minute, and the page fetches the
hosts again each minute on its own, with the time of the fetch under the
table as **Updated**. The page needs the Beszel
integration: its URL and a token (see [Settings](settings.md#integrations)).
Without one, the page says so.

---

Previous: [Dashboard](dashboard.md) · Next: [Workflows](workflows.md) · [Contents](../README.md)
