<?php

namespace Task\Artifacts;

// Artifacts reads the artifact table: the files a run left behind, and the
// files of the artifact tree, which are the ones that hold a commit.
//
// The table is this module's. A run page lists the artifacts of one run,
// newest first. The bytes of a file are not here: they sit outside the
// directory this may open, and Task\Host::artifact() reads them.
class Artifacts {
	/** LIMIT is how many artifacts a run page lists. */
	const LIMIT = 200;

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** get is one artifact by ID, or null. */
	function get($id) {
		if ($id === "") {
			return null;
		}

		$row = $this->db->get("SELECT * FROM artifact WHERE id = ?", $id);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Artifact($row);
	}

	/**
	 * find is the artifacts a filter lists, newest first.
	 *
	 * The filter is an array and a field it leaves out narrows nothing: `run`,
	 * `folder`, `job` and `path`, where a path is a file of the artifact tree
	 * and the rows under it are its versions.
	 */
	function find($filter = array()) {
		$limit = isset($filter["limit"]) ? (int)$filter["limit"] : self::LIMIT;
		$where = array();
		$args = array();

		$columns = array(
			"run_id" => "run",
			"folder" => "folder",
			"job" => "job",
			"path" => "path",
		);

		foreach ($columns as $column => $field) {
			$value = isset($filter[$field]) ? (string)$filter[$field] : "";

			if ($value !== "") {
				$where[] = $column . " = ?";
				$args[] = $value;
			}
		}

		$sql = "SELECT * FROM artifact";

		if (count($where) > 0) {
			$sql .= " WHERE " . implode(" AND ", $where);
		}

		$sql .= " ORDER BY created_at DESC, id DESC";

		if ($limit > 0) {
			$sql .= " LIMIT ?";
			$args[] = $limit;
		}

		return self::rows(\Task\DB::all($this->db, $sql, $args));
	}

	/**
	 * latest is the newest version of every file of the artifact tree, by path.
	 *
	 * A file of the tree is one that holds a commit, and its rows are its
	 * versions; the newest is the one the tree holds, which is what says
	 * whether a reader may see an entry of it at all.
	 *
	 * The version is counted by rowid and not by created_at: two versions
	 * written in the same millisecond read as the same instant, and the order
	 * they were inserted in is the order they were written in.
	 */
	function latest() {
		$sql = "SELECT * FROM artifact WHERE id IN (
			SELECT id FROM (
				SELECT id, ROW_NUMBER() OVER (PARTITION BY path ORDER BY rowid DESC) AS version
				FROM artifact WHERE commit_id != ''
			) WHERE version = 1
		) ORDER BY path";

		return self::rows($this->db->get_all($sql));
	}

	/** rows turns what the database answered into artifacts. */
	static function rows($found) {
		$out = array();

		if (!is_array($found)) {
			return $out;
		}

		foreach ($found as $row) {
			$out[] = new \Task\Model\Artifact($row);
		}

		return $out;
	}
}
