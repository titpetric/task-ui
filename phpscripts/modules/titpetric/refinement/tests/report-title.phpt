name: RefinementManager::reportTitle
description: >
  The first heading of the report, which is what the thread is called
  once a run has written one. ATX and setext both, the first one
  anywhere in the document so a report opening with a badge still
  answers, and empty when there is no heading at all — the absence of a
  title, which leaves the thread the name it had.
---
<?php

$cases = array(
	"# Users and their login methods\n\nThree tables.\n",
	"# Users and their login methods #\n",
	"#Nospace\n\n# The real one\n",
	"Users and their login methods\n=============================\n\nThree tables.\n",
	"<!-- generated -->\n\n# After a comment\n",
	"## Only a second level\n",
	"No heading at all.\n",
	"",
	"#   Padded   \n",
	"# " . str_repeat("x", 200) . "\n",
);

foreach ($cases as $markdown) {
	$title = RefinementManager::reportTitle($markdown);

	echo "[", $title, "] ", strlen($title), "\n";
}
---
[Users and their login methods] 29
[Users and their login methods] 29
[The real one] 12
[Users and their login methods] 29
[After a comment] 15
[] 0
[] 0
[] 0
[Padded] 6
[xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx...] 120
