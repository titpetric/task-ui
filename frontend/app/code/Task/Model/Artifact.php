<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Artifact is a row of the artifact table.
class Artifact {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "artifact";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"run_id",
		"folder",
		"job",
		"name",
		"path",
		"byte_size",
		"content_type",
		"created_at",
		"commit_id",
		"user_name",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Run ID */
	var $run_id = "";
	/** Folder */
	var $folder = "";
	/** Job */
	var $job = "";
	/** Name */
	var $name = "";
	/** Path */
	var $path = "";
	/** Byte Size */
	var $byte_size = 0;
	/** Content Type */
	var $content_type = "";
	/** Created At */
	var $created_at = "";
	/** Commit ID */
	var $commit_id = "";
	/** User Name */
	var $user_name = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->run_id = \Task\Row::text($row, "run_id");
		$this->folder = \Task\Row::text($row, "folder");
		$this->job = \Task\Row::text($row, "job");
		$this->name = \Task\Row::text($row, "name");
		$this->path = \Task\Row::text($row, "path");
		$this->byte_size = \Task\Row::number($row, "byte_size");
		$this->content_type = \Task\Row::text($row, "content_type");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->commit_id = \Task\Row::text($row, "commit_id");
		$this->user_name = \Task\Row::text($row, "user_name");
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
	/** getJob is Job. */
	function getJob() {
		return $this->job;
	}

	/** setJob sets Job. */
	function setJob($value) {
		$this->job = $value;

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
	/** getPath is Path. */
	function getPath() {
		return $this->path;
	}

	/** setPath sets Path. */
	function setPath($value) {
		$this->path = $value;

		return $this;
	}
	/** getByteSize is Byte Size. */
	function getByteSize() {
		return $this->byte_size;
	}

	/** setByteSize sets Byte Size. */
	function setByteSize($value) {
		$this->byte_size = $value;

		return $this;
	}
	/** getContentType is Content Type. */
	function getContentType() {
		return $this->content_type;
	}

	/** setContentType sets Content Type. */
	function setContentType($value) {
		$this->content_type = $value;

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
	/** getCommitId is Commit ID. */
	function getCommitId() {
		return $this->commit_id;
	}

	/** setCommitId sets Commit ID. */
	function setCommitId($value) {
		$this->commit_id = $value;

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
			$out[] = new Artifact($row);
		}

		return $out;
	}
}
