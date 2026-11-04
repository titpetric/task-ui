<?php

// @route GET /settings/workspaces/{folder}/keys/generate

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workspace_key_generate is the form that makes an SSH key of one
// workspace: its name, its algorithm and its comment.
//
// The page reads and does not write. The form posts to the API, which gates it
// again and is the only place the private half is made; nothing of it is here.
class mod_workspace_key_generate implements Module {
	/** TITLE is the heading, the last crumb and what the browser tab says. */
	const TITLE = "Generate key";

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

		if (!$chrome->allowed("workspace/edit")) {
			$view->refused();

			return;
		}

		$folder = Task\Route::param("folder");
		$workspaces = new Task\Workspace($chrome->db);
		$row = $workspaces->row($folder);

		if ($row === null) {
			$view->missing(Task\Workspace::says($folder));

			return;
		}

		$chrome->heading(self::TITLE, Task\Workspace::crumbs($folder, self::TITLE));
		$chrome->asks($folder);

		$found = Task\Key::folder();
		$dir = Task\Row::text($found, "dir");
		$next = Task\Workspace::PAGE . "/" . $folder;
		$head = new Task\Head(self::TITLE, self::lead($folder, $dir));
		$steps = Task\Setup::steps($chrome, Task\Setup::WORKSPACES);

		$view->page(array(
			"head" => $head->data(),
			"form" => Task\Key::wsGenerate($folder, $dir, $next, false),
			"trail" => Task\Setup::trail($steps),
		), "workspace_key_generate.tpl");
	}

	/**
	 * lead is the sentence under the heading, as HTML. It is built here and
	 * not in the template: a template writes the newline that sits before a
	 * tag, so a sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($folder, $dir) {
		$name = htmlspecialchars($folder, ENT_QUOTES);
		$where = htmlspecialchars($dir, ENT_QUOTES);

		return "Generates an SSH key of the workspace " . $name . " in " . $where . ", named <span class=\"mono\">workspace_" . $name . "_&lt;name&gt;_&lt;algorithm&gt;</span>. Its runs get it in their ssh agent. A name is unique across the folder.";
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workspace_key_generate;

	$mod->serve();
}
