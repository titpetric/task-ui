<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// WorkspaceKeyGrant is a row of the workspace_key_grant table.
class WorkspaceKeyGrant {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "workspace_key_grant";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "workspace_id", "key_name", "role", "user_name", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Workspace ID */
	var $workspace_id = "";
	/** Key Name */
	var $key_name = "";
	/** Role */
	var $role = "";
	/** User Name */
	var $user_name = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->workspace_id = \Task\Row::text($row, "workspace_id");
		$this->key_name = \Task\Row::text($row, "key_name");
		$this->role = \Task\Row::text($row, "role");
		$this->user_name = \Task\Row::text($row, "user_name");
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
	/** getWorkspaceId is Workspace ID. */
	function getWorkspaceId() {
		return $this->workspace_id;
	}

	/** setWorkspaceId sets Workspace ID. */
	function setWorkspaceId($value) {
		$this->workspace_id = $value;

		return $this;
	}
	/** getKeyName is Key Name. */
	function getKeyName() {
		return $this->key_name;
	}

	/** setKeyName sets Key Name. */
	function setKeyName($value) {
		$this->key_name = $value;

		return $this;
	}
	/** getRole is Role. */
	function getRole() {
		return $this->role;
	}

	/** setRole sets Role. */
	function setRole($value) {
		$this->role = $value;

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
			$out[] = new WorkspaceKeyGrant($row);
		}

		return $out;
	}
}
