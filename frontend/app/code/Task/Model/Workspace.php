<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Workspace is a row of the workspace table.
class Workspace {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "workspace";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"folder",
		"title",
		"description",
		"created_at",
		"updated_at",
		"home_project_id",
		"home_url",
		"home_branch",
		"hours",
		"shallow",
		"sidebar_limit",
		"sort",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Folder */
	var $folder = "";
	/** Title */
	var $title = "";
	/** Description */
	var $description = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Home Project ID */
	var $home_project_id = "";
	/** Home URL */
	var $home_url = "";
	/** Home Branch */
	var $home_branch = "";
	/** Hours */
	var $hours = "";
	/** Shallow */
	var $shallow = false;
	/** Sidebar Limit */
	var $sidebar_limit = 0;
	/** Sort */
	var $sort = 0;

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->folder = \Task\Row::text($row, "folder");
		$this->title = \Task\Row::text($row, "title");
		$this->description = \Task\Row::text($row, "description");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->home_project_id = \Task\Row::text($row, "home_project_id");
		$this->home_url = \Task\Row::text($row, "home_url");
		$this->home_branch = \Task\Row::text($row, "home_branch");
		$this->hours = \Task\Row::text($row, "hours");
		$this->shallow = \Task\Row::flag($row, "shallow");
		$this->sidebar_limit = \Task\Row::number($row, "sidebar_limit");
		$this->sort = \Task\Row::number($row, "sort");
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
	/** getFolder is Folder. */
	function getFolder() {
		return $this->folder;
	}

	/** setFolder sets Folder. */
	function setFolder($value) {
		$this->folder = $value;

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
	/** getHomeProjectId is Home Project ID. */
	function getHomeProjectId() {
		return $this->home_project_id;
	}

	/** setHomeProjectId sets Home Project ID. */
	function setHomeProjectId($value) {
		$this->home_project_id = $value;

		return $this;
	}
	/** getHomeUrl is Home URL. */
	function getHomeUrl() {
		return $this->home_url;
	}

	/** setHomeUrl sets Home URL. */
	function setHomeUrl($value) {
		$this->home_url = $value;

		return $this;
	}
	/** getHomeBranch is Home Branch. */
	function getHomeBranch() {
		return $this->home_branch;
	}

	/** setHomeBranch sets Home Branch. */
	function setHomeBranch($value) {
		$this->home_branch = $value;

		return $this;
	}
	/** getHours is Hours. */
	function getHours() {
		return $this->hours;
	}

	/** setHours sets Hours. */
	function setHours($value) {
		$this->hours = $value;

		return $this;
	}
	/** getShallow is Shallow. */
	function getShallow() {
		return $this->shallow;
	}

	/** setShallow sets Shallow. */
	function setShallow($value) {
		$this->shallow = $value;

		return $this;
	}
	/** getSidebarLimit is Sidebar Limit. */
	function getSidebarLimit() {
		return $this->sidebar_limit;
	}

	/** setSidebarLimit sets Sidebar Limit. */
	function setSidebarLimit($value) {
		$this->sidebar_limit = $value;

		return $this;
	}
	/** getSort is Sort. */
	function getSort() {
		return $this->sort;
	}

	/** setSort sets Sort. */
	function setSort($value) {
		$this->sort = $value;

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
			$out[] = new Workspace($row);
		}

		return $out;
	}
}
