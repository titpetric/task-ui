<?php

namespace Task;

// Activity is when runs ran: a bar per day of the last thirty, a bar per hour
// of the last twenty four, a bar per three hours of the last week, each bar
// split by how its runs ended, the durations of each column as lines under it,
// and the latest finished runs as a line of their own.
//
// It is computation over rows and nothing else, so it is PHP. A page reads the
// runs a filter listed through of(), and draws what data() holds through the
// chart component.
//
// The clock is a parameter and not a reading, so a page charts the moment it
// draws and a fixture states which bucket a run lands in. Every slot is a wall
// clock reading in the zone of that clock, which is the server's: a run at one
// in the morning is on its own day here and the day before in UTC.
//
// Durations are milliseconds throughout: Task\Run::elapsed reads them and
// Task\Run::duration words them. The duration plot is the one exception, and
// NANOS says why.
class Activity {
	/** DAYS is how many days the first chart covers, today the last. */
	const DAYS = 30;

	/** DURATIONS is how many of the latest finished runs the duration line draws. */
	const DURATIONS = 30;

	/**
	 * LIMIT bounds how many of the latest runs a page reads for the charts:
	 * more than the days of DAYS hold on a busy instance, read once per page.
	 */
	const LIMIT = 5000;

	/**
	 * STUB is the least height of a bar with runs, in percent, so a quiet day
	 * beside a busy one still shows.
	 */
	const STUB = 4;

	/**
	 * HEADROOM keeps the slowest run off the top edge of a duration line, in
	 * percent.
	 */
	const HEADROOM = 10;

	/**
	 * NANOS is how many nanoseconds a millisecond holds, which the duration
	 * plot scales in because Go plots a time.Duration.
	 *
	 * The mean of one is whole nanoseconds, so a column whose level falls on a
	 * rounding boundary is decided by what that division threw away: the Go
	 * page writes the level a hair above the boundary and rounds up, and a
	 * plot of exact milliseconds sits on it and rounds to even. One column of
	 * one chart in a sweep of a hundred pages is enough to find it.
	 */
	const NANOS = 1000000;

	/** HOUR is the layout of a whole clock hour of a day. */
	const HOUR = "2006-01-02 15";

	/** WHEN is the layout a run on the duration line names its moment in. */
	const WHEN = "Mon 2006-01-02 15:04";

	/** now is the moment the charts end at, and zone the zone they count in. */
	var $now = null;

	var $zone = null;

	/** days is how many days the first chart covers. */
	var $days = 0;

	/** spans are the three charts: what each covers, and how it is drawn. */
	var $spans = array();

	/** edges are where each chart's slots start, and where its last one ends. */
	var $edges = array();

	/** marks are the same edges as unix seconds, and a run finds its slot by them. */
	var $marks = array();

	/** bars are each chart's slots: their runs, their durations and their hovers. */
	var $bars = array();

	/** stats are the finished runs of each chart's window. */
	var $stats = array();

	/** lines are each chart's durations as paths, empty where none of its runs finished. */
	var $lines = array();

	/** ticks are the time markers under each chart. */
	var $ticks = array();

	/** total is every run of the window, and failed the ones of them that broke. */
	var $total = 0;

	var $failed = 0;

	/** median is the middle duration of the finished runs. */
	var $median = 0;

	/** durations is the duration line, empty where a page draws none. */
	var $durations = array();

	/**
	 * The charts of the runs that started in a window of days up to now,
	 * counted in the zone of now. data() is empty where none did.
	 *
	 * durations draws the duration line, which is one job's alone: on a page
	 * of many jobs it would mix them.
	 */
	function __construct($runs, $now, $days = 0, $durations = true) {
		$this->now = $now;
		$this->zone = $now->location();
		$this->days = $days < 1 ? self::DAYS : $days;
		$this->spans = array(
			self::span("Last " . $this->days . " days", "day", $this->days, 0, "", true),
			self::span("Last 24 hours", "hour", 24, 1, "hours", false),
			self::span("Last 7 days", "3 hours", 56, 3, "wide", false),
		);

		foreach ($this->spans as $c => $span) {
			$this->edges[$c] = self::edges($span, $now, $this->zone);
			$this->marks[$c] = self::marks($this->edges[$c]);
			$this->bars[$c] = self::slots($span["slots"]);
			$this->stats[$c] = self::blank();
			$this->lines[$c] = array();
			$this->ticks[$c] = array();
		}

		$finished = $this->tally($runs);

		if ($this->total === 0) {
			return;
		}

		$this->draw();
		$this->median = self::median($finished);

		if ($durations) {
			$this->durations = $this->spark(self::latest($finished));
		}
	}

	/**
	 * data is the scope the chart component reads, empty where no run is in
	 * the window and a page draws no chart at all.
	 */
	function data() {
		if ($this->total === 0) {
			return array();
		}

		return array(
			"sum" => self::plural($this->total, "run") . " in " . $this->days . " days",
			"failed" => $this->failed > 0 ? $this->failed . " failed" : "",
			"median" => $this->median > 0 ? "median " . Run::duration($this->median) : "",
			"charts" => $this->charts(),
			"durations" => $this->durations,
		);
	}

	/**
	 * of is the runs the charts read, of the rows a filter listed: when each
	 * one started, how long it took and how it ended. A row that never
	 * started is on no chart.
	 */
	static function of($found) {
		$out = array();

		foreach ($found as $run) {
			$at = Clock::parse($run->getStartedAt());

			if ($at === null) {
				continue;
			}

			$out[] = array(
				"at" => $at,
				"unix" => $at->unix(),
				"nano" => $at->unixnano(),
				"elapsed" => Run::elapsed($run),
				"state" => $run->getState(),
			);
		}

		return $out;
	}

	/**
	 * rows is of() of the rows themselves, for a page that reads a window with
	 * Task\Run::charted and hydrates none of it.
	 *
	 * The same five values, from the two text columns and the state, with one
	 * parse of each stamp: of() reads a model, whose stamps were parsed into it
	 * and written back out as text, so the window was parsed three times over
	 * -- 274ms of hydration and 57ms of re-parsing on a 2s /runs, for five
	 * numbers. A row with no stamp is on no chart, as there.
	 */
	static function rows($found) {
		$out = array();

		foreach ($found as $row) {
			$at = Clock::parse(Row::stamp($row, "started_text"));

			if ($at === null) {
				continue;
			}

			$out[] = array(
				"at" => $at,
				"unix" => $at->unix(),
				"nano" => $at->unixnano(),
				"elapsed" => Run::took($at, Clock::parse(Row::stamp($row, "finished_text"))),
				"state" => Row::text($row, "state"),
			);
		}

		return $out;
	}

	/**
	 * tally counts the runs into every chart they fall in, and answers the
	 * finished ones oldest first, which the median and the duration line
	 * read.
	 *
	 * The first chart holds every other span, so a run outside its window is
	 * on none of them.
	 */
	function tally($runs) {
		$charts = count($this->spans);
		$bars = $this->bars;
		$stats = $this->stats;
		$finished = array();

		foreach ($runs as $seq => $run) {
			$unix = $run["unix"];

			if (self::slot($this->marks[0], $unix) < 0) {
				continue;
			}

			$state = $run["state"];
			$broke = self::broke($state);
			$ended = $state !== Run::RUNNING;
			$elapsed = $run["elapsed"];

			$this->total++;

			if ($broke) {
				$this->failed++;
			}

			if ($ended) {
				$run["seq"] = $seq;
				$finished[] = $run;
			}

			for ($c = 0; $c < $charts; $c++) {
				$k = self::slot($this->marks[$c], $unix);

				if ($k < 0) {
					continue;
				}

				if ($state === Run::OK) {
					$bars[$c][$k]["ok"]++;
				} elseif ($broke) {
					$bars[$c][$k]["failed"]++;
				} else {
					$bars[$c][$k]["other"]++;
				}

				if (!$ended) {
					continue;
				}

				if ($bars[$c][$k]["finished"] === 0 || $elapsed < $bars[$c][$k]["min"]) {
					$bars[$c][$k]["min"] = $elapsed;
				}

				$bars[$c][$k]["max"] = max($bars[$c][$k]["max"], $elapsed);
				$bars[$c][$k]["finished"]++;
				$bars[$c][$k]["spent"] += $elapsed;
				$stats[$c] = self::spend($stats[$c], $elapsed);
			}
		}

		$this->bars = $bars;
		$this->stats = $stats;

		usort($finished, array("Task\\Activity", "earlier"));

		return $finished;
	}

	/**
	 * draw finishes every chart once its runs are counted: the heights of the
	 * bars, the duration lines, the hovers and the time markers.
	 */
	function draw() {
		foreach ($this->spans as $c => $span) {
			$bars = self::durations(self::scale($this->bars[$c]));
			$slowest = self::slowest($bars);
			$n = count($bars);
			$near = max(4, intdiv($n * 2, 15));

			foreach ($bars as $i => $bar) {
				$when = self::label($span, $this->edges[$c], $i);
				$bars[$i]["edge"] = self::edge($i, $n, $near);
				$bars[$i]["title"] = self::counts($bar, $when);
				$bars[$i]["hitTitle"] = self::averageText($bar, $when);
			}

			$this->bars[$c] = $bars;
			$this->lines[$c] = self::paths($bars, $span["smooth"], $slowest);
			$this->ticks[$c] = self::ticks($span, $this->edges[$c]);
		}
	}

	/** charts is each chart as the component draws it, in the order a page lays them out. */
	function charts() {
		$out = array();

		foreach ($this->spans as $c => $span) {
			$bars = array();

			foreach ($this->bars[$c] as $bar) {
				$bars[] = self::view($bar);
			}

			$out[] = array(
				"class" => $span["class"] === "" ? "activity-chart" : "activity-chart " . $span["class"],
				"title" => $span["title"],
				"label" => "Runs per " . $span["unit"] . ", " . strtolower($span["title"]),
				"n" => (string)$span["slots"],
				"stats" => self::summary($this->stats[$c]),
				"bars" => $bars,
				"lines" => $this->lines[$c],
				"ticks" => $this->ticks[$c],
			);
		}

		return $out;
	}

	/**
	 * span is one chart: what it covers, how wide a slot is, the class its
	 * markup takes and whether its lines curve.
	 *
	 * hours is zero for a calendar day. The wide class takes a row of its
	 * own, which a week in three hour slots needs.
	 */
	static function span($title, $unit, $slots, $hours, $class, $smooth) {
		return array(
			"title" => $title,
			"unit" => $unit,
			"slots" => $slots,
			"hours" => $hours,
			"class" => $class,
			"smooth" => $smooth,
		);
	}

	/**
	 * edges is where a span's slots start, oldest first, and where the last
	 * one ends: one more than it has slots, and now is in the last.
	 *
	 * A day span counts calendar days from the midnight of now, an hour span
	 * whole clock hours back from the hour now is in. Both read a wall clock
	 * in the zone of now, the day a page means by a day.
	 */
	static function edges($span, $now, $zone) {
		$midnight = \DateTime::parse_in_location(Clock::DAY, $now->format(Clock::DAY), $zone);
		$out = array();

		if ($span["hours"] === 0) {
			$first = $midnight->addDate(0, 0, 1 - $span["slots"]);

			for ($i = 0; $i <= $span["slots"]; $i++) {
				$out[] = $first->addDate(0, 0, $i);
			}

			return $out;
		}

		$last = intdiv($now->hour(), $span["hours"]) * $span["hours"];

		for ($i = 0; $i <= $span["slots"]; $i++) {
			$out[] = self::clock($midnight, $last + $span["hours"] * ($i - $span["slots"] + 1), $zone);
		}

		return $out;
	}

	/**
	 * clock is a whole hour of a day, the hour counted from its midnight and
	 * on another day where it is negative or past the end of this one, which
	 * is how an hour out of range reads.
	 */
	static function clock($midnight, $hour, $zone) {
		$shift = (int)floor($hour / 24);
		$day = $midnight->addDate(0, 0, $shift);
		$stamp = $day->format(Clock::DAY) . " " . self::two($hour - $shift * 24);

		return \DateTime::parse_in_location(self::HOUR, $stamp, $zone);
	}

	/** marks is each edge as a unix second, the number slot() searches. */
	static function marks($edges) {
		$out = array();

		foreach ($edges as $edge) {
			$out[] = $edge->unix();
		}

		return $out;
	}

	/**
	 * slot is the slot a unix second falls in, of the marks of a chart, and
	 * -1 outside the window. The marks rise, so the search halves them.
	 */
	static function slot($marks, $unix) {
		$last = count($marks) - 1;

		if ($last < 1 || $unix < $marks[0] || $unix >= $marks[$last]) {
			return -1;
		}

		$low = 0;
		$high = $last;

		while ($low + 1 < $high) {
			$mid = intdiv($low + $high, 2);

			if ($marks[$mid] <= $unix) {
				$low = $mid;

				continue;
			}

			$high = $mid;
		}

		return $low;
	}

	/** slots is a chart's empty bars, one per column. */
	static function slots($n) {
		$out = array();

		for ($i = 0; $i < $n; $i++) {
			$out[] = array(
				"ok" => 0,
				"failed" => 0,
				"other" => 0,
				"finished" => 0,
				"spent" => 0,
				"min" => 0,
				"max" => 0,
				"average" => 0,
				"averageNanos" => 0,
				"minNanos" => 0,
				"maxNanos" => 0,
				"okStyle" => "",
				"failedStyle" => "",
				"otherStyle" => "",
				"edge" => "",
				"title" => "",
				"hitTitle" => "",
			);
		}

		return $out;
	}

	/** blank is a window with no finished run in it yet. */
	static function blank() {
		return array(
			"count" => 0,
			"fastest" => 0,
			"mean" => 0,
			"slowest" => 0,
			"spent" => 0,
		);
	}

	/** spend counts a finished run into a window's stats. */
	static function spend($stats, $elapsed) {
		if ($stats["count"] === 0 || $elapsed < $stats["fastest"]) {
			$stats["fastest"] = $elapsed;
		}

		$stats["slowest"] = max($stats["slowest"], $elapsed);
		$stats["spent"] += $elapsed;
		$stats["count"]++;
		$stats["mean"] = intdiv($stats["spent"], $stats["count"]);

		return $stats;
	}

	/**
	 * summary is a window's finished runs in one line, as the chart's caption
	 * draws them, with no text where none finished.
	 */
	static function summary($stats) {
		if ($stats["count"] === 0) {
			return array("text" => "", "title" => "");
		}

		return array(
			"text" => "min/avg/max/total: " . Run::duration($stats["fastest"]) . " " . Run::duration($stats["mean"]) . " " . Run::duration($stats["slowest"]) . " " . Run::duration($stats["spent"]),
			"title" => self::plural($stats["count"], "finished run") . ": the fastest, the mean, the slowest and their time together",
		);
	}

	/**
	 * scale sets the heights of the parts of each bar against the busiest bar
	 * of the chart; a bar with runs is at least STUB tall, and an empty one
	 * has no part at all.
	 */
	static function scale($bars) {
		$busiest = 0;

		foreach ($bars as $bar) {
			$busiest = max($busiest, $bar["ok"] + $bar["failed"] + $bar["other"]);
		}

		if ($busiest === 0) {
			return $bars;
		}

		foreach ($bars as $i => $bar) {
			$total = $bar["ok"] + $bar["failed"] + $bar["other"];

			if ($total === 0) {
				continue;
			}

			$unit = 100 / $busiest;

			if ($total * $unit < self::STUB) {
				$unit = self::STUB / $total;
			}

			$bars[$i]["okStyle"] = self::height($bar["ok"] * $unit);
			$bars[$i]["failedStyle"] = self::height($bar["failed"] * $unit);
			$bars[$i]["otherStyle"] = self::height($bar["other"] * $unit);
		}

		return $bars;
	}

	/**
	 * durations finishes the duration columns of each bar: the mean of its
	 * finished runs, which the hover words, and the three the plot scales, in
	 * the nanoseconds Go holds a duration in.
	 *
	 * The plot's mean is the one that cannot be converted: Go sums nanoseconds
	 * and divides those, which truncates, and a mean of whole milliseconds
	 * scaled afterwards is a different number. The fastest and the slowest
	 * convert exactly, since scaling both sides of a quotient by the same
	 * whole number leaves the quotient alone.
	 */
	static function durations($bars) {
		foreach ($bars as $i => $bar) {
			if ($bar["finished"] > 0) {
				$bars[$i]["average"] = $bar["spent"] / $bar["finished"];
				$bars[$i]["averageNanos"] = intdiv($bar["spent"] * self::NANOS, $bar["finished"]);
			}

			$bars[$i]["minNanos"] = $bar["min"] * self::NANOS;
			$bars[$i]["maxNanos"] = $bar["max"] * self::NANOS;
		}

		return $bars;
	}

	/** slowest is the longest run of the bars, which is the top of their scale. */
	static function slowest($bars) {
		$out = 0;

		foreach ($bars as $bar) {
			if ($bar["finished"] > 0) {
				$out = max($out, $bar["max"]);
			}
		}

		return $out;
	}

	/** anyFinished reports whether a run ended in any bar of a chart. */
	static function anyFinished($bars) {
		foreach ($bars as $bar) {
			if ($bar["finished"] > 0) {
				return true;
			}
		}

		return false;
	}

	/**
	 * paths is a chart's durations as three lines against its slowest run,
	 * and empty where none of its runs finished, which draws no line at all.
	 */
	static function paths($bars, $smooth, $slowest) {
		if ($slowest === 0 && !self::anyFinished($bars)) {
			return array();
		}

		$top = $slowest * self::NANOS;

		return array(
			"max" => self::path($bars, $smooth, "maxNanos", $top),
			"min" => self::path($bars, $smooth, "minNanos", $top),
			"avg" => self::path($bars, $smooth, "averageNanos", $top),
			"label" => "Average, slowest and fastest run per column, the slowest " . Run::duration($slowest),
		);
	}

	/**
	 * path is one duration of every bar as a line, a point at the middle of
	 * each column: a bar with no finished run sits on the baseline, so the
	 * line runs the width of the chart.
	 *
	 * A smooth line is a monotone cubic curve, after Fritsch and Carlson: it
	 * passes through every point and never swings past one, below the
	 * baseline or over the slowest run. Any other line goes straight from
	 * point to point, since a curve through a ragged hour reads as a trend
	 * that is not there. One bar alone is a dot.
	 */
	static function path($bars, $smooth, $field, $slowest) {
		$n = count($bars);

		if ($n === 0) {
			return "";
		}

		$step = 100 / $n;
		$xs = array();
		$ys = array();

		foreach ($bars as $i => $bar) {
			$xs[] = ($i + 0.5) * $step;
			$ys[] = $bar["finished"] > 0 ? self::level($bar[$field], $slowest) : 100;
		}

		if ($n === 1) {
			return "M " . self::point($xs[0], $ys[0]) . " L " . self::point($xs[0], $ys[0]);
		}

		$out = "M " . self::point($xs[0], $ys[0]);

		if (!$smooth) {
			for ($i = 1; $i < $n; $i++) {
				$out .= " L " . self::point($xs[$i], $ys[$i]);
			}

			return $out;
		}

		$third = $step / 3;
		$tangents = self::tangents($ys, $step);

		for ($i = 0; $i < $n - 1; $i++) {
			$out .= " C " . self::point($xs[$i] + $third, $ys[$i] + $tangents[$i] * $third) . " " . self::point($xs[$i + 1] - $third, $ys[$i + 1] - $tangents[$i + 1] * $third) . " " . self::point($xs[$i + 1], $ys[$i + 1]);
		}

		return $out;
	}

	/**
	 * tangents is the slope of a curve at each of its points, spaced step
	 * apart: zero at a peak, a valley or a flat, and the harmonic mean of the
	 * slopes either side elsewhere, so the curve stays between its points.
	 */
	static function tangents($ys, $step) {
		$n = count($ys);
		$slopes = array();

		for ($i = 0; $i < $n - 1; $i++) {
			$slopes[] = ($ys[$i + 1] - $ys[$i]) / $step;
		}

		$out = array_fill(0, $n, 0);
		$out[0] = $slopes[0];
		$out[$n - 1] = $slopes[$n - 2];

		for ($i = 1; $i < $n - 1; $i++) {
			$before = $slopes[$i - 1];
			$after = $slopes[$i];

			if ($before * $after <= 0) {
				continue;
			}

			$out[$i] = 2 * $before * $after / ($before + $after);
		}

		return $out;
	}

	/**
	 * level is where a duration sits in a plot, the slowest run under the
	 * headroom. Both are nanoseconds, as NANOS says.
	 */
	static function level($nanos, $slowest) {
		if ($slowest === 0) {
			return 100;
		}

		return 100 - $nanos / $slowest * (100 - self::HEADROOM);
	}

	/**
	 * spark draws the durations of the runs as a line, a run a point spread
	 * evenly from left to right, and the plot under it as a wash. One run
	 * alone is a flat line across.
	 */
	function spark($runs) {
		$n = count($runs);

		if ($n === 0) {
			return array();
		}

		$longest = 0;

		foreach ($runs as $run) {
			$longest = max($longest, $run["elapsed"]);
		}

		$xs = array();
		$ys = array();

		foreach ($runs as $i => $run) {
			$xs[] = $n === 1 ? 50 : $i * 100 / ($n - 1);
			$ys[] = $longest === 0 ? 100 : 100 - $run["elapsed"] / $longest * (100 - self::HEADROOM);
		}

		if ($n === 1) {
			$line = "M 0 " . self::number($ys[0]) . " L 100 " . self::number($ys[0]);
			$area = "M 0 100 L 0 " . self::number($ys[0]) . " L 100 " . self::number($ys[0]) . " L 100 100 Z";
		} else {
			$parts = array();

			foreach ($runs as $i => $run) {
				$parts[] = ($i === 0 ? "M " : "L ") . self::point($xs[$i], $ys[$i]);
			}

			$line = implode(" ", $parts);
			$area = "M 0 100 L " . substr($line, 2) . " L 100 100 Z";
		}

		return array(
			"label" => "Duration of the latest " . $n . " runs, the longest " . Run::duration($longest),
			"caption" => "duration of the latest " . $n . " runs",
			"line" => $line,
			"area" => $area,
			"longest" => Run::duration($longest),
			"points" => $this->points($runs, $xs),
		);
	}

	/**
	 * points is a run of the duration line each: the stretch of the plot that
	 * hovers it, where the run sits on the line, and what the hover says.
	 */
	function points($runs, $xs) {
		$n = count($runs);
		$out = array();

		foreach ($runs as $i => $run) {
			$from = $i > 0 ? ($xs[$i - 1] + $xs[$i]) / 2 : 0;
			$to = $i < $n - 1 ? ($xs[$i] + $xs[$i + 1]) / 2 : 100;
			$title = Run::duration($run["elapsed"]) . ", " . $run["at"]->in($this->zone)->format(self::WHEN) . " · " . $run["state"];
			$edge = self::edge($i, $n, 4);

			$out[] = array(
				"class" => $edge === "" ? "hit tip" : "hit tip " . $edge,
				"style" => "left:" . self::number($from) . "%;width:" . self::number($to - $from) . "%",
				"at" => self::number($xs[$i]),
				"title" => $title,
				"failed" => self::broke($run["state"]),
			);
		}

		return $out;
	}

	/** latest is the newest DURATIONS of the finished runs, which the line draws. */
	static function latest($finished) {
		return count($finished) > self::DURATIONS ? array_slice($finished, -self::DURATIONS) : $finished;
	}

	/**
	 * median is the middle duration of the runs, and the mean of the middle
	 * two of an even count.
	 */
	static function median($runs) {
		if (count($runs) === 0) {
			return 0;
		}

		$all = array();

		foreach ($runs as $run) {
			$all[] = $run["elapsed"];
		}

		sort($all);

		$mid = intdiv(count($all), 2);

		if (count($all) % 2 === 0) {
			return intdiv($all[$mid - 1] + $all[$mid], 2);
		}

		return $all[$mid];
	}

	/**
	 * earlier orders two runs by when they started, the earlier first, and by
	 * the order they arrived in where that is the same, so the line is the
	 * same on every reload.
	 *
	 * The fraction of the second decides, not the second: time.Time.Before is
	 * what orders the Go line and a column holds the fraction to the
	 * millisecond, so four runs of one second read as four orderings here and
	 * one there. Whole seconds place a run in its slot and nothing else.
	 */
	static function earlier($one, $other) {
		if ($one["nano"] === $other["nano"]) {
			return $one["seq"] - $other["seq"];
		}

		return $one["nano"] < $other["nano"] ? -1 : 1;
	}

	/** label names a slot for its hover: the day, or the day and the hours. */
	static function label($span, $edges, $i) {
		if ($span["hours"] === 0) {
			return $edges[$i]->format("Mon " . Clock::DAY);
		}

		return $edges[$i]->format("Mon ") . self::two($edges[$i]->hour()) . ":00–" . self::two($edges[$i + 1]->hour()) . ":00";
	}

	/**
	 * ticks mark the time under a chart: a day a week for the days, today the
	 * last; every third clock hour for the hours, labelled at 03:00, 09:00,
	 * 15:00 and 21:00; and each day at its noon for the week.
	 */
	static function ticks($span, $edges) {
		$n = $span["slots"];
		$out = array();

		if ($span["hours"] === 0) {
			for ($i = $n - 1; $i >= 0; $i -= 7) {
				$out[] = array("label" => $edges[$i]->format("Jan 2"), "at" => self::middle($i, $n));
			}

			return array_reverse($out);
		}

		if ($span["hours"] === 1) {
			for ($i = 0; $i < $n; $i++) {
				$clock = $edges[$i]->hour();

				if ($clock % 3 !== 0) {
					continue;
				}

				$out[] = array(
					"label" => $clock % 6 === 3 ? self::two($clock) . ":00" : "",
					"at" => self::middle($i, $n),
				);
			}

			return $out;
		}

		for ($i = 1; $i < $n; $i++) {
			if ($edges[$i]->hour() === 12) {
				$out[] = array("label" => $edges[$i]->format("Mon"), "at" => self::number($i * 100 / $n));
			}
		}

		return $out;
	}

	/** view is a bar as the component draws it: its hovers, and each part's height. */
	static function view($bar) {
		return array(
			"class" => "tip " . $bar["edge"],
			"tip" => $bar["title"],
			"hit" => $bar["hitTitle"],
			"ok" => $bar["okStyle"],
			"failed" => $bar["failedStyle"],
			"other" => $bar["otherStyle"],
		);
	}

	/**
	 * counts says what a bar holds and when, for its hover: its runs first,
	 * then the ones of them that did not end ok.
	 */
	static function counts($bar, $when) {
		$total = $bar["ok"] + $bar["failed"] + $bar["other"];

		if ($total === 0) {
			return "no runs, " . $when;
		}

		$text = self::plural($total, "run") . ", " . $when;

		if ($bar["failed"] > 0) {
			$text .= " · " . $bar["failed"] . " failed";
		}

		if ($bar["other"] > 0) {
			$text .= " · " . $bar["other"] . " running or stopped";
		}

		return $text;
	}

	/** averageText says a bar's durations and when, for the hover of its column. */
	static function averageText($bar, $when) {
		if ($bar["finished"] === 0) {
			return "no finished runs, " . $when;
		}

		if ($bar["finished"] === 1) {
			return Run::duration($bar["average"]) . ", 1 run, " . $when;
		}

		return "avg " . Run::duration($bar["average"]) . " · min " . Run::duration($bar["min"]) . " · max " . Run::duration($bar["max"]) . " over " . self::plural($bar["finished"], "run") . ", " . $when;
	}

	/** broke reports whether a state is one a run ended badly in. */
	static function broke($state) {
		return $state === Run::FAILED || $state === Run::KILLED;
	}

	/**
	 * edge anchors the hover of slot i of n near either end of the chart, so
	 * that the hover stays within the chart: the first near slots open to the
	 * right, the last near slots to the left, and the rest sit under their own
	 * column.
	 */
	static function edge($i, $n, $near) {
		if ($i < $near) {
			return "tip-start";
		}

		if ($i >= $n - $near) {
			return "tip-end";
		}

		return "";
	}

	/** height is the height of a part of a bar, as its style attribute reads. */
	static function height($percent) {
		return "height:" . self::number($percent) . "%";
	}

	/** middle is where the column of slot i of n sits, in percent. */
	static function middle($i, $n) {
		return self::number(($i + 0.5) * 100 / $n);
	}

	/** point is one point of a path. */
	static function point($x, $y) {
		return self::number($x) . " " . self::number($y);
	}

	/** number writes a coordinate, or a percentage, of the 100 by 100 plot. */
	static function number($v) {
		return sprintf("%.2f", min(max($v, 0), 100));
	}

	/** plural says a count of something, the word plural for any count but one. */
	static function plural($n, $word) {
		return $n === 1 ? "1 " . $word : $n . " " . $word . "s";
	}

	/** two writes a number of an hour or a day in two digits. */
	static function two($n) {
		return $n < 10 ? "0" . $n : (string)$n;
	}
}
