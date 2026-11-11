# Dashboard

Generated from the module by `atkins docs`. DO NOT EDIT.

mod_dashboard is the page the instance opens on: every terminal running right now, as a tile with the last screen of its output, and the menu of what the reader may add.

The tiles follow the runs topic, so a run that starts adds one and one that ends marks it without a reload.

## Pages

| Method | Path | Answered by                              |
|--------|------|------------------------------------------|
| `GET`  | `/`  | [`mod_dashboard.php`](mod_dashboard.php) |

## The data model

No page of this module reads a table of its own. The chrome every page draws reads the instance's, which `docs/phpfront.md` lists once.

## Operations

This module owns no data model of its own. It reads the instance's tables through `code/Task`.

---

The tables come from `tables()` on `mod_dashboard`. The operations are the classes of `storage/`, which hold every statement this module runs.
