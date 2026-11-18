<?php

namespace Task\Threads;

// Work is what Claude did on one thread: each run's status, text, thinking,
// tool calls and results, grouped by the message that started the run.
class Work {
	/** CUT is how many bytes of a tool's input or result a card shows. */
	const CUT = 4000;

	/** byQuestion is the thread's runs, by the ID of the message that started each. */
	static function byQuestion($events, $chrome) {
		$order = [];
		$runs = [];

		foreach ($events as $row) {
			$question = isset($row["question_id"]) ? $row["question_id"] : "";
			$run = isset($row["run_id"]) ? $row["run_id"] : "";
			$key = $question . "\x1f" . $run;

			if (!isset($runs[$key])) {
				$order[] = $key;
				$runs[$key] = self::run($question, $run);
			}

			$runs[$key]["events"][] = self::event($row, $chrome);
		}

		$out = [];

		foreach ($order as $key) {
			$out[$runs[$key]["question"]][] = $runs[$key];
		}

		return $out;
	}

	/** run is one run of a thread as the template reads it. */
	static function run($question, $id) {
		return [
			"question" => $question,
			"id" => $id,
			"short" => substr($id, 0, 8),
			"events" => [],
		];
	}

	/** event is one record of a run as the template reads it. */
	static function event($row, $chrome) {
		$kind = isset($row["kind"]) ? $row["kind"] : "";
		$body = isset($row["body"]) ? $row["body"] : "";
		$data = json_decode(isset($row["data"]) ? $row["data"] : "", true);

		if (!is_array($data)) {
			$data = [];
		}

		$out = [
			"kind" => $kind,
			"body" => $body,
			"prose" => "",
			"tool" => "",
			"input" => "",
			"error" => false,
			"state" => "",
		];

		switch ($kind) {

			case "text":
			case "result":
				$out["prose"] = $chrome->message($body);
				break;
			case "tool_use":
				$out["tool"] = isset($data["name"]) ? $data["name"] : "Tool";
				$out["input"] = self::json(isset($data["input"]) ? $data["input"] : []);
				break;
			case "tool_result":
				$out["body"] = self::cut($body);
				$out["error"] = !empty($data["is_error"]);
				break;
			case "status":
				$out["state"] = isset($data["state"]) ? $data["state"] : "";

				if ($out["body"] === "") {
					$out["body"] = $out["state"];
				}

				break;
			case "error":
				$out["error"] = true;
				break;
		}

		return $out;
	}

	/** json is a tool's input as the card shows it. */
	static function json($value) {
		$text = json_encode($value);

		return $text === false ? "" : self::cut($text);
	}

	/** cut is text as a card holds it, marked where it was cut. */
	static function cut($text) {
		if (strlen($text) <= self::CUT) {
			return $text;
		}

		return substr($text, 0, self::CUT) . "\n... (truncated)";
	}
}
