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

What a thread shows above everything is what the last run embedded: the
diagram, the rendering, the screenshot. A refinement is reviewed by
looking at it, so the pictures get the width of the column and the run's
prose folds behind **show thinking**. A report that embeds nothing has
no picture to review, and then its prose is the output and stands open.

The report lists no files. Under it the page carries one link, **The
files of this thread**, which opens the folder in
[Artifacts](artifacts.md): the preview, the versions and the download
are there already, and a list in the report would be a second copy of
that listing going stale at the next rename. Beside it, once a run has
written one, **The scratch pad** opens `SCRATCHPAD.md`, which is the
runs' own notes: which tool drew which artifact, the command and the
flags, and the script under `bin/` that regenerates it. A run reads it
before it reads anything else, so the second run does not work out again
what the first already knew.

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

### Rules, and Apply all

Every correction is a **rule**: a thing the artifacts have to satisfy,
not a thing that happened once. The thread keeps them, and the **⋯**
trash button on a request retires one. A rule that was a mistake stops
counting from the moment it goes, and the files it produced stay where
they are.

Running the corrections one after another is what drifts. Each run sees
the folder as the run before it left it and the comment you just wrote,
so by the tenth the first correction is a thing nobody has read in an
hour. **Apply all**, above the comment box, is the answer: it folds the
design and every surviving rule into one request and runs it once, over
the artifacts that are already there, so the result satisfies all of
them together.

What it sends reads as the design, then the rules numbered in the order
they were asked for, then where to start from — the folder and the
scratch pad. A later rule wins where two disagree, and the run is told
so. The requests it folded stay on the page, behind **N folded into the
request below**, with the run each one started still linked.

Every request on the page is drawn as the markdown it was written as, so
a list reads as a list and a path in backticks reads as code. Its
headings come down to the smallest one and set as a bold line: a request
is a paragraph inside a thread, and a flattened one carrying `## The
design` would otherwise outrank the thread's own headings.

It composes from the thread and not from its own last answer, so a
second **Apply all** after two more corrections carries all of them
again. Delete a rule, press **Apply all**, and what comes back is the
design as it should have been.

Deleting the first request deletes the thread. It is the design the
rules correct, and without it they correct nothing. The folder in the
artifacts is not touched.

### Rolling back

The artifact tree is a git repository and every revision of it is a
commit, which the thread's runs are told: each request carries the
folder's recent history, the commits newest first with the paths each
one touched. So going back is an ordinary correction — *put the diagram
back the way it was before the layout change* — and the run does it by
reading the file at that commit and writing what it finds. Nothing is
erased either way: a rollback is a new revision that says what an old
one said.

### Forking

The rules generalise and the subject does not. *Every id is a ULID*,
*no foreign keys*, *the diagram is generated and the generation is
repeatable* are not about users and their logins; they are how the work
is done, and typing them again for the next subject is the thing worth
not doing.

**Fork**, under the comment box, takes the subject of the new work and
opens a thread with every rule of this one already in it. The new thread
starts on an empty folder of its own and its first run is the flattened
request: the subject you typed, then the rules. From there it is a
thread like any other, and retiring a rule in it leaves the thread it
came from alone.

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
