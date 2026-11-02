<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// PermissionPolicy is a row of the permission_policy table.
class PermissionPolicy {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "permission_policy";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "subject_kind", "subject", "path", "effect", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Subject Kind */
	var $subject_kind = "";
	/** Subject */
	var $subject = "";
	/** Path */
	var $path = "";
	/** Effect */
	var $effect = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->subject_kind = \Task\Row::text($row, "subject_kind");
		$this->subject = \Task\Row::text($row, "subject");
		$this->path = \Task\Row::text($row, "path");
		$this->effect = \Task\Row::text($row, "effect");
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
	/** getSubjectKind is Subject Kind. */
	function getSubjectKind() {
		return $this->subject_kind;
	}

	/** setSubjectKind sets Subject Kind. */
	function setSubjectKind($value) {
		$this->subject_kind = $value;

		return $this;
	}
	/** getSubject is Subject. */
	function getSubject() {
		return $this->subject;
	}

	/** setSubject sets Subject. */
	function setSubject($value) {
		$this->subject = $value;

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
	/** getEffect is Effect. */
	function getEffect() {
		return $this->effect;
	}

	/** setEffect sets Effect. */
	function setEffect($value) {
		$this->effect = $value;

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
			$out[] = new PermissionPolicy($row);
		}

		return $out;
	}
}
