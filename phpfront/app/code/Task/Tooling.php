<?php

namespace Task;

// Tooling is what atkins has recorded the tree running, as the panel on the
// runs page shows it: a section per database atkins records to, each with a
// chart per window, the newest commands, and the commands by their time.
//
// It reports commands and not runs. A run of a job is one row of the run
// table and the twenty commands inside it are twenty rows here, so the time a
// repository spends on go, npm or docker is atkins' to report and task-ui's
// to draw. Task\Host::tooling() answers the report, which is the only part
// that is not PHP: reading it means running atkins.
//
// The charts are the activity charts' own spans, so the two line up column for
// column, and they count their slots in the zone of the clock they are given,
// which is the server's.
//
// Durations are milliseconds throughout, as atkins reports them, and
// Task\Run::duration words them.
class Tooling {
	/** WINDOWS are the day charts an atkins that lists no samples draws, the longer first. */
	const WINDOWS = array(30, 7);

	/** STUB is the least height of a bar with anything in it, in percent, so a quiet command next to a busy one still shows. */
	const STUB = 4;

	/** LATEST is how many of the newest commands the list under the charts names. */
	const LATEST = 20;

	/**
	 * TOP is how many commands a window draws a bar each for, the rest drawn
	 * together as other. Beyond a handful of colours a slot's group stops
	 * reading as one.
	 */
	const TOP = 5;

	/** COLORS is how many series colours the palette holds. */
	const COLORS = 6;

	/** OTHER is the colour class of the commands past the top ones. */
	const OTHER = "s-other";

	/** now is the moment the charts end at, and zone the zone they count in. */
	var $now = null;

	var $zone = null;

	/**
	 * palette is the colour of each command, which it keeps across the windows
	 * and the scopes of the panel: a reader following one command down the
	 * page follows one colour.
	 */
	var $palette = array();

	/** scopes are the sections of the panel, the project's database first. */
	var $scopes = array();

	/**
	 * The panel of a report, as Task\Host::tooling() answers it. data() is
	 * empty for a tree nobody has run a job in through atkins, which is the
	 * usual answer and draws no panel at all.
	 */
	function __construct($report, $now) {
		$this->now = $now;
		$this->zone = $now->location();

		foreach ($report as $scope) {
			$this->scopes[] = $this->scope($scope);
		}
	}

	/** data is the scope the tooling component reads. */
	function data() {
		if (count($this->scopes) === 0) {
			return array();
		}

		return array("scopes" => $this->scopes);
	}

	/**
	 * spans are the three charts, the activity charts' own: the month and the
	 * day side by side, the week under them on a row of its own.
	 *
	 * The class is the extra one the figure takes, which is the week's width
	 * and nothing else: the activity chart's own classes are its template's.
	 */
	static function spans() {
		return array(
			Activity::span("Last 30 days", "day", 30, 0, "", false),
			Activity::span("Last 24 hours", "hour", 24, 1, "", false),
			Activity::span("Last 7 days", "3 hours", 56, 3, "wide", false),
		);
	}

	/** days is how many calendar days a span reaches back. */
	static function days($span) {
		if ($span["hours"] === 0) {
			return $span["slots"];
		}

		return max(1, intdiv($span["slots"] * $span["hours"], 24));
	}

	/**
	 * scope is one database as the panel draws it: its charts, the newest
	 * commands and the table behind the toggle.
	 *
	 * The samples are the finer series, and an atkins that lists none leaves
	 * the charts to the daily rollups, which are days wide. The newest
	 * commands are read after the charts, so each one is already in the colour
	 * its command holds on them.
	 *
	 * The stamps are read here, once, and the charts are given them: three
	 * charts over the same samples each parsing the same string was 470ms of a
	 * 2s /runs, and a parse crosses into the host twice.
	 */
	function scope($scope) {
		$charts = array();
		$latest = array();

		if (count($scope["samples"]) > 0) {
			$stamps = self::stamps($scope["samples"]);

			foreach (self::spans() as $span) {
				$chart = $this->sampled($scope, $stamps, $span);

				if ($chart !== null) {
					$charts[] = $chart;
				}
			}

			$latest = $this->latest($scope["samples"]);
		} else {
			foreach (self::WINDOWS as $days) {
				$chart = $this->daily($scope, $days);

				if ($chart !== null) {
					$charts[] = $chart;
				}
			}
		}

		$commands = self::commands($scope["commands"]);

		return array(
			"database" => $scope["database"],
			"name" => $scope["name"],
			"repository" => $scope["repository"],
			"sum" => Activity::plural($scope["runs"], "command") . " in " . $scope["days"] . " days · " . Run::duration($scope["total_ms"]) . " of tooling",
			"charts" => $charts,
			"latest" => $latest,
			"toggle" => Activity::plural(count($commands), "command") . " by the time they have taken",
			"commands" => $commands,
		);
	}

	/**
	 * stamps is the unix time of each sample, by the index the sample holds, 0
	 * for one whose stamp does not read. Every chart of a scope reads the same
	 * series, so the strings are parsed once for all of them.
	 */
	static function stamps($samples) {
		$out = array();

		foreach ($samples as $i => $sample) {
			$out[$i] = Clock::seconds($sample["at"]);
		}

		return $out;
	}

	/**
	 * sampled is one chart of a span from the samples, a slot per span slot in
	 * the zone of the clock. Null for a window the report is too short to
	 * fill: a week of history drawn as a month is a month of empty bars with a
	 * week at the end.
	 *
	 * stamps is what scope read of the samples' own timestamps, by index.
	 */
	function sampled($scope, $stamps, $span) {
		if ($scope["days"] < self::days($span)) {
			return null;
		}

		$edges = Activity::edges($span, $this->now, $this->zone);
		$marks = Activity::marks($edges);
		$slots = array();

		for ($i = 0; $i < $span["slots"]; $i++) {
			$slots[] = array(
				"label" => Activity::label($span, $edges, $i),
				"title" => "",
				"runs" => 0,
				"ms" => 0,
			);
		}

		$rows = array();

		foreach ($scope["samples"] as $i => $sample) {
			$at = $stamps[$i];

			if ($at === 0) {
				continue;
			}

			$slot = Activity::slot($marks, $at);

			if ($slot < 0) {
				continue;
			}

			$rows[] = array(
				"slot" => $slot,
				"name" => $sample["command"],
				"runs" => 1,
				"ms" => $sample["duration_ms"],
			);
			$slots[$slot]["runs"]++;
			$slots[$slot]["ms"] += $sample["duration_ms"];
		}

		foreach ($slots as $i => $slot) {
			$slots[$i]["title"] = self::slotTitle($slot["label"], $slot["runs"], $slot["ms"]);
		}

		$chart = $this->chart($slots, $rows);
		$chart["class"] = $span["class"] === "" ? "tooling-chart" : "tooling-chart " . $span["class"];
		$chart["title"] = $span["title"];
		$chart["unit"] = $span["unit"];
		$chart["by"] = "command";
		$chart["density"] = $span["slots"] > 30 ? "fine" : "";
		$chart["ticks"] = Activity::ticks($span, $edges);

		return $this->drawn($chart);
	}

	/**
	 * daily is one chart of days from the daily series, for an atkins that
	 * lists no samples. Null for a window the report is too short to fill.
	 */
	function daily($scope, $days) {
		if ($scope["days"] < $days) {
			return null;
		}

		$daily = self::days_of($scope, $days);
		$slots = array();
		$at = array();

		foreach ($daily as $i => $day) {
			$slots[] = array(
				"label" => $day["date"],
				"title" => self::slotTitle($day["date"], $day["runs"], $day["total_ms"]),
				"runs" => $day["runs"],
				"ms" => $day["total_ms"],
			);
			$at[$day["date"]] = $i;
		}

		$series = self::series($scope, $daily);
		$rows = array();

		foreach ($series["rows"] as $row) {
			$row["slot"] = isset($at[$row["date"]]) ? $at[$row["date"]] : 0;
			$rows[] = $row;
		}

		$chart = $this->chart($slots, $rows);
		$chart["class"] = "tooling-chart";
		$chart["title"] = "Last " . $days . " days";
		$chart["unit"] = "day";
		$chart["by"] = $series["by"];
		$chart["density"] = $days <= 7 ? "sparse" : "";
		$chart["ticks"] = self::dayTicks($daily);

		return $this->drawn($chart);
	}

	/**
	 * chart draws the slots: the window's top names by their total, which the
	 * legend lists and each slot draws a bar of in that order, the rest
	 * together as other.
	 *
	 * Every bar is against the tallest one command took in a slot, and not
	 * against the busiest slot, so a group of bars reads as the commands
	 * inside it and not as the slot's total.
	 */
	function chart($slots, $rows) {
		$runs = 0;
		$total = 0;
		$busiest = 0;

		foreach ($slots as $slot) {
			$busiest = max($busiest, $slot["ms"]);
			$runs += $slot["runs"];
			$total += $slot["ms"];
		}

		$drawn = $this->legend($rows);
		$legend = $drawn["entries"];
		$sums = self::sums($slots, $rows, $drawn["rank"], count($legend));
		$tallest = 0;

		foreach ($sums as $slot) {
			foreach ($slot as $sum) {
				$tallest = max($tallest, $sum["ms"]);
			}
		}

		$bars = array();
		$n = count($slots);

		foreach ($slots as $index => $slot) {
			$edge = Activity::edge($index, $n, max(4, intdiv($n * 2, 15)));
			$bars[] = array(
				"class" => "tip " . $edge,
				"title" => $slot["title"],
				"segments" => self::segments($slot["label"], $sums[$index], $legend, $tallest, $edge),
			);
		}

		return array(
			"runs" => $runs,
			"total" => $total,
			"busiest" => $busiest,
			"n" => $n,
			"bars" => $bars,
			"legend" => $legend,
		);
	}

	/**
	 * legend is the window's top names by their total, each in its colour,
	 * with the rest together as other, and the rank each top name draws its
	 * bar at.
	 *
	 * The names that already have a colour keep it, and a name that has none
	 * takes a colour no other name of this legend holds, so two commands under
	 * one chart never share one.
	 */
	function legend($rows) {
		$totals = array();
		$names = array();

		foreach ($rows as $row) {
			if (!isset($totals[$row["name"]])) {
				$totals[$row["name"]] = array("runs" => 0, "total" => 0);
				$names[] = $row["name"];
			}

			$totals[$row["name"]]["runs"] += $row["runs"];
			$totals[$row["name"]]["total"] += $row["ms"];
		}

		usort($names, function($one, $other) use ($totals) {
			if ($totals[$one]["total"] !== $totals[$other]["total"]) {
				return $totals[$one]["total"] > $totals[$other]["total"] ? -1 : 1;
			}

			return strcmp($one, $other);
		});

		$top = count($names) > self::TOP ? array_slice($names, 0, self::TOP) : $names;
		$taken = array();

		foreach ($top as $name) {
			if (isset($this->palette[$name])) {
				$taken[$this->palette[$name]] = true;
			}
		}

		$out = array();
		$rank = array();

		foreach ($top as $index => $name) {
			$color = $this->color($name, $taken);
			$taken[$color] = true;
			$rank[$name] = $index;
			$out[] = self::entry($name, $color, $totals[$name]["runs"], $totals[$name]["total"]);
		}

		if (count($names) === count($top)) {
			return array("entries" => $out, "rank" => $rank);
		}

		$rest = array("runs" => 0, "total" => 0);

		foreach (array_slice($names, count($top)) as $name) {
			$rest["runs"] += $totals[$name]["runs"];
			$rest["total"] += $totals[$name]["total"];
		}

		$out[] = self::entry("other", self::OTHER, $rest["runs"], $rest["total"]);

		return array("entries" => $out, "rank" => $rank);
	}

	/** entry is one line of a legend as the markup draws it. */
	static function entry($name, $color, $runs, $total) {
		return array(
			"name" => $name,
			"color" => $color,
			"total" => Run::duration($total),
			"runs" => Activity::plural($runs, "run"),
		);
	}

	/**
	 * color is the colour class of a name, and sets a name it has not seen to
	 * the colour fewest names hold that none of taken holds, taken being the
	 * colours of the names beside it in its window.
	 */
	function color($name, $taken) {
		if (isset($this->palette[$name])) {
			return $this->palette[$name];
		}

		$held = array();

		foreach ($this->palette as $color) {
			$held[$color] = isset($held[$color]) ? $held[$color] + 1 : 1;
		}

		$best = "";

		for ($index = 0; $index < self::COLORS; $index++) {
			$color = "s" . $index;

			if (isset($taken[$color])) {
				continue;
			}

			if ($best === "" || self::holds($held, $color) < self::holds($held, $best)) {
				$best = $color;
			}
		}

		if ($best === "") {
			$best = "s" . (count($this->palette) % self::COLORS);
		}

		$this->palette[$name] = $best;

		return $best;
	}

	/** holds is how many names hold a colour. */
	static function holds($held, $color) {
		return isset($held[$color]) ? $held[$color] : 0;
	}

	/**
	 * sums is each slot's rows summed per legend entry: the top names in their
	 * order, then the rest as other. A slot nothing ran in sums to nothing.
	 */
	static function sums($slots, $rows, $rank, $entries) {
		$out = array();

		foreach ($slots as $i => $slot) {
			$out[$i] = array();
		}

		foreach ($rows as $row) {
			if (count($out[$row["slot"]]) === 0) {
				for ($k = 0; $k < $entries; $k++) {
					$out[$row["slot"]][] = array("runs" => 0, "ms" => 0);
				}
			}

			$index = isset($rank[$row["name"]]) ? $rank[$row["name"]] : count($rank);
			$out[$row["slot"]][$index]["runs"] += $row["runs"];
			$out[$row["slot"]][$index]["ms"] += $row["ms"];
		}

		return $out;
	}

	/**
	 * segments is one slot's group: a bar per legend entry, in its order, each
	 * against the window's tallest. The bar of a command that did not run then
	 * is empty and says nothing; a slot nothing ran in has no bars at all, and
	 * its column says so.
	 */
	static function segments($when, $sums, $legend, $tallest, $edge) {
		$sum = 0;

		foreach ($sums as $each) {
			$sum += $each["ms"];
		}

		if ($sum <= 0) {
			return array();
		}

		$out = array();

		foreach ($legend as $index => $entry) {
			$ms = $sums[$index]["ms"];

			if ($ms <= 0) {
				$out[] = array(
					"class" => "",
					"style" => "",
					"title" => "",
				);

				continue;
			}

			$out[] = array(
				"class" => "tip " . $edge . " " . $entry["color"],
				"style" => self::height($ms, $tallest),
				"title" => $when . " · " . $entry["name"] . ": " . Run::duration($ms) . ", " . Activity::plural($sums[$index]["runs"], "run"),
			);
		}

		return $out;
	}

	/**
	 * height is the height of a bar, in percent of the tallest, with a floor
	 * so a quiet command beside a busy one still shows.
	 */
	static function height($ms, $tallest) {
		if ($ms <= 0 || $tallest <= 0) {
			return "";
		}

		$percent = $ms / $tallest * 100;

		if ($percent < self::STUB) {
			$percent = self::STUB;
		}

		return sprintf("height: %.1f%%", $percent);
	}

	/**
	 * drawn is a chart as the component reads it: the caption, the classes of
	 * the group of bars and what it says to a reader who hears it.
	 *
	 * Everything a branch would decide inside an attribute is decided here, so
	 * the markup holds no condition: a stray space in a style breaks the bar
	 * it draws.
	 */
	function drawn($chart) {
		$stats = Activity::plural($chart["runs"], "command") . " · " . Run::duration($chart["total"]);

		if ($chart["busiest"] > 0) {
			$stats .= " · busiest " . $chart["unit"] . " " . Run::duration($chart["busiest"]);
		}

		return array(
			"class" => $chart["class"],
			"title" => $chart["title"],
			"stats" => $stats,
			"bars" => array(
				"class" => $chart["density"] === "" ? "tooling-bars" : "tooling-bars " . $chart["density"],
				"style" => "--n: " . $chart["n"] . "; --k: " . count($chart["legend"]),
				"label" => "Time spent per " . $chart["unit"] . " by " . $chart["by"] . ", " . strtolower($chart["title"]),
				"slots" => $chart["bars"],
			),
			"ticks" => $chart["ticks"],
			"legend" => $chart["legend"],
		);
	}

	/**
	 * latest is the newest commands, newest first, each in the colour its
	 * command holds on the charts, at its clock time today and with its
	 * weekday before.
	 */
	function latest($samples) {
		$today = $this->now->format(Clock::DAY);
		$n = count($samples);
		$take = $n > self::LATEST ? self::LATEST : $n;
		$out = array();

		for ($i = $n - 1; $i >= $n - $take; $i--) {
			$sample = $samples[$i];
			$at = Clock::parse($sample["at"]);
			$when = "";

			if ($at !== null) {
				$local = $at->in($this->zone);
				$when = $local->format(Clock::DAY) === $today ? $local->format("15:04") : $local->format("Mon 15:04");
			}

			$out[] = array(
				"stamp" => Clock::read($sample["at"]),
				"when" => $when,
				"color" => isset($this->palette[$sample["command"]]) ? $this->palette[$sample["command"]] : self::OTHER,
				"title" => $sample["job"] === "" ? $sample["command"] : $sample["command"] . ", job " . $sample["job"],
				"command" => $sample["command"],
				"job" => $sample["job"],
				"took" => Run::duration($sample["duration_ms"]),
				"failed" => $sample["failed"] || $sample["exit_code"] !== 0,
				"exit" => $sample["exit_code"],
			);
		}

		return $out;
	}

	/** commands is the command keys by their total time, the table behind the toggle. */
	static function commands($found) {
		$out = array();

		foreach ($found as $command) {
			$out[] = array(
				"command" => $command["command"],
				"tool" => $command["tool"],
				"runs" => $command["runs"],
				"total" => Run::duration($command["total_ms"]),
				"average" => Run::duration($command["average_ms"]),
				"max" => Run::duration($command["max_ms"]),
				"failures" => $command["failures"],
				"last" => Clock::ago($command["last_run_at"]),
			);
		}

		return $out;
	}

	/** slotTitle is what a hover over a slot says, and a slot nothing ran in says so. */
	static function slotTitle($label, $runs, $ms) {
		if ($runs === 0) {
			return $label . ": nothing ran";
		}

		return $label . ": " . Activity::plural($runs, "command") . " in " . Run::duration($ms);
	}

	/**
	 * days_of is the total time per day of the last days, oldest first, with a
	 * zero for a day nothing ran: a chart that left them out would draw a busy
	 * fortnight and a quiet one the same width.
	 *
	 * The window counts back from the report's last day, which is the day
	 * atkins made it, because the rows are dates and a reader asks for the
	 * last seven days of a calendar.
	 */
	static function days_of($scope, $days) {
		$last = self::date($scope["to"]);

		if ($last === null || $days <= 0) {
			return array();
		}

		$first = $last->addDate(0, 0, 1 - $days)->format(Clock::DAY);
		$totals = array();

		foreach ($scope["usage"] as $row) {
			if ($row["date"] < $first) {
				continue;
			}

			if (!isset($totals[$row["date"]])) {
				$totals[$row["date"]] = array("runs" => 0, "total_ms" => 0);
			}

			$totals[$row["date"]]["runs"] += $row["runs"];
			$totals[$row["date"]]["total_ms"] += $row["total_ms"];
		}

		$out = array();

		for ($i = 1 - $days; $i <= 0; $i++) {
			$date = $last->addDate(0, 0, $i)->format(Clock::DAY);
			$held = isset($totals[$date]) ? $totals[$date] : array("runs" => 0, "total_ms" => 0);

			$out[] = array(
				"date" => $date,
				"runs" => $held["runs"],
				"total_ms" => $held["total_ms"],
			);
		}

		return $out;
	}

	/**
	 * series is the rows of a window's days, by command where atkins reports
	 * the command series and by tool where it does not, which is what a chart
	 * of days says its bars are of.
	 */
	static function series($scope, $daily) {
		if (count($daily) === 0) {
			return array("rows" => array(), "by" => "command");
		}

		$first = $daily[0]["date"];
		$rows = array();

		if (count($scope["command_usage"]) > 0) {
			foreach ($scope["command_usage"] as $row) {
				if ($row["date"] >= $first) {
					$rows[] = array(
						"slot" => 0,
						"date" => $row["date"],
						"name" => $row["command"],
						"runs" => $row["runs"],
						"ms" => $row["total_ms"],
					);
				}
			}

			return array("rows" => $rows, "by" => "command");
		}

		foreach ($scope["usage"] as $row) {
			if ($row["date"] >= $first) {
				$rows[] = array(
					"slot" => 0,
					"date" => $row["date"],
					"name" => $row["tool"],
					"runs" => $row["runs"],
					"ms" => $row["total_ms"],
				);
			}
		}

		return array("rows" => $rows, "by" => "tool");
	}

	/**
	 * dayTicks marks the days under a chart of days, centred on their column:
	 * on a week every day by its weekday, on a longer window a day a week and
	 * the last day the last, by its date.
	 */
	static function dayTicks($daily) {
		$n = count($daily);
		$step = $n <= 7 ? 1 : 7;
		$layout = $n <= 7 ? "Mon" : "Jan 2";
		$out = array();

		for ($i = $n - 1; $i >= 0; $i -= $step) {
			$day = self::date($daily[$i]["date"]);

			if ($day === null) {
				continue;
			}

			$out[] = array("label" => $day->format($layout), "at" => Activity::middle($i, $n));
		}

		return array_reverse($out);
	}

	/** date reads a day of the report, or null for one it does not hold. */
	static function date($value) {
		if (!is_string($value) || trim($value) === "") {
			return null;
		}

		return \DateTime::parse_in_location(Clock::DAY, trim($value), \Time\Location::utc());
	}
}
