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

		return array(
			"id" => $id,
			"href" => \Task\Route::href("/threads/" . \Task\Route::pathEscape($id)),
			"kind" => self::kind($note),
			"author" => $chrome->who()->of($note->getUserName()),
			"said" => \Task\Note::excerpt(Talk::said($note)),
			"state" => $note->getState(),
			"working" => $note->getState() === \Task\Note::WORKING,
			"run" => $run,
			"short" => substr($run, 0, 8),
			"ago" => \Task\Clock::ago(self::lastAt($note, $replies)),
			"stamp" => \Task\Clock::read(self::lastAt($note, $replies)),
			"messages" => count($replies) + 1,
			"heard" => self::heard($note, $replies),
		);
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
