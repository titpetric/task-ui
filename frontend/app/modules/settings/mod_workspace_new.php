<?php

// @route GET /settings/workspaces/new

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workspace_new is the form that adds a workspace: the folder it is made
// under the root, the repository cloned into its home, and what the sidebar
// calls it.
//
// The page reads and does not write. The form posts to /api/workspaces, which
// gates it again, makes the directories and runs the clone; nothing of that is
// here.
class mod_workspace_new implements Module {
	/** TITLE is the heading, the last crumb and what the browser tab says. */
	const TITLE = "Add workspace";

	/** LEAD is the sentence under the heading, as HTML: it names the directories. */
	const LEAD = "Makes a directory under the root, the folder you name, with an empty <code>home/</code>, the <code>HOME</code> of the workspace's runs, and a <code>src/</code> for its projects. With a repository URL, it clones the repository into the home, where atkins finds <code>.atkins/skills</code> and Claude its <code>CLAUDE.md</code>. If the clone fails, you see git's output here and task-ui keeps nothing. Then clone projects into it with Add project.";

	/**
	 * NEXT is the page the workspace that was added opens, which the script
	 * fills the folder of: the API answers the row and `{folder}` is the field
	 * of it the link takes.
	 */
	const NEXT = "/settings/workspaces/{folder}?notice=Workspace+added";

	/** RECURSIVE is the id of the Recursive toggle, which its label points at. */
	const RECURSIVE = "w-recursive";

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
		return array();
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("workspace/edit")) {
			$view->refused();

			return;
		}

		$chrome->heading(self::TITLE, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::WORKSPACES, "href" => Task\Workspace::PAGE),
			array("text" => self::TITLE, "href" => ""),
		));

		$head = new Task\Head(self::TITLE, self::LEAD);
		$steps = Task\Setup::steps($chrome, Task\Setup::WORKSPACES);

		$view->page(array(
			"head" => $head->data(),
			"next" => self::NEXT,
			"recursive" => array("id" => self::RECURSIVE, "checked" => true),
			"cancel" => Task\Route::href(Task\Workspace::PAGE),
			"trail" => Task\Setup::trail($steps),
		), "workspace_new.tpl");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workspace_new;

	$mod->serve();
}
