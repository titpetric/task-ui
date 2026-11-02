<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// RunSummary is a row of the run_summary table.
class RunSummary {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "run_summary";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("run_id", "summary", "user_name", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("run_id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** Run ID */
	var $run_id = "";
	/** Summary */
	var $summary = "";
	/** User Name */
	var $user_name = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->run_id = \Task\Row::text($row, "run_id");
		$this->summary = \Task\Row::text($row, "summary");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->created_at = \Task\Row::stamp($row, "created_at");
	}
	/** getRunId is Run ID. */
	function getRunId() {
		return $this->run_id;
	}

	/** setRunId sets Run ID. */
	function setRunId($value) {
		$this->run_id = $value;

		return $this;
	}
	/** getSummary is Summary. */
	function getSummary() {
		return $this->summary;
	}

	/** setSummary sets Summary. */
	function setSummary($value) {
		$this->summary = $value;

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
			$out[] = new RunSummary($row);
		}

		return $out;
	}
}
