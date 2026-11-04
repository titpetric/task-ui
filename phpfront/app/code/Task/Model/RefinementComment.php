<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// RefinementComment is a row of the refinement_comment table.
class RefinementComment {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "refinement_comment";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "thread_id", "user", "body", "note_id", "kind", "collapsed_into", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Thread ID */
	var $thread_id = "";
	/** User */
	var $user = "";
	/** Body */
	var $body = "";
	/** Note ID */
	var $note_id = "";
	/** Kind */
	var $kind = "";
	/** Collapsed Into */
	var $collapsed_into = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->thread_id = \Task\Row::text($row, "thread_id");
		$this->user = \Task\Row::text($row, "user");
		$this->body = \Task\Row::text($row, "body");
		$this->note_id = \Task\Row::text($row, "note_id");
		$this->kind = \Task\Row::text($row, "kind");
		$this->collapsed_into = \Task\Row::text($row, "collapsed_into");
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
	/** getUser is User. */
	function getUser() {
		return $this->user;
	}

	/** setUser sets User. */
	function setUser($value) {
		$this->user = $value;

		return $this;
	}
	/** getBody is Body. */
	function getBody() {
		return $this->body;
	}

	/** setBody sets Body. */
	function setBody($value) {
		$this->body = $value;

		return $this;
	}
	/** getNoteId is Note ID. */
	function getNoteId() {
		return $this->note_id;
	}

	/** setNoteId sets Note ID. */
	function setNoteId($value) {
		$this->note_id = $value;

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
	/** getCollapsedInto is Collapsed Into. */
	function getCollapsedInto() {
		return $this->collapsed_into;
	}

	/** setCollapsedInto sets Collapsed Into. */
	function setCollapsedInto($value) {
		$this->collapsed_into = $value;

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
			$out[] = new RefinementComment($row);
		}

		return $out;
	}
}
