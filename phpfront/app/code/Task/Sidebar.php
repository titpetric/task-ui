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

		return is_array($row) ? new Model\Sidebar($row) : null;
	}

	/** order writes the position and the hidden flag of each link, in one transaction. */
	function order($hrefs, $hidden) {
		$now = Clock::stamp();

		$this->db->begin();

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
				"href" => $row->getHref(),
				"icon" => self::icon($row->getIcon()),
				"active" => self::active($row->getHref(), $path),
				"flows" => $row->getHref() === "/workflows",
				"pulse" => $row->getHref() === "/runs",
			);
		}

		return $out;
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
