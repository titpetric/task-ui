name: chrome
description: Task\Chrome is what every page shares - the catalogue the reader sees, the named workspaces, the roles and the maps the body carries, who the header's prompt runs as and where it asks.
---
<?php

// Chrome reads the whole user table and the whole workspace table, because a
// page draws every reader's avatar and links a folder under its workspace. So
// each answer here is narrowed to the rows this fixture wrote: one connection
// serves the suite and another fixture's user is as real as these. The
// catalogue is no table - it is a host static this fixture owns.

$db = Task\DB::writer();

// The admin's row names no folder, so they see every one; the runner's names
// one, so they see that one alone.
$db->replace("user", array("id" => "fx-chrome-u1", "name" => "fx-chrome-ann", "role" => "admin", "admin" => 1));
$db->replace("user", array("id" => "fx-chrome-u2", "name" => "fx-chrome-bob", "role" => "runner", "folders" => '["fx-chrome/seen"]'));

$db->replace("workspace", array("id" => "fx-chrome-w1", "folder" => "fx-chrome-alpha", "sort" => 1));

// A key a role holds is a key every user of that role brings, which is what
// the runs-as select says beside a name.
$db->replace("key_info", array("name" => "fx-chrome-deploy", "purpose" => "ssh", "algo" => "ed25519", "fingerprint" => "fx:chrome:1", "scope" => "instance"));
$db->replace("key_role", array("id" => "fx-chrome-kr1", "key_name" => "fx-chrome-deploy", "role" => "runner"));

$db->replace("run", array("id" => "fx-chrome-r1", "folder" => "fx-chrome/seen", "job" => "plan", "dir" => "", "user_name" => "fx-chrome-ann", "state" => "ok"));

$plan = Task\Fixture::job("plan");
$dump = Task\Fixture::job("dump");

Task\Host::$folders = array(
	Task\Fixture::folder("fx-chrome/seen", array($plan)),
	Task\Fixture::folder("fx-chrome/hidden", array($dump)),
);
Task\Host::$login = true;
Task\Host::$whoami = "fx-chrome-ann";

// mine is the entries of a list this fixture wrote, in the order it answered.
function mine($values) {
	$out = array();

	foreach ($values as $value) {
		if (strpos($value, "fx-chrome") === 0) {
			$out[] = $value;
		}
	}

	return $out;
}

// ids names the folders of a catalogue.
function ids($folders) {
	$out = array();

	foreach ($folders as $folder) {
		$out[] = Task\Row::text($folder, "id");
	}

	return $out;
}

$chrome = new Task\Chrome;

echo implode(",", ids($chrome->folders())), "\n";
echo implode(",", mine($chrome->workspaces())), "\n";
echo implode(",", mine(array_keys($chrome->roles()))), "\n";
echo $chrome->roles()["fx-chrome-ann"], " ", $chrome->roles()["fx-chrome-bob"], "\n";

// The runs-as select is an admin's: every user, each with the instance SSH keys
// their role brings, the reader picked, in the short form the header draws.
$runAs = $chrome->runAs();

foreach ($runAs["users"] as $option) {
	if (strpos($option["name"], "fx-chrome") === 0) {
		echo $option["name"], " ", $option["role"], " [", $option["keys"], "] ", $option["said"], "\n";
	}
}

echo $runAs["selected"], " ", var_export($runAs["short"], true), "\n";

// The scope the theme reads, for the facts the body and the header carry. The
// prompt asks at the root, because this page named no place; the two maps are
// the whole instance's, so the shape is read and not the whole of either.
$data = $chrome->data();

echo $data["prompt"]["folder"], " ", $data["prompt"]["label"], " ", $data["prompt"]["href"], "\n";
echo var_export($data["loginRequired"], true), "\n";
echo var_export(strpos($data["roles"], '"fx-chrome-bob":"runner"') !== false, true), " ", var_export(strpos($data["workspaces"], '"fx-chrome-alpha":') !== false, true), "\n";

// A page that names a place says so, and the prompt asks there: the label is
// the place and the href the page of the longest folder the reader sees that
// holds it. A folder in no named workspace hangs under the root.
$chrome->asks("fx-chrome/seen", "docker");

echo $chrome->prompt["folder"], " ", $chrome->prompt["dir"], " ", $chrome->prompt["label"], " ", $chrome->prompt["href"], "\n";

// A run's page asks about the run, in the run's place, and names it by the
// start of its ID.
$runs = new Task\Run;
$run = $runs->get("fx-chrome-r1");

$chrome->asksOfRun($run);

echo $chrome->prompt["folder"], " ", $chrome->prompt["run"], " ", $chrome->prompt["label"], " ", $chrome->prompt["href"], "\n";

// Nobody signed in reads no roles, sees no folder and picks no user: who
// someone is goes no further than the page needs, and the body carries an
// empty map as {} rather than as [].
Task\Host::$whoami = "";

$nobody = new Task\Chrome;

echo "[", implode(",", array_keys($nobody->roles())), "] [", implode(",", ids($nobody->folders())), "] ", Task\Text::object($nobody->roles()), "\n";
echo count($nobody->runAs()["users"]), "\n";
---
fx-chrome/seen,fx-chrome/hidden
fx-chrome-alpha
fx-chrome-ann,fx-chrome-bob
admin runner
fx-chrome-ann admin [] no keys
fx-chrome-bob runner [fx-chrome-deploy] keys: fx-chrome-deploy
fx-chrome-ann true
internal internal /f/internal
true
true true
fx-chrome/seen docker fx-chrome/seen/docker /f/internal/fx-chrome/seen/docker
fx-chrome/seen fx-chrome-r1 run fx-c… /runs/fx-chrome-r1
[] [] {}
0
