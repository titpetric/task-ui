name: thread-feed
description: The rail lists a prompt thread as one entry, links it to its conversation, drops the Claude run it owns, keeps place visibility, and maps a note's state to a dot.
---
<?php

require "ext/autoload.php";

$db = Task\DB::writer();
$folder = "fx-feed";
$place = $folder . "/app";
$stamp = function($ago) {
	return Task\Clock::real()->add(\Time\Duration::parse("-" . $ago . "s"))->format(Task\Clock::STORED);
};

// A prompt thread in a working run, with one reply, then a board note and a
// chat. The chat is the bell's and not the rail's.
$db->replace("note", [
	"id" => "fx-feed-thread",
	"channel" => "notes",
	"user_name" => "ops",
	"body" => "why did the deploy fail",
	"state" => "working",
	"run_id" => "fx-feed-run",
	"workspace_id" => "fx-feed",
	"dir" => $place,
	"created_at" => $stamp(30),
	"updated_at" => $stamp(30),
]);
$db->replace("note", [
	"id" => "fx-feed-reply",
	"channel" => "notes",
	"parent_id" => "fx-feed-thread",
	"user_name" => "claude",
	"body" => "Looking at the log.",
	"run_id" => "fx-feed-run",
	"workspace_id" => "fx-feed",
	"dir" => $place,
	"created_at" => $stamp(10),
	"updated_at" => $stamp(10),
]);
$db->replace("note", [
	"id" => "fx-feed-board",
	"channel" => "notes",
	"user_name" => "alice",
	"body" => "a board note",
	"created_at" => $stamp(60),
	"updated_at" => $stamp(60),
]);
$db->replace("note", [
	"id" => "fx-feed-chat",
	"channel" => "chat",
	"user_name" => "alice",
	"body" => "hello",
	"created_at" => $stamp(90),
	"updated_at" => $stamp(90),
]);

// The thread's run is Claude's and is dropped; an ordinary run stays.
$db->replace("run", [
	"id" => "fx-feed-run",
	"folder" => $folder,
	"job" => "@claude-prompt",
	"dir" => "app",
	"user_name" => "ops",
	"state" => "ok",
	"started_at" => $stamp(30),
	"term_cols" => 120,
	"term_rows" => 30,
]);
$db->replace("run", [
	"id" => "fx-feed-build",
	"folder" => $folder,
	"job" => "build",
	"dir" => "",
	"user_name" => "ops",
	"state" => "ok",
	"started_at" => $stamp(120),
	"term_cols" => 120,
	"term_rows" => 30,
]);

$ops = new Task\Model\User(["id" => "fx-feed-u", "name" => "ops", "role" => "runner", "admin" => true]);
$access = new Task\Access($ops);
$roles = ["ops" => "runner"];
$known = [Task\Fixture::folder($place, [])];
$rail = new Task\Recent($db);

$found = [];
foreach ($rail->entries($access, $roles, $known, []) as $entry) {
	$found[$entry["href"]] = $entry;
}

$thread = $found["/threads/fx-feed-thread"];

echo $thread["href"], " ", var_export($thread["thread"], true), " ", $thread["name"], " ", $thread["place"], " ", $thread["state"], " ", $thread["children"], "\n";
echo var_export(isset($found["/runs/fx-feed-run"]), true), " ", var_export(isset($found["/runs/fx-feed-build"]), true), "\n";

// Without prompt/view the thread is not listed; the board note and the run are.
$viewerUser = new Task\Model\User(["id" => "fx-feed-v", "name" => "viewer", "role" => "viewer"]);
$noteView = new Task\Policy(["note/view" => Task\Policy::ALLOW]);
$viewer = new Task\Access($viewerUser, $noteView);
$seen = array_column($rail->entries($viewer, [], $known, []), "href");

echo var_export(in_array("/threads/fx-feed-thread", $seen, true), true), " ", var_export(in_array("/threads/fx-feed-board", $seen, true), true), " ", var_export(in_array("/runs/fx-feed-run", $seen, true), true), " ", var_export(in_array("/runs/fx-feed-build", $seen, true), true), "\n";

// A place is visible to a reader whose folder list holds it, or holds none.
$listed = new Task\Model\User(["id" => "fx-feed-l", "name" => "lister", "role" => "runner", "folders" => '["fx-feed/app"]']);
$lister = new Task\Access($listed);

echo var_export(Task\Recent::seesPlace($access, Task\Route::INTERNAL, $known, []), true), "\n";
echo var_export(Task\Recent::seesPlace($lister, $place, $known, []), true), "\n";
echo var_export(Task\Recent::seesPlace($lister, "other/app", $known, []), true), "\n";
echo var_export(Task\Recent::threadState("working"), true), " ", var_export(Task\Recent::threadState("open"), true), " ", var_export(Task\Recent::threadState("done"), true), " ", Task\Recent::threadState("failed"), "\n";

// The feed lists the prompt and the board note, roots only, and the live
// list holds the thread the agent is in.
$notes = new Task\Threads\Notes($db);
$feed = array_filter($notes->feed(), function($note) {
	return strpos($note->getId(), "fx-feed-") === 0;
});
$live = array_filter($notes->live(), function($note) {
	return $note->getId() === "fx-feed-thread";
});
echo count($feed), " ", count($live), "\n";
---
/threads/fx-feed-thread true why did the deploy fail fx-feed/app running 1
false true
false true false true
true
true
false
'running' 'queued' 'ok' failed
2 1
