<?php

namespace Task;

// Chart is a month of measures drawn the way rrdtool draws them: a line a
// series over a day axis with a gridline a week, a value axis of three or four
// gridlines in the chart's unit, and a gap where a series has no data.
//
// It is computation over the records Beszel answers with and nothing else, so
// it is PHP. A page hands it its series and draws what data() holds through the
// line-chart component, which writes the drawing raw.
//
// The drawing is frontend/view/chart.go here: the same geometry, the same
// gridlines, the same gap handling and the same legend, so the two front ends
// draw one picture. Keep them in step.
//
// A point with no measure is a gap. Go writes one as a NaN; this runtime has no
// is_nan, so a gap is a null and nothing compares it.
class Chart {
	/** The drawing's size in its own units. */
	const WIDTH = 400;

	const HEIGHT = 140;

	/** The margins of the plot: the value labels on the left, the dates under it. */
	const LEFT = 52;

	const RIGHT = 6;

	const TOP = 8;

	const BOTTOM = 20;

	/**
	 * STEP is the span of a record in seconds, where the series have no two
	 * points to tell it by: the hub's eight hour records.
	 */
	const STEP = 28800;

	/** MONTH is how far back a chart of the inventory reaches, Beszel's month. */
	const MONTH = "-720h";

	/** BUSIEST is how many series a chart of containers draws: the busiest of the month. */
	const BUSIEST = 5;

	/**
	 * EPOCH is how many seconds the unix epoch is past the first second of the
	 * year 1, which is the one a Go time truncates against.
	 */
	const EPOCH = 62135596800;

	/** NICE are the multiples a value axis steps in, one of them times a power of ten. */
	const NICE = array(1, 2, 2.5, 5, 10);

	/** SHARE is the top of a share's own axis, its hundred, in four steps. */
	const SHARE = 100;

	/** title is the caption over the drawing. */
	var $title = "";

	/** svg is the drawing. */
	var $svg = "";

	/** legend is each series' least, mean and most, worded in the chart's unit. */
	var $legend = array();

	/** total is the stack's own legend row, null on a chart that is not stacked. */
	var $total = null;

	/** empty says no series has a point in the span, and the page says so instead. */
	var $empty = false;

	/** from and to are the span, as unix seconds. */
	var $from = 0;

	var $to = 0;

	/** top is the value the top edge of the plot stands for. */
	var $top = 0;

	/**
	 * The chart of the series over the span from from to to, both times in the
	 * zone the date axis counts its days in.
	 *
	 * stacked draws each series as an area over the one before, and the series
	 * then share their times. A chart of one series alone is drawn as an area
	 * too, stacked or not.
	 *
	 * A series is what series() holds and a point what point() holds.
	 */
	function __construct($title, $unit, $from, $to, $stacked, $series) {
		$this->title = $title;
		$this->from = $from->unix();
		$this->to = $to->unix();
		$kept = array();

		foreach ($series as $one) {
			$one["points"] = $this->within($one["points"]);
			$kept[] = $one;
		}

		$drawn = $stacked ? self::stack($kept) : $kept;
		$largest = (float)$unit["max"];
		$points = 0;

		foreach ($drawn as $one) {
			foreach ($one["points"] as $point) {
				if ($point["v"] === null) {
					continue;
				}

				$largest = max($largest, $point["v"]);
				$points++;
			}
		}

		if ($points === 0) {
			$this->empty = true;

			return;
		}

		$scaled = self::scale($unit, $largest);
		$divisor = $scaled["divisor"];
		$suffix = $scaled["suffix"];
		$nice = self::nice($largest / $divisor);
		$top = $nice["top"];
		$step = $nice["step"];

		// A share reads against its own hundred and not against the busiest
		// moment of the month, so two of them read the same.
		// Both sides are cast: a float literal written here reads as an int
		// once the formatter has been over the file, and === would refuse it.
		if ((float)$unit["max"] === (float)self::SHARE && $largest <= self::SHARE) {
			$top = self::SHARE;
			$step = 25;
		}

		$this->top = $top * $divisor;
		$this->svg = $this->draw($from, $step * $divisor, $divisor, $suffix, $stacked, $kept, $drawn);

		foreach ($kept as $one) {
			$row = self::legendOf($one["name"], $one["color"], $one["points"], $divisor, $suffix);

			if ($row === null) {
				continue;
			}

			$this->legend[] = $row;
		}

		if (!$stacked || count($kept) < 2) {
			return;
		}

		$this->total = self::legendOf("Total", -1, $drawn[count($drawn) - 1]["points"], $divisor, $suffix);
	}

	/** data is the scope the line-chart component reads. */
	function data() {
		return array(
			"title" => $this->title,
			"svg" => $this->svg,
			"legend" => $this->legend,
			"total" => $this->total,
			"empty" => $this->empty,
		);
	}

	/**
	 * draw is the drawing: the value gridlines with their labels, the day axis,
	 * the series from the last to the first so that the first lies on top, and
	 * the axis under them.
	 */
	function draw($from, $step, $divisor, $suffix, $stacked, $series, $drawn) {
		$svg = sprintf('<svg class="chart-svg" viewBox="0 0 %g %g" role="img" aria-label="%s">', self::WIDTH, self::HEIGHT, self::escape($this->title));

		for ($v = 0; $v <= $this->top + $step / 1000; $v += $step) {
			$y = $this->y($v);
			$svg .= sprintf('<line class="grid" x1="%g" x2="%g" y1="%.1f" y2="%.1f"/>', self::LEFT, self::WIDTH - self::RIGHT, $y, $y);
			$svg .= sprintf('<text class="tick" x="%g" y="%.1f" text-anchor="end">%s</text>', self::LEFT - 4, $y + 3, self::word($v, $divisor, $suffix));
		}

		$svg .= $this->dates($from);
		$limit = self::gap($drawn);

		for ($i = count($drawn) - 1; $i >= 0; $i--) {
			$one = $drawn[$i];

			foreach (self::runs($one["points"], $limit) as $run) {
				$below = array();

				if ($stacked && $i > 0) {
					$below = self::pointsAt($drawn[$i - 1]["points"], $run);
				}

				if ($stacked || count($series) === 1) {
					$svg .= sprintf('<path class="area s%d" d="%s"/>', $one["color"], $this->band($run, $below));
				}

				if (count($run) === 1) {
					$svg .= sprintf('<circle class="dot s%d" cx="%.1f" cy="%.1f" r="1.6"/>', $one["color"], $this->x($run[0]["at"]), $this->y($run[0]["v"]));

					continue;
				}

				$svg .= sprintf('<path class="line s%d" d="%s"/>', $one["color"], $this->line($run));
			}
		}

		$svg .= sprintf('<line class="axis" x1="%g" x2="%g" y1="%g" y2="%g"/>', self::LEFT, self::WIDTH - self::RIGHT, self::HEIGHT - self::BOTTOM, self::HEIGHT - self::BOTTOM);

		return $svg . "</svg>";
	}

	/**
	 * dates is the day axis: a gridline and a date every Monday, and a tick
	 * under every day of the span.
	 *
	 * The days are the days of the span's own zone, which is the zone of the
	 * clock the page was drawn with, so a week starts where the server's Monday
	 * does.
	 */
	function dates($from) {
		$zone = $from->location();
		$midnight = \DateTime::parse_in_location(Clock::DAY, $from->format(Clock::DAY), $zone);
		$svg = "";

		for ($day = $midnight; $day->unix() <= $this->to; $day = $day->addDate(0, 0, 1)) {
			if ($day->unix() < $this->from) {
				continue;
			}

			$x = $this->x($day->unix());

			// The day of the week comes off the layout and not off weekday(),
			// which answers a host value no string is identical to.
			if ($day->format("Monday") === "Monday") {
				$svg .= sprintf('<line class="grid week" x1="%.1f" x2="%.1f" y1="%g" y2="%g"/>', $x, $x, self::TOP, self::HEIGHT - self::BOTTOM);
				$svg .= sprintf('<text class="tick" x="%.1f" y="%g" text-anchor="middle">%s</text>', $x, self::HEIGHT - 5, $day->format("Jan 2"));
			}

			$svg .= sprintf('<line class="day" x1="%.1f" x2="%.1f" y1="%g" y2="%g"/>', $x, $x, self::HEIGHT - self::BOTTOM, self::HEIGHT - self::BOTTOM + 3);
		}

		return $svg;
	}

	/** x is where a unix second sits across the plot. */
	function x($at) {
		$span = $this->to - $this->from;

		return self::LEFT + ($at - $this->from) / $span * (self::WIDTH - self::LEFT - self::RIGHT);
	}

	/** y is where a value sits down the plot. */
	function y($v) {
		return self::HEIGHT - self::BOTTOM - $v / $this->top * (self::HEIGHT - self::TOP - self::BOTTOM);
	}

	/** line is the path through a run of points. */
	function line($run) {
		$d = "";

		foreach ($run as $i => $point) {
			$cmd = $i === 0 ? "M" : "L";
			$d .= sprintf("%s%.1f %.1f", $cmd, $this->x($point["at"]), $this->y($point["v"]));
		}

		return $d;
	}

	/**
	 * band is the area between a run and the points under it, and the base line
	 * where there are none.
	 */
	function band($run, $below) {
		$d = $this->line($run);

		for ($i = count($run) - 1; $i >= 0; $i--) {
			$v = 0;

			if (isset($below[$i]) && $below[$i]["v"] !== null) {
				$v = $below[$i]["v"];
			}

			$d .= sprintf("L%.1f %.1f", $this->x($run[$i]["at"]), $this->y($v));
		}

		return $d . "Z";
	}

	/** within keeps the points of the span. */
	function within($points) {
		$kept = array();

		foreach ($points as $point) {
			if ($point["at"] < $this->from || $point["at"] > $this->to) {
				continue;
			}

			$kept[] = $point;
		}

		return $kept;
	}

	/**
	 * from is where a chart of the inventory starts: Beszel's month before the
	 * clock the page was drawn with.
	 */
	static function from($now) {
		$month = \Time\Duration::parse(self::MONTH);

		return $now->add($month);
	}

	/** series is a line of a chart: its name, one of the six series colours, and its points, oldest first. */
	static function series($name, $color, $points) {
		return array(
			"name" => $name,
			"color" => $color,
			"points" => $points,
		);
	}

	/** point is a measure at a unix second; a null value is a gap. */
	static function point($at, $v) {
		return array("at" => $at, "v" => $v);
	}

	/**
	 * stack sums the series in order: each point the sum of its own and those
	 * of the series before it at its time, a missing one counting zero.
	 */
	static function stack($series) {
		$stacked = array();
		$sums = array();

		foreach ($series as $one) {
			$points = array();

			foreach ($one["points"] as $point) {
				$at = $point["at"];
				$v = $point["v"] === null ? 0 : $point["v"];
				$sums[$at] = isset($sums[$at]) ? $sums[$at] + $v : $v;
				$points[] = self::point($at, $sums[$at]);
			}

			$stacked[] = self::series($one["name"], $one["color"], $points);
		}

		return $stacked;
	}

	/**
	 * pointsAt is, for each point of a run, the point of others at its time,
	 * null where others has none.
	 */
	static function pointsAt($others, $run) {
		$at = array();

		foreach ($others as $point) {
			$at[$point["at"]] = $point["v"];
		}

		$found = array();

		foreach ($run as $point) {
			$v = isset($at[$point["at"]]) ? $at[$point["at"]] : null;
			$found[] = self::point($point["at"], $v);
		}

		return $found;
	}

	/**
	 * gap is the longest span between two points a line bridges, in seconds:
	 * one and a half of the shortest span between two points of a series.
	 */
	static function gap($series) {
		$step = 0;

		foreach ($series as $one) {
			$points = $one["points"];
			$n = count($points);

			for ($i = 1; $i < $n; $i++) {
				$d = $points[$i]["at"] - $points[$i - 1]["at"];

				if ($d > 0 && ($step === 0 || $d < $step)) {
					$step = $d;
				}
			}
		}

		if ($step === 0) {
			$step = self::STEP;
		}

		return $step * 3 / 2;
	}

	/** runs splits points at a gap, at a null and at a span past limit. */
	static function runs($points, $limit) {
		$all = array();
		$run = array();

		foreach ($points as $i => $point) {
			if ($point["v"] === null || ($i > 0 && $point["at"] - $points[$i - 1]["at"] > $limit)) {
				if (count($run) > 0) {
					$all[] = $run;
				}

				$run = array();
			}

			if ($point["v"] === null) {
				continue;
			}

			$run[] = $point;
		}

		if (count($run) > 0) {
			$all[] = $run;
		}

		return $all;
	}

	/**
	 * legendOf words a series' least, mean and most; null for a series with no
	 * point in the span.
	 */
	static function legendOf($name, $color, $points, $divisor, $suffix) {
		$values = array();
		$sum = 0;

		foreach ($points as $point) {
			if ($point["v"] === null) {
				continue;
			}

			$values[] = $point["v"];
			$sum += $point["v"];
		}

		$n = count($values);

		if ($n === 0) {
			return null;
		}

		sort($values);

		return array(
			"name" => $name,
			"color" => $color,
			"min" => self::word($values[0], $divisor, $suffix),
			"avg" => self::word($sum / $n, $divisor, $suffix),
			"max" => self::word($values[$n - 1], $divisor, $suffix),
		);
	}

	/**
	 * nice is a top at or over largest and a step that divides it in three or
	 * four: one, two, two and a half or five times a power of ten.
	 */
	static function nice($largest) {
		if ($largest <= 0) {
			return array("top" => 1, "step" => 0.5);
		}

		$raw = $largest / 3;
		$magnitude = self::magnitude($raw);
		$step = 0;

		foreach (self::NICE as $m) {
			$step = $m * $magnitude;

			if ($step >= $raw) {
				break;
			}
		}

		return array("top" => $step * ceil($largest / $step - 1e-09), "step" => $step);
	}

	/**
	 * magnitude is the largest power of ten at or under a positive value, which
	 * is what the floor of its base ten logarithm names. The runtime has no
	 * log10, so the exponent is counted.
	 */
	static function magnitude($value) {
		$exponent = 0;
		$scaled = $value;

		while ($scaled >= 10) {
			$scaled = $scaled / 10;
			$exponent++;
		}

		while ($scaled < 1) {
			$scaled = $scaled * 10;
			$exponent--;
		}

		return pow(10, $exponent);
	}

	/**
	 * unit is what a chart measures: the suffixes it words a value in, each
	 * base times the one before, and a fixed top of its scale or zero to fit
	 * the data.
	 */
	static function unit($suffixes, $max = 0, $base = 1024) {
		return array(
			"suffixes" => $suffixes,
			"max" => $max,
			"base" => $base,
		);
	}

	/** percent is the unit of a share, 0 to 100. */
	static function percent() {
		return self::unit(array("%"), self::SHARE);
	}

	/** share is the unit of a percentage with no top of its own, which fits the data. */
	static function share() {
		return self::unit(array("%"));
	}

	/** cores is the unit of a count of busy cores. */
	static function cores() {
		return self::unit(array(" cores"));
	}

	/** gib is the unit of a size in GB as Beszel counts one, 2^30 bytes. */
	static function gib($max = 0) {
		return self::unit(array(" GiB", " TiB"), $max);
	}

	/** megabytes is the unit of a size in MB, as Beszel counts container memory. */
	static function megabytes() {
		return self::unit(array(" MB", " GB"));
	}

	/** rate is the unit of traffic in bytes a second. */
	static function rate() {
		return self::unit(array(" B/s", " KB/s", " MB/s", " GB/s"));
	}

	/**
	 * scale is the divisor and the suffix a unit words its values in: the
	 * largest suffix the largest value of the chart reaches.
	 */
	static function scale($unit, $largest) {
		$suffixes = $unit["suffixes"];
		$last = count($suffixes) - 1;
		$divisor = 1;

		for ($i = 0; $i < $last; $i++) {
			if ($largest < $divisor * $unit["base"]) {
				return array("divisor" => $divisor, "suffix" => $suffixes[$i]);
			}

			$divisor = $divisor * $unit["base"];
		}

		return array("divisor" => $divisor, "suffix" => $suffixes[$last]);
	}

	/**
	 * word words a value by a divisor and a suffix: one decimal under ten, whole
	 * above.
	 *
	 * Both are cast: a division of two integers answers an integer here, and an
	 * integer is not identical to the float it equals, so a whole value would
	 * read as a tenth.
	 */
	static function word($v, $divisor, $suffix) {
		$scaled = (float)$v / (float)$divisor;

		if ($scaled < 10 && $scaled !== (float)(int)$scaled) {
			return sprintf("%.1f", $scaled) . $suffix;
		}

		return sprintf("%.0f", $scaled) . $suffix;
	}

	/**
	 * step is the span of the records of a history, in seconds: the shortest
	 * between two of a host, rounded to the hour, and the hub's eight hours
	 * where no host has two.
	 */
	static function step($history) {
		$step = 0;

		foreach ($history as $samples) {
			$n = count($samples);

			for ($i = 1; $i < $n; $i++) {
				$d = Clock::seconds($samples[$i]["at"]) - Clock::seconds($samples[$i - 1]["at"]);

				if ($d >= 3600 && ($step === 0 || $d < $step)) {
					$step = $d;
				}
			}
		}

		if ($step === 0) {
			return self::STEP;
		}

		$over = $step % 3600;

		return $over + $over < 3600 ? $step - $over : $step - $over + 3600;
	}

	/**
	 * truncate rounds a unix second down to a multiple of a step, so that the
	 * records of several hosts a minute apart share one moment.
	 *
	 * The multiples are counted from the first second of the year 1 and not
	 * from the epoch, because that is what a Go time truncates against: a step
	 * the epoch does not divide, five hours for one, lands on another moment
	 * either way, and the two front ends would draw a different axis.
	 */
	static function truncate($at, $step) {
		if ($step <= 0) {
			return $at;
		}

		return $at - ($at + self::EPOCH) % $step;
	}

	/**
	 * busiest is the BUSIEST series of the highest mean, highest first, each in
	 * its colour, of the points a page holds by name.
	 */
	static function busiest($points) {
		$means = array();

		foreach ($points as $name => $all) {
			$sum = 0;

			foreach ($all as $point) {
				$sum += $point["v"];
			}

			$means[] = array("name" => $name, "avg" => $sum / count($all));
		}

		usort($means, array("Task\\Chart", "busier"));
		$series = array();

		foreach ($means as $i => $mean) {
			if ($i === self::BUSIEST) {
				break;
			}

			$series[] = self::series($mean["name"], $i, $points[$mean["name"]]);
		}

		return $series;
	}

	/** busier orders two series by their mean, the higher first, and by name where it is the same. */
	static function busier($one, $other) {
		if ($one["avg"] === $other["avg"]) {
			return $one["name"] < $other["name"] ? -1 : 1;
		}

		return $one["avg"] > $other["avg"] ? -1 : 1;
	}

	/** fill is the percentage of a total in use, 0 where there is no total. */
	static function fill($used, $total) {
		if ($total <= 0) {
			return 0;
		}

		return $used / $total * 100;
	}

	/**
	 * escape writes a string into an attribute the way the Go template writes
	 * one, which is not what htmlspecialchars writes: a quote is numeric there
	 * and named here.
	 */
	static function escape($text) {
		return str_replace(array("&", "'", "<", ">", "\""), array("&amp;", "&#39;", "&lt;", "&gt;", "&#34;"), $text);
	}
}
