name: thread-row
description: Task\Threads\Timeline::row draws a thread as one list row, and a working row names the run its reader may stop.
---
<?php

require "ext/autoload.php";

$db = Task\DB::writer();
$db->query("DELETE FROM note WHERE id LIKE 'fx-row-%'");
$db->query("DELETE FROM run WHERE id LIKE 'fx-row-%'");
$db->query("DELETE FROM user WHERE id LIKE 'fx-row-%'");
$db->replace("user", ["id" => "fx-row-u1", "name" => "fx-row-ann", "role" => "admin", "admin" => 1]);
Task\Host::$whoami = "fx-row-ann";

// A thread whose run the reader may end, and one that is over.
$db->replace("run", [
	"id" => "fx-row-r1", "folder" => "internal", "job" => "@claude-prompt",
	"user_id" => "fx-row-u1", "user_name" => "fx-row-ann", "state" => "running",
	"started_at" => "2026-11-16 10:00:00.000 +0000 UTC",
	"created_at" => "2026-11-16 10:00:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:00:00.000 +0000 UTC",
]);
$db->replace("note", [
	"id" => "fx-row-n1", "channel" => "notes", "parent_id" => "",
	"user_id" => "fx-row-u1", "user_name" => "fx-row-ann",
	"body" => "the build is still going", "mentions" => "",
	"state" => "working", "run_id" => "fx-row-r1",
	"workspace_id" => "fx-row-ws", "dir" => "internal",
	"created_at" => "2026-11-16 10:00:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:00:00.000 +0000 UTC",
]);
$db->replace("note", [
	"id" => "fx-row-n2", "channel" => "notes", "parent_id" => "",
	"user_id" => "fx-row-u1", "user_name" => "fx-row-ann",
	"body" => "the build is green", "mentions" => "",
	"state" => "done", "run_id" => "",
	"workspace_id" => "fx-row-ws", "dir" => "internal",
	"created_at" => "2026-11-16 10:05:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:05:00.000 +0000 UTC",
]);

$notes = new Task\Threads\Notes(Task\DB::writer());
$chrome = new Task\Chrome;

$working = Task\Threads\Timeline::row($notes->get("fx-row-n1"), [], $chrome);
$done = Task\Threads\Timeline::row($notes->get("fx-row-n2"), [], $chrome);
$chat = Task\Threads\Timeline::chat([
	"id" => "fx-row-c1", "peer" => "fx-row-ann", "last_spoken" => "",
	"last_body" => "hello", "last_audio_path" => "", "last_at" => "2026-11-16 10:00:00.000 +0000 UTC",
	"unread" => 0,
], $chrome);

// The working row names the run it may stop; a thread that is over and a
// chat name none.
echo var_export($working["working"], true), " ", var_export($working["can_end"], true), " ", $working["run"], "\n";
echo var_export($done["working"], true), " ", var_export($done["can_end"], true), " ", $done["state"], "\n";
echo var_export($chat["can_end"], true), " ", $chat["run"], " ", $chat["kind"], "\n";
echo $working["href"], " ", $working["kind"], " ", $working["place"], " ", $working["short"], "\n";
---
true true fx-row-r1
false false done
false  chat
/threads/fx-row-n1 prompt internal fx-row-r
