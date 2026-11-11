<?php

// @route GET /settings/workspaces/{folder}/keys/upload

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workspace_key_upload is the form that saves a private key already in use
// as an SSH key of one workspace.
//
// The page reads and does not write. The key is pasted or read from a file in
// the browser and posted to the API, which gates it again and is the only
// place the private half is stored; nothing of it is here.
class mod_workspace_key_upload implements Module {
	/** TITLE is the heading, the last crumb and what the browser tab says. */
	const TITLE = "Upload key";

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
			"form" => Task\Key::wsUpload($folder, $dir, $next, false),
			"trail" => Task\Setup::trail($steps),
		), "workspace_key_upload.tpl");
	}

	/**
	 * lead is the sentence under the heading, as text. It is built here and
	 * not in the template: a template writes the newline that sits before a
	 * tag, so a sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($folder, $dir) {
		$name = htmlspecialchars($folder, ENT_QUOTES);
		$where = htmlspecialchars($dir, ENT_QUOTES);

		return "Saves a private key you already use as an SSH key of the workspace " . $name . ". Its runs get it in their ssh agent. task-ui reads its algorithm, works out the public key and stores both in " . $where . ". The folder refuses a key it holds, under any name.";
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workspace_key_upload;

	$mod->serve();
}
