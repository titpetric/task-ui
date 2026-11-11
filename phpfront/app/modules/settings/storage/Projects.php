<?php

namespace Task\Settings;

// Projects reads the project table: which projects the sidebar hides, the
// folder of every project row, and one row by folder.
//
// The table is this module's, which draws the Projects page. What a project
// means to a page - its workspace, its crumbs, what a move of it leaves
// behind - is Task\Project.
class Projects {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** hidden is the folder of every project set not to show in the sidebar. */
	function hidden() {
		$out = array();

		foreach ($this->db->get_all("SELECT folder FROM project WHERE sidebar_hidden = 1") as $row) {
			$out[] = \Task\Row::text($row, "folder");
		}

		return $out;
	}

	/**
	 * folders is the folder of every project row, which the workspaces count
	 * beside the directories the host walked.
	 *
	 * A row whose directory is not there yet is a project of its workspace all
	 * the same: the clone failed and the page lists it with git's output.
	 */
	function folders() {
		$out = array();

		foreach ($this->db->get_all("SELECT folder FROM project ORDER BY folder") as $row) {
			$out[] = \Task\Row::text($row, "folder");
		}

		return $out;
	}

	/**
	 * row is the project a folder names, null where no row does.
	 *
	 * A directory somebody made by hand has no row, and the page of it draws
	 * the settings of a project the instance keeps nothing about.
	 */
	function row($folder) {
		$row = $this->db->get("SELECT * FROM project WHERE folder = ?", $folder);

		return is_array($row) ? $row : null;
	}
}
