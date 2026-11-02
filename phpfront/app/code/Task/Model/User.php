<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// User is a row of the user table.
class User {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "user";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"name",
		"role",
		"folders",
		"created_at",
		"updated_at",
		"admin",
		"mcp",
		"agent_reply",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Name */
	var $name = "";
	/** Role */
	var $role = "";
	/** Folders */
	var $folders = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Admin */
	var $admin = false;
	/** Mcp */
	var $mcp = false;
	/** Agent Reply */
	var $agent_reply = false;

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->name = \Task\Row::text($row, "name");
		$this->role = \Task\Row::text($row, "role");
		$this->folders = \Task\Row::text($row, "folders");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->admin = \Task\Row::flag($row, "admin");
		$this->mcp = \Task\Row::flag($row, "mcp");
		$this->agent_reply = \Task\Row::flag($row, "agent_reply");
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
	/** getRole is Role. */
	function getRole() {
		return $this->role;
	}

	/** setRole sets Role. */
	function setRole($value) {
		$this->role = $value;

		return $this;
	}
	/** getFolders is Folders. */
	function getFolders() {
		return $this->folders;
	}

	/** setFolders sets Folders. */
	function setFolders($value) {
		$this->folders = $value;

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
	/** getAdmin is Admin. */
	function getAdmin() {
		return $this->admin;
	}

	/** setAdmin sets Admin. */
	function setAdmin($value) {
		$this->admin = $value;

		return $this;
	}
	/** getMcp is Mcp. */
	function getMcp() {
		return $this->mcp;
	}

	/** setMcp sets Mcp. */
	function setMcp($value) {
		$this->mcp = $value;

		return $this;
	}
	/** getAgentReply is Agent Reply. */
	function getAgentReply() {
		return $this->agent_reply;
	}

	/** setAgentReply sets Agent Reply. */
	function setAgentReply($value) {
		$this->agent_reply = $value;

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
			$out[] = new User($row);
		}

		return $out;
	}
}
