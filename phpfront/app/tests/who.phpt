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
$runner = new Task\Who(new Task\Access($ops), $roles);

echo var_export($runner->mine("ops"), true), var_export($runner->mine("dev"), true), "\n";
echo var_export($runner->profile("ops"), true), var_export($runner->profile("dev"), true), "\n";
echo var_export($runner->profile("nobody"), true), " a name nobody holds\n";
echo implode(",", $runner->linkable()), "\n";

// Granted user/view, every known name links.
$seer = new Task\Who(new Task\Access($ops, new Task\Policy(array("user/view" => Task\Policy::ALLOW))), $roles);

echo implode(",", $seer->linkable()), "\n";
echo var_export($seer->edits("dev"), true), " edits needs its own rule\n";

// of draws the user: the avatar, the role and the menu beside them. Edit is
// there because a user edits their own, which is the same rule as the
// profile; another user's needs user/edit.
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
  Edit /users/ops/edit
GH 1
[OP][A][TP][AL][XY][?]
