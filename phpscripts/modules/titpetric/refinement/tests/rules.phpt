name: RefinementPrompt::rules
description: >
  The rule bodies, in order, out of the comment rows the store answered
  with. A body that is blank is no rule: a request with nothing in it
  asks for nothing, and numbering it would make the list lie about how
  many rules the thread holds.
---
<?php

$comments = array(
	array("body" => "Use ULID as primary keys."),
	array("body" => "   "),
	array("body" => "  Don't use foreign keys.  "),
	array("body" => "Run mig lint."),
);

$rules = RefinementPrompt::rules($comments);

echo count($rules), " rules\n";

foreach ($rules as $rule) {
	echo "[", $rule, "]\n";
}

echo count(RefinementPrompt::rules(array())), " for a thread with no correction\n";
---
3 rules
[Use ULID as primary keys.]
[Don't use foreign keys.]
[Run mig lint.]
0 for a thread with no correction
