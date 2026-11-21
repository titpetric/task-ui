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

		foreach ($runs as $key => $run) {
			$runs[$key]["stats"] = self::stats($run["events"]);
			$runs[$key]["tools"] = self::tools($run["events"]);
			$runs[$key]["progress"] = self::progress($run["events"]);
			$runs[$key]["status"] = self::status($run["events"]);
			$runs[$key]["agents"] = self::agents($run["events"]);
			$runs[$key]["todos"] = self::todos($run["events"]);
			$runs[$key]["trace"] = self::trace($run["events"]);
			$runs[$key]["events"] = self::withoutStatus($run["events"]);
		}

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
			"stats" => "",
			"tools" => 0,
			"progress" => "",
			"status" => [
				"body" => "",
				"state" => "",
				"tool" => "",
			],
			"agents" => [],
			"todos" => [],
			"trace" => self::noTrace(),
		];
	}

	/** agents is the subagents a run started, in order, with their task, kind and result. */
	static function agents($events) {
		$out = [];
		$at = [];

		foreach ($events as $event) {
			if ($event["kind"] === "tool_use" && $event["agent"] !== "") {
				$at[$event["tool_id"]] = count($out);
				$out[] = [
					"id" => $event["tool_id"],
					"task" => $event["agent"]["task"],
					"kind" => $event["agent"]["kind"],
					"prompt" => $event["agent"]["prompt"],
					"state" => "working",
					"result" => "",
				];

				continue;
			}

			if ($event["kind"] === "tool_result" && isset($at[$event["tool_id"]])) {
				$item = $at[$event["tool_id"]];
				$out[$item]["state"] = $event["error"] ? "failed" : "done";
				$out[$item]["result"] = $event["body"];
			}
		}

		foreach ($out as $index => $agent) {
			$out[$index]["className"] = "status-agent " . $agent["state"];
			$out[$index]["badge"] = $agent["state"] === "working" ? "badge signal" : ($agent["state"] === "failed" ? "badge bad" : "badge ok");
		}

		return $out;
	}

	/** todoList is the checklist a TodoWrite call wrote, as the answer draws it. */
	static function todoList($input) {
		$out = [];

		if (!is_array($input) || !isset($input["todos"]) || !is_array($input["todos"])) {
			return $out;
		}

		foreach ($input["todos"] as $todo) {
			if (!is_array($todo)) {
				continue;
			}

			$state = isset($todo["status"]) ? (string)$todo["status"] : "pending";

			if (!in_array($state, ["pending", "in_progress", "completed"], true)) {
				$state = "pending";
			}

			$text = isset($todo["content"]) ? (string)$todo["content"] : "";

			if ($state === "in_progress" && !empty($todo["activeForm"])) {
				$text = (string)$todo["activeForm"];
			}

			if ($text === "") {
				continue;
			}

			$out[] = [
				"text" => $text,
				"state" => $state,
				"className" => "dot " . ($state === "completed" ? "ok" : ($state === "in_progress" ? "running" : "pending")),
			];
		}

		return $out;
	}

	/** todos is the checklist a run wrote last, which replaces the one before it. */
	static function todos($events) {
		$out = [];

		foreach ($events as $event) {
			if (!empty($event["todo"])) {
				$out = $event["todo"];
			}
		}

		return $out;
	}

	/**
	 * stats is a run's context line: what the model read and wrote, and how
	 * many tools it called. A run still working has the tool count alone.
	 */
	static function stats($events) {
		$tools = 0;
		$final = false;
		$parts = [];

		foreach ($events as $event) {
			if ($event["kind"] === "tool_use") {
				$tools++;
			}

			if ($event["kind"] !== "result") {
				continue;
			}

			$final = true;

			if (!empty($event["input"])) {
				$parts[] = self::size($event["input"]) . " in";
			}

			if (!empty($event["cached"])) {
				$parts[] = self::size($event["cached"]) . " cached";
			}

			if (!empty($event["output"])) {
				$parts[] = self::size($event["output"]) . " out";
			}

			if (!empty($event["cost"])) {
				$parts[] = "$" . number_format($event["cost"], 2);
			}
		}

		if ($tools > 0 || !$final) {
			$parts[] = $tools . ($tools === 1 ? " tool call" : " tool calls");
		}

		return implode(", ", $parts);
	}

	/** tools is how many tool calls a run made. */
	static function tools($events) {
		$tools = 0;

		foreach ($events as $event) {
			if ($event["kind"] === "tool_use") {
				$tools++;
			}
		}

		return $tools;
	}

	/** noTrace is the trace of a run whose events hold no usable time. */
	static function noTrace() {
		return [
			"ms" => 0,
			"count" => 0,
			"states" => [],
			"segments" => [],
		];
	}

	/**
	 * trace is how a run spent its time: each gap between two events is a
	 * segment with the state and label the event before it set, and the states
	 * add those segments up.
	 */
	static function trace($events) {
		$segments = [];
		$state = "wait";
		$label = "Waiting";
		$last = 0;

		foreach ($events as $event) {
			$at = $event["at"];

			if ($last > 0 && $at > $last) {
				$segments[] = [
					"key" => $state,
					"label" => $label,
					"ms" => $at - $last,
					"span" => self::span($at - $last),
				];
			}

			if ($at > 0) {
				$last = $at;
			}

			$kind = self::category($event);

			if ($kind !== "") {
				$state = $kind;
				$label = self::traceLabel($event);
			}
		}

		$ms = [
			"think" => 0,
			"tools" => 0,
			"wait" => 0,
		];
		$count = [
			"think" => 0,
			"tools" => 0,
			"wait" => 0,
		];

		foreach ($segments as $segment) {
			$ms[$segment["key"]] = $ms[$segment["key"]] + $segment["ms"];
			$count[$segment["key"]] = $count[$segment["key"]] + 1;
		}

		$total = array_sum($ms);

		if ($total <= 0) {
			return self::noTrace();
		}

		$labels = [
			"think" => "Thinking",
			"tools" => "Tools",
			"wait" => "Waiting",
		];
		$states = [];
		$shares = 0;

		foreach ($labels as $key => $label) {
			$share = (int)round($ms[$key] * 100 / $total);
			$shares = $shares + $share;
			$states[] = [
				"key" => $key,
				"label" => $label,
				"ms" => $ms[$key],
				"count" => $count[$key],
				"share" => $share,
				"span" => self::span($ms[$key]),
			];
		}

		$states[count($states) - 1]["share"] = $states[count($states) - 1]["share"] + (100 - $shares);

		return [
			"ms" => $total,
			"count" => count($segments),
			"states" => $states,
			"segments" => $segments,
		];
	}

	/** traceLabel is the line a segment's tooltip names the work with. */
	static function traceLabel($event) {
		if ($event["kind"] === "status") {
			$body = trim($event["body"]);

			return $body === "" ? "Waiting" : $body;
		}

		if ($event["kind"] === "tool_use") {
			return $event["tool"] === "" ? "Tool" : $event["tool"];
		}

		if ($event["kind"] === "tool_result") {
			return "Tool result";
		}

		if ($event["kind"] === "text") {
			$parts = explode("\n", $event["body"]);
			$line = trim($parts[0]);

			if ($line === "") {
				return "Writing";
			}

			return strlen($line) > 64 ? substr($line, 0, 61) . "..." : $line;
		}

		return "Waiting";
	}

	/** category is the trace a status or an activity event belongs to, "" for one that names none. */
	static function category($event) {
		if ($event["kind"] === "status") {
			switch ($event["state"]) {
				case "thinking":
				case "writing":
					return "think";
				case "tool":
				case "result":
					return "tools";
				case "starting":
				case "done":
					return "wait";
			}

			return "";
		}

		if ($event["kind"] === "tool_use" || $event["kind"] === "tool_result") {
			return "tools";
		}

		if ($event["kind"] === "text") {
			return "think";
		}

		return "";
	}

	/** span is a duration as a line reads it: milliseconds, seconds, or minutes and seconds. */
	static function span($ms) {
		if ($ms >= 60000) {
			return floor($ms / 60000) . "m " . (int)round(($ms % 60000) / 1000) . "s";
		}

		if ($ms >= 1000) {
			return number_format($ms / 1000, 1) . "s";
		}

		return $ms . "ms";
	}

	/** stamp is an event's time as milliseconds, 0 where the row holds none. */
	static function stamp($value) {
		$text = (string)$value;

		if ($text === "") {
			return 0;
		}

		$seconds = strtotime($text);

		if ($seconds === false) {
			return 0;
		}

		$ms = 0;

		if (preg_match('/\.(\d+)/', $text, $found)) {
			$ms = (int)substr(str_pad($found[1], 3, "0"), 0, 3);
		}

		return $seconds * 1000 + $ms;
	}

	/** progress is the last text a live run printed, which its answer shows in place. */
	static function progress($events) {
		$out = "";

		foreach ($events as $event) {
			if ($event["kind"] === "text" && trim($event["prose"]) !== "") {
				$out = $event["prose"];
			}
		}

		return self::cut($out);
	}

	/** usage is what a thread read and wrote across every run, and how many tools it called. */
	static function usage($events) {
		$tools = 0;
		$input = 0;
		$cached = 0;
		$output = 0;
		$cost = 0;

		foreach ($events as $event) {
			if ($event["kind"] === "tool_use") {
				$tools++;
			}

			if ($event["kind"] !== "result") {
				continue;
			}

			$input += (int)$event["input"];
			$cached += (int)$event["cached"];
			$output += (int)$event["output"];
			$cost += (float)$event["cost"];
		}

		$tokens = $input + $cached;
		$parts = [];

		if ($tokens > 0) {
			$parts[] = self::size($tokens) . " context";
		}

		if ($output > 0) {
			$parts[] = self::size($output) . " out";
		}

		if ($tools > 0) {
			$parts[] = $tools . ($tools === 1 ? " tool call" : " tool calls");
		}

		if ($cost > 0) {
			$parts[] = "$" . number_format($cost, 2);
		}

		return [
			"tools" => $tools,
			"tokens" => $tokens,
			"output" => $output,
			"cost" => $cost,
			"label" => implode(", ", $parts),
		];
	}

	/** size is a token count as a reader scans it. */
	static function size($n) {
		if ($n >= 1000000) {
			return number_format($n / 1000000, 1) . "M";
		}

		if ($n >= 1000) {
			return number_format($n / 1000, 1) . "k";
		}

		return (string)$n;
	}

	/** status is the last state a run reported, or an empty one. */
	static function status($events) {
		$out = [
			"body" => "",
			"state" => "",
			"tool" => "",
		];

		foreach ($events as $event) {
			if ($event["kind"] !== "status") {
				continue;
			}

			$out = [
				"body" => $event["body"],
				"state" => $event["state"],
				"tool" => $event["tool"],
			];
		}

		return $out;
	}

	/** withoutStatus drops the records the page draws elsewhere. */
	static function withoutStatus($events) {
		$out = [];

		foreach ($events as $event) {
			if ($event["kind"] !== "status" && $event["kind"] !== "result") {
				$out[] = $event;
			}
		}

		return $out;
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
			"seq" => isset($row["seq"]) ? (int)$row["seq"] : 0,
			"at" => self::stamp(isset($row["created_at"]) ? $row["created_at"] : ""),
			"kind" => $kind,
			"body" => $body,
			"prose" => "",
			"tool" => "",
			"tool_id" => isset($row["tool_id"]) ? $row["tool_id"] : "",
			"input" => "",
			"agent" => "",
			"todo" => [],
			"error" => false,
			"state" => "",
		];

		switch ($kind) {

			case "text":
			case "result":
				$out["prose"] = $chrome->message($body);
				$out["input"] = isset($data["input_tokens"]) ? (int)$data["input_tokens"] : 0;
				$out["output"] = isset($data["output_tokens"]) ? (int)$data["output_tokens"] : 0;
				$out["cached"] = isset($data["cache_read_input_tokens"]) ? (int)$data["cache_read_input_tokens"] : 0;
				$out["cost"] = isset($data["cost_usd"]) ? (float)$data["cost_usd"] : 0;
				break;
			case "tool_use":
				$out["tool"] = isset($data["name"]) ? $data["name"] : "Tool";
				$out["input"] = self::json(isset($data["input"]) ? $data["input"] : []);

				// Task and Agent start a subagent, whose fields the card shows.
				if ($out["tool"] === "Task" || $out["tool"] === "Agent") {
					$out["agent"] = self::agent(isset($data["input"]) ? $data["input"] : []);
				}

				if ($out["tool"] === "TodoWrite") {
					$out["todo"] = self::todoList(isset($data["input"]) ? $data["input"] : []);
				}

				break;
			case "tool_result":
				$out["body"] = self::cut($body);
				$out["error"] = !empty($data["is_error"]);
				break;
			case "status":
				$out["state"] = isset($data["state"]) ? $data["state"] : "";
				$out["tool"] = isset($data["tool"]) ? $data["tool"] : "";

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

	/** agent is a subagent's task, kind and brief from a Task tool's input. */
	static function agent($input) {
		if (!is_array($input)) {
			return [
				"task" => "A subagent",
				"kind" => "",
				"prompt" => "",
			];
		}

		$task = isset($input["description"]) ? $input["description"] : "";

		return [
			"task" => $task === "" ? "A subagent" : $task,
			"kind" => isset($input["subagent_type"]) ? $input["subagent_type"] : "",
			"prompt" => isset($input["prompt"]) ? $input["prompt"] : "",
		];
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
