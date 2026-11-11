<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Host is a row of the host table.
class Host {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "host";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array("id", "name", "address", "source", "login", "added_by", "created_at");

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at");
	/** ID */
	var $id = "";
	/** Name */
	var $name = "";
	/** Address */
	var $address = "";
	/** Source */
	var $source = "";
	/** Login */
	var $login = "";
	/** Added By */
	var $added_by = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->name = \Task\Row::text($row, "name");
		$this->address = \Task\Row::text($row, "address");
		$this->source = \Task\Row::text($row, "source");
		$this->login = \Task\Row::text($row, "login");
		$this->added_by = \Task\Row::text($row, "added_by");
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
	/** getName is Name. */
	function getName() {
		return $this->name;
	}

	/** setName sets Name. */
	function setName($value) {
		$this->name = $value;

		return $this;
	}
	/** getAddress is Address. */
	function getAddress() {
		return $this->address;
	}

	/** setAddress sets Address. */
	function setAddress($value) {
		$this->address = $value;

		return $this;
	}
	/** getSource is Source. */
	function getSource() {
		return $this->source;
	}

	/** setSource sets Source. */
	function setSource($value) {
		$this->source = $value;

		return $this;
	}
	/** getLogin is Login. */
	function getLogin() {
		return $this->login;
	}

	/** setLogin sets Login. */
	function setLogin($value) {
		$this->login = $value;

		return $this;
	}
	/** getAddedBy is Added By. */
	function getAddedBy() {
		return $this->added_by;
	}

	/** setAddedBy sets Added By. */
	function setAddedBy($value) {
		$this->added_by = $value;

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
			$out[] = new Host($row);
		}

		return $out;
	}
}
