<?php

// @route GET /settings/instance

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_instance is what the server runs with, section by section: where it
// answers, how many runs it does at once, how much of its own environment a
// run starts with, where it keeps its data, what that data takes on disk, the
// tools it found in its PATH, its two keys and the services it reads from.
//
// The page reads. Almost every row is an environment variable, which a reader
// changes in the environment and restarts task-ui for; the backlog length and
// the parent environment are settings of their own and an admin saves those
// two here.
//
// What an admin alone sees is the server's own variable names and the measure
// of the disk, which is why the host is asked for the instance with the
// reader's flag: a reader who is no admin is answered neither.
class mod_instance implements Module {
	/**
	 * LEAD opens the sentence under the heading. It ends in who saves the two
	 * settings that are saved here, so the period is added after that.
	 */
	const LEAD = "What the server runs with. The environment sets most of these: change one there and restart task-ui. The backlog length and the parent environment are settings of their own, saved here";

	/** BY_ADMIN ends the sentence for a reader who may not save them. */
	const BY_ADMIN = " by an admin";

	/** PAGE is this page, which every form on it returns to. */
	const PAGE = "/settings/instance";

	/** The steps and pages the rows link out to. */
	const USERS = "/settings/users";

	const RETENTION = "/settings/retention";

	const INTEGRATIONS = "/settings/integrations";

	const KEYS = "/settings/keys";

	const UPLOAD = "/settings/keys/upload";

	/** The notices the four forms land with, one per section. */
	const BACKLOG_SAVED = "Backlog length saved";

	const FORWARD_SAVED = "Parent environment saved";

	const TOOLING_CHECKED = "Tooling checked";

	const TOOL_CHECKED = "Tool checked";

	/**
	 * The modes of the parent environment: the base variables alone, every one
	 * but the sensitive, and every one. Keep them in step with
	 * runner/environment.go.
	 */
	const NONE = "none";

	const AUTO = "auto";

	const ALL = "all";

	function name() {
		return "settings";
	}

	function title() {
		return "Settings";
	}

	function icon() {
		return "settings";
	}

	/**
	 * The tool table is the only one the page reads of its own. What the
	 * server runs with, the measure of the disk and the key folder are the
	 * host's, and none of the three is a table of this database.
	 */
	function tables() {
		return array("tool");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!Task\Sidebar::settings($chrome->access)) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::INSTANCE, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::INSTANCE, "href" => ""),
		));

		$admin = Task\User::admin($chrome->user);
		$instance = Task\Host::instance($admin);
		$head = new Task\Head(Task\Setup::INSTANCE, self::lead($admin));
		$steps = Task\Setup::steps($chrome, Task\Setup::INSTANCE);
		$tools = new Task\Toolset;

		$view->page(array(
			"head" => $head->data(),
			"admin" => $admin,
			"instance" => $instance,
			"site" => Task\Route::site(),
			"signin" => $this->signin($chrome, $instance),
			"users" => Task\Route::href(self::USERS),
			"retention" => Task\Route::href(self::RETENTION),
			"integrations" => Task\Route::href(self::INTEGRATIONS),
			"backlog" => self::lands(self::BACKLOG_SAVED, "runs"),
			"forward" => self::lands(self::FORWARD_SAVED, "environment"),
			"checked" => self::lands(self::TOOLING_CHECKED, "tooling"),
			"probed" => self::lands(self::TOOL_CHECKED, "tooling"),
			"modes" => self::modes(Task\Row::text($instance, "parent_env")),
			"base" => self::names($instance["env_base"]),
			"words" => self::names($instance["env_sensitive"]),
			"env" => self::env($instance),
			"storage" => self::storage($instance),
			"tooling" => $tools->groups(),
			"keys" => $this->keys($chrome),
			"trail" => Task\Setup::trail($steps),
		), "instance.tpl");
	}

	/**
	 * signin is the user auto login signs everyone in as, as the user
	 * component draws them, and nobody where the instance has no auto login.
	 */
	function signin($chrome, $instance) {
		$name = Task\Row::text($instance, "auto_login");

		if ($name === "") {
			return array();
		}

		$people = $chrome->who();

		return $people->of($name);
	}

	/**
	 * keys are the instance's two keys as the Keys section draws them: the
	 * encryption key of the secrets and the key git pulls with, each a row of
	 * a keys table, and nothing where the folder holds neither.
	 *
	 * A row's changes return to this page, which is a path of this instance
	 * with no mount in it: the redirect is the API's, and the API answers at
	 * the root.
	 *
	 * The folder is only read where the reader may see a key at all, because
	 * the section is theirs alone.
	 */
	function keys($chrome) {
		$shown = $chrome->allowed("key/view");
		$encrypt = array();
		$git = array();

		if ($shown) {
			$folder = Task\Key::folder();
			$named = Task\Row::text($folder, "instance");

			foreach ($folder["keys"] as $key) {
				if ($key["name"] === $named) {
					$encrypt = Task\Key::row($chrome, $key, self::PAGE);
				}

				if ($key["git"]) {
					$git = Task\Key::row($chrome, $key, self::PAGE);
				}
			}
		}

		return array(
			"shown" => $shown,
			"edits" => $chrome->allowed("key/edit"),
			"encrypt" => $encrypt,
			"git" => $git,
			"every" => Task\Route::href(self::KEYS),
			"upload" => Task\Route::href(self::UPLOAD),
			"next" => Task\Route::href(self::PAGE) . "#keys",
		);
	}

	/**
	 * lead is the sentence under the heading. It is built here and not in the
	 * template: a template writes the newline that sits before a tag, so a
	 * sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($admin) {
		$said = $admin ? self::LEAD : self::LEAD . self::BY_ADMIN;

		return $said . ".";
	}

	/**
	 * lands is where a form of this page returns to: the page itself, the
	 * notice it lands with and the section the change was made in, so the
	 * reader comes back to the row they saved.
	 */
	static function lands($notice, $section) {
		$page = Task\Route::href(self::PAGE);

		return $page . "?notice=" . Task\Route::queryEscape($notice) . "#" . $section;
	}

	/**
	 * modes are the three choices of the radio group, in the order the page
	 * lists them, with what each means under it and the one in effect checked.
	 * Keep the wording in step with runner/environment.go.
	 */
	static function modes($mode) {
		$forwards = array(
			array(
				"value" => self::NONE,
				"label" => "None",
				"note" => "Secure: the base variables alone, less the sensitive ones.",
			),
			array(
				"value" => self::AUTO,
				"label" => "Auto",
				"note" => "Filtered: every variable but the sensitive ones.",
			),
			array(
				"value" => self::ALL,
				"label" => "All",
				"note" => "Every variable, sensitive ones too.",
			),
		);
		$out = array();

		foreach ($forwards as $forward) {
			$forward["checked"] = $forward["value"] === $mode;
			$out[] = $forward;
		}

		return $out;
	}

	/**
	 * names is a list of variable names as the paragraph under the form reads
	 * them: the comma that goes before every one but the first, so the
	 * template writes the `<code>` around a name and nothing else.
	 */
	static function names($items) {
		$out = array();

		foreach ($items as $item) {
			$sep = count($out) === 0 ? "" : ", ";
			$out[] = array("sep" => $sep, "name" => $item);
		}

		return $out;
	}

	/**
	 * env is the server's own variables as the table lists them: whether a run
	 * gets each under the mode in effect, and the two flags the script reads
	 * to redraw the column when a reader picks another mode without saving.
	 *
	 * A reader who is no admin is answered no variable at all, and the table
	 * is not drawn for them.
	 */
	static function env($instance) {
		$mode = Task\Row::text($instance, "parent_env");
		$out = array();

		foreach ($instance["env"] as $name) {
			$out[] = array(
				"name" => $name["name"],
				"sensitive" => $name["sensitive"],
				"none" => self::word($name["none"]),
				"auto" => self::word($name["auto"]),
				"forwarded" => self::forwarded($name, $mode),
			);
		}

		return $out;
	}

	/**
	 * word is a flag as the row's attribute spells it, which the script
	 * compares against: a PHP boolean writes the empty string for false, and
	 * the attribute has to say the word.
	 */
	static function word($flag) {
		return $flag ? "true" : "false";
	}

	/**
	 * forwarded reports whether a run gets a variable under a mode: every one
	 * under all, and the flag the host answered for the other two. Keep it in
	 * step with runner/env_name.go.
	 */
	static function forwarded($name, $mode) {
		if ($mode === self::ALL) {
			return true;
		}

		if ($mode === self::AUTO) {
			return $name["auto"];
		}

		return $name["none"];
	}

	/**
	 * storage is what the instance keeps on disk as the table reads it: a row
	 * per item and the total under them, the sizes worded.
	 *
	 * A row still being measured is drawn as an ellipsis and carries no size,
	 * because nothing was counted yet and a zero would read as an item that
	 * holds nothing.
	 */
	static function storage($instance) {
		$table = $instance["storage"];
		$rows = array();

		foreach ($table["rows"] as $row) {
			$rows[] = self::measure($row);
		}

		return array("rows" => $rows, "total" => self::measure($table["total"]));
	}

	/** measure is one item of the storage table: where it is and what it takes. */
	static function measure($row) {
		return array(
			"name" => $row["name"],
			"where" => $row["where"],
			"size" => Task\Size::bytes($row["size"]),
			"files" => $row["files"],
			"pending" => $row["pending"],
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_instance;

	$mod->serve();
}
