<?php

// @route GET /settings/projects

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_projects is the directories under the src of every workspace: a block
// per workspace and a row per project, with what its pipeline reads, the
// checkout it is and what it takes on disk.
//
// The page reads and does not write. Add project opens the clone form, Pull
// now in a row's menu starts the git skill's pull, and Remove checkout deletes
// through the API and draws the page again.
class mod_projects implements Module {
	/** LEAD is the sentence under the heading, as text: what a block holds. */
	const LEAD = "The projects of each workspace, a block per workspace. task-ui pulls a git checkout before every run.";

	/** RELOAD is where Remove checkout goes: the page the row was in, read again. */
	const RELOAD = "reload";

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
		return array("workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("project/view")) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::PROJECTS, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::PROJECTS, "href" => ""),
		));

		$workspaces = new Task\Workspace($chrome->db);
		$named = $workspaces->names();
		$entries = Task\Host::entries($named);
		$spaces = Task\Host::spaces(self::dirs($entries));
		$instance = Task\Host::instance(false);
		$groups = Task\Project::groups($entries, $named);
		$steps = Task\Setup::steps($chrome, Task\Setup::PROJECTS);

		$view->page(array(
			"head" => $this->head($chrome),
			"groups" => $this->blocks($chrome, $groups, $spaces, Task\Row::text($instance, "root")),
			"trail" => Task\Setup::trail($steps),
		), "projects.tpl");
	}

	/** head is the page head: the heading, the lead and Add project. */
	function head($chrome) {
		$buttons = new Task\Buttons;
		$head = new Task\Head(Task\Setup::PROJECTS, self::LEAD);
		$href = Task\Route::href(Task\Project::PAGE . "/new");

		if ($chrome->allowed("project/edit")) {
			$buttons->primary("plus", "Add project", array("href" => $href));
		}

		$head->buttons($buttons->data());

		return $head->data();
	}

	/**
	 * dirs is the directory of every entry, which is what the page measures.
	 *
	 * The measure is the host's: a size is a walk of the tree and a reading of
	 * what git ignores there, and neither is PHP's to do.
	 */
	static function dirs($entries) {
		$out = array();

		foreach ($entries as $entry) {
			$out[] = Task\Row::text($entry, "dir");
		}

		return $out;
	}

	/**
	 * blocks is a block per workspace, in the order the settings list them:
	 * its head and the table of its projects.
	 *
	 * A workspace with nothing in it keeps its block, whose head says so and
	 * which draws no table at all.
	 */
	function blocks($chrome, $groups, $spaces, $root) {
		$out = array();

		foreach ($groups as $group) {
			$out[] = array(
				"block" => Task\Project::block($chrome, $group, $root),
				"table" => Task\Project::table($chrome, $group["entries"], $spaces, self::RELOAD),
			);
		}

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_projects;

	$mod->serve();
}
