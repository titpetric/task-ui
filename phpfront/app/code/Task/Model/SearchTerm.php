<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// SearchTerm is a row of the search_term table.
class SearchTerm {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "search_term";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = ["document_id", "ordinal", "term", "weight"];

	/** primary is the primary key columns. */
	public static $primary = ["document_id", "ordinal", "term"];

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = [];
	/** Document ID */
	var $document_id = "";
	/** Ordinal */
	var $ordinal = 0;
	/** Term */
	var $term = "";
	/** Weight */
	var $weight = 0;

	function __construct($row = []) {
		$this->document_id = \Task\Row::text($row, "document_id");
		$this->ordinal = \Task\Row::number($row, "ordinal");
		$this->term = \Task\Row::text($row, "term");
		$this->weight = \Task\Row::number($row, "weight");
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
	/** getTerm is Term. */
	function getTerm() {
		return $this->term;
	}

	/** setTerm sets Term. */
	function setTerm($value) {
		$this->term = $value;

		return $this;
	}
	/** getWeight is Weight. */
	function getWeight() {
		return $this->weight;
	}

	/** setWeight sets Weight. */
	function setWeight($value) {
		$this->weight = $value;

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
			$out[] = new SearchTerm($row);
		}

		return $out;
	}
}
