<?php

namespace Task\Settings;

// Refs reads the rows of this module's tables that name a folder: a variable
// and an env file, which a move of the project leaves where they were.
//
// Each owning module answers for its own tables, and Task\Project asks all of
// them. An ID may hold the underscore LIKE matches any character with, so the
// match is on the ID itself or on the ID and a slash as a prefix.
// storage/folder_ref_storage.go.
class Refs {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** of is the variables and the env files a folder and its children name. */
	function of($id) {
		$prefix = $id . "/";
		$length = strlen($prefix);
		$sql = "SELECT 'secret' AS kind, scope AS folder, name FROM secret
	WHERE scope = ? OR substr(scope, 1, ?) = ?
UNION ALL
SELECT 'env_file', folder, path FROM env_file
	WHERE folder = ? OR substr(folder, 1, ?) = ?
ORDER BY kind, folder, name";

		return $this->db->get_all($sql, $id, $length, $prefix, $id, $length, $prefix);
	}
}
