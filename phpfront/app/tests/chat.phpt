name: chat
description: A chat thread is private to its participants, listed with its peer and unread count, off the board timeline, and its recordings play from the chat route.
---
<?php

require "ext/autoload.php";

$db = Task\DB::writer();
$db->query("DELETE FROM note_participant WHERE note_id LIKE 'fx-chat-%'");
$db->query("DELETE FROM note WHERE id LIKE 'fx-chat-%'");
$db->query("DELETE FROM user WHERE id LIKE 'fx-chat-%'");

$db->replace("user", ["id" => "fx-chat-u1", "name" => "fx-chat-ann", "role" => "admin", "admin" => 1]);
$db->replace("user", ["id" => "fx-chat-u2", "name" => "fx-chat-bob", "role" => "admin", "admin" => 1]);

$root = "fx-chat-1";
$note = [
	"id" => $root,
	"channel" => "chat",
	"parent_id" => "",
	"user_id" => "fx-chat-u1",
	"user_name" => "fx-chat-ann",
	"body" => "hello",
	"mentions" => "",
	"state" => "",
	"audio_path" => "",
	"spoken" => "",
	"created_at" => "2020-01-01 10:00:00.000 +0000 UTC",
	"updated_at" => "2020-01-01 10:00:00.000 +0000 UTC",
];
$db->replace("note", $note);
$db->replace("note", array_merge($note, [
	"id" => "fx-chat-2",
	"parent_id" => $root,
	"user_id" => "fx-chat-u2",
	"user_name" => "fx-chat-bob",
	"body" => "a reply",
	"audio_path" => "voice/fx-chat.wav",
	"created_at" => "2020-01-01 10:01:00.000 +0000 UTC",
	"updated_at" => "2020-01-01 10:01:00.000 +0000 UTC",
]));
$db->replace("note_participant", ["note_id" => $root, "user_id" => "fx-chat-u1", "user_name" => "fx-chat-ann", "created_at" => "2020-01-01 10:00:00.000 +0000 UTC"]);
$db->replace("note_participant", ["note_id" => $root, "user_id" => "fx-chat-u2", "user_name" => "fx-chat-bob", "created_at" => "2020-01-01 10:00:00.000 +0000 UTC"]);

$notes = new Task\Threads\Notes;

echo var_export($notes->participant($root, "fx-chat-u1"), true), "\n";
echo var_export($notes->participant($root, "fx-chat-u3"), true), "\n";
echo $notes->peer($root, "fx-chat-u1"), "\n";
echo $notes->unread($root, "fx-chat-u1"), "\n";
$chats = $notes->chats("fx-chat-u1");
echo count($chats), " ", $chats[0]["peer"], " ", $chats[0]["unread"], "\n";
foreach ($notes->timeline() as $thread) {
	if ($thread->getId() === $root) {
		echo "board\n";
	}
}
echo Task\Threads\Talk::audio($notes->get("fx-chat-2")), "\n";
echo count(Task\Threads\Talk::spokenFor($notes->get($root), 1)), "\n";
---
true
false
fx-chat-bob
1
1 fx-chat-bob 1
/api/threads/messages/fx-chat-2/audio
0
