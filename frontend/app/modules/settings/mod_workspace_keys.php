<?php

// @route GET /settings/workspaces/{folder}/keys

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workspace_keys is the SSH keys of one workspace: the keys its runs get
// in their ssh agent, each one to every run of the workspace.
//
// The page reads and does not write. Generate and Upload open dialogs that
// post to the API, and Remove in a row's menu deletes through it; the private
// half is the keystore's and does not enter PHP.
class mod_workspace_keys implements Module {
	/** EMPTY_LINE is what the Keys head says where the workspace has none. */
	const EMPTY_LINE = "No keys added.";

	/** CRUMB is the last step of the trail, and what the browser tab says. */
	const CRUMB = "Keys";

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

		if (!$chrome->allowed("workspace/view")) {
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

		$title = "Keys of " . $folder;
		$here = Task\Workspace::PAGE . "/" . $folder . "/keys";

		$chrome->heading($title, Task\Workspace::crumbs($folder, self::CRUMB));
		$chrome->asks($folder);

		$found = Task\Key::folder();
		$dir = Task\Row::text($found, "dir");
		$keys = Task\Key::ofWorkspace($found["keys"], Task\Row::text($row, "id"));
		$edit = $chrome->allowed("workspace/edit");
		$steps = Task\Setup::steps($chrome, Task\Setup::WORKSPACES);

		$view->page(array(
			"head" => $this->head($chrome, $folder, $title, $edit && $dir !== ""),
			"dir" => $dir,
			"block" => $this->block($keys),
			"keys" => Task\Key::rows($chrome, $keys, $here, $folder),
			"dialogs" => $edit ? Task\Key::wsDialogs($folder, $dir, $here) : array(),
			"trail" => Task\Setup::trail($steps),
		), "workspace_keys.tpl");
	}

	/**
	 * head is the page head: the heading, the lead, the two buttons that open
	 * the add dialogs and the workspace's own menu.
	 */
	function head($chrome, $folder, $title, $adds) {
		$buttons = new Task\Buttons;
		$head = new Task\Head($title, self::lead($folder));
		$menu = $this->menu($chrome, $folder);

		if ($adds) {
			$buttons->item("upload", "Upload key", array("data-dialog-open" => "key-upload-dialog"));
			$buttons->primary("key", "Generate key", array("data-dialog-open" => "key-generate-dialog"));
		}

		$head->buttons($buttons->data());
		$head->menu($menu->data());

		return $head->data();
	}

	/**
	 * lead is the sentence under the heading, as text: a key added here is a
	 * key of the workspace and nothing narrows it further.
	 */
	static function lead($folder) {
		$where = htmlspecialchars($folder, ENT_QUOTES);

		return "The keys the workspace's runs get in their ssh agent, each one to every run in " . $where . ": a key added here is a key of this workspace and nothing narrows it further. Generate, upload and remove them here.";
	}

	/** menu is the workspace's own actions: its folder, its settings and its variables. */
	function menu($chrome, $folder) {
		$menu = new Task\Menu("Actions on " . $folder);
		$href = Task\Route::folderHref($folder, $chrome->workspaces());

		$menu->item("folder-open", "Open", array("href" => Task\Route::href($href)));
		$menu->item("settings", "Edit workspace", array("href" => Task\Route::href(Task\Workspace::PAGE . "/" . $folder)));
		$menu->when($chrome->allowed("secret/view"), "secret", "Secrets", array(
			"href" => Task\Route::href("/settings/secrets?workspace=" . Task\Route::queryEscape($folder)),
		));

		return $menu;
	}

	/** block is the head of the Keys section: how many there are, and the line for none. */
	function block($keys) {
		$found = count($keys);

		return array(
			"title" => "Keys",
			"note" => Task\Setup::says($found, "key", "keys"),
			"empty" => $found === 0 ? self::EMPTY_LINE : "",
			"buttons" => array(),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workspace_keys;

	$mod->serve();
}
