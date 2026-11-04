name: who
description: Who draws a user, and decides which mention links.
---
<?php

Task\Host::$rules = array(
	"paths" => array("user/all", "user/view", "user/edit"),
	"checks" => array("user/view", "user/edit"),
	"extras" => array(),
);

$ops = new Task\Model\User(array("id" => "01", "name" => "ops", "role" => "runner"));
$roles = array("ops" => "runner", "root" => "admin", "dev" => "viewer");

// A runner with no policy: their own profile opens, nobody else's.
$plain = new Task\Access($ops);
$runner = new Task\Who($plain, $roles);

echo var_export($runner->mine("ops"), true), var_export($runner->mine("dev"), true), "\n";
echo var_export($runner->profile("ops"), true), var_export($runner->profile("dev"), true), "\n";
echo var_export($runner->profile("nobody"), true), " a name nobody holds\n";
echo implode(",", $runner->linkable()), "\n";

// Granted user/view, every known name links.
$sees = new Task\Policy(array("user/view" => Task\Policy::ALLOW));
$granted = new Task\Access($ops, $sees);
$seer = new Task\Who($granted, $roles);

echo implode(",", $seer->linkable()), "\n";
echo var_export($seer->edits("dev"), true), " edits needs its own rule\n";

// of draws the user: the avatar, the role and the menu beside them. The
// profile is there because a reader reads their own; Edit is not, because the
// edit page is the user's permissions and those are nobody's own to change.
$drawn = $runner->of("ops");

echo $drawn["name"], " ", $drawn["role"], " ", $drawn["initials"], "\n";

foreach ($drawn["menu"]["groups"] as $group) {
	foreach ($group as $item) {
		echo "  ", $item["label"], " ", $item["href"], "\n";
	}
}

// A name nobody holds still draws, with the runs link alone.
$stranger = $runner->of("ghost");

echo $stranger["initials"], " ", count($stranger["menu"]["groups"][0]), "\n";

// avatar is the scope the avatar component reads, which of() answers a
// superset of: the same three keys, and no menu.
$face = Task\Who::avatar("ops", "runner");

echo $face["name"], " ", $face["role"], " ", $face["initials"], " ", count($face), "\n";

// ofRun is who started a run and how. The scheduler started this one outside
// its window and it killed the run before it, and the window is the stronger
// fact, so the component draws the out of hours badge and not the forced one.
$ran = new Task\Model\Run(array(
	"id" => "01",
	"user_name" => "ops",
	"scheduled" => 1,
	"forced" => 1,
	"reason" => "asked for by hand",
));
$start = $runner->ofRun($ran);

echo $start["user"]["name"], " ", var_export($start["scheduled"], true), " ", var_export($start["forced"], true), " [", $start["reason"], "]\n";

// A run started by hand carries no badge at all.
$byHand = new Task\Model\Run(array("id" => "02", "user_name" => "ops"));
$plainStart = $runner->ofRun($byHand);

echo var_export($plainStart["scheduled"], true), " ", var_export($plainStart["forced"], true), " [", $plainStart["reason"], "]\n";

// The initials are the first letters of the first two words, or the first
// two of one word.
foreach (array("ops", "a", "tit.petric", "ada lovelace", "x_y", "") as $name) {
	echo "[", Task\Who::initials($name), "]";
}

echo "\n";
---
truefalse
truefalse
false a name nobody holds
ops
ops,root,dev
false edits needs its own rule
ops runner OP
  Profile /users/ops
  Runs by this user /runs?user=ops
GH 1
ops runner OP 3
ops true true [asked for by hand]
false false []
[OP][A][TP][AL][XY][?]
