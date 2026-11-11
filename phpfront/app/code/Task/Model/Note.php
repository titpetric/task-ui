<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Note is a row of the note table.
class Note {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "note";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"channel",
		"parent_id",
		"user_id",
		"user_name",
		"body",
		"mentions",
		"state",
		"run_id",
		"created_at",
		"updated_at",
		"about_run_id",
		"about_workflow_run_id",
		"area",
		"edited_at",
		"deleted_at",
		"audio_path",
		"spoken",
		"workspace_id",
		"dir",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at", "edited_at", "deleted_at");
	/** ID */
	var $id = "";
	/** Channel */
	var $channel = "";
	/** Parent ID */
	var $parent_id = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** Body */
	var $body = "";
	/** Mentions */
	var $mentions = "";
	/** State */
	var $state = "";
	/** Run ID */
	var $run_id = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** About Run ID */
	var $about_run_id = "";
	/** About Workflow Run ID */
	var $about_workflow_run_id = "";
	/** Area */
	var $area = "";
	/** Edited At */
	var $edited_at = "";
	/** Deleted At */
	var $deleted_at = "";
	/** Audio Path */
	var $audio_path = "";
	/** Spoken */
	var $spoken = "";
	/** Workspace ID */
	var $workspace_id = "";
	/** Dir */
	var $dir = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->channel = \Task\Row::text($row, "channel");
		$this->parent_id = \Task\Row::text($row, "parent_id");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->body = \Task\Row::text($row, "body");
		$this->mentions = \Task\Row::text($row, "mentions");
		$this->state = \Task\Row::text($row, "state");
		$this->run_id = \Task\Row::text($row, "run_id");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->about_run_id = \Task\Row::text($row, "about_run_id");
		$this->about_workflow_run_id = \Task\Row::text($row, "about_workflow_run_id");
		$this->area = \Task\Row::text($row, "area");
		$this->edited_at = \Task\Row::stamp($row, "edited_at");
		$this->deleted_at = \Task\Row::stamp($row, "deleted_at");
		$this->audio_path = \Task\Row::text($row, "audio_path");
		$this->spoken = \Task\Row::text($row, "spoken");
		$this->workspace_id = \Task\Row::text($row, "workspace_id");
		$this->dir = \Task\Row::text($row, "dir");
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
	/** getChannel is Channel. */
	function getChannel() {
		return $this->channel;
	}

	/** setChannel sets Channel. */
	function setChannel($value) {
		$this->channel = $value;

		return $this;
	}
	/** getParentId is Parent ID. */
	function getParentId() {
		return $this->parent_id;
	}

	/** setParentId sets Parent ID. */
	function setParentId($value) {
		$this->parent_id = $value;

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
	/** getBody is Body. */
	function getBody() {
		return $this->body;
	}

	/** setBody sets Body. */
	function setBody($value) {
		$this->body = $value;

		return $this;
	}
	/** getMentions is Mentions. */
	function getMentions() {
		return $this->mentions;
	}

	/** setMentions sets Mentions. */
	function setMentions($value) {
		$this->mentions = $value;

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
	/** getRunId is Run ID. */
	function getRunId() {
		return $this->run_id;
	}

	/** setRunId sets Run ID. */
	function setRunId($value) {
		$this->run_id = $value;

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
	/** getAboutRunId is About Run ID. */
	function getAboutRunId() {
		return $this->about_run_id;
	}

	/** setAboutRunId sets About Run ID. */
	function setAboutRunId($value) {
		$this->about_run_id = $value;

		return $this;
	}
	/** getAboutWorkflowRunId is About Workflow Run ID. */
	function getAboutWorkflowRunId() {
		return $this->about_workflow_run_id;
	}

	/** setAboutWorkflowRunId sets About Workflow Run ID. */
	function setAboutWorkflowRunId($value) {
		$this->about_workflow_run_id = $value;

		return $this;
	}
	/** getArea is Area. */
	function getArea() {
		return $this->area;
	}

	/** setArea sets Area. */
	function setArea($value) {
		$this->area = $value;

		return $this;
	}
	/** getEditedAt is Edited At. */
	function getEditedAt() {
		return $this->edited_at;
	}

	/** setEditedAt sets Edited At. */
	function setEditedAt($value) {
		$this->edited_at = $value;

		return $this;
	}
	/** getDeletedAt is Deleted At. */
	function getDeletedAt() {
		return $this->deleted_at;
	}

	/** setDeletedAt sets Deleted At. */
	function setDeletedAt($value) {
		$this->deleted_at = $value;

		return $this;
	}
	/** getAudioPath is Audio Path. */
	function getAudioPath() {
		return $this->audio_path;
	}

	/** setAudioPath sets Audio Path. */
	function setAudioPath($value) {
		$this->audio_path = $value;

		return $this;
	}
	/** getSpoken is Spoken. */
	function getSpoken() {
		return $this->spoken;
	}

	/** setSpoken sets Spoken. */
	function setSpoken($value) {
		$this->spoken = $value;

		return $this;
	}
	/** getWorkspaceId is Workspace ID. */
	function getWorkspaceId() {
		return $this->workspace_id;
	}

	/** setWorkspaceId sets Workspace ID. */
	function setWorkspaceId($value) {
		$this->workspace_id = $value;

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
			$out[] = new Note($row);
		}

		return $out;
	}
}
