<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// NoteParticipant is a row of the note_participant table.
class NoteParticipant {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "note_participant";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("note_id", "user_id", "user_name", "last_read_at", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("note_id", "user_id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("last_read_at", "created_at");
	/** Note ID */
	var $note_id = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** Last Read At */
	var $last_read_at = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->note_id = \Task\Row::text($row, "note_id");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->last_read_at = \Task\Row::stamp($row, "last_read_at");
		$this->created_at = \Task\Row::stamp($row, "created_at");
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
	/** getUserId is User ID. */
	function getUserId() {
		return $this->user_id;
	}

	/** setUserId sets User ID. */
	function setUserId($value) {
		$this->user_id = $value;

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
	/** getLastReadAt is Last Read At. */
	function getLastReadAt() {
		return $this->last_read_at;
	}

	/** setLastReadAt sets Last Read At. */
	function setLastReadAt($value) {
		$this->last_read_at = $value;

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
			$out[] = new NoteParticipant($row);
		}

		return $out;
	}
}
