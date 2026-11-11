<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// WorkflowStep is a row of the workflow_step table.
class WorkflowStep {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "workflow_step";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"workflow_id",
		"position",
		"folder",
		"job",
		"disabled",
		"prompt",
		"dir",
		"workflow_ref",
		"target",
		"args",
		"kind",
		"markers",
		"tries",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array();
	/** ID */
	var $id = "";
	/** Workflow ID */
	var $workflow_id = "";
	/** Position */
	var $position = 0;
	/** Folder */
	var $folder = "";
	/** Job */
	var $job = "";
	/** Disabled */
	var $disabled = false;
	/** Prompt */
	var $prompt = "";
	/** Dir */
	var $dir = "";
	/** Workflow Ref */
	var $workflow_ref = "";
	/** Target */
	var $target = "";
	/** Args */
	var $args = "";
	/** Kind */
	var $kind = "";
	/** Markers */
	var $markers = "";
	/** Tries */
	var $tries = 0;

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->workflow_id = \Task\Row::text($row, "workflow_id");
		$this->position = \Task\Row::number($row, "position");
		$this->folder = \Task\Row::text($row, "folder");
		$this->job = \Task\Row::text($row, "job");
		$this->disabled = \Task\Row::flag($row, "disabled");
		$this->prompt = \Task\Row::text($row, "prompt");
		$this->dir = \Task\Row::text($row, "dir");
		$this->workflow_ref = \Task\Row::text($row, "workflow_ref");
		$this->target = \Task\Row::text($row, "target");
		$this->args = \Task\Row::text($row, "args");
		$this->kind = \Task\Row::text($row, "kind");
		$this->markers = \Task\Row::text($row, "markers");
		$this->tries = \Task\Row::number($row, "tries");
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
	/** getDisabled is Disabled. */
	function getDisabled() {
		return $this->disabled;
	}

	/** setDisabled sets Disabled. */
	function setDisabled($value) {
		$this->disabled = $value;

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
	/** getDir is Dir. */
	function getDir() {
		return $this->dir;
	}

	/** setDir sets Dir. */
	function setDir($value) {
		$this->dir = $value;

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
	/** getTries is Tries. */
	function getTries() {
		return $this->tries;
	}

	/** setTries sets Tries. */
	function setTries($value) {
		$this->tries = $value;

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
			$out[] = new WorkflowStep($row);
		}

		return $out;
	}
}
