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

An admin runs a shell command on the hosts at once, like pssh: pick the
hosts, by default every one with an ssh login (**All**, **None** and
**Invert** set them at once), write the command or take a
quick one (`uptime`, `df -h /`, `free -h`, the containers, the kernel, the
OS), and **Run**, or Ctrl+Enter. Each host runs it over its login, on a
terminal, for at most two minutes, and the page shows a row a host: its
exit code, how long it took, and what it printed, both streams, with the
colours a program prints on a terminal. A long output shows its last screen,
the final 24 lines; **Expand** shows the whole output, as on a job run's steps. Pagers are off (`PAGER=cat`), so
`systemctl status` and the like end on their own.

**Run and summarize** runs the command the same way, and after the last
host finishes, sends Claude every output as one markdown report: a
section a host with its login, exit code and output, without colours.
Claude's summary, one or two short paragraphs and a table with a row a
host when that helps, leads with what needs attention, and shows above the
rows, with the report to download as `.md`. The button shows when the
Claude Code integration has a token for you (see
[Settings](settings.md#integrations)); Claude only reads the report and
runs no tools, for at most three minutes.

The logins sign in as `debian` or `root`, so the panel is for admins alone.

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
hosts again each minute on its own, the time it drew them under the
table as **Live updated**. The page needs the Beszel
integration: its URL and a token (see [Settings](settings.md#integrations)).
Without one, the page says so.

---

Previous: [Dashboard](dashboard.md) · Next: [Workflows](workflows.md) · [Contents](../README.md)
