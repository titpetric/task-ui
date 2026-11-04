name: workspace
description: Task\Workspace reads the named workspaces the sidebar lists, and the rank of each one the body carries for the script.
---
<?php

require "ext/autoload.php";

// Both readers answer the whole table, so each answer is narrowed to the rows
// this fixture wrote: one connection serves the suite, and another fixture's
// workspace is as real as these. Nothing here reads a rank as a number for the
// same reason - a foreign row shifts every one of them.

$db = Task\DB::writer();

// The default workspace is no named one: its projects are the root's, which is
// why both readers leave it out. The rows are written with replace, because a
// folder is unique and another fixture may have written this one.
//
// The sort column runs against the names on purpose, so the two orders tell
// each other apart, and two rows share a sort so the tie is read too.
$db->replace("workspace", array("id" => "fx-ws-0", "folder" => Task\Route::INTERNAL, "sort" => 0));
$db->replace("workspace", array("id" => "fx-ws-1", "folder" => "fx-ws-alpha", "sort" => 3));
$db->replace("workspace", array("id" => "fx-ws-2", "folder" => "fx-ws-bravo", "sort" => 2));
$db->replace("workspace", array("id" => "fx-ws-3", "folder" => "fx-ws-charlie", "sort" => 1));
$db->replace("workspace", array("id" => "fx-ws-4", "folder" => "fx-ws-delta", "sort" => 1));

$workspaces = new Task\Workspace;

// mine is the entries of a list this fixture wrote, in the order it answered.
function mine($values) {
	$out = array();

	foreach ($values as $value) {
		if (strpos($value, "fx-ws-") === 0) {
			$out[] = $value;
		}
	}

	return $out;
}

// The sidebar's order is the sort column, and the folder where two share one.
$order = mine($workspaces->folders());

echo implode(",", $order), "\n";

// The ranks come back by folder, over the order already read and keyed in it:
// a rank is a workspace's place in that order.
$ranks = Task\Workspace::ranked($workspaces->folders());

echo implode(",", mine(array_keys($ranks))), "\n";

// A rank is a workspace's place in the sidebar's order, so the ranks rise
// along it however many workspaces the instance holds.
$rising = true;

for ($i = 1; $i < count($order); $i++) {
	if ($ranks[$order[$i]] <= $ranks[$order[$i - 1]]) {
		$rising = false;
	}
}

echo var_export($rising, true), " ", var_export(isset($ranks[Task\Route::INTERNAL]), true), "\n";

// The body carries the ranks as one attribute: a JSON object with its keys
// sorted, and {} for an instance that names no workspace, which is the shape a
// Go map marshals to. Only the rows this fixture wrote are ranked here,
// because the table is the whole instance's.
echo Task\Text::object(Task\Workspace::ranked($order)), "\n";
echo Task\Text::object(array()), "\n";

// A folder's page sits under its workspace where the workspace is named, and
// under the root where it is not. The root itself is its own page.
echo Task\Route::folderHref(Task\Route::INTERNAL, $order), "\n";
echo Task\Route::folderHref("fx-ws-alpha", $order), "\n";
echo Task\Route::folderHref("fx-ws-alpha/deploy", $order), "\n";
echo Task\Route::folderHref("fx-ws-elsewhere", $order), "\n";

// A directory inside a folder hangs off the folder's page.
echo Task\Route::placeHref("fx-ws-alpha", "", $order), "\n";
echo Task\Route::placeHref("fx-ws-alpha", "docker", $order), "\n";
echo Task\Route::placeHref("fx-ws-elsewhere", "docker", $order), "\n";

// What a title names a place by, which is the folder alone without one.
echo Task\Route::place("fx-ws-alpha", ""), " ", Task\Route::place("fx-ws-alpha", "docker"), "\n";
---
fx-ws-charlie,fx-ws-delta,fx-ws-bravo,fx-ws-alpha
fx-ws-charlie,fx-ws-delta,fx-ws-bravo,fx-ws-alpha
true false
{"fx-ws-alpha":3,"fx-ws-bravo":2,"fx-ws-charlie":0,"fx-ws-delta":1}
{}
/f/internal
/f/fx-ws-alpha
/f/fx-ws-alpha/deploy
/f/internal/fx-ws-elsewhere
/f/fx-ws-alpha
/f/fx-ws-alpha/docker
/f/internal/fx-ws-elsewhere/docker
fx-ws-alpha fx-ws-alpha/docker
