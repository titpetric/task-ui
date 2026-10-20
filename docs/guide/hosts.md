# Hosts

The hosts Beszel monitors, inline in task-ui, the largest first by cores,
then RAM: whether each is up, who task-ui signs in to it as over ssh, its
cores, how busy its CPU is, and how full its memory and disk are, with what
it uses of each and the size, in GB. A share turns amber past 75% and red
past 90%. A host whose agent sends no core count shows its threads.

The SSH column is the first of `debian@` and `root@` that signs in to the
host with the instance keys, without a password; the button beside it
copies the `ssh` command. task-ui tries every ten minutes, and a host it
reaches slowly fills in on the next view.

task-ui reads the hub at most once a minute. The page needs the Beszel
integration: its URL and a token (see [Settings](settings.md#integrations)).
Without one, the page says so.

---

Previous: [Dashboard](dashboard.md) · Next: [Workflows](workflows.md) · [Contents](../README.md)
