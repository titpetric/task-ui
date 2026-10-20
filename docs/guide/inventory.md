# Inventory

The hosts Beszel monitors, inline in task-ui, the largest first by cores,
then RAM: whether each is up, who task-ui signs in to it as over ssh, its
cores, how busy its CPU is, and how full its memory and disk are, with what
it uses of each and the size, in GB. A share turns amber past 75% and red
past 90%. A host whose agent sends no core count shows its threads.

The SSH column is the first of `debian@` and `root@` that signs in to the
host with the instance keys, without a password; the button beside it
copies the `ssh` command. task-ui tries every ten minutes, and a host it
reaches slowly fills in on the next view.

## Running a command on the hosts

An admin runs a shell command on the hosts at once, like pssh: pick the
hosts, by default every one with an ssh login, write the command or take a
quick one (`uptime`, `df -h /`, `free -h`, the containers, the kernel, the
OS), and **Run**, or Ctrl+Enter. Each host runs it over its login, for at
most two minutes, and the page shows a card a host: its exit code, how
long it took, and what it printed, both streams.

The logins sign in as `debian` or `root`, so the panel is for admins alone.

task-ui reads the hub at most once a minute. The page needs the Beszel
integration: its URL and a token (see [Settings](settings.md#integrations)).
Without one, the page says so.

---

Previous: [Dashboard](dashboard.md) · Next: [Workflows](workflows.md) · [Contents](../README.md)
