name: note
description: Task\Note reads the note table - the threads of a channel, the prompts about a run, the replies under either, the thread an agent worked a run in, and the cards the note component draws.
---
<?php

require "ext/autoload.php";

// One connection serves the suite and fixtures run together, so this fixture
// owns the rows named fx-note-* and reads no others. The channel column takes
// the board's channel or the chat's and nothing else, so a query cannot narrow
// to this fixture and mine() keeps what it owns.

$run = "fx-note-run";

$db = Task\DB::writer();

// Two board threads with a reply each, a thread of the other channel, and a
// prompt thread with its answer. Every note holds a channel, replies among
// them; a prompt is the one that holds a workspace and the run it is about.
$rows = array(
	array("id" => "fx-note-1", "channel" => "notes", "parent" => "", "body" => "first"),
	array("id" => "fx-note-2", "channel" => "notes", "parent" => "fx-note-1", "body" => "under the first"),
	array("id" => "fx-note-3", "channel" => "notes", "parent" => "", "body" => "second"),
	array("id" => "fx-note-4", "channel" => "notes", "parent" => "fx-note-3", "body" => "under the second", "run" => "fx-note-reply-run"),
	array("id" => "fx-note-5", "channel" => "chat", "parent" => "", "body" => "the other channel"),
	array("id" => "fx-note-6", "channel" => "notes", "parent" => "", "body" => "why did this run fail", "workspace" => "fx-note-ws", "about" => $run),
	array("id" => "fx-note-7", "channel" => "notes", "parent" => "fx-note-6", "body" => "it ran out of disk", "workspace" => "fx-note-ws", "about" => $run, "run" => "fx-note-answer-run"),
);

foreach ($rows as $row) {
	$db->replace("note", array(
		"id" => $row["id"],
		"channel" => $row["channel"],
		"parent_id" => $row["parent"],
		"user_name" => "admin",
		"body" => $row["body"],
		"mentions" => "",
		"state" => "",
		"run_id" => isset($row["run"]) ? $row["run"] : "",
		"about_run_id" => isset($row["about"]) ? $row["about"] : "",
		"workspace_id" => isset($row["workspace"]) ? $row["workspace"] : "",
		"created_at" => "2026-11-02 10:01:00.000 +0000 UTC",
		"updated_at" => "2026-11-02 10:01:00.000 +0000 UTC",
	));
}

// A page reads through the reader, which refuses a write.
$notes = new Task\Note;

// mine is the rows this fixture owns, in the order the query answered.
function mine($found) {
	$out = array();

	foreach ($found as $note) {
		if (strpos($note->getId(), "fx-note-") === 0) {
			$out[] = $note->getId();
		}
	}

	return implode(",", $out);
}

// The threads of a channel, newest first. The prompt is not among them: it
// holds a workspace, so it belongs to the page it was asked from and not to
// the board.
$roots = $notes->threads(Task\Note::NOTES);

echo mine($roots), "\n";
echo mine($notes->threads("chat")), "\n";

// ids is what replies() names, and the replies come back keyed by the thread
// they sit under.
echo implode(",", Task\Note::ids(array($notes->get("fx-note-1"), $notes->get("fx-note-3")))), "\n";

$replies = $notes->replies(array("fx-note-1", "fx-note-3"));

echo mine($replies["fx-note-1"]), " ", mine($replies["fx-note-3"]), "\n";
echo count($notes->replies(array())), " ", count($notes->replies(array("fx-note-nothing"))), "\n";

// The prompts about a run, which is the board's other half: a workspace, and
// the run the question is about.
echo mine($notes->prompts($run)), "\n";
echo "[", mine($notes->prompts("fx-note-no-run")), "]\n";
echo mine($notes->replies(array("fx-note-6"))["fx-note-6"]), "\n";

// One note by ID, and null for an ID that names none.
echo $notes->get("fx-note-6")->getBody(), "\n";
echo var_export($notes->get("fx-note-nothing"), true), "\n";

// The thread an agent worked a run in. The run belongs to a reply, so the
// thread above it is what the page draws; a run that names none answers null.
echo $notes->byRun("fx-note-answer-run")->getId(), "\n";
echo $notes->byRun("fx-note-reply-run")->getId(), "\n";
echo var_export($notes->byRun(""), true), " ", var_export($notes->byRun("fx-note-no-run"), true), "\n";

// The writes, through the writer. A reply holds the channel of the thread it
// sits under: the column takes the board's channel or the chat's, and a thread
// is told from a reply by the parent.
$writes = new Task\Note($db);
$posted = $writes->post("admin", "a written note for @admin");
$reply = $writes->post("admin", "a written reply", $posted->getId());

echo $posted->getChannel(), " ", $reply->getChannel(), " ", var_export($reply->getParentId() === $posted->getId(), true), "\n";

// The column is read off the body by the write itself, spaced, so a LIKE for
// "% name %" matches a whole name and not a prefix of one. A body that names
// nobody writes none.
$held = $writes->get($posted->getId());

echo "[", $held->getMentions(), "] [", $writes->get($reply->getId())->getMentions(), "]\n";
echo "[", Task\Note::mentions("A note for @admin and @ops"), "]\n";
echo "[", Task\Note::mentions("@admin @admin twice is once"), "]\n";
echo "[", Task\Note::mentions("@a.b @c-d @e_f are all names"), "]\n";

// An edit rewrites the column with the body, so a name added by a rewrite is
// found and one taken out is not.
$writes->edit($posted->getId(), "rewritten for @ops");

$edited = $writes->get($posted->getId());

echo $edited->getBody(), " [", $edited->getMentions(), "] ", var_export($edited->getEditedAt() !== "", true), "\n";

// Dropping a thread takes the replies under it with it.
$writes->drop($posted->getId());

echo var_export($writes->get($posted->getId()), true), " ", var_export($writes->get($reply->getId()), true), "\n";

// The cards the note component draws, which the board and a run's prompt box
// both read. The reader is an admin, so every note is theirs to rewrite; the
// agent is the user whose mentions run one, and the run a reply worked in is a
// row, where the run the prompt's answer names is none.
$db->replace("user", array("id" => "fx-note-u1", "name" => "fx-note-ann", "role" => "admin", "admin" => 1));
$db->replace("user", array("id" => "fx-note-u2", "name" => "fx-note-bot", "role" => "runner", "agent_reply" => 1, "mcp" => 1));

$db->replace("run", array("id" => "fx-note-reply-run", "folder" => "fx-note", "job" => "build", "user_name" => "fx-note-ann", "state" => "ok"));

// A board thread that names the agent, the agent's answer under it, and a
// reply of the reader's own.
$cards = array(
	array("id" => "fx-note-8", "parent" => "", "by" => "fx-note-ann", "body" => "@fx-note-bot look at the build", "state" => "working", "run" => ""),
	array("id" => "fx-note-9", "parent" => "fx-note-8", "by" => "fx-note-bot", "body" => "the cache was cold", "state" => "", "run" => "fx-note-reply-run"),
	array("id" => "fx-note-10", "parent" => "fx-note-8", "by" => "fx-note-ann", "body" => "thanks", "state" => "", "run" => ""),
);

foreach ($cards as $row) {
	$db->replace("note", array(
		"id" => $row["id"],
		"channel" => "notes",
		"parent_id" => $row["parent"],
		"user_name" => $row["by"],
		"body" => $row["body"],
		"mentions" => "",
		"state" => $row["state"],
		"run_id" => $row["run"],
		"about_run_id" => "",
		"workspace_id" => "",
		"created_at" => "2026-11-02 10:02:00.000 +0000 UTC",
		"updated_at" => "2026-11-02 10:02:00.000 +0000 UTC",
	));
}

Task\Host::$whoami = "fx-note-ann";

$chrome = new Task\Chrome;
$seen = new Task\Seen($chrome);

// labels names the items a card's menu offers, in the order it offers them.
function labels($menu) {
	$out = array();

	foreach ($menu["groups"] as $group) {
		foreach ($group as $item) {
			$out[] = $item["label"];
		}
	}

	return implode(",", $out);
}

// agents is the agents this fixture wrote, because the query reads every user
// an agent may reply as and another fixture's agent is as real as this one.
function agents() {
	$users = new Task\User;
	$out = array();

	foreach ($users->agents() as $agent) {
		if (strpos($agent->getName(), "fx-note-") === 0) {
			$out[] = $agent;
		}
	}

	return $out;
}

// The names a note names with @, once each in the order the text names them. A
// job ID is no mention, and a name that ends a sentence drops the dot.
echo "[", implode(",", Task\Note::names("@admin and @ops, ask @admin again")), "]\n";
echo "[", implode(",", Task\Note::names("run @host:dev1 and tell @fx-note-bot.")), "]\n";
echo "[", implode(",", Task\Note::names("nobody is named here")), "]\n";

// The agent whose replies answer a thread: claude on a prompt, the agent an
// instruction names on a board note, and nobody where no agent is named.
$agents = agents();
$under = $notes->replies(array("fx-note-8"))["fx-note-8"];

echo "[", Task\Note::answerer($notes->get("fx-note-6"), array(), $agents), "]\n";
echo "[", Task\Note::answerer($notes->get("fx-note-8"), $under, $agents), "]\n";
echo "[", Task\Note::answerer($notes->get("fx-note-1"), $notes->replies(array("fx-note-1"))["fx-note-1"], $agents), "]\n";
echo "[", Task\Note::agentIn("fx-note-bot", "@fx-note-bot talking to itself", $agents), "]\n";

// One thread as the component reads it: the note on top, its replies oldest
// first, and the field under it where the reader may write one.
$thread = Task\Note::thread($notes->get("fx-note-8"), $notes->replies(array("fx-note-8")), $chrome, "fx-note-bot", $seen);

echo $thread["id"], " ", var_export($thread["working"], true), " ", count($thread["replies"]), " ", $thread["author"]["name"], " ", var_export($thread["edited"], true), "\n";
echo labels($thread["menu"]), "\n";

// The agent's answer is no row to rewrite: it changes when its question runs
// again. The run it worked in links, because the reader sees that run.
$answer = $thread["replies"][0];

echo $answer["id"], " ", var_export($answer["edits"], true), " ", $answer["run"], " ", $answer["short"], "\n";
echo labels($answer["menu"]), "\n";

// A reply that is nobody's answer is the reader's to rewrite, and links no run.
echo $thread["replies"][1]["id"], " ", var_export($thread["replies"][1]["edits"], true), " [", $thread["replies"][1]["run"], "]\n";
echo labels($thread["replies"][1]["menu"]), "\n";

// A card links the run its agent worked in only for a reader who sees it: the
// run of the prompt's answer is gone, so the card with a seen links nowhere
// and the card without one links it.
echo "[", Task\Note::card($notes->get("fx-note-7"), $chrome, $notes->get("fx-note-6"), Task\Note::CLAUDE, $seen)["run"], "]\n";
echo "[", Task\Note::card($notes->get("fx-note-7"), $chrome, $notes->get("fx-note-6"), Task\Note::CLAUDE)["run"], "]\n";

// A prompt belongs to the page it was asked from, so removing it needs the
// prompt permission and removing a board note the note one. An admin holds
// both; nobody signed in holds neither, and rewrites nothing.
echo var_export(Task\Note::isPrompt($notes->get("fx-note-6")), true), " ", var_export(Task\Note::isPrompt($notes->get("fx-note-8")), true), "\n";
echo var_export(Task\Note::deletes($notes->get("fx-note-6"), $chrome), true), " ", var_export(Task\Note::deletes($notes->get("fx-note-8"), $chrome), true), "\n";

Task\Host::$whoami = "";

$nobody = new Task\Chrome;

echo var_export(Task\Note::edits($notes->get("fx-note-8"), $nobody, $notes->get("fx-note-8"), ""), true), " ", var_export(Task\Note::deletes($notes->get("fx-note-8"), $nobody), true), "\n";
echo "[", labels(Task\Note::card($notes->get("fx-note-8"), $nobody, $notes->get("fx-note-8"), "", $seen)["menu"]), "]\n";
---
fx-note-3,fx-note-1
fx-note-5
fx-note-1,fx-note-3
fx-note-2 fx-note-4
0 0
fx-note-6
[]
fx-note-7
why did this run fail
NULL
fx-note-6
fx-note-3
NULL NULL
notes notes true
[ admin ] []
[ admin ops ]
[ admin ]
[ a.b c-d e_f ]
rewritten for @ops [ ops ] true
NULL NULL
[admin,ops]
[fx-note-bot]
[]
[claude]
[fx-note-bot]
[]
[]
fx-note-8 true 2 fx-note-ann false
Edit,Delete
fx-note-9 false fx-note-reply-run fx-note-
Open its run,Delete
fx-note-10 true []
Edit,Delete
[]
[fx-note-answer-run]
true false
true true
false false
[]
