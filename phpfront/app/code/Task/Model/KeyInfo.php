<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// KeyInfo is a row of the key_info table.
class KeyInfo {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "key_info";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"name",
		"purpose",
		"algo",
		"bits",
		"fingerprint",
		"scope",
		"comment",
		"created_at",
		"updated_at",
		"backup",
		"backup_fingerprint",
		"rotated_at",
	);

	/** primary is the primary key columns. */
	public static $primary = array("name");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at", "rotated_at");
	/** Name */
	var $name = "";
	/** Purpose */
	var $purpose = "";
	/** Algo */
	var $algo = "";
	/** Bits */
	var $bits = 0;
	/** Fingerprint */
	var $fingerprint = "";
	/** Scope */
	var $scope = "";
	/** Comment */
	var $comment = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Backup */
	var $backup = "";
	/** Backup Fingerprint */
	var $backup_fingerprint = "";
	/** Rotated At */
	var $rotated_at = "";

	function __construct($row = array()) {
		$this->name = \Task\Row::text($row, "name");
		$this->purpose = \Task\Row::text($row, "purpose");
		$this->algo = \Task\Row::text($row, "algo");
		$this->bits = \Task\Row::number($row, "bits");
		$this->fingerprint = \Task\Row::text($row, "fingerprint");
		$this->scope = \Task\Row::text($row, "scope");
		$this->comment = \Task\Row::text($row, "comment");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->backup = \Task\Row::text($row, "backup");
		$this->backup_fingerprint = \Task\Row::text($row, "backup_fingerprint");
		$this->rotated_at = \Task\Row::stamp($row, "rotated_at");
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
	/** getPurpose is Purpose. */
	function getPurpose() {
		return $this->purpose;
	}

	/** setPurpose sets Purpose. */
	function setPurpose($value) {
		$this->purpose = $value;

		return $this;
	}
	/** getAlgo is Algo. */
	function getAlgo() {
		return $this->algo;
	}

	/** setAlgo sets Algo. */
	function setAlgo($value) {
		$this->algo = $value;

		return $this;
	}
	/** getBits is Bits. */
	function getBits() {
		return $this->bits;
	}

	/** setBits sets Bits. */
	function setBits($value) {
		$this->bits = $value;

		return $this;
	}
	/** getFingerprint is Fingerprint. */
	function getFingerprint() {
		return $this->fingerprint;
	}

	/** setFingerprint sets Fingerprint. */
	function setFingerprint($value) {
		$this->fingerprint = $value;

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
	/** getComment is Comment. */
	function getComment() {
		return $this->comment;
	}

	/** setComment sets Comment. */
	function setComment($value) {
		$this->comment = $value;

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
	/** getBackup is Backup. */
	function getBackup() {
		return $this->backup;
	}

	/** setBackup sets Backup. */
	function setBackup($value) {
		$this->backup = $value;

		return $this;
	}
	/** getBackupFingerprint is Backup Fingerprint. */
	function getBackupFingerprint() {
		return $this->backup_fingerprint;
	}

	/** setBackupFingerprint sets Backup Fingerprint. */
	function setBackupFingerprint($value) {
		$this->backup_fingerprint = $value;

		return $this;
	}
	/** getRotatedAt is Rotated At. */
	function getRotatedAt() {
		return $this->rotated_at;
	}

	/** setRotatedAt sets Rotated At. */
	function setRotatedAt($value) {
		$this->rotated_at = $value;

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
			$out[] = new KeyInfo($row);
		}

		return $out;
	}
}
