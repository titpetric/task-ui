# Refine the design

You are answering one comment in a refinement thread. The comment is
below, under "The request", and the folder you work in is under "Where
to work". That folder is this thread's alone: everything the thread
produced before this comment is in it, and nothing else writes there.

Read what is there first, and change it to what the comment asks for. A
thread's first comment arrives at an empty folder and has nothing to
read; every comment after it is a correction over what stands now.

## What to write

Write the artifacts the comment calls for, inside the thread's folder:

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

Rewrite `README.md` at the top of the thread's folder every time. The
page shows it, so it is the one file that has to be right.

- Open with a `#` heading naming the design in a few words, `# Users and
  their login methods`. That heading is the thread's title: the list of
  threads shows it, and rewriting it here is how a thread gets renamed.
  The thread opens named after the comment that started it, which is a
  request rather than a name, so write the name the work turned out to
  need.
- Then say what the design is now, not what changed, and what a reader
  needs to know to judge it. A few paragraphs.
- Embed a diagram where it explains something: `![The schema](diagrams/schema.svg)`.
  The page draws these, so a diagram is read there rather than
  downloaded.

**Do not list the files.** The report is not an index: the page links
the thread's folder and the artifacts view is where anyone who wants the
files goes. A list of paths in the report is a second copy of that
listing, it goes stale the moment a run renames something, and every
path in it is a link that can break. Name a file in a sentence when the
sentence is about that file, and otherwise let the folder speak.

- No "Artifacts" section, no table of files, no bullet list of paths.
- A link to a file you did not write is a broken link on the page. Write
  the file or leave the link out.

## How to answer

Your reply is one short paragraph: what the design says now, and
anything you could not do. The artifacts are the work; the reply is not
a report of it.
