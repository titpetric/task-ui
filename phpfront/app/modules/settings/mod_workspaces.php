<?php

// @route GET /settings/workspaces

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workspaces is the directories under the root: a row per workspace with
// its home repository, how many projects sit under its src and how many
// variables reach it.
//
// The page reads and does not write. Move up and Move down put the order to
// the API, Delete removes a workspace through it, and Import opens the dialog
// that posts a bundle; each gates the change again.
class mod_workspaces implements Module {
	/** LEAD is the sentence under the heading, as HTML: it names the directories. */
	const LEAD = "Directories under the root, each with a home, the <code>HOME</code> of the workspace's runs and a checkout of its home repository when it has one, and its projects side by side under <code>src/</code>. The folder name is the projects' namespace; <code>internal</code>, the root itself; the instance runs without it after Delete.";

	/** BLOCKED is what the Secrets count says a delete waits for. */
	const BLOCKED = "delete them before deleting the workspace";

	/** RELOAD is where Delete goes: the page it is read on, read again. */
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

		$chrome->heading(Task\Setup::WORKSPACES, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::WORKSPACES, "href" => ""),
		));

		$workspaces = new Task\Workspace($chrome->db);
		$rows = $workspaces->all();
		$named = $workspaces->folders();
		$entries = Task\Host::entries($workspaces->names());
		$projects = new Task\Project($chrome->db);
		$secrets = new Task\Secret($chrome->db);
		$nested = Task\Workspace::nested($entries, $projects->folders(), $named);
		$counts = Task\Workspace::counts($secrets->scopes(), $named);
		$homes = Task\Workspace::homes($rows);
		$instance = Task\Host::instance(false);
		$steps = Task\Setup::steps($chrome, Task\Setup::WORKSPACES);
		$admin = Task\User::admin($chrome->user);

		$view->page(array(
			"head" => $this->head($admin),
			"rows" => $this->rows($chrome, $rows, $nested, $counts, $homes, $named),
			"root" => Task\Row::text($instance, "root"),
			"imports" => $admin,
			"trail" => Task\Setup::trail($steps),
		), "workspaces.tpl");
	}

	/**
	 * head is the page head: the heading, the lead, Import for an admin and Add
	 * workspace.
	 *
	 * Import is an admin's alone because a bundle carries the secrets and the
	 * SSH keys of the workspace that was exported, which this instance would
	 * then hold.
	 */
	function head($admin) {
		$buttons = new Task\Buttons;
		$head = new Task\Head(Task\Setup::WORKSPACES, self::LEAD);

		$buttons->when($admin, "upload", "Import", array("data-dialog-open" => "import-dialog"));
		$buttons->primary("plus", "Add workspace", array("href" => Task\Route::href(Task\Workspace::PAGE . "/new")));
		$head->buttons($buttons->data());

		return $head->data();
	}

	/**
	 * rows is a row per workspace, in the order the settings list them: the
	 * default one first and the named ones in the sidebar's order.
	 *
	 * first is the row Move up stops at and last the row Move down stops at:
	 * the default workspace does not move at all, so the named ones start one
	 * further down where a row of it is first.
	 */
	function rows($chrome, $rows, $nested, $counts, $homes, $named) {
		$last = count($rows);
		$first = 0;
		$out = array();

		if ($last > 0 && $rows[0]->getFolder() === Task\Route::INTERNAL) {
			$first = 1;
		}

		foreach ($rows as $at => $row) {
			$out[] = $this->row($chrome, $row, $nested, $counts, $homes, $named, $at, $first, $last);
		}

		return $out;
	}

	/** row is one workspace: the cells the table draws and the menu beside them. */
	function row($chrome, $row, $nested, $counts, $homes, $named, $at, $first, $last) {
		$folder = $row->getFolder();
		$ids = isset($nested[$folder]) ? $nested[$folder] : array();
		$found = isset($counts[$folder]) ? $counts[$folder] : 0;
		$links = $this->links($folder, $named);
		$menu = $this->menu($chrome, $folder, $ids, $found, $links, $at, $first, $last);

		return array(
			"folder" => $folder,
			"href" => $links["href"],
			"edit" => $links["edit"],
			"add" => $links["add"],
			"variables" => $links["secrets"],
			"title" => $row->getTitle(),
			"description" => $row->getDescription(),
			"home" => isset($homes[$folder]) ? $homes[$folder] : "",
			"projects" => count($ids),
			"names" => implode(", ", $ids),
			"secrets" => $found,
			"blocked" => self::BLOCKED,
			"adds" => $chrome->allowed("project/edit"),
			"addsSecret" => $chrome->allowed("secret/edit"),
			"menu" => $menu->data(),
		);
	}

	/**
	 * links are the pages of one workspace: its folder, its settings, the form
	 * that clones a project into it, its SSH keys and the variables that reach
	 * it.
	 */
	function links($folder, $named) {
		$page = Task\Workspace::PAGE . "/" . $folder;
		$escaped = Task\Route::queryEscape($folder);

		return array(
			"href" => Task\Route::href(Task\Route::folderHref($folder, $named)),
			"edit" => Task\Route::href($page),
			"keys" => Task\Route::href($page . "/keys"),
			"add" => Task\Route::href(Task\Place::PROJECTS . "/new") . "?workspace=" . $escaped,
			"secrets" => Task\Route::href(Task\Place::SECRETS) . "?workspace=" . $escaped,
		);
	}

	/**
	 * menu is a row's actions: the pages of the workspace, where it sits in the
	 * order, and the Delete that removes it.
	 *
	 * The default workspace does not move: its projects are the root's and the
	 * order is of the named ones. A workspace that holds variables is not
	 * deleted either, and the item says so and opens them instead: nothing here
	 * removes somebody's credential as a side effect of removing a directory.
	 */
	function menu($chrome, $folder, $ids, $found, $links, $at, $first, $last) {
		$menu = new Task\Menu("Actions on " . $folder);
		$default = $folder === Task\Route::INTERNAL;
		$deletes = $chrome->allowed("workspace/delete");

		$menu->item("folder-open", "Open", array("href" => $links["href"]));
		$menu->item("settings", "Edit workspace", array("href" => $links["edit"]));
		$menu->when($chrome->allowed("project/edit"), "plus", "Add project", array("href" => $links["add"]));
		$menu->when($chrome->allowed("workspace/view"), "key", "SSH keys", array("href" => $links["keys"]));
		$menu->when($chrome->allowed("secret/view"), "secret", "Secrets", array("href" => $links["secrets"]));
		$menu->sep();
		$menu->when(!$default && $at !== $first, "up", "Move up", array("data-workspace-move" => "up", "data-folder" => $folder));
		$menu->when(!$default && $at + 1 !== $last, "down", "Move down", array("data-workspace-move" => "down", "data-folder" => $folder));
		$menu->when($deletes && $found > 0, "secret", Task\Workspace::blocked($found), array("href" => $links["secrets"]));
		$menu->whenDanger($deletes && $found === 0, "trash", "Delete", array(
			"data-delete" => "/api/workspaces/" . $folder,
			"data-confirm" => Task\Workspace::confirm($folder, $ids),
			"data-confirm-ok" => "Delete",
			"data-next" => self::RELOAD,
		));

		return $menu;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workspaces;

	$mod->serve();
}
