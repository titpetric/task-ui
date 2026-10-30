# Modules

A module is a plugin. It sits in the sidebar beside Dashboard and Runs,
opens at `/m/<vendor>/<name>`, and its pages are pages of the instance:
the same header, the same sidebar, the same theme. The body is the
module's, written in PHP and run inside task-ui by an interpreter the
binary holds. Nothing loads in an iframe and nothing is fetched from
another host.

task-ui finds its modules at startup. Any directory under the modules
folder or the workspace root with a `modinfo.php` in it is a module, and
that directory is its root. The instance records each one in the
`module` table with the name, the title, the icon and the root, and adds
a sidebar row for it. [Settings, Sidebar](settings.md#sidebar) shows or
hides the row and moves it.

The footer of a module's page prints how long the body took to render.

The modules that ship with the instance live in the binary and arrive in
the data folder on first start, so a fresh instance has them without a
checkout. A module in a project's checkout is found the same way, so a
repository ships one of its own.

Writing one is [Modules](../modules.md), beside the development docs.

## Refinement

`/m/titpetric/refinement` turns a conversation into artifacts. The page
opens on one field: what you want designed. **Start a thread** records
the thread and your first comment, and starts a Claude run that has the
artifact tree and the comment, and nothing else to go on.

The run writes its output into a folder of the instance's artifacts
named after the thread, which is the thread's alone: a new thread starts
on an empty directory and nothing another thread wrote is in its way. It
writes a `README.md` at the top of that folder saying what the design is
now, with a diagram embedded where one explains something. The thread's
page shows that report, so an SVG the run drew is on the page rather
than behind a download.

The report lists no files. Under it the page carries one link, **The
files of this thread**, which opens the folder in
[Artifacts](artifacts.md): the preview, the versions and the download
are there already, and a list in the report would be a second copy of
that listing going stale at the next rename.

A thread then takes corrections. The comment box at the bottom of the
thread is the same field as the first one: a comment starts a Claude run
of its own, with the thread's folder as it stands now and the comment
you just wrote. The run applies the correction and rewrites the report.
Three corrections in a row that each produce a run:

```
use ULID as primary keys
don't use foreign keys
run mig lint on the schema and fix violations
```

Every run is an ordinary run. It is on [Runs](runs.md) with its output
to replay, its artifacts are in [Artifacts](artifacts.md) with their
history, and Claude's reply lands in [Notes](notes.md). A correction
that fails fails visibly, and the thread keeps the artifacts of the
comment before it.

The artifacts belong to the instance. A later thread sees what an
earlier one left, because the run's context names the artifact tree the
same way every Claude run's does.

| Column           | What it holds                                                        |
|------------------|----------------------------------------------------------------------|
| the thread       | who opened it, when, and the `README.md` of its artifacts as it stands |
| a comment        | who wrote it, when, and the run it started                           |
| an artifact      | the path in the artifact tree, and the comment that produced it      |
| a correction     | the comment that names a change, and the run that applied it         |

---

Previous: [Notes](notes.md) · Next: [Projects and jobs](projects.md) · [Contents](../README.md)
