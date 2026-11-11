<?php

// @route GET /settings/sidebar

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_sidebar is the sidebar as a table: a row per item, in the order the
// sidebar lists them, with a grip to drag it, the link it opens and a switch
// that leaves it out.
//
// Every row is here, the hidden ones too, because this is the page that
// unhides one. The sidebar beside the table shows the same rows less the
// hidden, so the two disagree until Save, which is the point: the table is
// what the sidebar would be.
//
// The page reads and does not write. The form carries data-sidebar, which the
// script puts to /api/sidebar and then reloads.
class mod_sidebar implements Module {
	/** LEAD is the sentence under the heading, as HTML: it emphasises a column. */
	const LEAD = "The sidebar's top block, in the order it lists them. Drag a row by its grip, or focus the grip and press the up and down arrows. <strong>Shown</strong> leaves an item out of the sidebar without deleting it: the page behind it still answers, and its permission still decides who may open it.";

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
		return array("sidebar");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!Task\Sidebar::settings($chrome->access)) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::SIDEBAR, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::SIDEBAR, "href" => ""),
		));

		$admin = Task\User::admin($chrome->user);
		$store = new Task\Sidebar;
		$stored = $store->all();
		$items = count($stored) > 0 ? $stored : Task\Sidebar::defaults();
		$head = new Task\Head(Task\Setup::SIDEBAR, self::lead($admin));
		$steps = Task\Setup::steps($chrome, Task\Setup::SIDEBAR);
		$rows = array();

		foreach (Task\Sidebar::rows($items, Task\Host::modules()) as $row) {
			$row["number"] = count($rows) + 1;
			$row["menu"] = array();

			if ($admin) {
				$menu = self::menu($row);
				$row["menu"] = $menu->data();
			}

			$rows[] = $row;
		}

		$view->page(array(
			"head" => $head->data(),
			"admin" => $admin,
			"rows" => $rows,
			"trail" => Task\Setup::trail($steps),
		), "sidebar.tpl");
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
		$menu->item("up", "Move up", array("data-row-move" => "up"));
		$menu->item("down", "Move down", array("data-row-move" => "down"));

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
