<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// KeyRead is a row of the key_read table.
class KeyRead {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "key_read";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "key_name", "user_name", "token_id", "via", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Key Name */
	var $key_name = "";
	/** User Name */
	var $user_name = "";
	/** Token ID */
	var $token_id = "";
	/** Via */
	var $via = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->key_name = \Task\Row::text($row, "key_name");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->token_id = \Task\Row::text($row, "token_id");
		$this->via = \Task\Row::text($row, "via");
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
	/** getKeyName is Key Name. */
	function getKeyName() {
		return $this->key_name;
	}

	/** setKeyName sets Key Name. */
	function setKeyName($value) {
		$this->key_name = $value;

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
	/** getTokenId is Token ID. */
	function getTokenId() {
		return $this->token_id;
	}

	/** setTokenId sets Token ID. */
	function setTokenId($value) {
		$this->token_id = $value;

		return $this;
	}
	/** getVia is Via. */
	function getVia() {
		return $this->via;
	}

	/** setVia sets Via. */
	function setVia($value) {
		$this->via = $value;

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
			$out[] = new KeyRead($row);
		}

		return $out;
	}
}
