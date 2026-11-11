<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// RefinementArtifact is a row of the refinement_artifact table.
class RefinementArtifact {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "refinement_artifact";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "thread_id", "comment_id", "path", "kind", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Thread ID */
	var $thread_id = "";
	/** Comment ID */
	var $comment_id = "";
	/** Path */
	var $path = "";
	/** Kind */
	var $kind = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->thread_id = \Task\Row::text($row, "thread_id");
		$this->comment_id = \Task\Row::text($row, "comment_id");
		$this->path = \Task\Row::text($row, "path");
		$this->kind = \Task\Row::text($row, "kind");
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
	/** getThreadId is Thread ID. */
	function getThreadId() {
		return $this->thread_id;
	}

	/** setThreadId sets Thread ID. */
	function setThreadId($value) {
		$this->thread_id = $value;

		return $this;
	}
	/** getCommentId is Comment ID. */
	function getCommentId() {
		return $this->comment_id;
	}

	/** setCommentId sets Comment ID. */
	function setCommentId($value) {
		$this->comment_id = $value;

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
	/** getKind is Kind. */
	function getKind() {
		return $this->kind;
	}

	/** setKind sets Kind. */
	function setKind($value) {
		$this->kind = $value;

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
			$out[] = new RefinementArtifact($row);
		}

		return $out;
	}
}
