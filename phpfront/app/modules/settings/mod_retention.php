<?php

// @route GET /settings/retention

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_retention is what the instance keeps of what it records: a row per area
// with what it holds now and the rule that keeps it, and the last run of the
// job that applies them.
//
// The page reads and does not write. The form puts the rules to /api/retention
// and lands back here; a reader who may see a step of the settings reads the
// table with the selects turned off, and an admin saves it.
class mod_retention implements Module {
	/** LEAD is the sentence under the heading, as HTML: it names a job in code. */
	const LEAD = "How long task-ui keeps what it records, one row per area. <code>@cleanup</code> applies the rows in this order every midnight, oldest first: run history takes a run's recording, summary and threads with it.";

	/** CHANGES is what the sentence says further to a reader who may not save. */
	const CHANGES = " An admin changes the rules.";

	/** CLEANUP is the job that applies the rules, of the internal folder. */
	const CLEANUP = "@cleanup";

	/** WHAT titles the run the button starts, as the run list names it. */
	const WHAT = "internal / cleanup";

	/** STARTS is the hover of the button that starts it now. */
	const STARTS = "Run @cleanup now, with the saved rules";

	/** SAVED is the notice the page lands with after the rules are stored. */
	const SAVED = "Data retention saved";

	/** PAGE is this page, which the form returns to. */
	const PAGE = "/settings/retention";

	/** FAILED starts what the page answers when the walk of the disk failed. */
	const FAILED = "measuring the data failed: ";

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
		return array("setting", "run");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!Task\Sidebar::settings($chrome->access)) {
			$view->refused();

			return;
		}

		$usage = Task\Host::usage();

		if ($usage["error"] !== "") {
			$this->failed($usage["error"]);

			return;
		}

		$chrome->heading(Task\Setup::RETENTION, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::RETENTION, "href" => ""),
		));

		$admin = Task\User::admin($chrome->user);
		$head = new Task\Head(Task\Setup::RETENTION, self::lead($admin));

		if ($admin) {
			$buttons = new Task\Buttons;

			$buttons->item("", "Clean up now", self::cleans());
			$head->buttons($buttons->data());
		}

		$steps = Task\Setup::steps($chrome, Task\Setup::RETENTION);

		$view->page(array(
			"head" => $head->data(),
			"admin" => $admin,
			"next" => Task\Route::href(self::PAGE) . "?notice=" . Task\Route::queryEscape(self::SAVED),
			"runs" => Task\Route::href("/runs") . "?folder=" . Task\Route::INTERNAL . "&job=" . self::CLEANUP,
			"last" => $this->last(),
			"rows" => $this->rows($usage),
			"trail" => Task\Setup::trail($steps),
		), "retention.tpl");
	}

	/**
	 * failed answers for an instance whose walk of the disk did not finish,
	 * which is no page: the table would say every area holds nothing, and that
	 * reads as an instance that is empty rather than one that could not look.
	 */
	function failed($why) {
		http_response_code(500);
		echo self::FAILED . $why . "\n";
	}

	/**
	 * rows are the table's rows, and none for an instance with no cleanup
	 * wired: there is nothing to apply, so the table holds nothing to set.
	 */
	function rows($usage) {
		if ($usage["off"]) {
			return array();
		}

		$store = new Task\Retention;

		return Task\Retention::rows($usage["areas"], $store->rules());
	}

	/**
	 * last is the last run of @cleanup as the line under the heading reads it:
	 * when it ran, how it ended, and nothing at all where it never has.
	 */
	function last() {
		$runs = new Task\Run;
		$found = $runs->find(array(
			"folder" => Task\Route::INTERNAL,
			"job" => self::CLEANUP,
			"limit" => 1,
		));

		if (count($found) === 0) {
			return array(
				"ran" => false,
				"href" => "",
				"stamp" => "",
				"ago" => "",
				"state" => Task\Run::label(null),
			);
		}

		$run = $found[0];
		$started = $run->getStartedAt();

		return array(
			"ran" => true,
			"href" => Task\Route::href("/runs/" . $run->getId()),
			"stamp" => Task\Clock::read($started),
			"ago" => Task\Clock::ago($started),
			"state" => Task\Run::label($run),
		);
	}

	/**
	 * lead is the sentence under the heading. It is HTML and built here, not in
	 * the template: a template writes the newline that sits before a tag, so a
	 * sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($admin) {
		return $admin ? self::LEAD : self::LEAD . self::CHANGES;
	}

	/**
	 * cleans is what Clean up now carries: the job the script starts and the
	 * title the run it starts is listed under.
	 *
	 * data-run carries no value, as every button the script starts a run from
	 * does: what is read of it is whether it is there.
	 */
	static function cleans() {
		return array(
			"data-run" => null,
			"data-folder" => Task\Route::INTERNAL,
			"data-job" => self::CLEANUP,
			"data-title" => self::WHAT,
			"title" => self::STARTS,
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_retention;

	$mod->serve();
}
