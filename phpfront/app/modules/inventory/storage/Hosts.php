<?php

namespace Task\Inventory;

// Hosts reads and writes the host table: the machines of the inventory, the
// ones a person entered and the ones the Beszel hub reported.
//
// The table is this module's and no other page reads it, so every statement
// over it lives here. The address is the key: a machine goes in once, under
// the address ssh signs in to.
//
// Nothing here removes a row on its own. A host the hub stops reporting, one
// that is down and one somebody rebuilt all stay until a person removes them.
class Hosts {
	/** MANUAL is the source of a host a person entered. */
	const MANUAL = "manual";

	/** BESZEL is the source of a host the hub reported. */
	const BESZEL = "beszel";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** all is every host, by name. */
	function all() {
		$rows = $this->db->get_all("SELECT * FROM host ORDER BY name");
		$out = array();

		if (!is_array($rows)) {
			return $out;
		}

		foreach ($rows as $row) {
			$out[] = new \Task\Model\Host($row);
		}

		return $out;
	}

	/** find is the host of an address, or false where no row holds it. */
	function find($address) {
		return $this->db->get("SELECT * FROM host WHERE address = ?", $address);
	}

	/**
	 * add writes a host and returns its ID, or "" for an address the
	 * inventory holds. The name is the address where none is given.
	 */
	function add($name, $address, $source, $login, $by) {
		if (is_array($this->find($address))) {
			return "";
		}

		$id = \Task\Id::make();

		$this->db->insert("host", array(
			"id" => $id,
			"name" => $name === "" ? $address : $name,
			"address" => $address,
			"source" => $source,
			"login" => $login,
			"added_by" => $by,
			"created_at" => \Task\Clock::stamp(),
		));

		return $id;
	}

	/** drop takes a host out by ID. */
	function drop($id) {
		$this->db->query("DELETE FROM host WHERE id = ?", $id);
	}

	/**
	 * seen writes a host the hub reports under an address no row holds, and
	 * returns its ID or "". A row already there keeps the name it has.
	 */
	function seen($name, $address) {
		return $this->add($name, $address, self::BESZEL, "", "");
	}
}
