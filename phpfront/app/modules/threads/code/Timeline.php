<?php

namespace Task\Threads;

// Timeline is the list the Threads page draws: one row per thread, newest
// first, whatever started it.
//
// A row says who said the first thing, when, what it was, how it is going and
// how many messages came after. What it does not say is the conversation: that
// is Talk, on the thread's own page.
class Timeline {
	/** NOTE, PROMPT and VOICE are what started a thread, which the row names. */
	const NOTE = "note";

	const PROMPT = "prompt";

	const VOICE = "voice";

	/** row is one thread as components/thread-row.tpl reads it. */
	static function row($note, $replies, $chrome, $seen = null) {
		$id = $note->getId();
		$run = $note->getRunId();

		if ($seen !== null && !$seen->run($run)) {
			$run = "";
		}

		return [
			"id" => $id,
			"href" => \Task\Route::href("/threads/" . \Task\Route::pathEscape($id)),
			"kind" => self::kind($note),
			"author" => $chrome->who()->of($note->getUserName()),
			"said" => \Task\Note::excerpt(Talk::opening($note, $replies)),
			"state" => $note->getState(),
			"working" => $note->getState() === \Task\Note::WORKING,
			"place" => $note->getDir(),
			"run" => $run,
			"short" => substr($run, 0, 8),
			"ago" => \Task\Clock::ago(self::lastAt($note, $replies)),
			"stamp" => \Task\Clock::read(self::lastAt($note, $replies)),
			"messages" => count($replies) + 1 - (isset($replies[0]) && Talk::joins($note, $replies[0]) ? 1 : 0),
			"heard" => self::heard($note, $replies),
		];
	}

	/** visible reports whether the reader may read a root note of the feed. */
	static function visible($note, $chrome) {
		if (\Task\Note::isPrompt($note)) {
			return $chrome->allowed("prompt/view") && \Task\Note::seesPlace($chrome, $note->getDir());
		}

		return $chrome->allowed("note/view");
	}

	/** chat is one chat thread as the Chats section reads it. */
	static function chat($row, $chrome) {
		$said = $row["last_spoken"] !== "" ? $row["last_spoken"] : $row["last_body"];

		return [
			"id" => $row["id"],
			"href" => \Task\Route::href("/threads/" . \Task\Route::pathEscape($row["id"])),
			"peer" => $chrome->who()->of($row["peer"]),
			"said" => \Task\Note::excerpt($said),
			"ago" => \Task\Clock::ago($row["last_at"]),
			"stamp" => \Task\Clock::read($row["last_at"]),
			"heard" => $row["last_audio_path"] !== "",
			"unread" => (int)$row["unread"],
		];
	}

	/** kind names what started the thread: a recording, a prompt asked from a page, or a note. */
	static function kind($note) {
		if ($note->getAudioPath() !== "") {
			return self::VOICE;
		}

		return \Task\Note::isPrompt($note) ? self::PROMPT : self::NOTE;
	}

	/** heard reports whether any message of the thread was spoken or read out. */
	static function heard($note, $replies) {
		if ($note->getAudioPath() !== "") {
			return true;
		}

		foreach ($replies as $reply) {
			if ($reply->getAudioPath() !== "") {
				return true;
			}
		}

		return false;
	}

	/** lastAt is when the thread last moved: its newest reply, or the thread itself. */
	static function lastAt($note, $replies) {
		$at = $note->getCreatedAt();

		foreach ($replies as $reply) {
			$at = $reply->getCreatedAt();
		}

		return $at;
	}
}
