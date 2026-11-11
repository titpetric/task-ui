<?php

// @route GET /settings/secrets/{id}/reads

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_secret_reads is the log of one variable: the runs that got it in their
// environment, newest first, a page at a time.
//
// No value is on the page. The log records that a run read a variable and not
// what it read, so the page draws the run, the job and who started it, and a
// secret's value stays where it was encrypted.
class mod_secret_reads implements Module {
	/** HEADING is what the block above the table is called. */
	const HEADING = "Reads";

	/** ORDER is the faint note beside the heading, which the order is. */
	const ORDER = "newest first";

	/** BACK is the button on the right of the head, back to the list. */
	const BACK = "Secrets";

	/** SHORT is how much of a run's ID the cell shows, the rest in its title. */
	const SHORT = 8;

	function name() {
		return "settings";
	}

	function title() {
		return "Settings";
	}

	function icon() {
		return "settings";
	}

	function tables() {
		return array("secret", "secret_read");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("secret/view")) {
			$view->refused();

			return;
		}

		$id = Task\Route::param("id");
		$secrets = new Task\Secret($chrome->db);
		$secret = $secrets->byId($id);

		if ($secret === null) {
			$view->missing(Task\Secret::MISSING);

			return;
		}

		$name = $secret->getName();
		$title = "Reads of " . $name;

		$chrome->heading($title, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::SECRETS, "href" => Task\Secret::PAGE),
			array("text" => $name, "href" => ""),
		));

		// One row more than the page holds: the row that tells whether the log
		// goes on, which paging drops again.
		$before = Task\Route::input("before");
		$after = Task\Route::input("after");
		$found = $secrets->reads($id, $before, $after, Task\Secret::READS + 1);
		$page = Task\Pager::paged($found, Task\Secret::READS, $before, $after);
		$who = $chrome->who();
		$rows = array();

		foreach ($page["rows"] as $read) {
			$rows[] = $this->row($read, $who);
		}

		$head = new Task\Head($title);
		$buttons = new Task\Buttons;

		$head->badge($secret->getScope());
		$head->said(self::said($name));
		$buttons->item("", self::BACK, array("href" => Task\Route::href(Task\Secret::PAGE)));
		$head->buttons($buttons->data());

		$steps = Task\Setup::steps($chrome, Task\Setup::SECRETS);
		$path = Task\Secret::PAGE . "/" . $id . "/reads";

		$view->page(array(
			"head" => $head->data(),
			"heading" => self::HEADING,
			"order" => self::ORDER,
			"name" => $name,
			"rows" => $rows,
			"pager" => Task\Pager::of($path, $page["prev"], $page["next"]),
			"trail" => Task\Setup::trail($steps),
		), "secret_reads.tpl");
	}

	/**
	 * said is the sentence under the heading: what the log holds, and how much
	 * of it the instance keeps.
	 *
	 * The thousand is storage.SecretReadKeep, which the Go template writes out
	 * as well: the cap belongs to the writer of the log and no page sets it.
	 */
	static function said($name) {
		return "The runs that got " . $name . " in their environment, newest first. task-ui keeps the last 1000 reads of each secret. The value never shows.";
	}

	/**
	 * row is one read of the table: when a run got the variable, the run, the
	 * job it ran and who started it.
	 *
	 * The run is linked by its whole ID and shown by the first eight
	 * characters, the whole one in the link's title, because a column of
	 * twenty-six character IDs says nothing a reader reads.
	 */
	function row($read, $who) {
		$run = $read->getRunId();
		$at = $read->getCreatedAt();

		return array(
			"ago" => Task\Clock::ago($at),
			"stamp" => Task\Clock::read($at),
			"run" => $run,
			"runHref" => Task\Route::href("/runs/" . $run),
			"short" => substr($run, 0, self::SHORT),
			"key" => self::key($read),
			"user" => $who->of($read->getUserName()),
		);
	}

	/**
	 * key is the job the run ran, spelled the way every other page spells one:
	 * the folder, the directory inside it where the job came from one, and the
	 * job's own name.
	 */
	static function key($read) {
		$dir = $read->getDir();
		$inside = $dir === "" ? "" : $dir . "/";

		return $read->getFolder() . "/" . $inside . $read->getJob();
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_secret_reads;

	$mod->serve();
}
