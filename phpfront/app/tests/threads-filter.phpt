name: threads-filter
description: Task\Threads\Notes::feed narrows the threads by place, user, state, origin and words.
---
<?php

require "ext/autoload.php";

$db = Task\DB::writer();
$db->query("DELETE FROM note WHERE id LIKE 'fx-filter-%'");

// Every body holds the word the filter searches for, so a filter that names
// nothing else still lists this fixture's threads alone.
foreach ([
	["id" => "fx-filter-c1", "user_name" => "fx-filter-ann", "state" => "done", "dir" => "fx-filter/app", "origin" => "folder", "body" => "fxfilterword the build broke"],
	["id" => "fx-filter-c2", "user_name" => "fx-filter-bob", "state" => "working", "dir" => "fx-filter/app/sub", "origin" => "orb", "body" => "fxfilterword deploy is slow"],
	["id" => "fx-filter-c3", "user_name" => "fx-filter-ann", "state" => "done", "dir" => "fx-filter/other", "origin" => "threads", "body" => "fxfilterword the build is green"],
] as $row) {
	$db->replace("note", $row + [
		"parent_id" => "",
		"channel" => "notes",
		"user_id" => "",
		"mentions" => "",
		"workspace_id" => "",
		"created_at" => "2026-11-16 10:00:00.000 +0000 UTC",
		"updated_at" => "2026-11-16 10:00:00.000 +0000 UTC",
	]);
}

// ids is the threads a filter answers, newest first, as one line.
function ids($rows) {
	$out = [];

	foreach ($rows as $row) {
		$out[] = $row->getId();
	}

	return implode(",", $out);
}

$notes = new Task\Threads\Notes($db);
$base = ["q" => "fxfilterword"];

echo ids($notes->feed($base)), "\n";
echo ids($notes->feed(["dir" => "fx-filter/app"] + $base)), "\n";
echo ids($notes->feed(["state" => "done"] + $base)), "\n";
echo ids($notes->feed(["user" => "fx-filter-bob"] + $base)), "\n";
echo ids($notes->feed(["origin" => "threads"] + $base)), "\n";
---
fx-filter-c3,fx-filter-c2,fx-filter-c1
fx-filter-c2,fx-filter-c1
fx-filter-c3,fx-filter-c1
fx-filter-c2
fx-filter-c3
