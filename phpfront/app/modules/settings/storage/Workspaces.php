<?php

namespace Task\Settings;

// Workspaces reads the workspace table: the folder of every workspace, the
// rows themselves, each title, and the sidebar limit of the ones that set one.
//
// The table is this module's, which draws the Workspaces page. What a
// workspace means to a page - where its home is, which projects are inside it
// - is Task\Workspace.
class Workspaces {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** folders is the folder of every named workspace, in the sidebar's order. */
	function folders() {
		$out = array();

		foreach ($this->db->get_all("SELECT folder FROM workspace WHERE folder != ? ORDER BY sort, folder", \Task\Route::INTERNAL) as $row) {
			$out[] = \Task\Row::text($row, "folder");
		}

		return $out;
	}

	/**
	 * names is the folder of every workspace row, the default one first and
	 * the named ones in the sidebar's order after it.
	 */
	function names() {
		$out = array();

		foreach ($this->db->get_all("SELECT folder FROM workspace ORDER BY CASE WHEN folder = ? THEN 0 ELSE 1 END, sort, folder", \Task\Route::INTERNAL) as $row) {
			$out[] = \Task\Row::text($row, "folder");
		}

		return $out;
	}

	/**
	 * all is every workspace row, the default one first and the named ones in
	 * the sidebar's order after it, which is storage/workspace_storage.go:69.
	 */
	function all() {
		return \Task\Model\Workspace::rows($this->db->get_all("SELECT * FROM workspace ORDER BY CASE WHEN folder = ? THEN 0 ELSE 1 END, sort, folder", \Task\Route::INTERNAL));
	}

	/** row is one workspace row by folder, or null. */
	function row($folder) {
		$row = $this->db->get("SELECT * FROM workspace WHERE folder = ?", $folder);

		return is_array($row) ? $row : null;
	}

	/** titles is the title of every workspace, by folder. */
	function titles() {
		$out = array();

		foreach ($this->db->get_all("SELECT folder, title FROM workspace") as $row) {
			$out[\Task\Row::text($row, "folder")] = \Task\Row::text($row, "title");
		}

		return $out;
	}

	/**
	 * limits is how many projects the sidebar lists for a workspace that
	 * limits them, by folder; a workspace that lists every one is not here.
	 */
	function limits() {
		$out = array();

		foreach ($this->db->get_all("SELECT folder, sidebar_limit FROM workspace WHERE sidebar_limit > 0") as $row) {
			$out[\Task\Row::text($row, "folder")] = \Task\Row::number($row, "sidebar_limit");
		}

		return $out;
	}
}
