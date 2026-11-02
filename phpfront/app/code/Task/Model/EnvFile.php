<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// EnvFile is a row of the env_file table.
class EnvFile {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "env_file";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "folder", "path", "content", "created_at", "updated_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Folder */
	var $folder = "";
	/** Path */
	var $path = "";
	/** Content */
	var $content = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->folder = \Task\Row::text($row, "folder");
		$this->path = \Task\Row::text($row, "path");
		$this->content = \Task\Row::text($row, "content");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
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
	/** getPath is Path. */
	function getPath() {
		return $this->path;
	}

	/** setPath sets Path. */
	function setPath($value) {
		$this->path = $value;

		return $this;
	}
	/** getContent is Content. */
	function getContent() {
		return $this->content;
	}

	/** setContent sets Content. */
	function setContent($value) {
		$this->content = $value;

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
			$out[] = new EnvFile($row);
		}

		return $out;
	}
}
