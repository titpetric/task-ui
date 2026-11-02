<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Setting is a row of the setting table.
class Setting {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "setting";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("name", "value", "updated_at");

	/** primary is the primary key columns. */
	public static $primary = array("name");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("updated_at");
	/** Name */
	var $name = "";
	/** Value */
	var $value = "";
	/** Updated At */
	var $updated_at = "";

	function __construct($row = array()) {
		$this->name = \Task\Row::text($row, "name");
		$this->value = \Task\Row::text($row, "value");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
	}
	/** getName is Name. */
	function getName() {
		return $this->name;
	}

	/** setName sets Name. */
	function setName($value) {
		$this->name = $value;

		return $this;
	}
	/** getValue is Value. */
	function getValue() {
		return $this->value;
	}

	/** setValue sets Value. */
	function setValue($value) {
		$this->value = $value;

		return $this;
	}
	/** getUpdatedAt is Updated At. */
	function getUpdatedAt() {
		return $this->updated_at;
	}

	/** setUpdatedAt sets Updated At. */
	function setUpdatedAt($value) {
		$this->updated_at = $value;

		return $this;
	}

	/** values is this row as the database takes it, keyed by column. */
	function values() {
		return \Task\Row::values($this, self::$stamps);
	}

	/** rows turns what the database answered into models. */
	static function rows($found) {
		$out = array();

		if (!is_array($found)) {
			return $out;
		}

		foreach ($found as $row) {
			$out[] = new Setting($row);
		}

		return $out;
	}
}
