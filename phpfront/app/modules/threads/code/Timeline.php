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
	static function row($note, $replies, $chrome, $seen = null, $selectedID = "", $usage = null) {
		$id = $note->getId();
		$run = $note->getRunId();
		$selected = $id === $selectedID;
		$unread = 0;
		$classes = ["thread-row"];

		if ($seen !== null && !$seen->run($run)) {
			$run = "";
		}

		if ($note->getState() !== "") {
			$classes[] = $note->getState();
		}

		if ($unread) {
			$classes[] = "unread";
		}

		if ($selected) {
			$classes[] = "selected";
		}

		$at = self::lastAt($note, $replies);
		$count = count($replies) - (isset($replies[0]) && Talk::joins($note, $replies[0]) ? 1 : 0);

		return [
			"id" => $id,
			"className" => implode(" ", $classes),
			"href" => \Task\Route::href("/threads/" . \Task\Route::pathEscape($id)),
			"kind" => self::kind($note),
			"person" => $chrome->who()->of($note->getUserName()),
			"said" => Conversation::title($note, $replies),
			"state" => $note->getState(),
			"working" => $note->getState() === \Task\Note::WORKING,
			"place" => $note->getDir(),
			"run" => $run,
			"short" => substr($run, 0, 8),
			"can_end" => $run !== "" && $note->getState() === \Task\Note::WORKING && self::ends($chrome, $run),
			"ago" => \Task\Clock::ago($at),
			"stamp" => \Task\Clock::read($at),
			"messages" => $count + 1,
			"replies" => $count,
			"tools" => is_array($usage) ? (int)($usage["tools"] ?? 0) : 0,
			"tokens" => is_array($usage) ? (int)($usage["tokens"] ?? 0) : 0,
			"cost" => is_array($usage) && (float)($usage["cost"] ?? 0) > 0 ? "$" . number_format((float)$usage["cost"], 2) : "",
			"context" => self::context($usage),
			"heard" => self::heard($note, $replies),
			"unread" => $unread,
			"selected" => $selected,
		];
	}

	/** visible reports whether the reader may read a root note of the feed. */
	static function visible($note, $chrome) {
		if (\Task\Note::isPrompt($note)) {
			return $chrome->allowed("prompt/view") && \Task\Note::seesPlace($chrome, $note->getDir());
		}

		return $chrome->allowed("note/view");
	}

	/** chat is one chat thread as components/thread-row.tpl reads it. */
	static function chat($row, $chrome, $selectedID = "") {
		$said = $row["last_spoken"] !== "" ? $row["last_spoken"] : $row["last_body"];
		$classes = ["thread-row"];

		if ((int)$row["unread"] > 0) {
			$classes[] = "unread";
		}

		if ($row["id"] === $selectedID) {
			$classes[] = "selected";
		}

		return [
			"id" => $row["id"],
			"className" => implode(" ", $classes),
			"href" => \Task\Route::href("/threads/" . \Task\Route::pathEscape($row["id"])),
			"kind" => "chat",
			"person" => $chrome->who()->of($row["peer"]),
			"said" => \Task\Note::excerpt($said),
			"state" => "",
			"working" => false,
			"place" => "",
			"run" => "",
			"short" => "",
			"can_end" => false,
			"messages" => 0,
			"replies" => 0,
			"ago" => \Task\Clock::ago($row["last_at"]),
			"stamp" => \Task\Clock::read($row["last_at"]),
			"heard" => $row["last_audio_path"] !== "",
			"unread" => (int)$row["unread"],
			"selected" => $row["id"] === $selectedID,
		];
	}

	/** context is the share of the context window a thread used, as a row scans it. */
	static function context($usage) {
		$tokens = is_array($usage) ? (int)($usage["tokens"] ?? 0) : 0;

		return $tokens > 0 ? Work::size($tokens) . " ctx" : "";
	}

	/** ends reports whether the reader may stop the run behind a working thread. */
	static function ends($chrome, $id) {
		$runs = new \Task\Runs\Runs;
		$run = $runs->get($id);

		return $run !== null && !\Task\Run::ended($run->getState()) && \Task\Run::canEnd($chrome, $run);
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
