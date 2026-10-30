# Refine the design

You are answering one comment in a refinement thread. The comment is
below, under "The request". Everything the thread produced before it is
already in the artifact tree, at `$TASK_UI_ARTIFACTS`, and the artifacts
belong to the instance rather than to this thread: a later thread reads
what an earlier one left.

Work in the artifact tree. Read what is there first, and change it to
what the comment asks for. A thread's first comment usually has nothing
to read; every comment after it is a correction over what stands now.

## What to write

Write the artifacts the comment calls for, under `$TASK_UI_ARTIFACTS`:

- A schema goes in `schema/` as one `*.up.sql` per table, named for it.
  Singular table names, a ULID primary key in `CHAR(26)`, no foreign
  keys, `created_at` and `updated_at` as native timestamps, indexes
  named `idx_<table>_<columns>` and unique ones `uidx_<table>_<columns>`.
  Each file opens with a `-- <table>: <one line>` comment and comments
  its columns.
- A diagram goes in `diagrams/` as an SVG, drawn with `d2` or
  `plantuml`, both of which are on the PATH. Write the source beside it
  so the next run can change the diagram rather than redraw it.
- Anything else the comment names goes where it belongs, in a directory
  that says what it is.

## The report

Rewrite `README.md` at the top of the artifact tree every time. It is
the index of the thread and the page shows it, so it is the one file
that has to be right.

- Open with a `#` heading naming the design in a few words, `# Users and
  their login methods`. That heading is the thread's title: the list of
  threads shows it, and rewriting it here is how a thread gets renamed.
  The thread opens named after the comment that started it, which is a
  request rather than a name, so write the name the work turned out to
  need.
- Then a sentence on what the design is now, not what changed.
- List every artifact the thread has produced, each as a markdown link
  to its path relative to the tree: `[schema/user.up.sql](schema/user.up.sql)`.
- Embed every image: `![The schema](diagrams/schema.svg)`. The page
  draws these, so a diagram is read there rather than downloaded.
- Say what a reader needs to know to judge it, in a few lines. No
  changelog, no list of what you did.

A link that names a file you did not write is a broken link on the page.
Write the file or leave the link out.

## How to answer

Your reply is one short paragraph: what the design says now, and
anything you could not do. The artifacts are the work; the reply is not
a report of it.
