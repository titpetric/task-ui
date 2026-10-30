name: RefinementPrompt::flatten
description: >
  The request an Apply all sends: the design, then every rule numbered in
  the order it was asked for, then where the run starts from. The
  numbers are how a later rule says it beats an earlier one. A thread
  with no correction has nothing to flatten, so the rules section is not
  there at all.
---
<?php

echo RefinementPrompt::flatten("  A schema for users and their logins.  ", array(
	"Use ULID as primary keys.",
	"  Don't use foreign keys.  ",
	"Run mig lint and fix what it reports.",
)), "\n";

echo "--- with no rules ---\n";
echo RefinementPrompt::flatten("A schema for users and their logins.", array()), "\n";
---
## The design

A schema for users and their logins.

## The rules

Every correction this thread has made, as a rule the artifacts have to
satisfy at once. A later rule wins where two disagree.

1. Use ULID as primary keys.
2. Don't use foreign keys.
3. Run mig lint and fix what it reports.

## Where this starts

The thread's folder already holds the artifacts, and your scratch pad,
`SCRATCHPAD.md` at the top of it, says which tools made them and how to
run them again. Read both before you change anything, run the pipelines
the scratch pad names rather than starting over, and produce artifacts
that satisfy every rule together rather than applying them one after
another.
--- with no rules ---
## The design

A schema for users and their logins.

## Where this starts

The thread's folder already holds the artifacts, and your scratch pad,
`SCRATCHPAD.md` at the top of it, says which tools made them and how to
run them again. Read both before you change anything, run the pipelines
the scratch pad names rather than starting over, and produce artifacts
that satisfy every rule together rather than applying them one after
another.
