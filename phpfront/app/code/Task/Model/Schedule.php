<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Schedule is a row of the schedule table.
class Schedule {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "schedule";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"folder",
		"job",
		"user_id",
		"user_name",
		"expression",
		"enabled",
		"last_run_at",
		"next_run_at",
		"created_at",
		"updated_at",
		"target",
		"args",
		"triage",
		"workflow_id",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("last_run_at", "next_run_at", "created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Folder */
	var $folder = "";
	/** Job */
	var $job = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** Expression */
	var $expression = "";
	/** Enabled */
	var $enabled = 0;
	/** Last Run At */
	var $last_run_at = "";
	/** Next Run At */
	var $next_run_at = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Target */
	var $target = "";
	/** Args */
	var $args = "";
	/** Triage */
	var $triage = 0;
	/** Workflow ID */
	var $workflow_id = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->folder = \Task\Row::text($row, "folder");
		$this->job = \Task\Row::text($row, "job");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->expression = \Task\Row::text($row, "expression");
		$this->enabled = \Task\Row::number($row, "enabled");
		$this->last_run_at = \Task\Row::stamp($row, "last_run_at");
		$this->next_run_at = \Task\Row::stamp($row, "next_run_at");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->target = \Task\Row::text($row, "target");
		$this->args = \Task\Row::text($row, "args");
		$this->triage = \Task\Row::number($row, "triage");
		$this->workflow_id = \Task\Row::text($row, "workflow_id");
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
	/** getFolder is Folder. */
	function getFolder() {
		return $this->folder;
	}

	/** setFolder sets Folder. */
	function setFolder($value) {
		$this->folder = $value;

		return $this;
	}
	/** getJob is Job. */
	function getJob() {
		return $this->job;
	}

	/** setJob sets Job. */
	function setJob($value) {
		$this->job = $value;

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
	/** getExpression is Expression. */
	function getExpression() {
		return $this->expression;
	}

	/** setExpression sets Expression. */
	function setExpression($value) {
		$this->expression = $value;

		return $this;
	}
	/** getEnabled is Enabled. */
	function getEnabled() {
		return $this->enabled;
	}

	/** setEnabled sets Enabled. */
	function setEnabled($value) {
		$this->enabled = $value;

		return $this;
	}
	/** getLastRunAt is Last Run At. */
	function getLastRunAt() {
		return $this->last_run_at;
	}

	/** setLastRunAt sets Last Run At. */
	function setLastRunAt($value) {
		$this->last_run_at = $value;

		return $this;
	}
	/** getNextRunAt is Next Run At. */
	function getNextRunAt() {
		return $this->next_run_at;
	}

	/** setNextRunAt sets Next Run At. */
	function setNextRunAt($value) {
		$this->next_run_at = $value;

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
	/** getTarget is Target. */
	function getTarget() {
		return $this->target;
	}

	/** setTarget sets Target. */
	function setTarget($value) {
		$this->target = $value;

		return $this;
	}
	/** getArgs is Args. */
	function getArgs() {
		return $this->args;
	}

	/** setArgs sets Args. */
	function setArgs($value) {
		$this->args = $value;

		return $this;
	}
	/** getTriage is Triage. */
	function getTriage() {
		return $this->triage;
	}

	/** setTriage sets Triage. */
	function setTriage($value) {
		$this->triage = $value;

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
			$out[] = new Schedule($row);
		}

		return $out;
	}
}
