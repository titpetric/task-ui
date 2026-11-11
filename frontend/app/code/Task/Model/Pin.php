<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Pin is a row of the pin table.
class Pin {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "pin";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "user_name", "folder", "job", "dir", "args", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** User Name */
	var $user_name = "";
	/** Folder */
	var $folder = "";
	/** Job */
	var $job = "";
	/** Dir */
	var $dir = "";
	/** Args */
	var $args = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->folder = \Task\Row::text($row, "folder");
		$this->job = \Task\Row::text($row, "job");
		$this->dir = \Task\Row::text($row, "dir");
		$this->args = \Task\Row::text($row, "args");
		$this->created_at = \Task\Row::stamp($row, "created_at");
	}
	/** getId is ID. */
	function getId() {
		return $this->id;
	}

	/** setId sets ID. */
	function setId($value) {
		$this->id = $value;

		return $this;
	}
	/** getUserName is User Name. */
	function getUserName() {
		return $this->user_name;
	}

	/** setUserName sets User Name. */
	function setUserName($value) {
		$this->user_name = $value;

		return $this;
	}
	/** getFolder is Folder. */
	function getFolder() {
		return $this->folder;
	}

	/** setFolder sets Folder. */
	function setFolder($value) {
		$this->folder = $value;

		return $this;
	}
	/** getJob is Job. */
	function getJob() {
		return $this->job;
	}

	/** setJob sets Job. */
	function setJob($value) {
		$this->job = $value;

		return $this;
	}
	/** getDir is Dir. */
	function getDir() {
		return $this->dir;
	}

	/** setDir sets Dir. */
	function setDir($value) {
		$this->dir = $value;

		return $this;
	}
	/** getArgs is Args. */
	function getArgs() {
		return $this->args;
	}

	/** setArgs sets Args. */
	function setArgs($value) {
		$this->args = $value;

		return $this;
	}
	/** getCreatedAt is Created At. */
	function getCreatedAt() {
		return $this->created_at;
	}

	/** setCreatedAt sets Created At. */
	function setCreatedAt($value) {
		$this->created_at = $value;

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
			$out[] = new Pin($row);
		}

		return $out;
	}
}
