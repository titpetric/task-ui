<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// MeetingParticipant is a row of the meeting_participant table.
class MeetingParticipant {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "meeting_participant";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = ["meeting_id", "user_id", "user_name", "joined_at", "last_seen_at", "left_at", "joins"];

	/** primary is the primary key columns. */
	public static $primary = ["meeting_id", "user_id"];

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = ["joined_at", "last_seen_at", "left_at"];
	/** Meeting ID */
	var $meeting_id = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** Joined At */
	var $joined_at = "";
	/** Last Seen At */
	var $last_seen_at = "";
	/** Left At */
	var $left_at = "";
	/** Joins */
	var $joins = 0;

	function __construct($row = []) {
		$this->meeting_id = \Task\Row::text($row, "meeting_id");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->joined_at = \Task\Row::stamp($row, "joined_at");
		$this->last_seen_at = \Task\Row::stamp($row, "last_seen_at");
		$this->left_at = \Task\Row::stamp($row, "left_at");
		$this->joins = \Task\Row::number($row, "joins");
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
	/** getJoinedAt is Joined At. */
	function getJoinedAt() {
		return $this->joined_at;
	}

	/** setJoinedAt sets Joined At. */
	function setJoinedAt($value) {
		$this->joined_at = $value;

		return $this;
	}
	/** getLastSeenAt is Last Seen At. */
	function getLastSeenAt() {
		return $this->last_seen_at;
	}

	/** setLastSeenAt sets Last Seen At. */
	function setLastSeenAt($value) {
		$this->last_seen_at = $value;

		return $this;
	}
	/** getLeftAt is Left At. */
	function getLeftAt() {
		return $this->left_at;
	}

	/** setLeftAt sets Left At. */
	function setLeftAt($value) {
		$this->left_at = $value;

		return $this;
	}
	/** getJoins is Joins. */
	function getJoins() {
		return $this->joins;
	}

	/** setJoins sets Joins. */
	function setJoins($value) {
		$this->joins = $value;

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
			$out[] = new MeetingParticipant($row);
		}

		return $out;
	}
}
