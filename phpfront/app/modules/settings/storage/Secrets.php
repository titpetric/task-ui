<?php

namespace Task\Settings;

// Secrets reads the secret and secret_read tables: the environment variables
// and secrets a workspace's runs are given, and which run got which.
//
// The tables are this module's, which draws the Secrets page. No value is
// here: a secret's value is encrypted with the instance's key and the key does
// not enter PHP, so nothing answers more than what a page counts or lists.
// What a row means to a page is Task\Secret.
class Secrets {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** count is how many variables the settings list, the integrations' apart. */
	function count() {
		$row = $this->db->get("SELECT COUNT(*) AS found FROM secret WHERE scope NOT LIKE ? AND scope NOT LIKE ?", \Task\Secret::OWNED, \Task\Secret::PERSONAL);

		return \Task\Row::number($row, "found");
	}

	/**
	 * scopes is the scope of every variable the settings list, one entry per
	 * variable, which the keys page counts by workspace.
	 */
	function scopes() {
		$out = array();

		foreach ($this->db->get_all("SELECT scope FROM secret WHERE scope NOT LIKE ? AND scope NOT LIKE ?", \Task\Secret::OWNED, \Task\Secret::PERSONAL) as $row) {
			$out[] = \Task\Row::text($row, "scope");
		}

		return $out;
	}

	/**
	 * all is every variable the settings list, in the order the page draws
	 * them: those of every workspace first, then each workspace's, then a
	 * folder's, and by name inside each. storage/secret_storage.go:81.
	 */
	function all() {
		$sql = "SELECT * FROM secret WHERE scope NOT LIKE ? AND scope NOT LIKE ? ORDER BY CASE WHEN scope = ? THEN 0 WHEN scope LIKE ? THEN 1 ELSE 2 END, scope, name";

		return \Task\Model\Secret::rows($this->db->get_all($sql, \Task\Secret::OWNED, \Task\Secret::PERSONAL, \Task\Workspace::EVERY, \Task\Workspace::SCOPE . "%"));
	}

	/** counts is how many runs got each variable, by its ID. */
	function counts() {
		$out = array();

		foreach ($this->db->get_all("SELECT secret_id, COUNT(*) AS found FROM secret_read GROUP BY secret_id") as $row) {
			$out[\Task\Row::text($row, "secret_id")] = \Task\Row::number($row, "found");
		}

		return $out;
	}

	/**
	 * last is the newest read of each variable, by its ID.
	 *
	 * The newest is the greatest rowid and not the greatest timestamp, for the
	 * reason reads() gives: the reads of one run share a millisecond.
	 */
	function last() {
		$sql = "SELECT * FROM secret_read WHERE rowid IN (SELECT MAX(rowid) FROM secret_read GROUP BY secret_id)";
		$out = array();

		foreach (\Task\Model\SecretRead::rows($this->db->get_all($sql)) as $read) {
			$out[$read->getSecretId()] = $read;
		}

		return $out;
	}

	/**
	 * byId is one variable the settings list, or null.
	 *
	 * An integration's own setting is no variable of the settings, so an ID
	 * that names one answers as an ID that names nothing.
	 */
	function byId($id) {
		$row = $this->db->get("SELECT * FROM secret WHERE id = ? AND scope NOT LIKE ? AND scope NOT LIKE ?", $id, \Task\Secret::OWNED, \Task\Secret::PERSONAL);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Secret($row);
	}

	/**
	 * edited is the variable a form was opened on: the row an ID names, or the
	 * row a scope and a name name where no ID was given, null where neither
	 * names one. frontend/secret.go:127.
	 *
	 * The integration scope is not left out here, which byId() leaves out: the
	 * form opens whatever row it was pointed at, and a service's own setting
	 * has a form like any other variable.
	 */
	function edited($id, $scope, $name) {
		if ($id !== "") {
			$row = $this->db->get("SELECT * FROM secret WHERE id = ?", $id);
		} else {
			$row = $this->db->get("SELECT * FROM secret WHERE scope = ? AND name = ?", $scope, $name);
		}

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Secret($row);
	}

	/**
	 * reads is a page of the runs that got a variable, newest first, or oldest
	 * first for a page read after a cursor, which is reversed here.
	 *
	 * The order is the row's and not the timestamp's: the reads of one run are
	 * written in one transaction and share a millisecond, so a timestamp
	 * orders none of them and rowid orders all of them. The cursor is a read's
	 * ID, and a cursor whose row is gone keeps nothing.
	 */
	function reads($id, $before, $after, $limit) {
		$sql = "SELECT * FROM secret_read WHERE secret_id = ?";
		$args = array($id);
		$order = " ORDER BY rowid DESC";

		if ($before !== "") {
			$sql .= " AND rowid < (SELECT rowid FROM secret_read WHERE id = ?)";
			$args[] = $before;
		} elseif ($after !== "") {
			$sql .= " AND rowid > (SELECT rowid FROM secret_read WHERE id = ?)";
			$args[] = $after;
			$order = " ORDER BY rowid ASC";
		}

		$sql .= $order;

		if ($limit > 0) {
			$sql .= " LIMIT ?";
			$args[] = $limit;
		}

		$out = \Task\Model\SecretRead::rows(\Task\DB::all($this->db, $sql, $args));

		if ($before === "" && $after !== "") {
			$out = array_reverse($out);
		}

		return $out;
	}
}
