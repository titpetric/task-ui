<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// SearchDocument is a row of the search_document table.
class SearchDocument {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "search_document";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"kind",
		"path",
		"file",
		"folder_id",
		"workspace",
		"project",
		"target",
		"name",
		"dir",
		"title",
		"hash",
		"size",
		"modified_at",
		"indexed_at",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("modified_at", "indexed_at");
	/** ID */
	var $id = "";
	/** Kind */
	var $kind = "";
	/** Path */
	var $path = "";
	/** File */
	var $file = "";
	/** Folder ID */
	var $folder_id = "";
	/** Workspace */
	var $workspace = "";
	/** Project */
	var $project = "";
	/** Target */
	var $target = "";
	/** Name */
	var $name = "";
	/** Dir */
	var $dir = "";
	/** Title */
	var $title = "";
	/** Hash */
	var $hash = "";
	/** Size */
	var $size = 0;
	/** Modified At */
	var $modified_at = "";
	/** Indexed At */
	var $indexed_at = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->kind = \Task\Row::text($row, "kind");
		$this->path = \Task\Row::text($row, "path");
		$this->file = \Task\Row::text($row, "file");
		$this->folder_id = \Task\Row::text($row, "folder_id");
		$this->workspace = \Task\Row::text($row, "workspace");
		$this->project = \Task\Row::text($row, "project");
		$this->target = \Task\Row::text($row, "target");
		$this->name = \Task\Row::text($row, "name");
		$this->dir = \Task\Row::text($row, "dir");
		$this->title = \Task\Row::text($row, "title");
		$this->hash = \Task\Row::text($row, "hash");
		$this->size = \Task\Row::number($row, "size");
		$this->modified_at = \Task\Row::stamp($row, "modified_at");
		$this->indexed_at = \Task\Row::stamp($row, "indexed_at");
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
	/** getPath is Path. */
	function getPath() {
		return $this->path;
	}

	/** setPath sets Path. */
	function setPath($value) {
		$this->path = $value;

		return $this;
	}
	/** getFile is File. */
	function getFile() {
		return $this->file;
	}

	/** setFile sets File. */
	function setFile($value) {
		$this->file = $value;

		return $this;
	}
	/** getFolderId is Folder ID. */
	function getFolderId() {
		return $this->folder_id;
	}

	/** setFolderId sets Folder ID. */
	function setFolderId($value) {
		$this->folder_id = $value;

		return $this;
	}
	/** getWorkspace is Workspace. */
	function getWorkspace() {
		return $this->workspace;
	}

	/** setWorkspace sets Workspace. */
	function setWorkspace($value) {
		$this->workspace = $value;

		return $this;
	}
	/** getProject is Project. */
	function getProject() {
		return $this->project;
	}

	/** setProject sets Project. */
	function setProject($value) {
		$this->project = $value;

		return $this;
	}
	/** getTarget is Target. */
	function getTarget() {
		return $this->target;
	}

	/** setTarget sets Target. */
	function setTarget($value) {
		$this->target = $value;

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
	/** getDir is Dir. */
	function getDir() {
		return $this->dir;
	}

	/** setDir sets Dir. */
	function setDir($value) {
		$this->dir = $value;

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
	/** getHash is Hash. */
	function getHash() {
		return $this->hash;
	}

	/** setHash sets Hash. */
	function setHash($value) {
		$this->hash = $value;

		return $this;
	}
	/** getSize is Size. */
	function getSize() {
		return $this->size;
	}

	/** setSize sets Size. */
	function setSize($value) {
		$this->size = $value;

		return $this;
	}
	/** getModifiedAt is Modified At. */
	function getModifiedAt() {
		return $this->modified_at;
	}

	/** setModifiedAt sets Modified At. */
	function setModifiedAt($value) {
		$this->modified_at = $value;

		return $this;
	}
	/** getIndexedAt is Indexed At. */
	function getIndexedAt() {
		return $this->indexed_at;
	}

	/** setIndexedAt sets Indexed At. */
	function setIndexedAt($value) {
		$this->indexed_at = $value;

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
			$out[] = new SearchDocument($row);
		}

		return $out;
	}
}
