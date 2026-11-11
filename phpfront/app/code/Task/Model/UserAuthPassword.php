<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// UserAuthPassword is a row of the user_auth_password table.
class UserAuthPassword {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "user_auth_password";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "user_id", "hash", "kind", "used_at", "created_at", "updated_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("used_at", "created_at", "updated_at");
	/** ID */
	var $id = "";
	/** User ID */
	var $user_id = "";
	/** Hash */
	var $hash = "";
	/** Kind */
	var $kind = "";
	/** Used At */
	var $used_at = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->hash = \Task\Row::text($row, "hash");
		$this->kind = \Task\Row::text($row, "kind");
		$this->used_at = \Task\Row::stamp($row, "used_at");
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
	/** getUserId is User ID. */
	function getUserId() {
		return $this->user_id;
	}

	/** setUserId sets User ID. */
	function setUserId($value) {
		$this->user_id = $value;

		return $this;
	}
	/** getHash is Hash. */
	function getHash() {
		return $this->hash;
	}

	/** setHash sets Hash. */
	function setHash($value) {
		$this->hash = $value;

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
	/** getUsedAt is Used At. */
	function getUsedAt() {
		return $this->used_at;
	}

	/** setUsedAt sets Used At. */
	function setUsedAt($value) {
		$this->used_at = $value;

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
			$out[] = new UserAuthPassword($row);
		}

		return $out;
	}
}
