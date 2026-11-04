<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// WorkflowRunStep is a row of the workflow_run_step table.
class WorkflowRunStep {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "workflow_run_step";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"workflow_run_id",
		"position",
		"folder",
		"job",
		"run_id",
		"state",
		"created_at",
		"updated_at",
		"prompt",
		"answer",
		"workflow_ref",
		"workflow_run_ref",
		"kind",
		"markers",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Workflow Run ID */
	var $workflow_run_id = "";
	/** Position */
	var $position = 0;
	/** Folder */
	var $folder = "";
	/** Job */
	var $job = "";
	/** Run ID */
	var $run_id = "";
	/** State */
	var $state = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Prompt */
	var $prompt = "";
	/** Answer */
	var $answer = "";
	/** Workflow Ref */
	var $workflow_ref = "";
	/** Workflow Run Ref */
	var $workflow_run_ref = "";
	/** Kind */
	var $kind = "";
	/** Markers */
	var $markers = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->workflow_run_id = \Task\Row::text($row, "workflow_run_id");
		$this->position = \Task\Row::number($row, "position");
		$this->folder = \Task\Row::text($row, "folder");
		$this->job = \Task\Row::text($row, "job");
		$this->run_id = \Task\Row::text($row, "run_id");
		$this->state = \Task\Row::text($row, "state");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->prompt = \Task\Row::text($row, "prompt");
		$this->answer = \Task\Row::text($row, "answer");
		$this->workflow_ref = \Task\Row::text($row, "workflow_ref");
		$this->workflow_run_ref = \Task\Row::text($row, "workflow_run_ref");
		$this->kind = \Task\Row::text($row, "kind");
		$this->markers = \Task\Row::text($row, "markers");
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
	/** getWorkflowRunId is Workflow Run ID. */
	function getWorkflowRunId() {
		return $this->workflow_run_id;
	}

	/** setWorkflowRunId sets Workflow Run ID. */
	function setWorkflowRunId($value) {
		$this->workflow_run_id = $value;

		return $this;
	}
	/** getPosition is Position. */
	function getPosition() {
		return $this->position;
	}

	/** setPosition sets Position. */
	function setPosition($value) {
		$this->position = $value;

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
	/** getRunId is Run ID. */
	function getRunId() {
		return $this->run_id;
	}

	/** setRunId sets Run ID. */
	function setRunId($value) {
		$this->run_id = $value;

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
	/** getPrompt is Prompt. */
	function getPrompt() {
		return $this->prompt;
	}

	/** setPrompt sets Prompt. */
	function setPrompt($value) {
		$this->prompt = $value;

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
	/** getWorkflowRef is Workflow Ref. */
	function getWorkflowRef() {
		return $this->workflow_ref;
	}

	/** setWorkflowRef sets Workflow Ref. */
	function setWorkflowRef($value) {
		$this->workflow_ref = $value;

		return $this;
	}
	/** getWorkflowRunRef is Workflow Run Ref. */
	function getWorkflowRunRef() {
		return $this->workflow_run_ref;
	}

	/** setWorkflowRunRef sets Workflow Run Ref. */
	function setWorkflowRunRef($value) {
		$this->workflow_run_ref = $value;

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
	/** getMarkers is Markers. */
	function getMarkers() {
		return $this->markers;
	}

	/** setMarkers sets Markers. */
	function setMarkers($value) {
		$this->markers = $value;

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
			$out[] = new WorkflowRunStep($row);
		}

		return $out;
	}
}
