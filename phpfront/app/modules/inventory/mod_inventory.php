<?php

// @route GET /inventory

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_inventory is the hosts page: what Beszel monitors largest first, the
// pools the hosts sum into, and with the hosts permission the form that runs
// one command on many of them at once.
//
// Every fact of a host is the host's and not a row of this instance, so they
// come over the binding in one call: Task\Host::inventory() reads the hub, the
// ssh logins and what each host last told, sums the pools and hands back one
// answer. The page draws it and computes no fact a host decides.
//
// The reader's hosts permission goes into that call, because it is what makes
// task-ui read a host now rather than answer out of what the host last told.
class mod_inventory implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "The hosts task-ui keeps, largest first. For each host: whether it's up, the ssh login task-ui uses, its cores, and how busy its CPU, memory and disk are. Add a host by the address ssh signs in to and task-ui proves the login before keeping it; Beszel's hosts come in on their own, and every host stays until someone removes it. With Run on Hosts, you can run a command on many hosts at once, like pssh: each host gets a run of its own, with a recording, in Runs and the recents.";

	/** COMMANDS are the quick commands over the field, which a button fills it with. */
	const COMMANDS = array(
		"uptime",
		"df -h /",
		"free -h",
		"docker ps --format '{{.Names}}\t{{.Status}}'",
		"uname -r",
		"cat /etc/os-release | grep PRETTY",
	);

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
		return array("host", "run", "note", "user", "workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("inventory/view")) {
			$view->refused();

			return;
		}

		$chrome->heading("Inventory", array(array("text" => "Inventory", "href" => "")));

		$hosts = $chrome->allowed("hosts");
		$drops = $chrome->allowed("inventory/delete");
		$store = new Task\Inventory\Hosts(Task\DB::writer());
		$held = $store->all();
		$facts = Task\Host::inventory($hosts, self::names($held));

		// A host the hub reports and the inventory does not hold goes in on
		// this read, so the list keeps it once the hub stops answering.
		$held = $this->sync($store, $held, $facts["hosts"]);

		if ($facts["error"] !== "") {
			$chrome->reports($facts["error"]);
		}

		$logins = $facts["logins"];
		$last = array("batch" => null, "prompt" => null);

		if ($hosts) {
			$last = $this->batch($chrome);
		}

		$now = Task\Clock::now();
		$rows = array();

		foreach ($facts["hosts"] as $system) {
			$rows[] = $this->row($system, $logins, $hosts, isset($held[$system["host"]]) ? $held[$system["host"]] : null, $drops);
		}

		$head = new Task\Head("Inventory", self::LEAD);

		$view->page(array(
			"head" => $head->data(),
			"off" => $facts["off"],
			"integrations" => $chrome->allowed("integration/view"),
			"add" => $chrome->allowed("inventory/edit") ? Task\Route::href("/inventory/hosts") : "",
			"form" => $this->form($facts, $logins, $hosts),
			"batch" => $last["batch"],
			"prompt" => $last["prompt"],
			"cluster" => $this->cluster($facts, $now),
			"logins" => count($logins) > 0,
			"rows" => $rows,
			"now" => $now->format("15:04"),
		));
	}

	/** names is the name of each held host by the address it answers on, the inventory the binding folds the hub into. */
	static function names($held) {
		$out = array();

		foreach ($held as $host) {
			$out[$host->address] = $host->name;
		}

		return $out;
	}

	/**
	 * sync writes a host the hub reports and the inventory does not hold, and
	 * returns every held host by address.
	 *
	 * The hub is a source and the rows are the list: a system nobody entered
	 * is entered here, once, and a row the hub stopped reporting is left
	 * where it is. Nothing of a row is rewritten, so the name a person gave a
	 * host outlives the hub renaming it.
	 */
	function sync($store, $held, $systems) {
		$by = array();

		foreach ($held as $host) {
			$by[$host->address] = $host;
		}

		foreach ($systems as $system) {
			$address = $system["host"];

			if ($address === "" || isset($by[$address])) {
				continue;
			}

			$id = $store->seen($system["name"], $address);

			if ($id === "") {
				continue;
			}

			$row = $store->find($address);

			if (is_array($row)) {
				$by[$address] = new Task\Model\Host($row);
			}
		}

		return $by;
	}

	/**
	 * form is the Run on the hosts block: the quick commands, the hosts to
	 * pick and the two buttons.
	 *
	 * It draws for a reader with the hosts permission, and only where task-ui
	 * has an ssh login to a host, because a picker with nothing in it runs
	 * nothing. The reader's last batch and the prompt box about it sit in this
	 * block too, so an instance with no login draws neither.
	 *
	 * Run and summarize wants a token for Claude as well.
	 */
	function form($facts, $logins, $hosts) {
		if (!$hosts || count($logins) === 0) {
			return null;
		}

		$picks = array();

		foreach ($facts["hosts"] as $system) {
			$host = $system["host"];

			if (!isset($logins[$host])) {
				continue;
			}

			$picks[] = array("host" => $host, "login" => $logins[$host]);
		}

		return array(
			"commands" => self::COMMANDS,
			"picks" => $picks,
			"summarize" => Task\Host::claude(),
		);
	}

	/**
	 * batch is the reader's last Run on the hosts: the runs of it, one a host,
	 * and the prompt box about them.
	 *
	 * The batch is read from the found run's batch ID and not from the one row
	 * the filter answered, so the whole batch is listed whichever run of it is
	 * the newest; a run that is in no batch is its own. The runs are listed
	 * oldest first, the order they were started in.
	 *
	 * The runner is ahead of the table, so a session is laid over each row
	 * before a cell of it is read.
	 */
	function batch($chrome) {
		$runs = new Task\Run;
		$last = $runs->find(array(
			"folder" => Task\Route::INTERNAL,
			"job" => Task\Run::HOST_ALL,
			"user" => $chrome->name(),
			"limit" => 1,
		));

		if (count($last) === 0) {
			return array("batch" => null, "prompt" => null);
		}

		$found = array();

		if ($last[0]->getBatchId() !== "") {
			$of = $runs->find(array("batch" => $last[0]->getBatchId(), "limit" => Task\Run::BATCH_LIMIT));
			$found = array_reverse($of);
		}

		if (count($found) === 0) {
			$found = $last;
		}

		$rows = array();
		$live = false;

		foreach ($found as $run) {
			$session = Task\Host::session($run->getId());

			Task\Run::overlay($run, $session);

			if (!Task\Run::ended($run->getState())) {
				$live = true;
			}

			$rows[] = $this->result($run);
		}

		$first = $found[0];
		$seen = new Task\Seen($chrome);

		return array(
			"batch" => array(
				"live" => $live,
				"args" => implode("", Task\Run::argList($first)),
				"count" => count($found),
				"ago" => Task\Clock::ago($first->getStartedAt()),
				"id" => $first->getBatchId(),
				"rows" => $rows,
			),
			"prompt" => Task\Prompt::panel($chrome, $first, $seen),
		);
	}

	/**
	 * result is one run of the batch: the host it ran on, how it ended, how
	 * long it took and its output.
	 *
	 * A queued run says what it waits for, a running one draws the last screen
	 * of its terminal, and one that ended has its recording read by the script.
	 * Only a running run has a screen, so only one is asked for.
	 */
	function result($run) {
		$id = $run->getId();
		$state = $run->getState();
		$ended = Task\Run::ended($state);
		$running = !$ended && $state !== Task\Run::QUEUED;

		return array(
			"id" => $id,
			"target" => $run->getTarget(),
			"state" => Task\Run::label($run),
			"duration" => Task\Run::duration(Task\Run::elapsed($run)),
			"attach" => $ended ? "View" : "Attach",
			"queued" => $state === Task\Run::QUEUED,
			"running" => $running,
			"frame" => $running ? Task\Host::frame($id) : "",
			"cast" => $run->getCastPath(),
			"cols" => $run->getTermCols(),
			"rows" => $run->getTermRows(),
		);
	}

	/**
	 * cluster is the pools the hosts sum into, the legend of which colour each
	 * host is drawn in, and the month the stacked charts draw.
	 *
	 * An instance Beszel monitors no host for has no cluster at all: the pools
	 * would be sums of nothing and the page draws none of it. The arithmetic is
	 * the binding's, so a figure arrives here already worded.
	 */
	function cluster($facts, $now) {
		if (count($facts["hosts"]) === 0) {
			return null;
		}

		$cluster = $facts["cluster"];
		$pools = array();

		foreach ($cluster["pools"] as $pool) {
			$pools[] = self::pool($pool);
		}

		return array(
			"hosts" => $cluster["hosts"],
			"count" => count($cluster["hosts"]),
			"up" => $cluster["up"],
			"down" => $cluster["down"],
			"left" => $cluster["left"],
			"pools" => $pools,
			"charts" => $this->charts($facts, $now),
		);
	}

	/**
	 * pool is one bar: its figure, its share as a whole percentage, and the
	 * segment of each host that holds a part of it.
	 *
	 * A segment of no width draws nothing and is left out here rather than in
	 * the template, which holds no arithmetic. The share is left off the
	 * containers, which are a count and not a part of anything.
	 */
	static function pool($pool) {
		$segments = array();

		foreach ($pool["segments"] as $segment) {
			if ($segment["width"] <= 0) {
				continue;
			}

			$segments[] = array(
				"color" => $segment["color"],
				"width" => sprintf("%.3f", $segment["width"]),
				"fill" => sprintf("%.1f", $segment["fill"]),
				"title" => $segment["title"],
			);
		}

		return array(
			"name" => $pool["name"],
			"figure" => $pool["figure"],
			"share" => $pool["name"] !== "Containers",
			"pct" => sprintf("%.0f", $pool["percent"]),
			"segments" => $segments,
		);
	}

	/**
	 * charts is the cluster's month: the busy cores and the memory in use of
	 * every host that is in the pools, stacked, each in its legend colour.
	 *
	 * The hosts share one axis of times, which is every moment any host of the
	 * cluster has a record at, truncated to the span of the records; a host
	 * with no record at one of them has a gap there and not a zero, so a host
	 * Beszel lost sight of for a day leaves a hole in the stack rather than a
	 * dip. A hub that keeps no records draws no chart at all.
	 */
	function charts($facts, $now) {
		if (count($facts["history"]) === 0) {
			return array();
		}

		$records = $facts["history"];
		$step = Task\Chart::step($records);
		$systems = array();

		foreach ($facts["hosts"] as $system) {
			$systems[$system["name"]] = $system;
		}

		$times = self::times($facts["cluster"]["hosts"], $systems, $records, $step);
		$cores = array();
		$memory = array();

		foreach ($facts["cluster"]["hosts"] as $host) {
			if ($host["left"]) {
				continue;
			}

			$cpus = isset($systems[$host["name"]]) ? $systems[$host["name"]]["cpus"] : 0;
			$busy = array();
			$used = array();

			foreach (self::records($host["name"], $systems, $records) as $sample) {
				$at = Task\Chart::truncate(Task\Clock::seconds($sample["at"]), $step);
				$busy[$at] = $sample["cpu"] / 100 * $cpus;
				$used[$at] = $sample["memory_used"];
			}

			$cores[] = Task\Chart::series($host["name"], $host["color"], self::points($times, $busy));
			$memory[] = Task\Chart::series($host["name"], $host["color"], self::points($times, $used));
		}

		$from = Task\Chart::from($now);
		$busiest = new Task\Chart("Busy cores", Task\Chart::cores(), $from, $now, true, $cores);
		$fullest = new Task\Chart("Memory used", Task\Chart::gib(), $from, $now, true, $memory);

		return array($busiest->data(), $fullest->data());
	}

	/**
	 * times is every moment a host of the cluster has a record at, truncated
	 * to the span of the records, each once and oldest first.
	 *
	 * Every host of the cluster is read and not only the ones in the pools,
	 * because the axis is the cluster's: a host that left still says when the
	 * hub was recording.
	 */
	static function times($hosts, $systems, $records, $step) {
		$seen = array();
		$times = array();

		foreach ($hosts as $host) {
			foreach (self::records($host["name"], $systems, $records) as $sample) {
				$at = Task\Chart::truncate(Task\Clock::seconds($sample["at"]), $step);

				if (isset($seen[$at])) {
					continue;
				}

				$seen[$at] = true;
				$times[] = $at;
			}
		}

		sort($times);

		return $times;
	}

	/** records are the hub's records of a host of the cluster, which it holds by host ID. */
	static function records($name, $systems, $records) {
		$id = isset($systems[$name]) ? $systems[$name]["id"] : "";

		return isset($records[$id]) ? $records[$id] : array();
	}

	/** points is a series over the shared times, a gap at a time the host has no record at. */
	static function points($times, $values) {
		$points = array();

		foreach ($times as $at) {
			$v = isset($values[$at]) ? $values[$at] : null;
			$points[] = Task\Chart::point($at, $v);
		}

		return $points;
	}

	/**
	 * row is one host of the table: its line, the meters of how busy it is, the
	 * sizes behind them and the menu of what the reader may do to it.
	 *
	 * The meters are the agent's last word, which every host has; the sizes are
	 * the hub's last minute of measures, which a host without one has none of,
	 * and then both size cells are empty.
	 *
	 * open is the hosts permission: it opens the expanded row, and makes the
	 * line a button. href is the host's page, which the menu links and the
	 * script fetches the expanded row from, so the row carries the mount of
	 * the front end that drew it.
	 *
	 * entry is the inventory row this host is held by, and drops Allow
	 * Deletes on Inventory: together they draw the Remove button, which posts
	 * to the row's own path. A host the hub still reports comes back on the
	 * next read, which the question on the form says.
	 */
	function row($system, $logins, $open, $entry = null, $drops = false) {
		$name = $system["name"];
		$host = $system["host"];
		$login = isset($logins[$host]) ? $logins[$host] : "";
		$href = Task\Route::href("/inventory/hosts/" . $name);
		$menu = new Task\Menu("Actions on " . $name);

		$status = $system["status"] === "" ? "no measures" : $system["status"];

		$menu->detail("Address", $host, true);
		$menu->detail("Status", $status);

		if ($entry !== null) {
			$menu->detail("Source", $entry->source === Task\Inventory\Hosts::BESZEL ? "Beszel" : "entered here");
		}

		$menu->when($open, "server", "Open the host's page", array("href" => $href));
		$menu->when($login !== "", "copy", "Copy ssh " . $login, array("data-copy" => "ssh " . $login));

		return array(
			"name" => $name,
			"host" => $host,
			"href" => $href,
			"drop" => $drops && $entry !== null ? Task\Route::href("/inventory/hosts/" . Task\Route::pathEscape($entry->id) . "/delete") : "",
			"asked" => "Remove " . $name . " from the inventory? A host Beszel reports comes back on the next read; one nobody monitors is gone until it is entered again.",
			"status" => $status,
			"up" => $system["up"],
			"open" => $open,
			"login" => $login,
			"copy" => Task\Copy::of("ssh " . $login, "ssh " . $login, $login),
			"cpus" => $system["cpus"],
			"cpu" => Task\Meter::of($system["cpu"]),
			"memory" => Task\Meter::of($system["memory"]),
			"disk" => Task\Meter::of($system["disk"]),
			"measured" => $system["measured"],
			"memoryUsed" => sprintf("%.1f", $system["memory_used"]),
			"memoryTotal" => sprintf("%.1f", $system["memory_total"]),
			"diskUsed" => sprintf("%.0f", $system["disk_used"]),
			"diskTotal" => sprintf("%.0f", $system["disk_total"]),
			"menu" => $menu->data(),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_inventory;

	$mod->serve();
}
