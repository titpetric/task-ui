<?php

// @route GET /inventory/hosts/{name}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_host is the page of one host of the inventory, which is the expanded row
// of its line on the Inventory page as well: the containers Beszel measures and
// docker ps sizes, side by side, the durable filesystems, and the host's month
// in charts.
//
// Every fact is the host's own and none of it is a row of this instance, so it
// all comes over the binding in one call: Task\Host::host() reads the hub, the
// ssh login and what the host last told, merges the two views of a container
// and hands back one answer. The page words that answer and decides nothing.
//
// The reader's hosts permission goes into the call, because it is what makes
// task-ui read the host now rather than answer out of what it kept. The page is
// gated on that permission outright: the row it draws is the one the Inventory
// page opens only for a reader who has it.
class mod_host implements Module {
	/** LEAD is the sentence under the heading, as HTML: it links back to the page the host is a line of. */
	const LEAD = 'A host of the <a href="%s">inventory</a>, as its line on the Inventory page opens: its containers, its durable filesystems and its month in charts.';

	/** MISSING is what the page of a host Beszel does not monitor says. */
	const MISSING = "Beszel monitors no host ";

	/** COLORS is how many colours the series of a chart take turns at, which is hostinfo's count. */
	const COLORS = 6;

	function name() {
		return "inventory";
	}

	function title() {
		return "Inventory";
	}

	function icon() {
		return "server";
	}

	function tables() {
		return array();
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);
		$hosts = $chrome->allowed("hosts");

		if (!$hosts) {
			$view->refused();

			return;
		}

		$name = Task\Route::param("name");
		$facts = Task\Host::host($name, $hosts);

		if ($facts["error"] !== "") {
			$view->missing($facts["error"]);

			return;
		}

		if (!$facts["found"]) {
			$view->missing(self::MISSING . $name . ".");

			return;
		}

		$chrome->heading($name, array(
			array("text" => "Inventory", "href" => "/inventory"),
			array("text" => $name, "href" => ""),
		));

		$now = Task\Clock::now();
		$scope = $this->scope($facts, $now);

		$view->page($scope, "host.tpl");
	}

	/**
	 * scope is what the template reads: the head, the host's own facts as one
	 * line, its containers, its durable filesystems and its charts.
	 *
	 * The two notes that say why a cell has a dash are the host's and not a
	 * container's, so they are written once and the rows read them from here.
	 *
	 * The charts of the host and those of its containers are one list, as the
	 * Go page draws them: one after the other in the one row of charts, and the
	 * page says Beszel holds no records only where there is neither.
	 */
	function scope($facts, $now) {
		$system = $facts["system"];
		$containers = array();

		foreach ($facts["containers"] as $container) {
			$containers[] = self::container($container);
		}

		$filesystems = array();

		foreach ($facts["filesystems"] as $filesystem) {
			$filesystems[] = self::filesystem($filesystem);
		}

		$charts = array_merge($this->charts($facts, $now), $this->containerCharts($facts, $now));
		$head = new Task\Head($system["name"], sprintf(self::LEAD, Task\Route::href("/inventory")));

		return array(
			"head" => $head->data(),
			"name" => $system["name"],
			"host" => $system["host"],
			"login" => $facts["login"],
			"cpus" => $system["cpus"],
			"up" => $system["up"],
			"uptime" => $system["uptime"],
			"status" => $system["status"],
			"agent" => $system["agent"],
			"read" => self::read($facts["read"]),
			"count" => count($containers),
			"containers" => $containers,
			"measured" => $facts["measured"],
			"sized" => $facts["sized"],
			"known" => $facts["measured"] || $facts["sized"],
			"unsized" => $facts["unsized"],
			"measureNote" => self::measureNote($facts),
			"sizeNote" => self::sizeNote($facts),
			"filesystems" => $filesystems,
			"fromDF" => $facts["from_df"],
			"charts" => $charts,
		);
	}

	/**
	 * container is one row of the Containers table: what docker calls it, how
	 * it is doing, how much it takes and how much it has written.
	 *
	 * A container Beszel holds no measures of, or docker ps did not list, has a
	 * dash in those cells instead and the title of the cell says why.
	 *
	 * Each number cell carries the measure itself as its sort key, because the
	 * script sorts the table on that and not on the words; a dash sorts as -1.
	 * The sizes are counts of bytes and the rest are floats, which number()
	 * writes the way the Go page writes one.
	 */
	static function container($container) {
		$state = $container["state"];
		$measured = $container["measured"];

		return array(
			"name" => $container["name"],
			"image" => $container["image"],
			"status" => $container["status"],
			"faint" => $state !== "" && $state !== "running",
			"measured" => $measured,
			"cpuV" => self::number($container["cpu"]),
			"cpu" => sprintf("%.1f", $container["cpu"]),
			"memoryV" => self::number($container["memory"]),
			"memory" => Task\Size::megabytes($container["memory"]),
			"shares" => $container["memory_share"] > 0,
			"share" => sprintf("%.1f", $container["memory_share"]),
			"sized" => $container["sized"],
			"sizeV" => $container["size_rw"],
			"size" => Task\Size::bytes($container["size_rw"]),
			"virtual" => Task\Size::bytes($container["size_virtual"]),
			"traffic" => $measured && $container["net"] > 0,
			"netV" => self::number($container["net"]),
			"net" => Task\Size::rate($container["net"]),
		);
	}

	/** filesystem is one row of the Filesystems table: where it is mounted, what it is, and how full. */
	static function filesystem($filesystem) {
		return array(
			"mount" => $filesystem["mount"],
			"device" => $filesystem["device"],
			"type" => $filesystem["type"],
			"size" => Task\Size::bytes($filesystem["size"]),
			"used" => Task\Size::bytes($filesystem["used"]),
			"free" => Task\Size::bytes($filesystem["free"]),
			"use" => Task\Meter::of($filesystem["percent"]),
		);
	}

	/** measureNote says why a container has no measures. */
	static function measureNote($facts) {
		return $facts["measured"] ? "Beszel reports no measures of the container" : "Beszel reports no containers of the host";
	}

	/** sizeNote says why a container has no size. */
	static function sizeNote($facts) {
		return $facts["sized"] ? "docker ps does not list the container" : $facts["unsized"];
	}

	/**
	 * charts is the host's month as Beszel recorded it: how busy its CPU was,
	 * the memory in use against its RAM, how full each filesystem the agent
	 * watches was, and the traffic both ways.
	 *
	 * A hub that kept no record of the host draws no chart at all, and the page
	 * says so instead. The memory is read against the largest RAM of the month
	 * and not against the busiest moment, so that a host whose memory never
	 * filled still draws against the whole of it.
	 *
	 * The root filesystem is always drawn, because Beszel reports it of every
	 * host; the others are the ones the agent was told to watch, each by the
	 * name it was given.
	 */
	function charts($facts, $now) {
		if (count($facts["samples"]) === 0) {
			return array();
		}

		$cpu = array();
		$used = array();
		$sent = array();
		$recv = array();
		$disks = array("/" => array());
		$ram = 0;

		foreach ($facts["samples"] as $sample) {
			$at = Task\Clock::seconds($sample["at"]);
			$cpu[] = Task\Chart::point($at, $sample["cpu"]);
			$used[] = Task\Chart::point($at, $sample["memory_used"]);
			$ram = max($ram, $sample["memory"]);
			$sent[] = Task\Chart::point($at, $sample["sent"]);
			$recv[] = Task\Chart::point($at, $sample["recv"]);
			$disks["/"][] = Task\Chart::point($at, Task\Chart::fill($sample["disk_used"], $sample["disk"]));

			foreach ($sample["extra"] as $mount) {
				$name = $mount["name"];

				if (!isset($disks[$name])) {
					$disks[$name] = array();
				}

				$disks[$name][] = Task\Chart::point($at, Task\Chart::fill($mount["used"], $mount["total"]));
			}
		}

		$from = Task\Chart::from($now);
		$busy = new Task\Chart("CPU", Task\Chart::percent(), $from, $now, false, array(Task\Chart::series("CPU", 0, $cpu)));
		$memory = new Task\Chart("Memory", Task\Chart::gib($ram), $from, $now, false, array(Task\Chart::series("Used", 0, $used)));
		$disk = new Task\Chart("Disk", Task\Chart::percent(), $from, $now, false, self::mounts($disks));
		$net = new Task\Chart("Network", Task\Chart::rate(), $from, $now, false, array(Task\Chart::series("Sent", 0, $sent), Task\Chart::series("Received", 1, $recv)));

		return array($busy->data(), $memory->data(), $disk->data(), $net->data());
	}

	/** mounts are the filesystems of the records as series, by name and each in its own colour. */
	static function mounts($disks) {
		$names = array_keys($disks);

		sort($names);

		$series = array();

		foreach ($names as $i => $name) {
			$series[] = Task\Chart::series($name, $i % self::COLORS, $disks[$name]);
		}

		return $series;
	}

	/**
	 * containerCharts is the month of the host's busiest containers: the five
	 * of the highest mean memory, and the five of the highest mean CPU.
	 *
	 * The two read the same records and so hold the same containers, but in
	 * their own order and so in their own colours: a container is the colour of
	 * where it stands in the chart it is in.
	 */
	function containerCharts($facts, $now) {
		if (count($facts["container_samples"]) === 0) {
			return array();
		}

		$memory = array();
		$cpu = array();

		foreach ($facts["container_samples"] as $sample) {
			$at = Task\Clock::seconds($sample["at"]);

			foreach ($sample["containers"] as $stat) {
				$name = $stat["name"];

				if (!isset($memory[$name])) {
					$memory[$name] = array();
					$cpu[$name] = array();
				}

				$memory[$name][] = Task\Chart::point($at, $stat["memory"]);
				$cpu[$name][] = Task\Chart::point($at, $stat["cpu"]);
			}
		}

		$from = Task\Chart::from($now);
		$fullest = new Task\Chart("Containers by memory", Task\Chart::megabytes(), $from, $now, false, Task\Chart::busiest($memory));
		$busiest = new Task\Chart("Containers by CPU", Task\Chart::share(), $from, $now, false, Task\Chart::busiest($cpu));

		return array($fullest->data(), $busiest->data());
	}

	/** read is how long since the host last answered over ssh, and "" where it never has. */
	static function read($stamp) {
		return $stamp === "" ? "" : Task\Clock::ago($stamp);
	}

	/**
	 * number writes a float into a sort key the way the Go page writes one,
	 * which is Go's %v: the shortest decimal that reads back as the same float,
	 * in exponent form above a million and below a ten-thousandth, the exponent
	 * signed and at least two digits.
	 *
	 * json_encode answers the same shortest digits, because the runtime's
	 * encoder is Go's, but not in the same shape: it writes a plain decimal up
	 * to 1e21 where %v turns at 1e6, and it trims a small negative exponent to
	 * one digit where %v pads it. The sort keys are compared against the Go page
	 * byte for byte, so the digits are read back out of it and written again.
	 *
	 * This takes a float. A count of bytes is an integer on both sides and is
	 * written as itself.
	 */
	static function number($value) {
		$text = json_encode($value);
		$sign = "";

		if (strpos($text, "-") === 0) {
			$sign = "-";
			$text = substr($text, 1);
		}

		$exponent = 0;
		$e = strpos($text, "e");

		if ($e !== false) {
			$exponent = (int)substr($text, $e + 1);
			$text = substr($text, 0, $e);
		}

		$dot = strpos($text, ".");
		$digits = $text;
		$point = strlen($text);

		if ($dot !== false) {
			$digits = substr($text, 0, $dot) . substr($text, $dot + 1);
			$point = $dot;
		}

		$leading = ltrim($digits, "0");
		$point = $point - (strlen($digits) - strlen($leading)) + $exponent;
		$digits = rtrim($leading, "0");

		if ($digits === "") {
			return $sign . "0";
		}

		if ($point - 1 < -4 || $point - 1 >= 6) {
			return $sign . self::exponent($digits, $point - 1);
		}

		if ($point <= 0) {
			return $sign . "0." . str_repeat("0", -$point) . $digits;
		}

		if ($point >= strlen($digits)) {
			return $sign . $digits . str_repeat("0", $point - strlen($digits));
		}

		return $sign . substr($digits, 0, $point) . "." . substr($digits, $point);
	}

	/** exponent is the digits of a float in exponent form: one digit, the rest, and the power of ten. */
	static function exponent($digits, $power) {
		$out = substr($digits, 0, 1);

		if (strlen($digits) > 1) {
			$out .= "." . substr($digits, 1);
		}

		$mark = "+";

		if ($power < 0) {
			$mark = "-";
			$power = -$power;
		}

		$pad = $power < 10 ? "0" : "";

		return $out . "e" . $mark . $pad . $power;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_host;

	$mod->serve();
}
