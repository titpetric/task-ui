name: talk
description: Task\Threads\Talk shows a spoken note as its transcript and links the raw recording.
---
<?php

require "ext/autoload.php";

// A note the orb left when no agent answered: the body holds the prompt's
// words, the spoken column the transcript, and the audio path the clip.
$db = Task\DB::writer();
$db->query("DELETE FROM note WHERE id LIKE 'fx-talk-%'");
$db->query("DELETE FROM user WHERE id LIKE 'fx-talk-%'");
$db->replace("user", array("id" => "fx-talk-u1", "name" => "fx-talk-ann", "role" => "admin", "admin" => 1));
$db->replace("note", array(
	"id" => "fx-talk-1",
	"channel" => "notes",
	"parent_id" => "",
	"user_id" => "fx-talk-u1",
	"user_name" => "fx-talk-ann",
	"body" => "The recording is at voice/2026-11-16T10-00-00.wav, already transcribed as the words above.",
	"mentions" => "",
	"state" => "",
	"audio_path" => "voice/2026-11-16T10-00-00.wav",
	"spoken" => "ship the release",
	"created_at" => "2026-11-16 10:00:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:00:00.000 +0000 UTC",
));

Task\Host::$whoami = "fx-talk-ann";

$notes = new Task\Threads\Notes;
$chrome = new Task\Chrome;
$seen = new Task\Seen($chrome);
$note = $notes->get("fx-talk-1");
$messages = Task\Threads\Talk::messages($note, array(), $chrome, $seen);

echo Task\Threads\Talk::said($note), "\n";
echo Task\Threads\Talk::audio($note), "\n";
echo count($messages), " ", $messages[0]["audio"], "\n";
echo var_export(strpos($messages[0]["body"], "ship the release") !== false, true), "\n";

// A clip whose transcript was posted as the next message is one message.
$db->replace("note", array(
	"id" => "fx-talk-2",
	"channel" => "notes",
	"parent_id" => "",
	"user_id" => "fx-talk-u1",
	"user_name" => "fx-talk-ann",
	"body" => "A recording.",
	"mentions" => "",
	"state" => "",
	"audio_path" => "voice/2026-11-16T10-00-01.wav",
	"spoken" => "",
	"created_at" => "2026-11-16 10:01:00.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:01:00.000 +0000 UTC",
));
$db->replace("note", array(
	"id" => "fx-talk-3",
	"channel" => "notes",
	"parent_id" => "fx-talk-2",
	"user_id" => "fx-talk-u1",
	"user_name" => "fx-talk-ann",
	"body" => "ship the release",
	"mentions" => "",
	"state" => "",
	"audio_path" => "",
	"spoken" => "ship the release",
	"created_at" => "2026-11-16 10:01:01.000 +0000 UTC",
	"updated_at" => "2026-11-16 10:01:01.000 +0000 UTC",
));
$joined = Task\Threads\Talk::messages($notes->get("fx-talk-2"), array($notes->get("fx-talk-3")), $chrome, $seen);
echo count($joined), " ", $joined[0]["audio"], " ", var_export(strpos($joined[0]["body"], "ship the release") !== false, true), "\n";
echo Task\Threads\Talk::opening($notes->get("fx-talk-2"), array($notes->get("fx-talk-3"))), "\n";

Task\Host::$whoami = "";
---
ship the release
/artifacts/files/voice/2026-11-16T10-00-00.wav?raw
1 /artifacts/files/voice/2026-11-16T10-00-00.wav?raw
true
1 /artifacts/files/voice/2026-11-16T10-00-01.wav?raw true
ship the release
