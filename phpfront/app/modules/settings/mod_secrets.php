<?php

// @route GET /settings/secrets

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_secrets lists every variable the runs of this instance are given: the
// secrets, whose values are encrypted and never drawn, and the env entries,
// whose values are.
//
// A value that is drawn is the vault's and not the row's. The row holds
// armour, the instance's encryption key opens it, and the key does not enter
// PHP: Task\Host::variables() answers the value of each env entry by row ID.
//
// The walk stops at the first variable it will not open, so that row and the
// ones after it answer the armour they hold rather than nothing, which is what
// the Go page prints of them, and what stopped the walk is the line above the
// table.
class mod_secrets implements Module {
	/**
	 * TITLE is the heading over the lead.
	 *
	 * The tab and the last crumb say Task\Setup::SECRETS instead, which is the
	 * step's own name: the heading names what the page holds and the crumb
	 * names where it sits.
	 */
	const TITLE = "Secrets and env";

	/** LEAD is the sentence under the heading. */
	const LEAD = "Variables for the environment of runs, secret or plain env, applying to every workspace or to one. A secret's value is never shown again; an env value is. Runs record what they get, and <code>task-ui get-secret NAME</code> reads either.";

	/** ADD is the button of the page's own action. */
	const ADD = "Add variable";

	/** NO_KEY is why Add is off: nothing can be encrypted without the key. */
	const NO_KEY = "The instance has no encryption key yet";

	/** HEADING is what the block above the table is called. */
	const HEADING = "Variables";

	/** EMPTY_LINE is what the block says where the instance has no variable. */
	const EMPTY_LINE = "No variables yet.";

	/** MASK is what a secret's value column draws in place of the value. */
	const MASK = "••••••";

	/** ENV_HEADING is the block of what the integrations add to a run. */
	const ENV_HEADING = "Integration environment";

	/** ENV_NOTE is the faint note beside it, which says whose runs it is. */
	const ENV_NOTE = "read only, computed from the integrations for your runs";

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
		return array("secret", "secret_read", "workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("secret/view")) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::SECRETS, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::SECRETS, "href" => ""),
		));

		$secrets = new Task\Secret($chrome->db);
		$workspaces = new Task\Workspace($chrome->db);
		$rows = $workspaces->all();
		$titles = Task\Secret::titles($rows);
		$found = $secrets->all();

		// ?workspace= draws what reaches one workspace: the variables of every
		// workspace, its own and its folders'. The filter is the name itself
		// and not a row, as the Go page takes it: a name that names no
		// workspace keeps the variables of every workspace and nothing else.
		$only = Task\Route::input("workspace");

		if ($only !== "") {
			$folders = $workspaces->folders();
			$found = Task\Secret::reaching($found, $only, $folders);
		}

		$vault = Task\Host::variables();
		$values = isset($vault["values"]) ? $vault["values"] : array();
		$counts = $secrets->counts();
		$last = $secrets->last();
		$key = Task\Key::folder();
		$steps = Task\Setup::steps($chrome, Task\Setup::SECRETS);
		$head = new Task\Head(self::TITLE, self::LEAD);
		$buttons = new Task\Buttons;

		$buttons->primary("plus", self::ADD, self::action($key["instance"], $only));
		$head->buttons($buttons->data());

		$drawn = array();

		foreach ($found as $secret) {
			$drawn[] = $this->row($secret, $chrome, $titles, $values, $counts, $last);
		}

		$view->page(array(
			"head" => $head->data(),
			"workspace" => $only,
			"every" => Task\Route::href(Task\Secret::PAGE),
			"block" => self::block(count($found)),
			"error" => isset($vault["error"]) ? $vault["error"] : "",
			"rows" => $drawn,
			"env" => self::env(),
			"envBlock" => array(
				"title" => self::ENV_HEADING,
				"note" => self::ENV_NOTE,
				"empty" => "",
				"buttons" => array(),
			),
			"trail" => Task\Setup::trail($steps),
		), "secrets.tpl");
	}

	/**
	 * action is the attributes of the Add button: the form of a new variable,
	 * opened on the workspace the list is filtered to.
	 *
	 * Without an encryption key the button is off rather than absent, which is
	 * what says the instance cannot encrypt a value yet. The workspace is
	 * escaped as the markup of a link escapes a value written into its query,
	 * because the Go template writes this one itself.
	 */
	static function action($instance, $workspace) {
		if ($instance === "") {
			return array("disabled" => self::NO_KEY);
		}

		$href = Task\Secret::PAGE . "/new";

		if ($workspace !== "") {
			$href .= "?scope=" . Task\Workspace::SCOPE . Task\Route::urlEscape($workspace);
		}

		return array("href" => Task\Route::href($href));
	}

	/** block is the head of the Variables section: how many there are, and the line for none. */
	static function block($found) {
		return array(
			"title" => self::HEADING,
			"note" => Task\Setup::says($found, "variable", "variables"),
			"empty" => $found === 0 ? self::EMPTY_LINE : "",
			"buttons" => array(),
		);
	}

	/**
	 * row is one variable of the table: its name and kind, its value, the
	 * scope its runs are in, how often it was read and its menu.
	 *
	 * A scope the titles do not name is drawn as it is written, which is a
	 * folder's own scope and a workspace that is gone alike.
	 */
	function row($secret, $chrome, $titles, $values, $counts, $last) {
		$id = $secret->getId();
		$name = $secret->getName();
		$scope = $secret->getScope();
		$title = isset($titles[$scope]) ? $titles[$scope] : "";
		$applies = $title === "" ? $scope : $title;
		$reads = isset($counts[$id]) ? $counts[$id] : 0;
		$secret_value = $secret->getIsSecret() !== 0;
		$value = isset($values[$id]) ? $values[$id] : "";
		$menu = $this->menu($secret, $chrome, $applies);

		return array(
			"name" => $name,
			"secret" => $secret_value,
			"value" => $secret_value ? "" : $value,
			"mask" => self::MASK,
			"applies" => $applies,
			"reads" => $reads,
			"readsHref" => Task\Route::href(Task\Secret::PAGE . "/" . $id . "/reads"),
			"read" => self::read($id, $last),
			"menu" => $menu->data(),
		);
	}

	/**
	 * read is the last run that got a variable: when it was, the run it was,
	 * and the job of it, "" where no run ever got the variable.
	 */
	static function read($id, $last) {
		if (!isset($last[$id])) {
			return array(
				"run" => "",
				"href" => "",
				"ago" => "",
				"title" => "",
			);
		}

		$row = $last[$id];
		$run = $row->getRunId();
		$at = $row->getCreatedAt();
		$dir = $row->getDir();
		$inside = $dir === "" ? "" : $dir . "/";
		$job = $row->getFolder() . "/" . $inside . $row->getJob();

		return array(
			"run" => $run,
			"href" => Task\Route::href("/runs/" . $run),
			"ago" => Task\Clock::ago($at),
			"title" => "Last read " . Task\Clock::read($at) . ": " . $job,
		);
	}

	/**
	 * menu is the actions on one variable: when it changed, its form, its log
	 * of reads, and removing it.
	 *
	 * The form is opened by ID and not by scope and name, because a name is
	 * not unique across the scopes.
	 */
	function menu($secret, $chrome, $applies) {
		$id = $secret->getId();
		$name = $secret->getName();
		$menu = new Task\Menu("Actions on " . $name);
		$edit = Task\Secret::PAGE . "/edit?id=" . Task\Route::queryEscape($id);

		$menu->detail("Changed", Task\Clock::read($secret->getUpdatedAt()));
		$menu->when($chrome->allowed("secret/edit"), "edit", "Edit", array("href" => Task\Route::href($edit)));
		$menu->item("list", "Reads", array("href" => Task\Route::href(Task\Secret::PAGE . "/" . $id . "/reads")));
		$menu->whenDanger($chrome->allowed("secret/delete"), "trash", "Delete", array(
			"data-delete" => "/api/secrets/" . $id,
			"data-confirm" => "Delete " . $name . " from " . $applies . "? Runs stop getting it.",
			"data-confirm-ok" => "Delete",
			"data-next" => "reload",
		));

		return $menu;
	}

	/**
	 * env is what the integrations add to a run of the reader, which is the
	 * second table: the name of each variable and its value, a credential's
	 * withheld.
	 *
	 * It is the reader's own, so two readers draw two tables: a Who setting
	 * names the readers a service reaches.
	 */
	static function env() {
		$out = array();

		foreach (Task\Host::integrationEnv() as $var) {
			$out[] = array(
				"name" => Task\Row::text($var, "name"),
				"value" => Task\Row::text($var, "value"),
				"redacted" => Task\Row::flag($var, "redacted"),
			);
		}

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_secrets;

	$mod->serve();
}
