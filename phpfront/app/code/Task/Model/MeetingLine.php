<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// MeetingLine is a row of the meeting_line table.
class MeetingLine {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "meeting_line";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = [
		"id",
		"meeting_id",
		"user_id",
		"user_name",
		"kind",
		"body",
		"start_ms",
		"end_ms",
		"created_at",
	];

	/** primary is the primary key columns. */
	public static $primary = ["id"];

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = ["created_at"];
	/** ID */
	var $id = "";
	/** Meeting ID */
	var $meeting_id = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** Kind */
	var $kind = "";
	/** Body */
	var $body = "";
	/** Start Ms */
	var $start_ms = 0;
	/** End Ms */
	var $end_ms = 0;
	/** Created At */
	var $created_at = "";

	function __construct($row = []) {
		$this->id = \Task\Row::text($row, "id");
		$this->meeting_id = \Task\Row::text($row, "meeting_id");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->kind = \Task\Row::text($row, "kind");
		$this->body = \Task\Row::text($row, "body");
		$this->start_ms = \Task\Row::number($row, "start_ms");
		$this->end_ms = \Task\Row::number($row, "end_ms");
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
	/** getMeetingId is Meeting ID. */
	function getMeetingId() {
		return $this->meeting_id;
	}

	/** setMeetingId sets Meeting ID. */
	function setMeetingId($value) {
		$this->meeting_id = $value;

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
	/** getKind is Kind. */
	function getKind() {
		return $this->kind;
	}

	/** setKind sets Kind. */
	function setKind($value) {
		$this->kind = $value;

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
	/** getStartMs is Start Ms. */
	function getStartMs() {
		return $this->start_ms;
	}

	/** setStartMs sets Start Ms. */
	function setStartMs($value) {
		$this->start_ms = $value;

		return $this;
	}
	/** getEndMs is End Ms. */
	function getEndMs() {
		return $this->end_ms;
	}

	/** setEndMs sets End Ms. */
	function setEndMs($value) {
		$this->end_ms = $value;

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
		$out = [];

		if (!is_array($found)) {
			return $out;
		}

		foreach ($found as $row) {
			$out[] = new MeetingLine($row);
		}

		return $out;
	}
}
