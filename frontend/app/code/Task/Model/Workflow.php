<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Workflow is a row of the workflow table.
class Workflow {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "workflow";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"name",
		"user_id",
		"user_name",
		"created_at",
		"updated_at",
		"description",
		"prompt",
		"prompt_folder",
		"run_as_user_id",
		"runs_over",
		"sidebar_hidden",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Name */
	var $name = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Description */
	var $description = "";
	/** Prompt */
	var $prompt = "";
	/** Prompt Folder */
	var $prompt_folder = "";
	/** Run As User ID */
	var $run_as_user_id = "";
	/** Runs Over */
	var $runs_over = "";
	/** Sidebar Hidden */
	var $sidebar_hidden = false;

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->name = \Task\Row::text($row, "name");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->description = \Task\Row::text($row, "description");
		$this->prompt = \Task\Row::text($row, "prompt");
		$this->prompt_folder = \Task\Row::text($row, "prompt_folder");
		$this->run_as_user_id = \Task\Row::text($row, "run_as_user_id");
		$this->runs_over = \Task\Row::text($row, "runs_over");
		$this->sidebar_hidden = \Task\Row::flag($row, "sidebar_hidden");
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
	/** getDescription is Description. */
	function getDescription() {
		return $this->description;
	}

	/** setDescription sets Description. */
	function setDescription($value) {
		$this->description = $value;

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
	/** getPromptFolder is Prompt Folder. */
	function getPromptFolder() {
		return $this->prompt_folder;
	}

	/** setPromptFolder sets Prompt Folder. */
	function setPromptFolder($value) {
		$this->prompt_folder = $value;

		return $this;
	}
	/** getRunAsUserId is Run As User ID. */
	function getRunAsUserId() {
		return $this->run_as_user_id;
	}

	/** setRunAsUserId sets Run As User ID. */
	function setRunAsUserId($value) {
		$this->run_as_user_id = $value;

		return $this;
	}
	/** getRunsOver is Runs Over. */
	function getRunsOver() {
		return $this->runs_over;
	}

	/** setRunsOver sets Runs Over. */
	function setRunsOver($value) {
		$this->runs_over = $value;

		return $this;
	}
	/** getSidebarHidden is Sidebar Hidden. */
	function getSidebarHidden() {
		return $this->sidebar_hidden;
	}

	/** setSidebarHidden sets Sidebar Hidden. */
	function setSidebarHidden($value) {
		$this->sidebar_hidden = $value;

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
			$out[] = new Workflow($row);
		}

		return $out;
	}
}
