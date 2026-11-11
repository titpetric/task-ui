<?php

namespace Task\Settings;

// Tools reads the tool table: what the last check recorded of the runtimes and
// the tools the image carries.
//
// The table is this module's, which draws the Instance page. What a row means
// and which category it belongs to is Task\Toolset.
class Tools {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** all is every tool the last check recorded, by category and then by name. */
	function all() {
		return \Task\Model\Tool::rows($this->db->get_all("SELECT * FROM tool ORDER BY category, name"));
	}
}
