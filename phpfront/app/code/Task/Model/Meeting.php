<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Meeting is a row of the meeting table.
class Meeting {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "meeting";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = [
		"id",
		"title",
		"user_id",
		"user_name",
		"state",
		"started_at",
		"ended_at",
		"limit_seconds",
		"empty_since",
		"summary",
		"summary_state",
		"summary_message",
		"summary_user_name",
		"created_at",
		"updated_at",
		"agent",
		"agent_id",
		"agent_note_id",
	];

	/** primary is the primary key columns. */
	public static $primary = ["id"];

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = ["started_at", "ended_at", "empty_since", "created_at", "updated_at"];
	/** ID */
	var $id = "";
	/** Title */
	var $title = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** State */
	var $state = "";
	/** Started At */
	var $started_at = "";
	/** Ended At */
	var $ended_at = "";
	/** Limit Seconds */
	var $limit_seconds = 0;
	/** Empty Since */
	var $empty_since = "";
	/** Summary */
	var $summary = "";
	/** Summary State */
	var $summary_state = "";
	/** Summary Message */
	var $summary_message = "";
	/** Summary User Name */
	var $summary_user_name = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Agent */
	var $agent = "";
	/** Agent ID */
	var $agent_id = "";
	/** Agent Note ID */
	var $agent_note_id = "";

	function __construct($row = []) {
		$this->id = \Task\Row::text($row, "id");
		$this->title = \Task\Row::text($row, "title");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->state = \Task\Row::text($row, "state");
		$this->started_at = \Task\Row::stamp($row, "started_at");
		$this->ended_at = \Task\Row::stamp($row, "ended_at");
		$this->limit_seconds = \Task\Row::number($row, "limit_seconds");
		$this->empty_since = \Task\Row::stamp($row, "empty_since");
		$this->summary = \Task\Row::text($row, "summary");
		$this->summary_state = \Task\Row::text($row, "summary_state");
		$this->summary_message = \Task\Row::text($row, "summary_message");
		$this->summary_user_name = \Task\Row::text($row, "summary_user_name");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->agent = \Task\Row::text($row, "agent");
		$this->agent_id = \Task\Row::text($row, "agent_id");
		$this->agent_note_id = \Task\Row::text($row, "agent_note_id");
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
	/** getState is State. */
	function getState() {
		return $this->state;
	}

	/** setState sets State. */
	function setState($value) {
		$this->state = $value;

		return $this;
	}
	/** getStartedAt is Started At. */
	function getStartedAt() {
		return $this->started_at;
	}

	/** setStartedAt sets Started At. */
	function setStartedAt($value) {
		$this->started_at = $value;

		return $this;
	}
	/** getEndedAt is Ended At. */
	function getEndedAt() {
		return $this->ended_at;
	}

	/** setEndedAt sets Ended At. */
	function setEndedAt($value) {
		$this->ended_at = $value;

		return $this;
	}
	/** getLimitSeconds is Limit Seconds. */
	function getLimitSeconds() {
		return $this->limit_seconds;
	}

	/** setLimitSeconds sets Limit Seconds. */
	function setLimitSeconds($value) {
		$this->limit_seconds = $value;

		return $this;
	}
	/** getEmptySince is Empty Since. */
	function getEmptySince() {
		return $this->empty_since;
	}

	/** setEmptySince sets Empty Since. */
	function setEmptySince($value) {
		$this->empty_since = $value;

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
	/** getSummaryState is Summary State. */
	function getSummaryState() {
		return $this->summary_state;
	}

	/** setSummaryState sets Summary State. */
	function setSummaryState($value) {
		$this->summary_state = $value;

		return $this;
	}
	/** getSummaryMessage is Summary Message. */
	function getSummaryMessage() {
		return $this->summary_message;
	}

	/** setSummaryMessage sets Summary Message. */
	function setSummaryMessage($value) {
		$this->summary_message = $value;

		return $this;
	}
	/** getSummaryUserName is Summary User Name. */
	function getSummaryUserName() {
		return $this->summary_user_name;
	}

	/** setSummaryUserName sets Summary User Name. */
	function setSummaryUserName($value) {
		$this->summary_user_name = $value;

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
	/** getAgent is Agent. */
	function getAgent() {
		return $this->agent;
	}

	/** setAgent sets Agent. */
	function setAgent($value) {
		$this->agent = $value;

		return $this;
	}
	/** getAgentId is Agent ID. */
	function getAgentId() {
		return $this->agent_id;
	}

	/** setAgentId sets Agent ID. */
	function setAgentId($value) {
		$this->agent_id = $value;

		return $this;
	}
	/** getAgentNoteId is Agent Note ID. */
	function getAgentNoteId() {
		return $this->agent_note_id;
	}

	/** setAgentNoteId sets Agent Note ID. */
	function setAgentNoteId($value) {
		$this->agent_note_id = $value;

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
			$out[] = new Meeting($row);
		}

		return $out;
	}
}
