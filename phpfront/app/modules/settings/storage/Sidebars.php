<?php

namespace Task\Settings;

// Sidebars reads and writes the sidebar table: the top block of the sidebar,
// in its order, and the rows the Sidebar settings page edits.
//
// The table is this module's. What a row draws and who sees it is
// Task\Sidebar.
class Sidebars {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** all is every row, hidden ones included, in the order the table holds. */
	function all() {
		return \Task\Model\Sidebar::rows($this->db->get_all("SELECT * FROM sidebar ORDER BY position, title"));
	}

	/** byHref is one row by its link, or null. */
	function byHref($href) {
		$row = $this->db->get("SELECT * FROM sidebar WHERE href = ?", $href);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Sidebar($row);
	}

	/** ensure writes a row for each link the table does not hold. */
	function ensure($rows) {
		$now = \Task\Clock::stamp();

		foreach ($rows as $row) {
			$this->db->query("INSERT OR IGNORE INTO sidebar (id, kind, href, title, icon, position, hidden, rule, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", ulid(), $row["kind"], $row["href"], $row["title"], $row["icon"], $row["position"], $row["hidden"] ? 1 : 0, $row["rule"], $now, $now);
		}
	}

	/** order writes the position and the hidden flag of each link, in one transaction. */
	function order($hrefs, $hidden) {
		$now = \Task\Clock::stamp();
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
}
