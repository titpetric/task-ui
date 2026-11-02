<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Secret is a row of the secret table.
class Secret {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "secret";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "scope", "name", "content", "created_at", "updated_at", "is_secret");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Scope */
	var $scope = "";
	/** Name */
	var $name = "";
	/** Content */
	var $content = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Is Secret */
	var $is_secret = 0;

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->scope = \Task\Row::text($row, "scope");
		$this->name = \Task\Row::text($row, "name");
		$this->content = \Task\Row::text($row, "content");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->is_secret = \Task\Row::number($row, "is_secret");
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
	/** getScope is Scope. */
	function getScope() {
		return $this->scope;
	}

	/** setScope sets Scope. */
	function setScope($value) {
		$this->scope = $value;

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
	/** getIsSecret is Is Secret. */
	function getIsSecret() {
		return $this->is_secret;
	}

	/** setIsSecret sets Is Secret. */
	function setIsSecret($value) {
		$this->is_secret = $value;

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
			$out[] = new Secret($row);
		}

		return $out;
	}
}
