<?php

namespace Task\Settings;

// Keys reads the key tables: the instance SSH keys, and the two tables that
// say who gets one, by the role a user acts with and by their name.
//
// The Keys page is this module's and the statements are here with it. The
// tables are the Go keystore's to write and migrate, under keystore/schema/,
// which is why this module's own schema/ creates none of them. No private
// material is here: the keystore holds it and it does not enter PHP.
class Keys {
	/** SSH is the purpose of a key a run gets, as keystore writes the column. */
	const SSH = "ssh";

	/** WORKSPACE starts the scope of a key of one workspace, which no run of another gets. */
	const WORKSPACE = "workspace:%";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** ssh is the name of every instance SSH key, by name; a workspace's own is no run's. */
	function ssh() {
		$out = array();

		foreach ($this->db->get_all("SELECT name FROM key_info WHERE purpose = ? AND scope NOT LIKE ? ORDER BY name", self::SSH, self::WORKSPACE) as $row) {
			$out[] = \Task\Row::text($row, "name");
		}

		return $out;
	}

	/** roles is the roles each key is granted to, by key name. */
	function roles() {
		return $this->links("SELECT key_name, role AS value FROM key_role");
	}

	/** holders is the users each key is granted to, by key name. */
	function holders() {
		return $this->links("SELECT key_name, user_name AS value FROM key_user");
	}

	/** links reads one of the two grant tables as the values of each key name. */
	private function links($statement) {
		$out = array();

		foreach ($this->db->get_all($statement) as $row) {
			$key = \Task\Row::text($row, "key_name");

			if (!isset($out[$key])) {
				$out[$key] = array();
			}

			$out[$key][] = \Task\Row::text($row, "value");
		}

		return $out;
	}
}
