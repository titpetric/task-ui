name: thread-work
description: Task\Threads\Work groups a thread's events by the message that started each run.
---
<?php

require "ext/autoload.php";

$db = Task\DB::writer();
$db->query("DELETE FROM note WHERE id LIKE 'fx-work-%'");
$db->query("DELETE FROM note_event WHERE note_id LIKE 'fx-work-%'");
$db->query("DELETE FROM user WHERE id LIKE 'fx-work-%'");
$db->replace("user", ["id" => "fx-work-u1", "name" => "fx-work-ann", "role" => "admin", "admin" => 1]);
$db->replace("note", [
	"id" => "fx-work-n1",
	"channel" => "notes",
	"parent_id" => "",
	"user_id" => "fx-work-u1",
	"user_name" => "fx-work-ann",
	"body" => "why did the build fail",
	"mentions" => "",
	"state" => "done",
	"workspace_id" => "fx-work-ws",
	"dir" => "fx-work/app",
	"created_at" => "2026-11-16 10:00:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:00:00.000 +0000 UTC",
]);
$db->replace("note", [
	"id" => "fx-work-n2",
	"channel" => "notes",
	"parent_id" => "fx-work-n1",
	"user_id" => "fx-work-u1",
	"user_name" => "fx-work-ann",
	"body" => "and the tests",
	"mentions" => "",
	"state" => "done",
	"created_at" => "2026-11-16 10:01:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:01:00.000 +0000 UTC",
]);
$db->replace("note_event", [
	"id" => "fx-work-e1", "note_id" => "fx-work-n1", "run_id" => "fx-work-r1",
	"question_id" => "fx-work-n1", "seq" => 1, "kind" => "text",
	"tool_id" => "", "body" => "**on it**", "data" => "{}",
]);
$db->replace("note_event", [
	"id" => "fx-work-e2", "note_id" => "fx-work-n1", "run_id" => "fx-work-r1",
	"question_id" => "fx-work-n1", "seq" => 2, "kind" => "tool_use",
	"tool_id" => "t1", "body" => "", "data" => "{\"name\":\"Bash\",\"input\":{\"command\":\"go test ./...\"}}",
]);
$db->replace("note_event", [
	"id" => "fx-work-e3", "note_id" => "fx-work-n1", "run_id" => "fx-work-r1",
	"question_id" => "fx-work-n1", "seq" => 3, "kind" => "tool_result",
	"tool_id" => "t1", "body" => "ok", "data" => "{\"is_error\":false}",
]);
$db->replace("note_event", [
	"id" => "fx-work-e4", "note_id" => "fx-work-n1", "run_id" => "fx-work-r2",
	"question_id" => "fx-work-n2", "seq" => 1, "kind" => "error",
	"tool_id" => "", "body" => "it ran out of disk", "data" => "{}",
]);
$db->replace("note_event", [
	"id" => "fx-work-e5", "note_id" => "fx-work-n1", "run_id" => "fx-work-r2",
	"question_id" => "fx-work-n2", "seq" => 2, "kind" => "status",
	"tool_id" => "", "body" => "Running Bash", "data" => "{\"state\":\"tool\",\"tool\":\"Bash\"}",
]);

$notes = new Task\Threads\Notes(Task\DB::writer());
$chrome = new Task\Chrome;
$events = $notes->events("fx-work-n1");
$work = Task\Threads\Work::byQuestion($events, $chrome);
$messages = Task\Threads\Talk::messages($notes->get("fx-work-n1"), [$notes->get("fx-work-n2")], $chrome, null, $work);

echo count($events), "\n";
echo count($work["fx-work-n1"]), " ", count($work["fx-work-n2"]), "\n";
echo $work["fx-work-n1"][0]["events"][0]["kind"], " ", var_export(strpos($work["fx-work-n1"][0]["events"][0]["prose"], "on it") !== false, true), "\n";
echo $work["fx-work-n1"][0]["events"][1]["tool"], " ", var_export(strpos($work["fx-work-n1"][0]["events"][1]["input"], "go test") !== false, true), "\n";
echo var_export($work["fx-work-n2"][0]["events"][0]["error"], true), "\n";
echo $work["fx-work-n2"][0]["status"]["body"], " ", $work["fx-work-n2"][0]["status"]["state"], " ", $work["fx-work-n2"][0]["status"]["tool"], "\n";
echo count($work["fx-work-n2"][0]["events"]), " ", $work["fx-work-n2"][0]["events"][0]["seq"], "\n";
echo count($messages[0]["work"]), " ", count($messages[1]["work"]), "\n";

// The reply a run wrote is the agent's, not the question that started the run.
$db->replace("note", [
	"id" => "fx-work-n4",
	"channel" => "notes",
	"parent_id" => "fx-work-n1",
	"user_id" => "fx-work-u1",
	"user_name" => "fx-work-ann",
	"body" => "the tests pass",
	"mentions" => "",
	"state" => "",
	"run_id" => "fx-work-r1",
	"created_at" => "2026-11-16 10:03:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:03:00.000 +0000 UTC",
]);
$replies = [$notes->get("fx-work-n2"), $notes->get("fx-work-n4")];
$answers = Task\Threads\Conversation::answers($replies, $work);
$talk = Task\Threads\Talk::messages($notes->get("fx-work-n1"), $replies, $chrome, null, $work, $answers);

echo count($answers), " ", var_export($talk[0]["agent"], true), " ", var_export($talk[1]["agent"], true), " ", var_export($talk[2]["agent"], true), "\n";

// A working thread carries the run it may end, and the last status as its
// pinned line.
$db->query("DELETE FROM run WHERE id LIKE 'fx-work-%'");
$db->replace("run", [
	"id" => "fx-work-r9", "folder" => "fx-work/app", "job" => "@claude-prompt",
	"user_id" => "fx-work-u1", "user_name" => "fx-work-ann", "state" => "running",
	"started_at" => "2026-11-16 10:02:00.000 +0000 UTC",
	"created_at" => "2026-11-16 10:02:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:02:00.000 +0000 UTC",
]);
$db->replace("note", [
	"id" => "fx-work-n3",
	"channel" => "notes",
	"parent_id" => "",
	"user_id" => "fx-work-u1",
	"user_name" => "fx-work-ann",
	"body" => "still going",
	"mentions" => "",
	"state" => "working",
	"run_id" => "fx-work-r9",
	"created_at" => "2026-11-16 10:02:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:02:00.000 +0000 UTC",
]);
$status = Task\Threads\Conversation::status($notes->get("fx-work-n3"), [], $chrome, []);

echo var_export($status["working"], true), " ", $status["run"], " ", $status["started"], " ", var_export($status["can_end"], true), "\n";

$notes->forget("fx-work-n1");
echo count($notes->events("fx-work-n1")), "\n";
---
5
1 1
text true
Bash true
true
Running Bash tool Bash
1 1
1 1
1 false false true
true fx-work-r9 1794823320000 false
0
