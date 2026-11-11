name: runs
description: The runs page lays the runner's session over a row - what it reports, how many watch it and the summary an ended run offers - and charts the whole window of a filter.
---
<?php

require "modules/runs/mod_runs.php";

$mod = new mod_runs;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo var_export($mod instanceof Module, true), "\n";

// One connection serves the suite and fixtures run together, so the rows are
// this fixture's own and every query narrows to its folders: fx-runs holds the
// rows a table draws, and fx-runs-now the rows the charts count, which start
// at this moment and would leave a window of days as the fixture ages.
$db = Task\DB::writer();

$db->replace("user", array("id" => "fx-runs-u1", "name" => "fx-runs-ann", "role" => "admin", "admin" => 1));

// at is a moment of the day the rows record, in the shape a write leaves.
function at($time) {
	return "2026-11-02 " . $time . " +0000 UTC";
}

// A run that ended, with a recording; one that waits for a slot; one in
// progress whose column holds the title it started under; and one the table
// still reads as running.
$db->replace("run", array(
	"id" => "fx-runs-1",
	"folder" => "fx-runs",
	"job" => "build",
	"user_name" => "fx-runs-ann",
	"state" => "ok",
	"cast_path" => "casts/fx-runs-1.cast",
	"started_at" => at("10:01:00.000"),
	"finished_at" => at("10:01:02.500"),
));
$db->replace("run", array(
	"id" => "fx-runs-2",
	"folder" => "fx-runs",
	"job" => "deploy",
	"user_name" => "fx-runs-ann",
	"state" => "queued",
));
$db->replace("run", array(
	"id" => "fx-runs-3",
	"folder" => "fx-runs",
	"job" => "test",
	"user_name" => "fx-runs-ann",
	"state" => "running",
	"title" => "the title the row holds",
	"started_at" => at("10:03:00.000"),
));
$db->replace("run", array(
	"id" => "fx-runs-4",
	"folder" => "fx-runs",
	"job" => "plan",
	"user_name" => "fx-runs-ann",
	"state" => "running",
	"started_at" => at("10:05:00.000"),
));

// The runner holds three of the four: the queued run waits for a slot, the
// running one has two windows watching it, and the fourth broke while the
// table still reads it as running.
Task\Host::$whoami = "fx-runs-ann";
Task\Host::$sessions = array(
	"fx-runs-2" => array(
		"live" => true,
		"attached" => 0,
		"message" => "waits for a free slot, number 1 in line",
		"state" => "queued",
		"exit_code" => 0,
		"title" => "",
		"finished_at" => "",
	),
	"fx-runs-3" => array(
		"live" => true,
		"attached" => 2,
		"message" => "",
		"state" => "running",
		"exit_code" => 0,
		"title" => "build the image",
		"finished_at" => "",
	),
	"fx-runs-4" => array(
		"live" => true,
		"attached" => 1,
		"message" => "",
		"state" => "failed",
		"exit_code" => 2,
		"title" => "",
		"finished_at" => at("10:05:04.500"),
	),
);

$runs = new Task\Runs\Runs;
$chrome = new Task\Chrome;
$workspaces = $chrome->workspaces();

// labels names the items a menu offers, in the order it offers them. The
// details on top are not items, so they are not here.
function labels($menu) {
	$out = array();

	foreach ($menu["groups"] as $group) {
		foreach ($group as $item) {
			$out[] = $item["label"];
		}
	}

	return implode(",", $out);
}

// attrs is what one item of a menu writes into the markup, escaped as the
// component draws it raw.
function attrs($menu, $label) {
	foreach ($menu["groups"] as $group) {
		foreach ($group as $item) {
			if ($item["label"] === $label) {
				return $item["attrs"];
			}
		}
	}

	return "";
}

// A run the runner no longer holds is the row alone: it ended, nobody watches
// it, and a reader with a token for Claude may read a summary of its output.
// The summary is read from the API at the root, the page it links under the
// theme's mount.
$one = $mod->row($runs->get("fx-runs-1"), $chrome, $workspaces, true);

echo $one["state"]["state"], " ", var_export($one["state"]["ended"], true), " [", $one["state"]["code"], "] ", $one["duration"], " ", $one["attach"], " ", $one["watching"], "\n";
echo labels($one["menu"]), "\n";
echo attrs($one["menu"], "Summary"), "\n";

// No token, no summary.
$plain = $mod->row($runs->get("fx-runs-1"), $chrome, $workspaces, false);

echo labels($plain["menu"]), "\n";

// A run that waits says what the runner reports, which the column does not
// hold, and may be stopped. A run that has not ended offers no summary.
$queued = $mod->row($runs->get("fx-runs-2"), $chrome, $workspaces, true);

echo $queued["state"]["state"], " [", $queued["message"], "] ", $queued["watching"], " ", $queued["attach"], " ", $queued["stamp"], "\n";
echo labels($queued["menu"]), "\n";

// A run in progress reads its title from the session and counts the windows
// watching it.
$live = $mod->row($runs->get("fx-runs-3"), $chrome, $workspaces, true);

echo $live["state"]["state"], " [", $live["title"], "] ", $live["watching"], " ", $live["attach"], "\n";
echo labels($live["menu"]), "\n";

// A run the runner has finished reads as finished before the table catches
// up: the state, the exit code and how long it took are the session's, and
// nothing may stop it.
$done = $mod->row($runs->get("fx-runs-4"), $chrome, $workspaces, true);

echo $done["state"]["state"], " ", var_export($done["state"]["ended"], true), " ", $done["state"]["code"], " ", $done["duration"], " ", $done["attach"], " ", $done["watching"], "\n";
echo labels($done["menu"]), "\n";

// The charts are of the filter and not of the page: the cursors come off and
// the limit becomes the charts' own, so both rows count where the filter
// named one page of one row after a cursor. The clock is the server's and the
// rows start now, so each lands in the last slot of every chart.
$db->replace("run", array(
	"id" => "fx-runs-now-1",
	"folder" => "fx-runs-now",
	"job" => "build",
	"user_name" => "fx-runs-ann",
	"state" => "running",
	"started_at" => Task\Clock::stamp(),
));
$db->replace("run", array(
	"id" => "fx-runs-now-2",
	"folder" => "fx-runs-now",
	"job" => "build",
	"user_name" => "fx-runs-ann",
	"state" => "running",
	"started_at" => Task\Clock::stamp(),
));

$seen = new Task\Seen($chrome);
$charts = $mod->activity($seen, $runs, array(
	"folder" => "fx-runs-now",
	"before" => "fx-runs-now-2",
	"limit" => 1,
));

echo implode(",", array_keys($charts)), "\n";
echo $charts["sum"], " [", $charts["failed"], "] [", $charts["median"], "] ", count($charts["charts"]), "\n";

// The duration line is one job's alone and this page mixes jobs, so the
// charts draw without it.
echo count($charts["durations"]), "\n";

// Today is the last slot of the day chart and the busiest of them, so its
// runs fill it. Both are in progress, so the part that ended ok is drawn at no
// height and the line of finished runs says nothing.
$day = $charts["charts"][0];

echo count($day["bars"]), " ", $day["n"], " ", $day["bars"][29]["class"], " [", $day["bars"][29]["other"], "] [", $day["bars"][29]["ok"], "] [", $day["stats"]["text"], "]\n";
---
runs Runs list
true
ok true [] 2.5s View 0
Open the run's page,Replay the recording,Summary
data-summary="/api/runs/fx-runs-1/summary" data-summary-page="/runs/fx-runs-1" data-summary-title="fx-runs / build" title="Claude&#039;s summary of the output"
Open the run's page,Replay the recording
queued [waits for a free slot, number 1 in line] 0 Attach -
Open the run's page,Stop,Kill
running [build the image] 2 Attach
Open the run's page,Stop,Kill
failed true 2 4.5s View 1
Open the run's page,Summary
sum,failed,median,charts,durations
2 runs in 30 days [] [] 3
0
30 30 tip tip-end [height:100.00%] [height:0.00%] []
