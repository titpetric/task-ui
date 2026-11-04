<?php

// @route GET /settings/projects/{rest...}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_project_edit is the settings of one project: what the sidebar calls it,
// who may run it and each of its jobs, which conventions it reads, how deep
// its folders are listed and which workspace it sits in.
//
// The page reads and does not write: the form saves through the API. A folder
// the catalogue does not hold still has a page, because what the instance
// keeps about a project outlives the directory a failed clone never made.
class mod_project_edit implements Module {
	/** CRUMB is the last step of the trail, which links nowhere. */
	const CRUMB = "Edit";

	/** LEAD is the sentence under the heading, as HTML. */
	const LEAD = "Settings task-ui keeps for the folder: a title, who can run it, and which sources to read. task-ui doesn't change the files in the folder.";

	/** TITLE is what the heading says before the folder's ID. */
	const TITLE = "Edit ";

	/** SAVED is where the form goes when the save worked. */
	const SAVED = "?notice=Project+saved";

	/** NOTICE is where Remove checkout goes, which says what went. */
	const NOTICE = "?notice=";

	/** REMOVED is what it says there, escaped into the query with the ID. */
	const REMOVED = "Removed ";

	/** FIELD is the id of the Recursive toggle, which its label names. */
	const FIELD = "pe-recursive";

	/** ALLOW is what nothing checked in Who can run means. */
	const ALLOW = "everyone who may run can";

	/** JOBS is what nothing checked in a job's own Who can run means. */
	const JOBS = "the folder's Who can run applies";

	/** DOT is what stands between a workspace's folder and what it is called. */
	const DOT = " · ";

	/** ROOT is what the option of the default workspace says it is. */
	const ROOT = "the root";

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
		return array("project", "workspace", "schedule", "secret", "env_file", "workflow", "workflow_step");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("project/view")) {
			$view->refused();

			return;
		}

		$id = self::asked();
		$found = self::held($chrome->all(), $id);

		if ($found === null && !Task\Folders::validID($id)) {
			$view->missing(Task\Project::says($id));

			return;
		}

		$named = $chrome->workspaces();

		$chrome->heading(self::TITLE . $id, Task\Project::crumbs($id, $named, self::CRUMB));
		$chrome->at($id);
		$chrome->asks($id);

		$projects = new Task\Project($chrome->db);
		$workspaces = new Task\Workspace($chrome->db);
		$row = $projects->row($id);
		$checkout = Task\Host::checkout(Task\Host::folderDir($id, $named));
		$users = $chrome->users->all();
		$allow = Task\Text::items(Task\Row::text($row, "allow"));
		$jobs = Task\Text::map(Task\Row::text($row, "job_allow"));
		$steps = Task\Setup::steps($chrome, Task\Setup::PROJECTS);

		$view->page(array(
			"head" => $this->head($chrome, $id, $checkout),
			"api" => "/api/projects/" . $id,
			"next" => Task\Route::href(Task\Project::PAGE) . self::SAVED,
			"back" => Task\Route::href(Task\Project::PAGE),
			"form" => self::form($row),
			"allow" => self::who("allow", $allow, $users, self::ALLOW),
			"move" => $this->move($projects, $workspaces, $id, $named),
			"recursive" => array("id" => self::FIELD, "checked" => !Task\Row::flag($row, "shallow")),
			"sidebar" => !Task\Row::flag($row, "sidebar_hidden"),
			"jobs" => self::perJob($found, $jobs, $users),
			"trail" => Task\Setup::trail($steps),
		), "project_edit.tpl");
	}

	/**
	 * head is the page head: the heading, the lead and the removal of the
	 * checkout, which only a checkout that is there has.
	 */
	function head($chrome, $id, $checkout) {
		$head = new Task\Head(self::TITLE . $id, self::LEAD);

		if (!Task\Row::flag($checkout, "found") || !$chrome->allowed("project/delete")) {
			return $head->data();
		}

		$menu = new Task\Menu("Actions on " . $id);

		$menu->danger("trash", "Remove checkout", self::removes($id, $checkout));
		$head->menu($menu->data());

		return $head->data();
	}

	/**
	 * asked is the folder ID the path names: what is left of it after the
	 * projects, with the slashes either side cut, which is what the Go route
	 * reads out of its wildcard.
	 *
	 * The path and not the parameter of the same name: a wildcard that matched
	 * nothing leaves no value in $_REQUEST at all, and a query spelling the
	 * name would stand in for it, so /settings/projects/?rest=widget would
	 * draw the page the Go route answers Not found for.
	 */
	static function asked() {
		$under = ltrim(Task\Project::PAGE, "/");
		$path = Task\Route::path();
		$rest = substr($path, strlen($under));

		return trim($rest, "/");
	}

	/**
	 * held is the catalogue folder an ID names, null where the catalogue holds
	 * none: a row of the project table is not a folder, and a page of one the
	 * host never discovered draws the settings and no jobs.
	 *
	 * The catalogue is the unfiltered one, because a folder the reader may not
	 * see is not a shorter one they may: the Go page resolves the ID through
	 * the catalogue itself, which matches the ID and nothing else.
	 */
	static function held($folders, $id) {
		foreach ($folders as $folder) {
			if (Task\Row::text($folder, "id") === $id) {
				return $folder;
			}
		}

		return null;
	}

	/**
	 * removes is what Remove checkout carries: the folder it deletes, what it
	 * asks first, and the projects it goes to saying what went.
	 *
	 * The delete is of the API at the root of the instance and carries no
	 * mount; the page it goes to is a link of this application and carries it.
	 */
	static function removes($id, $checkout) {
		$notice = Task\Route::queryEscape(self::REMOVED . $id);
		$page = Task\Route::href(Task\Project::PAGE);

		return array(
			"data-delete" => "/api/projects/" . $id,
			"data-confirm" => Task\Place::removeConfirm($id, Task\Row::text($checkout, "remote")),
			"data-confirm-ok" => "Remove checkout",
			"data-next" => $page . self::NOTICE . $notice,
		);
	}

	/**
	 * form is what the fields stand on: what the instance keeps about the
	 * project, the sources as the comma separated list the field takes.
	 *
	 * A directory with no row of its own fills none of them, which is the
	 * empty form the page draws over a project it keeps nothing about.
	 */
	static function form($row) {
		$sources = Task\Text::items(Task\Row::text($row, "sources"));

		return array(
			"title" => Task\Row::text($row, "title"),
			"description" => Task\Row::text($row, "description"),
			"sources" => implode(", ", $sources),
		);
	}

	/** who is the Who can run element: the names checked, over the users listed. */
	static function who($name, $named, $users, $empty) {
		$who = new Task\WhoField($name, "set", $named, $named, $users, $empty);

		return $who->data();
	}

	/**
	 * perJob is a row per job the folder declares: its name and who may run
	 * it, in the compact form a table cell takes.
	 *
	 * A folder the catalogue does not hold declares none and the page draws no
	 * table at all.
	 */
	static function perJob($found, $allow, $users) {
		$out = array();

		foreach (Task\Row::items($found, "jobs") as $job) {
			$id = Task\Row::text($job, "id");
			$named = isset($allow[$id]) ? $allow[$id] : array();
			$who = new Task\WhoField("job_allow." . $id, "set", $named, $named, $users, self::JOBS);

			$who->tight();

			$out[] = array("name" => Task\Row::text($job, "name"), "who" => $who->data());
		}

		return $out;
	}

	/**
	 * move is the Workspace select: the workspace the project sits in, the
	 * ones it may go to, and the rows a move leaves where they are.
	 *
	 * A named workspace and a folder deeper inside a project move nowhere and
	 * the page draws no select for them, which is the empty scope here.
	 */
	function move($projects, $workspaces, $id, $named) {
		$from = Task\Project::moves($id, $named);

		if ($from === "") {
			return array();
		}

		$rows = $workspaces->all();
		$left = $projects->refs($id);

		return array(
			"id" => $id,
			"from" => $from,
			"left" => implode("; ", $left),
			"options" => self::options($rows, $from),
		);
	}

	/**
	 * options is a workspace per row, the default one first whether a row
	 * names it or not: the root is where a project with no workspace sits.
	 *
	 * A workspace with no title is said as its folder alone, which is what the
	 * Workspaces page calls it too.
	 */
	static function options($rows, $from) {
		$root = Task\Route::INTERNAL;
		$out = array();

		$out[] = array(
			"value" => $root,
			"label" => $root . self::DOT . self::ROOT,
			"selected" => $from === $root,
		);

		foreach ($rows as $row) {
			$folder = $row->getFolder();

			if ($folder === $root) {
				continue;
			}

			$title = $row->getTitle();

			$out[] = array(
				"value" => $folder,
				"label" => $title === "" ? $folder : $folder . self::DOT . $title,
				"selected" => $folder === $from,
			);
		}

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_project_edit;

	$mod->serve();
}
