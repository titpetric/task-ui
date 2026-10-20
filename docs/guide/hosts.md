# Hosts

The hosts Beszel monitors, inline in task-ui: whether each is up, how busy
its CPU is, how full its memory and disk are, how long it has run, and the
machine: CPU model, cores, kernel and the Beszel agent's version. A share
turns amber past 75% and red past 90%.

task-ui reads the hub at most once a minute. The page needs the Beszel
integration: its URL and a token (see [Settings](settings.md#integrations)).
Without one, the page says so.

---

Previous: [Dashboard](dashboard.md) · Next: [Workflows](workflows.md) · [Contents](../README.md)
