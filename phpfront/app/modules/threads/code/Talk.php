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

	/** RECORDING is the body a clip without a transcript carries. */
	const RECORDING = "A recording.";

	/**
	 * messages is the thread and its replies, oldest first, as the template
	 * reads them.
	 *
	 * An answer the session read out is named by the recording the question
	 * was spoken into, `<clip>.<n>.reply.wav`, which is what the brief in
	 * docs/guide/voice.md tells the session to write. The row does not hold
	 * it, so the folder is listed once and the answers take what is there.
	 */
	static function messages($note, $replies, $chrome, $seen = null, $work = [], $answers = [], $status = []) {
		$out = [self::message($note, $chrome, $seen)];
		$spoken = self::spokenFor($note, count($replies));
		$at = 0;

		foreach ($replies as $reply) {
			$at = $at + 1;
			$said = self::message($reply, $chrome, $seen);
			$said["agent"] = $said["agent"] || in_array($reply->getId(), $answers, true);
			$said["className"] = self::className($said["mine"], $said["agent"], $said["working"]);

			if ($said["audio"] === "" && isset($spoken[$at])) {
				$said["audio"] = $spoken[$at];
			}

			if ($at === 1 && $said["audio"] === "" && self::joins($note, $reply)) {
				$out[0]["body"] = $said["body"];

				continue;
			}

			$out[] = $said;
		}

		self::hand($out, $answers);
		self::attach($out, $work, $answers, $status);

		return $out;
	}

	/** hand hides a question's run once its answer carries it. */
	static function hand(&$out, $answers) {
		foreach ($out as $index => $message) {
			$run = $message["run"];

			if ($run !== "" && isset($answers[$run]) && $answers[$run] !== $message["id"]) {
				$out[$index]["state"] = "";
				$out[$index]["working"] = false;
				$out[$index]["run"] = "";
				$out[$index]["short"] = "";
				$out[$index]["className"] = self::className($message["mine"], $message["agent"], false);
			}
		}
	}

	/** attach puts each run's work under the answer it wrote, and a live run's state in its bubble. */
	static function attach(&$out, $work, $answers, $status = []) {
		$at = [];

		foreach ($out as $index => $message) {
			$at[$message["id"]] = $index;
			$out[$index]["work"] = [];
		}

		foreach ($work as $question => $runs) {
			foreach ($runs as $run) {
				$id = isset($answers[$run["id"]]) ? $answers[$run["id"]] : $question;

				if (!isset($at[$id])) {
					$id = $question;
				}

				// The trace is the agent's work: a message somebody wrote is never enriched with it.
				if (isset($at[$id]) && $out[$at[$id]]["agent"]) {
					$out[$at[$id]]["work"][] = $run;

					if ($out[$at[$id]]["working"]) {
						$out[$at[$id]]["working_view"] = self::live($run, $status);
					} elseif ($out[$at[$id]]["body"] === "" && $run["progress"] !== "") {
						// A run that ended without an answer still printed its last text.
						$out[$at[$id]]["body"] = \Task\Host::markdown($run["progress"]);
					}
				}
			}
		}
	}

	/** live is a working run as its answer draws it: state, latest text, tools and subagents. */
	static function live($run, $status) {
		$label = isset($status["label"]) && $status["label"] !== "" ? $status["label"] : "Working";
		$progress = $run["progress"];

		return [
			"label" => $label,
			"state" => isset($status["state"]) ? $status["state"] : "working",
			"started" => isset($status["started"]) ? $status["started"] : "",
			"run" => $run["id"],
			"canEnd" => !empty($status["can_end"]),
			"agents" => $run["agents"],
			"todos" => isset($run["todos"]) ? $run["todos"] : [],
			"progress" => $progress,
			"progressHtml" => $progress === "" ? "" : self::quote($progress),
			"progressClass" => "said-text" . ($progress === "" ? " faint" : ""),
			"stats" => $run["stats"],
			"tools" => $run["tools"],
		];
	}

	/** quote is the latest text of a live run drawn as one block quote, each line cited. */
	static function quote($text) {
		return \Task\Host::markdown("> " . str_replace("\n", "\n> ", trim($text)));
	}

	/** joins reports whether a note and the reply after it are one recording. */
	static function joins($note, $reply) {
		return $note->getAudioPath() !== "" && $note->getSpoken() === "" && $note->getBody() === self::RECORDING && $reply->getAudioPath() === "" && $reply->getSpoken() !== "" && $reply->getUserId() === $note->getUserId();
	}

	/** opening is what a thread is read as: the transcript of its first message. */
	static function opening($note, $replies) {
		if (isset($replies[0]) && self::joins($note, $replies[0])) {
			return self::said($replies[0]);
		}

		return self::said($note);
	}

	/**
	 * spokenFor is the answers of a voice thread that were read out, by their
	 * place in the thread: an empty list for a thread nobody spoke.
	 */
	static function spokenFor($note, $replies) {
		$clip = $note->getAudioPath();

		if ($clip === "" || $replies === 0 || \Task\Note::isChat($note)) {
			return [];
		}

		$folder = dirname($clip);
		// Host::artifacts answers the names a folder holds without the folder
		// in front, so the candidate is matched on the name and played on the
		// whole path.
		$base = substr(basename($clip), 0, strlen(basename($clip)) - 4);
		$held = \Task\Host::artifacts($folder);
		$out = [];

		for ($at = 1; $at <= $replies; $at = $at + 1) {
			$name = $base . "." . $at . ".reply.wav";

			if (in_array($name, $held, true)) {
				$out[$at] = self::url($folder . "/" . $name);
			}
		}

		return $out;
	}

	/** className is the class of a message: whose side it is on, and whether it is being written. */
	static function className($mine, $agent, $working) {
		$classes = ["said"];

		if ($mine) {
			$classes[] = "mine";
		}

		if ($agent) {
			$classes[] = "agent";
		}

		if ($working) {
			$classes[] = "working";
		}

		return implode(" ", $classes);
	}

	/** message is one thing somebody said, as the thread template reads it. */
	static function message($note, $chrome, $seen = null) {
		$author = $note->getUserName();
		$run = $note->getRunId();
		$mine = $author === $chrome->name();
		$agent = $note->getParentId() !== "" && $note->getAgent() !== "";
		// A question carries the working state of the run it started; only the agent's answer is written by it.
		$working = $agent && $note->getState() === \Task\Note::WORKING;

		if ($seen !== null && !$seen->run($run)) {
			$run = "";
		}

		return [
			"id" => $note->getId(),
			"className" => self::className($mine, $agent, $working),
			"mine" => $mine,
			"author" => $chrome->who()->of($author),
			"body" => $chrome->message(self::said($note), $note->getId()),
			"audio" => self::audio($note),
			"state" => $note->getState(),
			"working" => $working,
			"agent" => $agent,
			"run" => $run,
			"short" => substr($run, 0, 8),
			"work" => [],
			"working_view" => [],
			"ago" => \Task\Clock::ago($note->getCreatedAt()),
			"stamp" => \Task\Clock::read($note->getCreatedAt()),
		];
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

		if ($path === "") {
			return "";
		}

		if (\Task\Note::isChat($note)) {
			return "/api/threads/messages/" . rawurlencode($note->getId()) . "/audio";
		}

		return self::url($path);
	}

	/** url is a file of the artifact tree as the player reads it, by byte range. */
	static function url($path) {
		return self::FILES . implode("/", array_map("rawurlencode", explode("/", $path))) . "?raw";
	}
}
