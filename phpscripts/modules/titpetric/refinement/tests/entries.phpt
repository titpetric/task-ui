name: RefinementManager::entries and designID
description: >
  How the page draws a thread. A comment that stands on its own is an
  entry; an Apply all is an entry with the comments it folded under it,
  which the page keeps behind a disclosure rather than behind nothing.
  The design is the oldest comment that corrects nothing and records no
  Apply all, and the page needs it because its delete button takes the
  thread with it.
---
<?php

$comments = array(
	array("id" => "01D", "kind" => "", "correction" => 0, "collapsed_into" => "01A"),
	array("id" => "01R1", "kind" => "", "correction" => 1, "collapsed_into" => "01A"),
	array("id" => "01R2", "kind" => "", "correction" => 1, "collapsed_into" => "01A"),
	array("id" => "01A", "kind" => "apply", "correction" => 0, "collapsed_into" => ""),
	array("id" => "01R3", "kind" => "", "correction" => 1, "collapsed_into" => ""),
);

foreach (RefinementManager::entries($comments) as $entry) {
	echo $entry["comment"]["id"], " folded=", $entry["folded_count"],
		" applied=", ($entry["applied"] ? "yes" : "no"), "\n";

	foreach ($entry["folded"] as $under) {
		echo "  ", $under["id"], "\n";
	}
}

echo "design ", RefinementManager::designID($comments), "\n";
echo "design of a thread with none [", RefinementManager::designID(array(
	array("id" => "01A", "kind" => "apply", "correction" => 0, "collapsed_into" => ""),
	array("id" => "01R", "kind" => "", "correction" => 1, "collapsed_into" => ""),
)), "]\n";
---
01A folded=3 applied=yes
  01D
  01R1
  01R2
01R3 folded=0 applied=no
design 01D
design of a thread with none []
