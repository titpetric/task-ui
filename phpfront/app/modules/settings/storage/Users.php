<?php

namespace Task\Settings;

// Users reads the user table: one user by ID or by name, every user by name,
// how many there are, how many are admins, and the agents a reply may be
// written as.
//
// The table is this module's, which draws the Users page. What a user may do -
// the role, the folders, the permissions - is Task\User.
class Users {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** get is one user by ID, or null. */
	function get($id) {
		$row = $this->db->get("SELECT * FROM user WHERE id = ?", $id);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\User($row);
	}

	/** byName is one user by name, or null. */
	function byName($name) {
		$row = $this->db->get("SELECT * FROM user WHERE name = ?", $name);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\User($row);
	}

	/** all is every user, by name. */
	function all() {
		return \Task\Model\User::rows($this->db->get_all("SELECT * FROM user ORDER BY name"));
	}

	/** count is how many users there are. */
	function count() {
		$row = $this->db->get("SELECT COUNT(*) AS found FROM user");

		return \Task\Row::number($row, "found");
	}

	/** admins is how many users carry the admin flag. */
	function admins() {
		$row = $this->db->get("SELECT COUNT(*) AS found FROM user WHERE admin = 1");

		return \Task\Row::number($row, "found");
	}

	/** agents is every user an agent may reply as. */
	function agents() {
		return \Task\Model\User::rows($this->db->get_all("SELECT * FROM user WHERE agent_reply = 1 AND mcp = 1 ORDER BY name"));
	}
}
