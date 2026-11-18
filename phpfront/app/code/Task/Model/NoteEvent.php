<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// NoteEvent is a row of the note_event table.
class NoteEvent {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "note_event";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = [
		"id",
		"note_id",
		"run_id",
		"question_id",
		"seq",
		"kind",
		"tool_id",
		"body",
		"data",
		"created_at",
	];

	/** primary is the primary key columns. */
	public static $primary = ["id"];

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = ["created_at"];
	/** ID */
	var $id = "";
	/** Note ID */
	var $note_id = "";
	/** Run ID */
	var $run_id = "";
	/** Question ID */
	var $question_id = "";
	/** Seq */
	var $seq = 0;
	/** Kind */
	var $kind = "";
	/** Tool ID */
	var $tool_id = "";
	/** Body */
	var $body = "";
	/** Data */
	var $data = "";
	/** Created At */
	var $created_at = "";

	function __construct($row = []) {
		$this->id = \Task\Row::text($row, "id");
		$this->note_id = \Task\Row::text($row, "note_id");
		$this->run_id = \Task\Row::text($row, "run_id");
		$this->question_id = \Task\Row::text($row, "question_id");
		$this->seq = \Task\Row::number($row, "seq");
		$this->kind = \Task\Row::text($row, "kind");
		$this->tool_id = \Task\Row::text($row, "tool_id");
		$this->body = \Task\Row::text($row, "body");
		$this->data = \Task\Row::text($row, "data");
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
	/** getNoteId is Note ID. */
	function getNoteId() {
		return $this->note_id;
	}

	/** setNoteId sets Note ID. */
	function setNoteId($value) {
		$this->note_id = $value;

		return $this;
	}
	/** getRunId is Run ID. */
	function getRunId() {
		return $this->run_id;
	}

	/** setRunId sets Run ID. */
	function setRunId($value) {
		$this->run_id = $value;

		return $this;
	}
	/** getQuestionId is Question ID. */
	function getQuestionId() {
		return $this->question_id;
	}

	/** setQuestionId sets Question ID. */
	function setQuestionId($value) {
		$this->question_id = $value;

		return $this;
	}
	/** getSeq is Seq. */
	function getSeq() {
		return $this->seq;
	}

	/** setSeq sets Seq. */
	function setSeq($value) {
		$this->seq = $value;

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
	/** getToolId is Tool ID. */
	function getToolId() {
		return $this->tool_id;
	}

	/** setToolId sets Tool ID. */
	function setToolId($value) {
		$this->tool_id = $value;

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
	/** getData is Data. */
	function getData() {
		return $this->data;
	}

	/** setData sets Data. */
	function setData($value) {
		$this->data = $value;

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
		$out = [];

		if (!is_array($found)) {
			return $out;
		}

		foreach ($found as $row) {
			$out[] = new NoteEvent($row);
		}

		return $out;
	}
}
