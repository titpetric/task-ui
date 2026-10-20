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
Claude's summary leads with what needs attention, and shows above the
rows, with the report to download as `.md`. The button shows when the
Claude Code integration has a token for you (see
[Settings](settings.md#integrations)); Claude only reads the report and
runs no tools, for at most three minutes.

The logins sign in as `debian` or `root`, so the panel is for admins alone.

task-ui reads the hub at most once a minute. The page needs the Beszel
integration: its URL and a token (see [Settings](settings.md#integrations)).
Without one, the page says so.

---

Previous: [Dashboard](dashboard.md) · Next: [Workflows](workflows.md) · [Contents](../README.md)
