<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Sidebar is a row of the sidebar table.
class Sidebar {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "sidebar";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"kind",
		"href",
		"title",
		"icon",
		"position",
		"hidden",
		"rule",
		"created_at",
		"updated_at",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Kind */
	var $kind = "";
	/** Href */
	var $href = "";
	/** Title */
	var $title = "";
	/** Icon */
	var $icon = "";
	/** Position */
	var $position = 0;
	/** Hidden */
	var $hidden = false;
	/** Rule */
	var $rule = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->kind = \Task\Row::text($row, "kind");
		$this->href = \Task\Row::text($row, "href");
		$this->title = \Task\Row::text($row, "title");
		$this->icon = \Task\Row::text($row, "icon");
		$this->position = \Task\Row::number($row, "position");
		$this->hidden = \Task\Row::flag($row, "hidden");
		$this->rule = \Task\Row::text($row, "rule");
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
	/** getKind is Kind. */
	function getKind() {
		return $this->kind;
	}

	/** setKind sets Kind. */
	function setKind($value) {
		$this->kind = $value;

		return $this;
	}
	/** getHref is Href. */
	function getHref() {
		return $this->href;
	}

	/** setHref sets Href. */
	function setHref($value) {
		$this->href = $value;

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
	/** getIcon is Icon. */
	function getIcon() {
		return $this->icon;
	}

	/** setIcon sets Icon. */
	function setIcon($value) {
		$this->icon = $value;

		return $this;
	}
	/** getPosition is Position. */
	function getPosition() {
		return $this->position;
	}

	/** setPosition sets Position. */
	function setPosition($value) {
		$this->position = $value;

		return $this;
	}
	/** getHidden is Hidden. */
	function getHidden() {
		return $this->hidden;
	}

	/** setHidden sets Hidden. */
	function setHidden($value) {
		$this->hidden = $value;

		return $this;
	}
	/** getRule is Rule. */
	function getRule() {
		return $this->rule;
	}

	/** setRule sets Rule. */
	function setRule($value) {
		$this->rule = $value;

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
			$out[] = new Sidebar($row);
		}

		return $out;
	}
}
