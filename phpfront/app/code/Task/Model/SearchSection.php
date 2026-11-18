<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// SearchSection is a row of the search_section table.
class SearchSection {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "search_section";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = ["document_id", "ordinal", "level", "heading", "breadcrumb", "line", "body"];

	/** primary is the primary key columns. */
	public static $primary = ["document_id", "ordinal"];

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = [];
	/** Document ID */
	var $document_id = "";
	/** Ordinal */
	var $ordinal = 0;
	/** Level */
	var $level = 0;
	/** Heading */
	var $heading = "";
	/** Breadcrumb */
	var $breadcrumb = "";
	/** Line */
	var $line = 0;
	/** Body */
	var $body = "";

	function __construct($row = []) {
		$this->document_id = \Task\Row::text($row, "document_id");
		$this->ordinal = \Task\Row::number($row, "ordinal");
		$this->level = \Task\Row::number($row, "level");
		$this->heading = \Task\Row::text($row, "heading");
		$this->breadcrumb = \Task\Row::text($row, "breadcrumb");
		$this->line = \Task\Row::number($row, "line");
		$this->body = \Task\Row::text($row, "body");
	}
	/** getDocumentId is Document ID. */
	function getDocumentId() {
		return $this->document_id;
	}

	/** setDocumentId sets Document ID. */
	function setDocumentId($value) {
		$this->document_id = $value;

		return $this;
	}
	/** getOrdinal is Ordinal. */
	function getOrdinal() {
		return $this->ordinal;
	}

	/** setOrdinal sets Ordinal. */
	function setOrdinal($value) {
		$this->ordinal = $value;

		return $this;
	}
	/** getLevel is Level. */
	function getLevel() {
		return $this->level;
	}

	/** setLevel sets Level. */
	function setLevel($value) {
		$this->level = $value;

		return $this;
	}
	/** getHeading is Heading. */
	function getHeading() {
		return $this->heading;
	}

	/** setHeading sets Heading. */
	function setHeading($value) {
		$this->heading = $value;

		return $this;
	}
	/** getBreadcrumb is Breadcrumb. */
	function getBreadcrumb() {
		return $this->breadcrumb;
	}

	/** setBreadcrumb sets Breadcrumb. */
	function setBreadcrumb($value) {
		$this->breadcrumb = $value;

		return $this;
	}
	/** getLine is Line. */
	function getLine() {
		return $this->line;
	}

	/** setLine sets Line. */
	function setLine($value) {
		$this->line = $value;

		return $this;
	}
	/** getBody is Body. */
	function getBody() {
		return $this->body;
	}

	/** setBody sets Body. */
	function setBody($value) {
		$this->body = $value;

		return $this;
	}

	/** values is this row as the database takes it, keyed by column. */
	function values() {
		return \Task\Row::values($this, self::$stamps);
	}

	/** rows turns what the database answered into models. */
	static function rows($found) {
		$out = [];

		if (!is_array($found)) {
			return $out;
		}

		foreach ($found as $row) {
			$out[] = new SearchSection($row);
		}

		return $out;
	}
}
