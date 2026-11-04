<?php

namespace Task;

// Sidebar reads and writes the sidebar table: the top block of the sidebar,
// in its order, and the rows the Sidebar settings page edits.
class Sidebar {
	/** VIEW and MODULE are the kinds a row may be. */
	const VIEW = "view";

	const MODULE = "module";

	/** SETTINGS is the one rule that is not a permission path. */
	const SETTINGS = "settings";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
	}

	/** all is every row, hidden ones included, in the order the table holds. */
	function all() {
		return Model\Sidebar::rows($this->db->get_all("SELECT * FROM sidebar ORDER BY position, title"));
	}

	/** byHref is one row by its link, or null. */
	function byHref($href) {
		$row = $this->db->get("SELECT * FROM sidebar WHERE href = ?", $href);

		if (!is_array($row)) {
			return null;
		}

		return new Model\Sidebar($row);
	}

	/** order writes the position and the hidden flag of each link, in one transaction. */
	function order($hrefs, $hidden) {
		$now = Clock::stamp();

		$this->db->begin();

		// A runtime failure in the loop is not catchable: an Error and a
		// fault the interpreter raises are neither an Exception. Without
		// this the frame leaves a write transaction open on the pooled
		// connection, and every other writer waits on it.
		defer($this->db->rollback);

		foreach ($hrefs as $position => $href) {
			$this->db->query("UPDATE sidebar SET position = ?, hidden = ?, updated_at = ? WHERE href = ?", $position, isset($hidden[$href]) && $hidden[$href] ? 1 : 0, $now, $href);
		}

		$this->db->commit();

		return true;
	}

	/**
	 * items is the top block as one user sees it: the table's order, less
	 * the hidden rows and the rows whose rule the user lacks.
	 *
	 * Two rows carry more than a link, and the template keys on the flag and
	 * not on the href, so renaming a page does not quietly empty either.
	 */
	function items($access, $path) {
		$settings = self::settings($access);
		$out = array();

		foreach ($this->all() as $row) {
			if ($row->getHidden() || !self::sees($row, $access, $settings)) {
				continue;
			}

			$out[] = array(
				"title" => $row->getTitle(),
				"href" => self::link($row),
				"icon" => self::icon($row->getIcon()),
				"active" => self::active($row->getHref(), $path),
				"flows" => $row->getHref() === "/workflows",
				"pulse" => $row->getHref() === "/runs",
			);
		}

		return $out;
	}

	/**
	 * defaults is the top block as it stood before the table held it: the
	 * views across all folders, in this order, which is the order a fresh
	 * instance starts with and what the settings page draws while the table is
	 * empty.
	 *
	 * The rows, their links and their symbols are model/sidebar.go's
	 * SidebarDefaults, a constant table no binding carries; keep them in step
	 * with it. The positions are the order of this list, which the instance
	 * writes when it fills the table.
	 */
	static function defaults() {
		$views = array(
			array(
				"title" => "Dashboard",
				"href" => "/",
				"icon" => "grid",
				"rule" => "",
			),
			array(
				"title" => "Inventory",
				"href" => "/inventory",
				"icon" => "server",
				"rule" => "",
			),
			array(
				"title" => "Workflows",
				"href" => "/workflows",
				"icon" => "flow",
				"rule" => "",
			),
			array(
				"title" => "Runs",
				"href" => "/runs",
				"icon" => "list",
				"rule" => "",
			),
			array(
				"title" => "Artifacts",
				"href" => "/artifacts",
				"icon" => "box",
				"rule" => "",
			),
			array(
				"title" => "Scheduled",
				"href" => "/schedules",
				"icon" => "clock",
				"rule" => "",
			),
			array(
				"title" => "Notes",
				"href" => "/notes",
				"icon" => "note",
				"rule" => "",
			),
			array(
				"title" => "Settings",
				"href" => "/settings",
				"icon" => "settings",
				"rule" => self::SETTINGS,
			),
		);
		$out = array();

		foreach ($views as $view) {
			$view["kind"] = self::VIEW;
			$row = new Model\Sidebar($view);
			$out[] = $row;
		}

		return $out;
	}

	/**
	 * rows are the rows of the Sidebar settings table: every row the table
	 * holds in its order, the hidden ones too, because this is the page that
	 * unhides one.
	 *
	 * $modules is Task\Host::modules(), which a row of a module is paired with
	 * by its link: the description under the title is the module's own. A
	 * module whose row outlived its directory is paired with nothing, and the
	 * row says what kind it is instead.
	 */
	static function rows($items, $modules) {
		$found = array();

		foreach ($modules as $module) {
			$found["/m/" . Row::text($module, "name")] = $module;
		}

		$out = array();

		foreach ($items as $item) {
			$href = $item->getHref();
			$module = isset($found[$href]) ? $found[$href] : array();

			$out[] = array(
				"href" => $href,
				"link" => self::link($item),
				"icon" => self::icon($item->getIcon()),
				"title" => self::title($item),
				"kind" => $item->getKind(),
				"hidden" => $item->getHidden(),
				"note" => self::note($item, $module),
			);
		}

		return $out;
	}

	/**
	 * title is what a row of the settings table shows, falling back to the
	 * link so a row with no title is still something to drag.
	 */
	static function title($item) {
		$title = $item->getTitle();

		return $title === "" ? $item->getHref() : $title;
	}

	/**
	 * note is the line under the title: a module's description, or what the
	 * instance calls the row's kind where the module is gone.
	 */
	static function note($item, $module) {
		$said = Row::text($module, "description");

		if ($said !== "") {
			return $said;
		}

		return $item->getKind() === self::MODULE ? "a module" : "";
	}

	/**
	 * link is a row's whole link, the prefix in it where the prefix applies.
	 *
	 * A module's page is another application on the same instance, mounted at
	 * /m/<vendor>/<name> and not under this one, so the prefix is this
	 * application's own rows alone. Every other link a template writes is
	 * {base} and a path, because every other link is a page of this one.
	 */
	static function link($row) {
		if ($row->getKind() === self::MODULE) {
			return $row->getHref();
		}

		return Route::href($row->getHref());
	}

	/**
	 * active reports whether a row is the page being read: the dashboard on
	 * its own path alone, every other row on anything under it, so
	 * /settings/keys lights Settings and / does not light everything.
	 */
	static function active($href, $path) {
		if ($href === "/") {
			return $path === "/";
		}

		return $path === $href || strpos($path, $href . "/") === 0;
	}

	/**
	 * sees reports whether a user may see a row. An empty rule is a row
	 * everyone signed in sees, and SETTINGS is the one value that is not a
	 * permission path.
	 */
	static function sees($row, $access, $settings) {
		$rule = $row->getRule();

		if ($rule === "") {
			return true;
		}

		if ($rule === self::SETTINGS) {
			return $settings;
		}

		return $access !== null && $access->allowed($rule);
	}

	/** settings reports whether any step of the settings opens for a user. */
	static function settings($access) {
		if ($access === null) {
			return false;
		}

		foreach (array("workspace", "project", "key", "secret", "integration", "user") as $section) {
			if ($access->allowed($section . "/view")) {
				return true;
			}
		}

		return false;
	}

	/**
	 * icon keeps a name to the characters a sprite symbol is named with, so a
	 * value from the database cannot reach the markup as anything else. A
	 * name with nothing left falls back to the dot every list already uses.
	 */
	static function icon($icon) {
		$kept = preg_replace('/[^a-z0-9\-]/', "", strtolower(trim($icon)));

		return $kept === "" ? "box" : $kept;
	}
}
