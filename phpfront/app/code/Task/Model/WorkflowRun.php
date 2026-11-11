<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// WorkflowRun is a row of the workflow_run table.
class WorkflowRun {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "workflow_run";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"workflow_id",
		"name",
		"user_id",
		"user_name",
		"state",
		"message",
		"started_at",
		"finished_at",
		"created_at",
		"updated_at",
		"input",
		"answer",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("started_at", "finished_at", "created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Workflow ID */
	var $workflow_id = "";
	/** Name */
	var $name = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** State */
	var $state = "";
	/** Message */
	var $message = "";
	/** Started At */
	var $started_at = "";
	/** Finished At */
	var $finished_at = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Input */
	var $input = "";
	/** Answer */
	var $answer = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->workflow_id = \Task\Row::text($row, "workflow_id");
		$this->name = \Task\Row::text($row, "name");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->state = \Task\Row::text($row, "state");
		$this->message = \Task\Row::text($row, "message");
		$this->started_at = \Task\Row::stamp($row, "started_at");
		$this->finished_at = \Task\Row::stamp($row, "finished_at");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->input = \Task\Row::text($row, "input");
		$this->answer = \Task\Row::text($row, "answer");
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
	/** getWorkflowId is Workflow ID. */
	function getWorkflowId() {
		return $this->workflow_id;
	}

	/** setWorkflowId sets Workflow ID. */
	function setWorkflowId($value) {
		$this->workflow_id = $value;

		return $this;
	}
	/** getName is Name. */
	function getName() {
		return $this->name;
	}

	/** setName sets Name. */
	function setName($value) {
		$this->name = $value;

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
	/** getMessage is Message. */
	function getMessage() {
		return $this->message;
	}

	/** setMessage sets Message. */
	function setMessage($value) {
		$this->message = $value;

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
	/** getFinishedAt is Finished At. */
	function getFinishedAt() {
		return $this->finished_at;
	}

	/** setFinishedAt sets Finished At. */
	function setFinishedAt($value) {
		$this->finished_at = $value;

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
	/** getInput is Input. */
	function getInput() {
		return $this->input;
	}

	/** setInput sets Input. */
	function setInput($value) {
		$this->input = $value;

		return $this;
	}
	/** getAnswer is Answer. */
	function getAnswer() {
		return $this->answer;
	}

	/** setAnswer sets Answer. */
	function setAnswer($value) {
		$this->answer = $value;

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
			$out[] = new WorkflowRun($row);
		}

		return $out;
	}
}
