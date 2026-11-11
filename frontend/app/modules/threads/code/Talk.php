<?php

namespace Task\Threads;

// Talk is one thread as a conversation: the messages in the order they were
// said, each on the side of whoever said it.
//
// The reader's own messages take one side and everyone else's the other, which
// is the shape a spoken thread wants: a question, an answer, a question. A
// message that was spoken carries the recording it was spoken into, and an
// answer that was read out carries what it was read out as, so a thread is
// read and listened to the same way it happened.
class Talk {
	/** FILES is where a recording of the artifact tree is played from. */
	const FILES = "/artifacts/files/";

	/**
	 * messages is the thread and its replies, oldest first, as the template
	 * reads them.
	 *
	 * An answer the session read out is named by the recording the question
	 * was spoken into, `<clip>.<n>.reply.wav`, which is what the brief in
	 * docs/guide/voice.md tells the session to write. The row does not hold
	 * it, so the folder is listed once and the answers take what is there.
	 */
	static function messages($note, $replies, $chrome, $seen = null) {
		$out = array(self::message($note, $chrome, $seen));
		$spoken = self::spokenFor($note, count($replies));
		$at = 0;

		foreach ($replies as $reply) {
			$at = $at + 1;
			$said = self::message($reply, $chrome, $seen);

			if ($said["audio"] === "" && isset($spoken[$at])) {
				$said["audio"] = $spoken[$at];
			}

			$out[] = $said;
		}

		return $out;
	}

	/**
	 * spokenFor is the answers of a voice thread that were read out, by their
	 * place in the thread: an empty list for a thread nobody spoke.
	 */
	static function spokenFor($note, $replies) {
		$clip = $note->getAudioPath();

		if ($clip === "" || $replies === 0) {
			return array();
		}

		$folder = dirname($clip);
		// Host::artifacts answers the names a folder holds without the folder
		// in front, so the candidate is matched on the name and played on the
		// whole path.
		$base = substr(basename($clip), 0, strlen(basename($clip)) - 4);
		$held = \Task\Host::artifacts($folder);
		$out = array();

		for ($at = 1; $at <= $replies; $at = $at + 1) {
			$name = $base . "." . $at . ".reply.wav";

			if (in_array($name, $held, true)) {
				$out[$at] = self::url($folder . "/" . $name);
			}
		}

		return $out;
	}

	/** message is one thing somebody said, as the thread template reads it. */
	static function message($note, $chrome, $seen = null) {
		$author = $note->getUserName();
		$run = $note->getRunId();

		if ($seen !== null && !$seen->run($run)) {
			$run = "";
		}

		return array(
			"id" => $note->getId(),
			"mine" => $author === $chrome->name(),
			"author" => $chrome->who()->of($author),
			"body" => $chrome->message(self::said($note)),
			"audio" => self::audio($note),
			"state" => $note->getState(),
			"working" => $note->getState() === \Task\Note::WORKING,
			"run" => $run,
			"short" => substr($run, 0, 8),
			"ago" => \Task\Clock::ago($note->getCreatedAt()),
			"stamp" => \Task\Clock::read($note->getCreatedAt()),
		);
	}

	/**
	 * said is what a message shows: the transcription where one was spoken,
	 * and the body otherwise.
	 *
	 * A spoken message's body carries what the agent needs beside the words -
	 * where the clip is, and the brief the session opens on - and none of that
	 * was said by anyone.
	 */
	static function said($note) {
		$spoken = $note->getSpoken();

		return $spoken === "" ? $note->getBody() : $spoken;
	}

	/** audio is the URL the message plays from, or the empty string where nothing was recorded. */
	static function audio($note) {
		$path = $note->getAudioPath();

		return $path === "" ? "" : self::url($path);
	}

	/** url is a file of the artifact tree as the player reads it, by byte range. */
	static function url($path) {
		return self::FILES . implode("/", array_map("rawurlencode", explode("/", $path))) . "?raw";
	}
}
