<?php

// @route GET /settings/telemetry

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_telemetry is the one form that configures what the instance records of
// its own work: the switch, the sample rate, how much it retains, and who may
// read the traces.
//
// The page reads and does not write. Anyone who may open a step of the
// settings reads it; an admin of an instance with a tracer wired saves it,
// and an instance with none shows these same fields, which are what it would
// record, with every one of them turned off.
class mod_telemetry implements Module {
	/** LEAD is the sentence under the heading, as HTML: it links to another step. */
	const LEAD = "What the instance records of its own work. With it on, every sampled request becomes a trace: the handler's spans, the database calls under it and, on a <a href=\"%s\">module</a> page, the spans of the PHP that ran inside the request, in one trace.";

	/** CHANGES is what the sentence says further to a reader who may not save. */
	const CHANGES = " An admin changes the settings.";

	/** DASHBOARD is the button that opens the traces, and what it says on hover. */
	const DASHBOARD = "Open the dashboard";

	const TRACES = "The traces, for an admin";

	/** SAVED is the notice the page lands with after the options are stored. */
	const SAVED = "Telemetry saved";

	/** PAGE is this page, which the form returns to. */
	const PAGE = "/settings/telemetry";

	/** MODULES is the step that lists what a module page is. */
	const MODULES = "/settings/sidebar";

	/** FAILED starts what the page answers when the stored options did not read. */
	const FAILED = "reading the telemetry settings failed: ";

	function name() {
		return "settings";
	}

	function title() {
		return "Settings";
	}

	function icon() {
		return "settings";
	}

	/** The options are the tracer's own and are in no table of this database. */
	function tables() {
		return array();
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!Task\Sidebar::settings($chrome->access)) {
			$view->refused();

			return;
		}

		$opts = Task\Host::telemetry();

		if (Task\Row::text($opts, "error") !== "") {
			$this->failed(Task\Row::text($opts, "error"));

			return;
		}

		$chrome->heading(Task\Setup::TELEMETRY, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::TELEMETRY, "href" => ""),
		));

		// The dashboard is mounted on the instance's own router and not inside
		// this application, so its path carries no prefix.
		$path = Task\Row::text($opts, "path");
		$admin = Task\User::admin($chrome->user) && Task\Row::flag($opts, "wired");
		$head = new Task\Head(Task\Setup::TELEMETRY, self::lead($admin));
		$buttons = new Task\Buttons;

		$buttons->item("", self::DASHBOARD, array("href" => $path, "title" => self::TRACES));
		$head->buttons($buttons->data());

		$steps = Task\Setup::steps($chrome, Task\Setup::TELEMETRY);

		$view->page(array(
			"head" => $head->data(),
			"admin" => $admin,
			"next" => Task\Route::href(self::PAGE) . "?notice=" . Task\Route::queryEscape(self::SAVED),
			"path" => $path,
			"rate" => self::rate(Task\Row::text($opts, "sample_rate")),
			"ignore" => self::lines(Task\Row::items($opts, "ignore_paths")),
			"networks" => self::lines(Task\Row::items($opts, "allowed_networks")),
			"opts" => $opts,
			"trail" => Task\Setup::trail($steps),
		), "telemetry.tpl");
	}

	/**
	 * failed answers for an instance whose stored options did not read, which
	 * is no page: the form would hold the defaults, and a save of it would
	 * write them over whatever is there.
	 */
	function failed($why) {
		http_response_code(500);
		echo self::FAILED . $why . "\n";
	}

	/**
	 * lead is the sentence under the heading. It is HTML and built here, not in
	 * the template: a template writes the newline that sits before a tag, so a
	 * sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($admin) {
		$said = sprintf(self::LEAD, Task\Route::href(self::MODULES));

		return $admin ? $said : $said . self::CHANGES;
	}

	/** lines is a list field as the textarea that edits it spells it: one per line. */
	static function lines($items) {
		$out = array();

		foreach ($items as $item) {
			$out[] = (string)$item;
		}

		return implode("\n", $out);
	}

	/**
	 * rate words the sample rate, which is the one option that is not a whole
	 * number: a percent of the requests, and a fraction of one where an admin
	 * typed one.
	 */
	static function rate($value) {
		$percent = (float)$value;

		if ($percent === (float)(int)$percent) {
			return (string)(int)$percent;
		}

		return (string)$percent;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_telemetry;

	$mod->serve();
}
