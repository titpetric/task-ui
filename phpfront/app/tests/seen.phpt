name: seen
description: Seen answers whether the reader sees a run, over the rows a filter read or by ID, and keeps the answer of each.
---
<?php

require "ext/autoload.php";

// A reader whose row names the folders they see, so the list a filter read
// holds both a run they see and a run they do not.
$db = Task\DB::writer();

$db->replace("user", array(
	"id" => "fx-seen-u1",
	"name" => "fx-seen-bob",
	"role" => "runner",
	"folders" => '["fx-seen"]',
));

// run writes one run of a folder, in the shape a write leaves.
function run($id, $folder) {
	$db = Task\DB::writer();

	$db->replace("run", array(
		"id" => $id,
		"folder" => $folder,
		"job" => "build",
		"dir" => "",
		"user_name" => "fx-seen-bob",
		"state" => "ok",
		"started_at" => "2026-11-02 10:01:00.000 +0000 UTC",
		"finished_at" => "2026-11-02 10:01:02.500 +0000 UTC",
	));
}

run("fx-seen-1", "fx-seen");
run("fx-seen-2", "fx-seen-elsewhere");
run("fx-seen-3", "fx-seen");

Task\Host::$whoami = "fx-seen-bob";

$chrome = new Task\Chrome;
$runs = new Task\Run;
$seen = new Task\Seen($chrome);

// ids names the runs of a list, in the order it holds them.
function ids($rows) {
	$out = array();

	foreach ($rows as $row) {
		$out[] = $row->getId();
	}

	return implode(",", $out);
}

// The rows are read by their own IDs, because the suite shares a connection
// and a filter of a folder reads whatever another fixture left in it.
$list = array($runs->get("fx-seen-1"), $runs->get("fx-seen-2"), $runs->get("fx-seen-3"));

// Of the list, the runs of the folder the row names, in the order they
// arrived.
echo ids($seen->of($list)), "\n";

// Every run of the list was answered, the ones dropped as well, so a page that
// asks about one of them by ID later reads nothing again.
echo implode(",", array_keys($seen->asked)), "\n";
echo var_export($seen->run("fx-seen-1"), true), " ", var_export($seen->run("fx-seen-2"), true), "\n";

// An empty ID is no run at all, and nothing is read for it.
echo var_export($seen->run(""), true), "\n";

// A reader who is nobody sees none of them.
Task\Host::$whoami = "";

$nobody = new Task\Chrome;
$hidden = new Task\Seen($nobody);

echo "[", ids($hidden->of($list)), "]\n";
---
fx-seen-1,fx-seen-3
fx-seen-1,fx-seen-2,fx-seen-3
true false
false
[]
