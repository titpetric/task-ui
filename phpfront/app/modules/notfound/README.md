# Not found

Generated from the module by `atkins docs`. DO NOT EDIT.

mod_notfound is the catch-all: the page for a path no other route claimed.

chi matches a literal and a parameter ahead of a wildcard, so this answers only what nothing else did. The Go router draws the same page from r.NotFound, behind the same guard, and the reason reads the same because Task\Route::path() drops the mount prefix: the cutover changes no string.

Without it an unrouted path under the prefix answers chi's own plain text, which is no page at all.

## Pages

| Method | Path         | Answered by                            |
|--------|--------------|----------------------------------------|
| `GET`  | `/{rest...}` | [`mod_notfound.php`](mod_notfound.php) |
| `POST` | `/{rest...}` | [`mod_notfound.php`](mod_notfound.php) |

## The data model

No page of this module reads a table of its own. The chrome every page draws reads the instance's, which `docs/phpfront.md` lists once.

## Operations

This module owns no data model of its own. It reads the instance's tables through `code/Task`.

---

The tables come from `tables()` on `mod_notfound`. The operations are the classes of `storage/`, which hold every statement this module runs.
