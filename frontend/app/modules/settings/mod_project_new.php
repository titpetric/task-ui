<?php

// @route GET /settings/projects/new

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_project_new is the form that clones a repository into a workspace.
//
// The page reads and does not write: the form posts to the API, which clones
// and keeps the project. The workspace the page was opened for is the one the
// select stands on and the page Cancel goes back to.
class mod_project_new implements Module {
	/** TITLE is the heading, the last crumb and what the button says. */
	const TITLE = "Add project";

	/** LEAD is the sentence under the heading, as HTML. */
	const LEAD = "Clones the repository into a workspace; the internal workspace, internal, is the root. If the clone works, the folder appears. If it fails, you see git's output here and task-ui keeps nothing. task-ui uses the instance SSH keys to reach private repositories.";

	/** NOTICE is what the page the form goes to says when the clone worked. */
	const NOTICE = "?notice=Project+added";

	/** FIELD is the id of the Recursive toggle, which its label names. */
	const FIELD = "p-recursive";

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
		return array("workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("project/edit")) {
			$view->refused();

			return;
		}

		$chrome->heading(self::TITLE, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::PROJECTS, "href" => Task\Project::PAGE),
			array("text" => self::TITLE, "href" => ""),
		));

		$workspaces = new Task\Workspace($chrome->db);
		$rows = $workspaces->all();
		$chosen = Task\Route::input("workspace");
		$back = self::back($chosen);
		$steps = Task\Setup::steps($chrome, Task\Setup::PROJECTS);

		$view->page(array(
			"head" => $this->head(),
			"options" => self::options($rows, $chosen),
			"back" => Task\Route::linked($back),
			"next" => $back . self::NOTICE,
			"recursive" => array("id" => self::FIELD, "checked" => true),
			"trail" => Task\Setup::trail($steps),
		), "project_new.tpl");
	}

	/** head is the page head: the heading and the lead, and no action beside them. */
	function head() {
		$head = new Task\Head(self::TITLE, self::LEAD);

		return $head->data();
	}

	/**
	 * back is where Cancel goes and where the form goes when it worked: the
	 * workspace the page was opened for, else the projects.
	 *
	 * The workspace is the query value as it arrived, which is not a folder
	 * the page read: a name nothing answers to is a link to a page that is
	 * not there, the way the Go page links it. Cancel carries it escaped as a
	 * link and data-next carries it as it is, which is what the two attributes
	 * of the Go markup hold: one is a link to html/template and one is not.
	 */
	static function back($chosen) {
		if ($chosen === "") {
			return Task\Route::href(Task\Project::PAGE);
		}

		return Task\Route::href(Task\Workspace::PAGE . "/" . $chosen);
	}

	/**
	 * options is the Workspace select: a workspace per row, in the order the
	 * settings list them, each said as its folder and what it is called.
	 *
	 * The one the page was opened for stands selected; opened for none, the
	 * first row does, which is the default workspace while there is one.
	 */
	static function options($rows, $chosen) {
		$out = array();

		foreach ($rows as $at => $row) {
			$folder = $row->getFolder();
			$title = $row->getTitle();

			if ($folder === Task\Route::INTERNAL) {
				$title = self::ROOT;
			}

			$out[] = array(
				"value" => $folder,
				"folder" => $folder,
				"title" => $title === "" ? $folder : $title,
				"selected" => $folder === $chosen || ($chosen === "" && $at === 0),
			);
		}

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_project_new;

	$mod->serve();
}
