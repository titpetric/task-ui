<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// UserAuthToken is a row of the user_auth_token table.
class UserAuthToken {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "user_auth_token";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"user_id",
		"name",
		"hash",
		"used_at",
		"created_at",
		"updated_at",
		"expires_at",
		"kind",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("used_at", "created_at", "updated_at", "expires_at");
	/** ID */
	var $id = "";
	/** User ID */
	var $user_id = "";
	/** Name */
	var $name = "";
	/** Hash */
	var $hash = "";
	/** Used At */
	var $used_at = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Expires At */
	var $expires_at = "";
	/** Kind */
	var $kind = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->name = \Task\Row::text($row, "name");
		$this->hash = \Task\Row::text($row, "hash");
		$this->used_at = \Task\Row::stamp($row, "used_at");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->expires_at = \Task\Row::stamp($row, "expires_at");
		$this->kind = \Task\Row::text($row, "kind");
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
	/** getName is Name. */
	function getName() {
		return $this->name;
	}

	/** setName sets Name. */
	function setName($value) {
		$this->name = $value;

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
	/** getExpiresAt is Expires At. */
	function getExpiresAt() {
		return $this->expires_at;
	}

	/** setExpiresAt sets Expires At. */
	function setExpiresAt($value) {
		$this->expires_at = $value;

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
			$out[] = new UserAuthToken($row);
		}

		return $out;
	}
}
