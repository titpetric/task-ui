<?php

namespace Task\Folders;

// Pins reads the pin table: what a reader asked the sidebar to list besides
// the folders it lists anyway.
//
// The table is this module's. A pin names a folder, a directory inside one, or
// a job of either with the arguments it runs with; what the sidebar does with
// one is Task\Folders, and what a page draws of one is Task\Pin.
class Pins {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** ofUser is every pin of one reader, in the order they were made. */
	function ofUser($name) {
		if ($name === "") {
			return array();
		}

		$rows = $this->db->get_all("SELECT * FROM pin WHERE user_name = ? ORDER BY rowid", $name);

		return \Task\Model\Pin::rows($rows);
	}
}
