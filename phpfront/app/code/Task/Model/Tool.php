<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Tool is a row of the tool table.
class Tool {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "tool";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("name", "category", "version", "path", "description", "updated_at");

	/** primary is the primary key columns. */
	public static $primary = array("name");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("updated_at");
	/** Name */
	var $name = "";
	/** Category */
	var $category = "";
	/** Version */
	var $version = "";
	/** Path */
	var $path = "";
	/** Description */
	var $description = "";
	/** Updated At */
	var $updated_at = "";

	function __construct($row = array()) {
		$this->name = \Task\Row::text($row, "name");
		$this->category = \Task\Row::text($row, "category");
		$this->version = \Task\Row::text($row, "version");
		$this->path = \Task\Row::text($row, "path");
		$this->description = \Task\Row::text($row, "description");
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
	/** getCategory is Category. */
	function getCategory() {
		return $this->category;
	}

	/** setCategory sets Category. */
	function setCategory($value) {
		$this->category = $value;

		return $this;
	}
	/** getVersion is Version. */
	function getVersion() {
		return $this->version;
	}

	/** setVersion sets Version. */
	function setVersion($value) {
		$this->version = $value;

		return $this;
	}
	/** getPath is Path. */
	function getPath() {
		return $this->path;
	}

	/** setPath sets Path. */
	function setPath($value) {
		$this->path = $value;

		return $this;
	}
	/** getDescription is Description. */
	function getDescription() {
		return $this->description;
	}

	/** setDescription sets Description. */
	function setDescription($value) {
		$this->description = $value;

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
			$out[] = new Tool($row);
		}

		return $out;
	}
}
