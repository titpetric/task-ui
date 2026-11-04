<?php

// @route GET /settings/workspaces/{folder}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workspace_edit is one workspace: its settings, the home its runs get as
// HOME, the projects under its src and the SSH keys of every run in it.
//
// The page reads and does not write. Save puts the settings to the API, the
// home dialog clones a repository into the home through it, Remove empties
// the home, Export downloads a bundle and Delete removes the workspace; each
// gates the change again.
class mod_workspace_edit implements Module {
	/** CRUMB is the last step of the trail. */
	const CRUMB = "Edit";

	/** LEAD is the sentence under the heading, as HTML: what the page holds. */
	const LEAD = "The workspace's settings, its home repository, its projects and the SSH keys its runs get. The workspace holds its home, the <code>HOME</code> of its runs, and its projects side by side under <code>src/</code>; its folder stays as added.";

	/** ROOT_LEAD is what the lead adds for the default workspace, which is the root. */
	const ROOT_LEAD = " The internal workspace is the root itself: deleting it removes the root's home and projects, and the instance runs without one.";

	/** RECURSIVE is the id of the Recursive toggle, which its label points at. */
	const RECURSIVE = "we-recursive";

	/** SAVED is where Save goes: this page again, with the notice. */
	const SAVED = "?notice=Workspace+saved";

	/** DELETED is where Delete goes: the Workspaces page, with the notice. */
	const DELETED = "/settings/workspaces?notice=Workspace+deleted";

	/** CLONED is where the home dialog goes once the clone is done. */
	const CLONED = "?notice=Home+repository+cloned";

	/** REMOVED is where Remove of the home goes: this page, with the notice. */
	const REMOVED = "?notice=Home+repository+removed";

	/** ROOT_DIR starts what the Settings heading says for the default workspace. */
	const ROOT_DIR = "the root, ";

	/** HOME_TITLE is the Home repository block's heading. */
	const HOME_TITLE = "Home repository";

	/** HOME_NOTE ends the home block's note, after the folder the home is. */
	const HOME_NOTE = "/, the HOME of the workspace's runs";

	/** HOME_EMPTY is the home block's line where no repository is cloned there. */
	const HOME_EMPTY = "The home is an empty directory.";

	/** PROJECTS_TITLE is the Projects block's heading. */
	const PROJECTS_TITLE = "Projects";

	/** ROOT_NOTE is the Projects block's note for the default workspace. */
	const ROOT_NOTE = "in src/ at the root";

	/** KEYS_TITLE is the Keys block's heading. */
	const KEYS_TITLE = "Keys";

	/** KEYS_NOTE is the Keys block's note: a key here is a key of the workspace. */
	const KEYS_NOTE = "SSH keys the workspace's runs get";

	/** KEYS_EMPTY is the Keys block's line where the workspace has none. */
	const KEYS_EMPTY = "No keys added.";

	/** RELOAD is where a row's delete goes: the page it stands on, read again. */
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
		return array("workspace", "project", "secret");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("workspace/view")) {
			$view->refused();

			return;
		}

		$asked = Task\Route::param("folder");
		$workspaces = new Task\Workspace($chrome->db);
		$row = $workspaces->row($asked);

		if ($row === null) {
			$view->missing(Task\Workspace::says($asked));

			return;
		}

		$folder = Task\Row::text($row, "folder");
		$named = $chrome->workspaces();
		$title = "Edit " . $folder;
		$here = Task\Workspace::PAGE . "/" . $folder;
		$href = Task\Route::folderHref($folder, $named);

		$chrome->heading($title, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::WORKSPACES, "href" => Task\Workspace::PAGE),
			array("text" => Task\Route::crumbName($folder), "href" => $href),
			array("text" => self::CRUMB, "href" => ""),
		));
		$chrome->asks($folder);

		$instance = Task\Host::instance(false);
		$root = Task\Row::text($instance, "root");
		$entries = Task\Host::entries($workspaces->names());
		$inside = Task\Workspace::inside($entries, $folder, $named);
		$spaces = Task\Host::spaces($this->dirs($inside));
		$projects = new Task\Project($chrome->db);
		$nested = Task\Workspace::nested($entries, $projects->folders(), $named);
		$secrets = new Task\Secret($chrome->db);
		$counts = Task\Workspace::counts($secrets->scopes(), $named);
		$found = isset($counts[$folder]) ? $counts[$folder] : 0;
		$ids = isset($nested[$folder]) ? $nested[$folder] : array();
		$keyfolder = Task\Key::folder();
		$dir = Task\Row::text($keyfolder, "dir");
		$keys = Task\Key::ofWorkspace($keyfolder["keys"], Task\Row::text($row, "id"));
		$edit = $chrome->allowed("workspace/edit");
		$admin = Task\User::admin($chrome->user);
		$steps = Task\Setup::steps($chrome, Task\Setup::WORKSPACES);

		$view->page(array(
			"head" => $this->head($chrome, $folder, $title, $ids, $found, $href, $dir !== "" && $admin),
			"settings" => $this->settings($row, $folder, $root),
			"home" => $this->home($chrome, $row, $folder, $named, $edit),
			"block" => $this->block($folder, $inside),
			"table" => Task\Project::table($chrome, $inside, $spaces, self::RELOAD),
			"keys" => $dir === "" ? array() : $this->keys($chrome, $keys, $here, $folder, $edit && $dir !== ""),
			"dialogs" => $edit && $dir !== "" ? Task\Key::wsDialogs($folder, $dir, $here) : array(),
			"export" => $dir === "" ? "" : $folder,
			"trail" => Task\Setup::trail($steps),
		), "workspace_edit.tpl");
	}

	/** dirs is the directory of each entry, which the measure of them is asked for. */
	function dirs($entries) {
		$out = array();

		foreach ($entries as $entry) {
			$out[] = Task\Row::text($entry, "dir");
		}

		return $out;
	}

	/**
	 * head is the page head: the heading, the lead, Add project, Export and the
	 * workspace's own menu.
	 *
	 * Export is an admin's alone and only where the instance has a key folder:
	 * the bundle carries the workspace's secrets and the keys its runs get.
	 */
	function head($chrome, $folder, $title, $ids, $found, $href, $exports) {
		$buttons = new Task\Buttons;
		$head = new Task\Head($title, self::lead($folder));
		$menu = $this->menu($chrome, $folder, $ids, $found, $href);
		$adds = Task\Route::href(Task\Place::PROJECTS . "/new") . "?workspace=" . Task\Route::queryEscape($folder);

		if ($chrome->allowed("project/edit")) {
			$buttons->primary("plus", "Add project", array("href" => $adds));
		}

		$buttons->when($exports, "download", "Export", array("data-dialog-open" => "export-dialog"));
		$head->buttons($buttons->data());
		$head->menu($menu->data());

		return $head->data();
	}

	/** lead is the sentence under the heading; the default workspace is the root and says so. */
	static function lead($folder) {
		if ($folder === Task\Route::INTERNAL) {
			return self::LEAD . self::ROOT_LEAD;
		}

		return self::LEAD;
	}

	/**
	 * menu is the workspace's actions: its folder and the Delete that removes
	 * it, which the variables reaching it stop.
	 */
	function menu($chrome, $folder, $ids, $found, $href) {
		$menu = new Task\Menu("Actions on " . $folder);
		$deletes = $chrome->allowed("workspace/delete");
		$secrets = Task\Route::href(Task\Place::SECRETS) . "?workspace=" . Task\Route::queryEscape($folder);

		$menu->item("folder-open", "Open", array("href" => Task\Route::href($href)));
		$menu->when($deletes && $found > 0, "secret", Task\Workspace::blocked($found), array("href" => $secrets));
		$menu->whenDanger($deletes && $found === 0, "trash", "Delete", array(
			"data-delete" => "/api/workspaces/" . $folder,
			"data-confirm" => Task\Workspace::confirm($folder, $ids),
			"data-confirm-ok" => "Delete",
			"data-next" => self::DELETED,
		));

		return $menu;
	}

	/**
	 * settings is the form that saves the workspace: what the sidebar calls it,
	 * how deep its folders are listed, how many of them, and when its runs may
	 * start.
	 *
	 * The heading says where the directory is. Sidebar is empty rather than 0
	 * where every project is listed, which is what the field means by all.
	 */
	function settings($row, $folder, $root) {
		$limit = Task\Row::number($row, "sidebar_limit");
		$page = Task\Workspace::PAGE . "/" . $folder;

		return array(
			"dir" => self::where($root, $folder),
			"action" => "/api/workspaces/" . $folder,
			"next" => $page . self::SAVED,
			"title" => Task\Row::text($row, "title"),
			"description" => Task\Row::text($row, "description"),
			"recursive" => array("id" => self::RECURSIVE, "checked" => !Task\Row::flag($row, "shallow")),
			"limit" => $limit > 0 ? $limit : "",
			"hours" => Task\Row::text($row, "hours"),
			"cancel" => Task\Route::href(Task\Workspace::PAGE),
		);
	}

	/**
	 * where is what the Settings heading says after the title: the directory of
	 * the workspace under the root.
	 *
	 * The default workspace is the root itself, which is no directory of its
	 * own, and the heading names it as the root.
	 */
	static function where($root, $folder) {
		if ($folder === Task\Route::INTERNAL) {
			return self::ROOT_DIR . $root;
		}

		return Task\Workspace::dir($root, $folder);
	}

	/**
	 * home is the workspace's home: the repository cloned there, what git knows
	 * about the checkout, and the pull of its git skill.
	 *
	 * A home with no repository is an empty directory and says so in the block
	 * head; the dialog that clones one into it is the same dialog that replaces
	 * what is there, which is why Clone and Replace are one item.
	 */
	function home($chrome, $row, $folder, $named, $edit) {
		$id = Task\Workspace::homeId($folder);
		$url = Task\Row::text($row, "home_url");
		$branch = Task\Row::text($row, "home_branch");
		$dir = Task\Host::folderDir($id, $named);
		$checkout = Task\Host::checkout($dir);
		$pull = Task\Project::pull($chrome, $id);
		$empty = $url === "";

		return array(
			"block" => $this->homeBlock($folder, $id, $pull, $empty, $edit),
			"empty" => $empty,
			"folder" => $id,
			"href" => Task\Route::href(Task\Route::folderHref($id, $named)),
			"url" => $url,
			"branch" => $branch,
			"checkout" => $this->checkout($checkout),
			"workspace" => $folder,
			"edit" => $edit,
			"clone" => $empty ? "Clone" : "Replace",
			"action" => "/api/workspaces/" . $folder . "/home",
			"next" => Task\Workspace::PAGE . "/" . $folder . self::CLONED,
		);
	}

	/** checkout is what git knows about the home's checkout, none where it holds none. */
	function checkout($found) {
		$when = Task\Row::text($found, "when");

		return array(
			"found" => Task\Row::flag($found, "found"),
			"branch" => Task\Row::text($found, "branch"),
			"commit" => Task\Row::text($found, "commit"),
			"subject" => Task\Row::text($found, "subject"),
			"ago" => $when === "" ? "" : Task\Clock::ago($when),
		);
	}

	/**
	 * homeBlock is the head of the Home repository block: where the home is,
	 * the line for an empty one, and the three items that act on it.
	 */
	function homeBlock($folder, $id, $pull, $empty, $edit) {
		$buttons = new Task\Buttons(true);
		$page = Task\Workspace::PAGE . "/" . $folder;
		$asks = "Remove the home repository of " . $folder . "? Its checkout goes, and the home becomes an empty directory. Runs keep HOME there.";

		if ($pull !== null) {
			$buttons->item("download", "Pull now", Task\Project::pulls($id, $pull));
		}

		$buttons->when($edit, "plus", $empty ? "Clone" : "Replace", array("data-dialog-open" => "home-replace-dialog"));
		$buttons->whenDanger($edit && !$empty, "", "Remove", array(
			"data-api" => "/api/workspaces/" . $folder . "/home",
			"data-method" => "DELETE",
			"data-confirm" => $asks,
			"data-confirm-ok" => "Remove",
			"data-next" => $page . self::REMOVED,
		));

		return array(
			"title" => self::HOME_TITLE,
			"note" => $id . self::HOME_NOTE,
			"empty" => $empty ? self::HOME_EMPTY : "",
			"buttons" => $buttons->data(),
		);
	}

	/**
	 * block is the head of the Projects block: where the workspace's projects
	 * are and the line for a workspace with none.
	 */
	function block($folder, $inside) {
		$note = $folder === Task\Route::INTERNAL ? self::ROOT_NOTE : "inside " . $folder . "/src/";
		$empty = count($inside) === 0 ? "No projects in " . $folder . " yet." : "";

		return array(
			"title" => self::PROJECTS_TITLE,
			"note" => $note,
			"empty" => $empty,
			"buttons" => array(),
		);
	}

	/**
	 * keys is the Keys block: the keys of the workspace and the two items that
	 * add one, which are the reader's where they may edit the workspace.
	 */
	function keys($chrome, $keys, $here, $folder, $adds) {
		$buttons = new Task\Buttons(true);
		$found = count($keys);

		$buttons->when($adds, "upload", "Upload key", array("data-dialog-open" => "key-upload-dialog"));
		$buttons->when($adds, "key", "Generate key", array("data-dialog-open" => "key-generate-dialog"));

		return array(
			"block" => array(
				"title" => self::KEYS_TITLE,
				"note" => self::KEYS_NOTE,
				"empty" => $found === 0 ? self::KEYS_EMPTY : "",
				"buttons" => $buttons->data(),
			),
			"rows" => Task\Key::rows($chrome, $keys, $here, $folder),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workspace_edit;

	$mod->serve();
}
