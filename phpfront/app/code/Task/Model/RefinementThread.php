<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// RefinementThread is a row of the refinement_thread table.
class RefinementThread {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "refinement_thread";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "title", "user", "report_markdown", "created_at", "updated_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Title */
	var $title = "";
	/** User */
	var $user = "";
	/** Report Markdown */
	var $report_markdown = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->title = \Task\Row::text($row, "title");
		$this->user = \Task\Row::text($row, "user");
		$this->report_markdown = \Task\Row::text($row, "report_markdown");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
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
	/** getTitle is Title. */
	function getTitle() {
		return $this->title;
	}

	/** setTitle sets Title. */
	function setTitle($value) {
		$this->title = $value;

		return $this;
	}
	/** getUser is User. */
	function getUser() {
		return $this->user;
	}

	/** setUser sets User. */
	function setUser($value) {
		$this->user = $value;

		return $this;
	}
	/** getReportMarkdown is Report Markdown. */
	function getReportMarkdown() {
		return $this->report_markdown;
	}

	/** setReportMarkdown sets Report Markdown. */
	function setReportMarkdown($value) {
		$this->report_markdown = $value;

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
			$out[] = new RefinementThread($row);
		}

		return $out;
	}
}
