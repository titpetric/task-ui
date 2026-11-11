<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Module is a row of the module table.
class Module {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "module";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"name",
		"vendor",
		"plugin",
		"root",
		"title",
		"description",
		"icon",
		"version",
		"created_at",
		"updated_at",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Name */
	var $name = "";
	/** Vendor */
	var $vendor = "";
	/** Plugin */
	var $plugin = "";
	/** Root */
	var $root = "";
	/** Title */
	var $title = "";
	/** Description */
	var $description = "";
	/** Icon */
	var $icon = "";
	/** Version */
	var $version = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->name = \Task\Row::text($row, "name");
		$this->vendor = \Task\Row::text($row, "vendor");
		$this->plugin = \Task\Row::text($row, "plugin");
		$this->root = \Task\Row::text($row, "root");
		$this->title = \Task\Row::text($row, "title");
		$this->description = \Task\Row::text($row, "description");
		$this->icon = \Task\Row::text($row, "icon");
		$this->version = \Task\Row::text($row, "version");
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
	/** getName is Name. */
	function getName() {
		return $this->name;
	}

	/** setName sets Name. */
	function setName($value) {
		$this->name = $value;

		return $this;
	}
	/** getVendor is Vendor. */
	function getVendor() {
		return $this->vendor;
	}

	/** setVendor sets Vendor. */
	function setVendor($value) {
		$this->vendor = $value;

		return $this;
	}
	/** getPlugin is Plugin. */
	function getPlugin() {
		return $this->plugin;
	}

	/** setPlugin sets Plugin. */
	function setPlugin($value) {
		$this->plugin = $value;

		return $this;
	}
	/** getRoot is Root. */
	function getRoot() {
		return $this->root;
	}

	/** setRoot sets Root. */
	function setRoot($value) {
		$this->root = $value;

		return $this;
	}
	/** getTitle is Title. */
	function getTitle() {
		return $this->title;
	}

	/** setTitle sets Title. */
	function setTitle($value) {
		$this->title = $value;

		return $this;
	}
	/** getDescription is Description. */
	function getDescription() {
		return $this->description;
	}

	/** setDescription sets Description. */
	function setDescription($value) {
		$this->description = $value;

		return $this;
	}
	/** getIcon is Icon. */
	function getIcon() {
		return $this->icon;
	}

	/** setIcon sets Icon. */
	function setIcon($value) {
		$this->icon = $value;

		return $this;
	}
	/** getVersion is Version. */
	function getVersion() {
		return $this->version;
	}

	/** setVersion sets Version. */
	function setVersion($value) {
		$this->version = $value;

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
			$out[] = new Module($row);
		}

		return $out;
	}
}
