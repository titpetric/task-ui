<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Migrations is a row of the migrations table.
class Migrations {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "migrations";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("project", "filename", "statement_index", "status");

	/** primary is the primary key columns. */
	public static $primary = array("project", "filename");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array();
	/** Project */
	var $project = "";
	/** Filename */
	var $filename = "";
	/** Statement Index */
	var $statement_index = 0;
	/** Status */
	var $status = "";

	function __construct($row = array()) {
		$this->project = \Task\Row::text($row, "project");
		$this->filename = \Task\Row::text($row, "filename");
		$this->statement_index = \Task\Row::number($row, "statement_index");
		$this->status = \Task\Row::text($row, "status");
	}
	/** getProject is Project. */
	function getProject() {
		return $this->project;
	}

	/** setProject sets Project. */
	function setProject($value) {
		$this->project = $value;

		return $this;
	}
	/** getFilename is Filename. */
	function getFilename() {
		return $this->filename;
	}

	/** setFilename sets Filename. */
	function setFilename($value) {
		$this->filename = $value;

		return $this;
	}
	/** getStatementIndex is Statement Index. */
	function getStatementIndex() {
		return $this->statement_index;
	}

	/** setStatementIndex sets Statement Index. */
	function setStatementIndex($value) {
		$this->statement_index = $value;

		return $this;
	}
	/** getStatus is Status. */
	function getStatus() {
		return $this->status;
	}

	/** setStatus sets Status. */
	function setStatus($value) {
		$this->status = $value;

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
			$out[] = new Migrations($row);
		}

		return $out;
	}
}
