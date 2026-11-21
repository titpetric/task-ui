<?php

// @route GET /settings/sidebar

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_sidebar is the sidebar as a table: a row per module that enabled a row
// in its modinfo.php, in the order the sidebar lists them, with a grip to drag
// it, the link it opens and a switch that leaves it out.
//
// Every enabled module is here, the hidden ones too, because this is the page
// that unhides one. The sidebar beside the table shows the same rows less the
// hidden, so the two disagree until Save, which is the point: the table is
// what the sidebar would be.
//
// The page seeds the rows the table does not hold and reads them back. The
// form carries data-sidebar, which the script puts to /api/sidebar and then
// reloads.
class mod_sidebar implements Module {
	/** LEAD is the sentence under the heading, as HTML: it emphasises a column. */
	const LEAD = "The sidebar's modules, in the order it lists them. A module calls <code>enableSideBar</code> in its modinfo.php to appear here. Drag a row by its grip, or focus the grip and press the up and down arrows. <strong>Show in sidebar</strong> leaves a module out without deleting it: its page still answers, and its permission still decides who may open it.";

	/** SAVES is what the sentence says further to a reader who may not save. */
	const SAVES = " An admin saves the order.";

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
		return ["sidebar"];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!Task\Sidebar::settings($chrome->access)) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::SIDEBAR, [
			["text" => "Settings", "href" => "/settings"],
			["text" => Task\Setup::SIDEBAR, "href" => ""],
		]);

		$admin = Task\User::admin($chrome->user);
		$store = new Task\Sidebar(Task\DB::writer());
		$store->seed();
		$items = $store->all();
		$head = new Task\Head(Task\Setup::SIDEBAR, self::lead($admin));
		$steps = Task\Setup::steps($chrome, Task\Setup::SIDEBAR);
		$rows = [];

		foreach ($store->rows($items) as $row) {
			$row["number"] = count($rows) + 1;
			$row["menu"] = [];

			if ($admin) {
				$menu = self::menu($row);
				$row["menu"] = $menu->data();
			}

			$rows[] = $row;
		}

		$view->page([
			"head" => $head->data(),
			"admin" => $admin,
			"rows" => $rows,
			"cancel" => Task\Route::href("/settings"),
			"trail" => Task\Setup::trail($steps),
		], "sidebar.tpl");
	}

	/**
	 * lead is the sentence under the heading. It is HTML and built here, not in
	 * the template: a template writes the newline that sits before a tag, so a
	 * sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($admin) {
		return $admin ? self::LEAD : self::LEAD . self::SAVES;
	}

	/**
	 * menu is a row's actions: the link and the kind on top, then the two ways
	 * to move it without a pointer.
	 */
	static function menu($row) {
		$menu = new Task\Menu("Actions on " . $row["title"]);
		$menu->detail("Link", $row["href"], true);
		$menu->detail("Kind", $row["kind"]);
		$menu->item("up", "Move up", ["data-row-move" => "up"]);
		$menu->item("down", "Move down", ["data-row-move" => "down"]);

		return $menu;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_sidebar;
	$mod->serve();
}
