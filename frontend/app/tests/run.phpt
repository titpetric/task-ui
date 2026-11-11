name: run
description: Task\Run reads the run table - the conditions a filter builds, the pages either side of a cursor, and the text a timestamp column arrives as.
---
<?php

require "ext/autoload.php";

// One connection serves the suite and fixtures run together, so every query
// here names this fixture's own folders. A fixture that read the whole table
// would read another fixture's rows.

$folder = "fx-run";
$other = "fx-run-other";

$db = Task\DB::writer();

// A minute apart, newest last. Among them the two rows the filters pick out:
// the job of a directory inside the folder, which carries the same name as
// the folder's own job, and a command on a host.
$rows = array(
	array("id" => "fx-run-1", "folder" => $folder, "job" => "build", "dir" => "", "state" => "ok", "at" => "10:01"),
	array("id" => "fx-run-2", "folder" => $folder, "job" => "build", "dir" => "sub", "state" => "failed", "at" => "10:02"),
	array("id" => "fx-run-3", "folder" => $folder, "job" => "test", "dir" => "", "state" => "ok", "at" => "10:03"),
	array("id" => "fx-run-4", "folder" => $folder, "job" => "@host:web1", "dir" => "", "state" => "ok", "at" => "10:04"),
	array("id" => "fx-run-5", "folder" => $folder, "job" => "deploy", "dir" => "", "state" => "running", "at" => "10:05"),
	array("id" => "fx-run-6", "folder" => $other, "job" => "build", "dir" => "", "state" => "ok", "at" => "10:06"),
);

foreach ($rows as $row) {
	$db->replace("run", array(
		"id" => $row["id"],
		"folder" => $row["folder"],
		"job" => $row["job"],
		"dir" => $row["dir"],
		"user_name" => "admin",
		"state" => $row["state"],
		"exit_code" => $row["state"] === "failed" ? 2 : 0,
		"started_at" => "2026-11-02 " . $row["at"] . ":00.000 +0000 UTC",
		"term_cols" => 120,
		"term_rows" => 30,
	));
}

// Two runs the first one dispatched, in a folder of their own so the filters
// above read the same rows they always did.
$kids = "fx-run-kids";
$dispatched = array(
	array("id" => "fx-run-k1", "at" => "10:07"),
	array("id" => "fx-run-k2", "at" => "10:08"),
);

foreach ($dispatched as $kid) {
	$db->replace("run", array(
		"id" => $kid["id"],
		"folder" => $kids,
		"job" => "@claude-prompt",
		"dir" => "",
		"user_name" => "admin",
		"state" => "ok",
		"started_at" => "2026-11-02 " . $kid["at"] . ":00.000 +0000 UTC",
		"term_cols" => 120,
		"term_rows" => 30,
		"parent_id" => "fx-run-1",
		"parent_kind" => "run",
	));
}

// A page reads through the reader, which refuses a write: a list is a list.
$runs = new Task\Run;

// ids names what a query answered, in the order it answered.
function ids($found) {
	$out = array();

	foreach ($found as $run) {
		$out[] = $run->getId();
	}

	return implode(",", $out);
}

// Newest first, a page at a time, and the pages either side of a cursor. A
// page read after a cursor comes back newest first too, reversed by find.
echo ids($runs->find(array("folder" => $folder, "limit" => 3))), "\n";
echo ids($runs->find(array("folder" => $folder, "limit" => 3, "before" => "fx-run-3"))), "\n";
echo ids($runs->find(array("folder" => $folder, "limit" => 2, "after" => "fx-run-2"))), "\n";
echo "[", ids($runs->find(array("folder" => $folder, "before" => "gone"))), "]\n";

// A job narrows to the folder's own by default, because dir is read whenever
// job is: the run of the directory inside it answers only when asked for.
echo ids($runs->find(array("folder" => $folder, "job" => "build"))), "\n";
echo ids($runs->find(array("folder" => $folder, "job" => "build", "dir" => "sub"))), "\n";

// @host:* is every command on a host and not a job by that name.
echo ids($runs->find(array("folder" => $folder, "job" => Task\Run::HOST_ALL))), "\n";
echo "[", ids($runs->find(array("folder" => $folder, "job" => "@host:web9"))), "]\n";

// The rest of the conditions: a state, the user who started it, and the
// folders a reader may see, which keeps a page from filling with rows the
// reader is about to be denied.
echo ids($runs->find(array("folder" => $folder, "state" => "running"))), "\n";
echo ids($runs->find(array("folders" => array($other), "limit" => 10))), "\n";
echo ids($runs->find(array("folders" => array($folder, $other), "state" => "failed"))), "\n";
echo "[", ids($runs->find(array("folder" => $folder, "user" => "nobody"))), "]\n";

// No limit reads the lot.
echo count($runs->find(array("folder" => $folder, "limit" => 0))), "\n";

// The parent conditions: the runs one run dispatched, newest first as every
// listing is, and the runs nothing dispatched, which is what the rail lists.
echo ids($runs->find(array("parent" => "fx-run-1", "limit" => 10))), "\n";
echo ids($runs->find(array("folders" => array($kids), "limit" => 10))), "\n";
echo "[", ids($runs->find(array("folders" => array($kids), "roots" => true, "limit" => 10))), "]\n";

// The children of one run, oldest first, as a run's own page lists them.
echo ids($runs->children("fx-run-1")), "\n";
echo "[", ids($runs->children("fx-run-9")), "] [", ids($runs->children("")), "]\n";

// The children of many runs at once, which the rail counts: the rows carry the
// parent and the columns the permission rule reads, and no parent asks nothing.
$found = $runs->childrenOf(array("fx-run-1", "fx-run-3"));

echo count($found), " ", Task\Row::text($found[0], "parent_id"), " ", Task\Row::text($found[0], "folder"), " ", Task\Row::text($found[0], "job"), "\n";
echo count($runs->childrenOf(array())), "\n";

// One run by ID, and nothing for an ID that names none.
$one = $runs->get("fx-run-2");

echo $one->getJob(), " ", $one->getState(), " ", $one->getExitCode(), "\n";
// The column is declared DATETIME and the driver renders what it parsed; the
// expression beside it carries no type, so the text arrives whole and a
// duration reads to the tenth of a second the Go page reads it to.
echo $one->getStartedAt(), "\n";
echo var_export($runs->get(""), true), " ", var_export($runs->get("nope"), true), "\n";
---
fx-run-5,fx-run-4,fx-run-3
fx-run-2,fx-run-1
fx-run-4,fx-run-3
[]
fx-run-1
fx-run-2
fx-run-4
[]
fx-run-5
fx-run-6
fx-run-2
[]
5
fx-run-k2,fx-run-k1
fx-run-k2,fx-run-k1
[]
fx-run-k1,fx-run-k2
[] []
2 fx-run-1 fx-run-kids @claude-prompt
0
build failed 2
2026-11-02 10:02:00.000 +0000 UTC
NULL NULL
