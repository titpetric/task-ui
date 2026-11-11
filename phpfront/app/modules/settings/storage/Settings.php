<?php

namespace Task\Settings;

// Settings reads the setting table: the instance's own values, by name.
//
// The table is this module's. What a value means - the retention rules, the
// sidebar's limits - is read off it by the class that owns that question.
class Settings {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/**
	 * values is the table by name.
	 *
	 * The whole table is read in one statement, as the cleanup reads it: the
	 * rows are a handful and the names of two of them per kind.
	 */
	function values() {
		$out = array();

		foreach ($this->db->get_all("SELECT name, value FROM setting") as $row) {
			$out[\Task\Row::text($row, "name")] = \Task\Row::text($row, "value");
		}

		return $out;
	}
}
