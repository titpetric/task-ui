<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// SecretRead is a row of the secret_read table.
class SecretRead {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "secret_read";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "secret_id", "name", "run_id", "folder", "dir", "job", "user_name", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Secret ID */
	var $secret_id = "";
	/** Name */
	var $name = "";
	/** Run ID */
	var $run_id = "";
	/** Folder */
	var $folder = "";
	/** Dir */
	var $dir = "";
	/** Job */
	var $job = "";
	/** User Name */
	var $user_name = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->secret_id = \Task\Row::text($row, "secret_id");
		$this->name = \Task\Row::text($row, "name");
		$this->run_id = \Task\Row::text($row, "run_id");
		$this->folder = \Task\Row::text($row, "folder");
		$this->dir = \Task\Row::text($row, "dir");
		$this->job = \Task\Row::text($row, "job");
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
	/** getSecretId is Secret ID. */
	function getSecretId() {
		return $this->secret_id;
	}

	/** setSecretId sets Secret ID. */
	function setSecretId($value) {
		$this->secret_id = $value;

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
	/** getRunId is Run ID. */
	function getRunId() {
		return $this->run_id;
	}

	/** setRunId sets Run ID. */
	function setRunId($value) {
		$this->run_id = $value;

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
	/** getDir is Dir. */
	function getDir() {
		return $this->dir;
	}

	/** setDir sets Dir. */
	function setDir($value) {
		$this->dir = $value;

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
			$out[] = new SecretRead($row);
		}

		return $out;
	}
}
