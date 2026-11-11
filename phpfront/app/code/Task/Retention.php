<?php

namespace Task;

// Retention is what the instance keeps of what it records: the areas the
// cleanup sweeps, in the order it sweeps them, and the rule that keeps each.
//
// The areas, their prose, the ages a rule takes and the size limits a page
// lists are the retention package's constant tables, which no binding
// carries; keep them in step with retention/area.go, retention/rule.go and
// retention/settings.go.
//
// A rule is a pair of plain setting rows, retention.<kind>.age and
// retention.<kind>.limit, so this reads the table the cleanup reads. The host
// answers what each area holds now, which is a walk of the disk and not a
// query.
class Retention {
	/** The ages a rule keeps an item for, shortest first. */
	const DAY = "24h";

	const WEEK = "7d";

	const MONTH = "31d";

	const QUARTER = "90d";

	const YEAR = "365d";

	const NEVER = "never";

	/** PREFIX starts the name of the two setting rows a kind's rule is kept as. */
	const PREFIX = "retention.";

	/** SIGNINS is the one area no rule governs: removed every cleanup. */
	const SIGNINS = "signins";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
	}

	/**
	 * areas are the rows of the table in order: the kinds with a rule, the
	 * fixed one among them. The cleanup applies them in this order, so run
	 * history takes a run's recording, summary and threads before the areas
	 * after it count them.
	 *
	 * `what` says what goes when an item expires and what stays. `sized` is an
	 * area whose items have a size, which the page shows beside the count and
	 * a size limit applies to unless `unlimited`. `fixed` is what the row says
	 * where the selects go for an area with no rule.
	 */
	static function areas() {
		return array(
			array(
				"kind" => "recordings",
				"title" => "Run recordings",
				"what" => "The terminal recording and output of a finished run, the .cast file Replay reads. The run, who started it and how it ended stay.",
				"one" => "recording",
				"many" => "recordings",
				"sized" => true,
				"unlimited" => false,
				"fixed" => "",
			),
			array(
				"kind" => "runs",
				"title" => "Run history",
				"what" => "The rows of finished runs older than the age, with their recordings, summaries and the prompt threads about them. The size is their recordings'.",
				"one" => "run",
				"many" => "runs",
				"sized" => true,
				"unlimited" => false,
				"fixed" => "",
			),
			array(
				"kind" => "prompts",
				"title" => "Prompt threads",
				"what" => "Prompt threads whose last reply is older than the age, with their replies. The runs stay; the notes people write on the board are not touched.",
				"one" => "thread",
				"many" => "threads",
				"sized" => false,
				"unlimited" => false,
				"fixed" => "",
			),
			array(
				"kind" => "workflow_runs",
				"title" => "Workflow run history",
				"what" => "The runs of workflows older than the age, with their steps, answers and the threads of their prompts. The workflows stay.",
				"one" => "run",
				"many" => "runs",
				"sized" => false,
				"unlimited" => false,
				"fixed" => "",
			),
			array(
				"kind" => "artifacts",
				"title" => "Artifact history",
				"what" => "The earlier versions of the artifact tree's files, which the history keeps when a run or a user writes a file again. The latest version of every file stays.",
				"one" => "version",
				"many" => "versions",
				"sized" => true,
				"unlimited" => false,
				"fixed" => "",
			),
			array(
				"kind" => "secret_reads",
				"title" => "Secret read log",
				"what" => "Which run got which secret in its environment. The log always keeps the last 1000 reads of each secret.",
				"one" => "read",
				"many" => "reads",
				"sized" => false,
				"unlimited" => false,
				"fixed" => "",
			),
			array(
				"kind" => self::SIGNINS,
				"title" => "Expired sign-ins",
				"what" => "Tokens and browser sessions past their expiry, which no longer sign in.",
				"one" => "row",
				"many" => "rows",
				"sized" => false,
				"unlimited" => false,
				"fixed" => "Removed every cleanup, no setting.",
			),
			array(
				"kind" => "tooling",
				"title" => "Tool history",
				"what" => "The runs and commands older than the age in atkins.db of each workspace's home and the instance's own copy: what the tools ran, which atkins prunes and vacuums.",
				"one" => "database",
				"many" => "databases",
				"sized" => true,
				"unlimited" => true,
				"fixed" => "",
			),
		);
	}

	/** ages are the ages a rule takes, shortest first. */
	static function ages() {
		return array(self::DAY, self::WEEK, self::MONTH, self::QUARTER, self::YEAR, self::NEVER);
	}

	/** label names an age as the page lists it: "31 days", "Never". */
	static function label($age) {
		$labels = array(
			self::DAY => "24 hours",
			self::WEEK => "7 days",
			self::MONTH => "31 days",
			self::QUARTER => "90 days",
			self::YEAR => "365 days",
			self::NEVER => "Never",
		);

		return isset($labels[$age]) ? $labels[$age] : "";
	}

	/** sizes are the size limits the page lists, in bytes, besides no limit. */
	static function sizes() {
		return array(100 << 20, 500 << 20, 1 << 30, 5 << 30, 10 << 30, 50 << 30);
	}

	/**
	 * sizeLabel names a size limit as the page lists it: whole units with no
	 * decimal, "100 MB", and the byte wording of any other.
	 */
	static function sizeLabel($bytes) {
		$units = array(
			array(1 << 40, "TB"),
			array(1 << 30, "GB"),
			array(1 << 20, "MB"),
			array(1 << 10, "KB"),
		);

		foreach ($units as $unit) {
			if ($bytes < $unit[0]) {
				continue;
			}

			if ($bytes % $unit[0] === 0) {
				return (int)($bytes / $unit[0]) . " " . $unit[1];
			}

			return sprintf("%.1f", $bytes / $unit[0]) . " " . $unit[1];
		}

		return Size::bytes($bytes);
	}

	/** kinds are the kinds of data with a rule, in the order the page shows them. */
	static function kinds() {
		$out = array();

		foreach (self::areas() as $area) {
			if ($area["fixed"] === "") {
				$out[] = $area["kind"];
			}
		}

		return $out;
	}

	/** blank is the rule a kind keeps with nothing stored: 31 days, no size limit. */
	static function blank() {
		return array("age" => self::MONTH, "limit" => 0);
	}

	/**
	 * values is the setting table by name, which the rules are read out of.
	 *
	 * The whole table is read in one statement, as the cleanup reads it: the
	 * rows are a handful and the names of two of them per kind.
	 */
	function values() {
		$out = array();

		foreach ($this->db->get_all("SELECT name, value FROM setting") as $row) {
			$out[Row::text($row, "name")] = Row::text($row, "value");
		}

		return $out;
	}

	/**
	 * rules is the rule of every kind by kind: the stored rows over the
	 * defaults, and the default where a row names an age that is not one or a
	 * limit that is not a number.
	 */
	function rules() {
		$values = $this->values();
		$out = array();

		foreach (self::kinds() as $kind) {
			$rule = self::blank();
			$age = isset($values[self::PREFIX . $kind . ".age"]) ? $values[self::PREFIX . $kind . ".age"] : "";
			$limit = isset($values[self::PREFIX . $kind . ".limit"]) ? $values[self::PREFIX . $kind . ".limit"] : "";

			if ($age !== "") {
				$rule["age"] = $age;
			}

			if (self::whole($limit)) {
				$rule["limit"] = (int)$limit;
			}

			$out[$kind] = self::valid($rule) ? $rule : self::blank();
		}

		return $out;
	}

	/**
	 * whole reports whether a stored limit is a decimal integer, which is what
	 * the Go side's parse takes: anything else leaves the rule's limit as it
	 * was, which is none.
	 */
	static function whole($text) {
		return preg_match('/^[+-]?[0-9]+$/', $text) === 1;
	}

	/** valid reports whether a rule names an age of ages() and no negative limit. */
	static function valid($rule) {
		return in_array($rule["age"], self::ages(), true) && $rule["limit"] >= 0;
	}

	/**
	 * rows are the table's rows, an area each in the order the cleanup sweeps
	 * them: what the area is, what it holds now, and the choices of the two
	 * selects its rule sets.
	 *
	 * $usage is Task\Host::usage()'s areas, which is a walk of the disk; a kind
	 * it does not answer for holds nothing.
	 */
	static function rows($usage, $rules) {
		$found = self::byKind($usage);
		$out = array();

		foreach (self::areas() as $area) {
			$kind = $area["kind"];
			$rule = isset($rules[$kind]) ? $rules[$kind] : self::blank();
			$holds = isset($found[$kind]) ? $found[$kind] : array("count" => 0, "bytes" => 0);

			$out[] = array(
				"area" => $area,
				"limited" => $area["sized"] && !$area["unlimited"],
				"holds" => self::holds($area, $holds),
				"ages" => self::ageChoices($rule),
				"limits" => self::limitChoices($rule),
			);
		}

		return $out;
	}

	/** byKind is the host's area usage by the kind each row names. */
	static function byKind($usage) {
		$out = array();

		foreach ($usage as $row) {
			$out[Row::text($row, "kind")] = $row;
		}

		return $out;
	}

	/**
	 * holds reads what an area holds as the page and the cleanup say it: the
	 * count, and the bytes after it for an area whose items have a size.
	 */
	static function holds($area, $usage) {
		$count = Row::number($usage, "count");
		$name = self::name($count, $area["one"], $area["many"]);

		if (!$area["sized"]) {
			return $name;
		}

		return $name . ", " . Size::bytes(Row::number($usage, "bytes"));
	}

	/** name counts the items of an area by name: "1 run", "12 runs". */
	static function name($count, $one, $many) {
		if ($count === 1) {
			return $count . " " . $one;
		}

		return $count . " " . $many;
	}

	/** ageChoices are the ages the Keep select lists, the rule's selected. */
	static function ageChoices($rule) {
		$out = array();

		foreach (self::ages() as $age) {
			$out[] = array(
				"value" => $age,
				"label" => self::label($age),
				"selected" => $age === $rule["age"],
			);
		}

		return $out;
	}

	/**
	 * limitChoices are the size limits to pick from: no limit first, then
	 * every size the page lists, with a rule's own limit at its place in them
	 * where the list lacks it.
	 */
	static function limitChoices($rule) {
		$limit = $rule["limit"];
		$sizes = self::sizes();

		if ($limit > 0 && !in_array($limit, $sizes, true)) {
			$sizes[] = $limit;

			sort($sizes);
		}

		$out = array(
			array(
				"value" => "0",
				"label" => "No limit",
				"selected" => $limit === 0,
			),
		);

		foreach ($sizes as $size) {
			$out[] = array(
				"value" => (string)$size,
				"label" => self::sizeLabel($size),
				"selected" => $size === $limit,
			);
		}

		return $out;
	}
}
