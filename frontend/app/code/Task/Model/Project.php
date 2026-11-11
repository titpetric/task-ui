<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Project is a row of the project table.
class Project {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "project";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"folder",
		"title",
		"description",
		"source_url",
		"branch",
		"allow",
		"commands",
		"sources",
		"job_allow",
		"created_at",
		"updated_at",
		"shallow",
		"sidebar_hidden",
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
	/** Source URL */
	var $source_url = "";
	/** Branch */
	var $branch = "";
	/** Allow */
	var $allow = "";
	/** Commands */
	var $commands = "";
	/** Sources */
	var $sources = "";
	/** Job Allow */
	var $job_allow = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Shallow */
	var $shallow = false;
	/** Sidebar Hidden */
	var $sidebar_hidden = false;

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->folder = \Task\Row::text($row, "folder");
		$this->title = \Task\Row::text($row, "title");
		$this->description = \Task\Row::text($row, "description");
		$this->source_url = \Task\Row::text($row, "source_url");
		$this->branch = \Task\Row::text($row, "branch");
		$this->allow = \Task\Row::text($row, "allow");
		$this->commands = \Task\Row::text($row, "commands");
		$this->sources = \Task\Row::text($row, "sources");
		$this->job_allow = \Task\Row::text($row, "job_allow");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->shallow = \Task\Row::flag($row, "shallow");
		$this->sidebar_hidden = \Task\Row::flag($row, "sidebar_hidden");
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
	/** getSourceUrl is Source URL. */
	function getSourceUrl() {
		return $this->source_url;
	}

	/** setSourceUrl sets Source URL. */
	function setSourceUrl($value) {
		$this->source_url = $value;

		return $this;
	}
	/** getBranch is Branch. */
	function getBranch() {
		return $this->branch;
	}

	/** setBranch sets Branch. */
	function setBranch($value) {
		$this->branch = $value;

		return $this;
	}
	/** getAllow is Allow. */
	function getAllow() {
		return $this->allow;
	}

	/** setAllow sets Allow. */
	function setAllow($value) {
		$this->allow = $value;

		return $this;
	}
	/** getCommands is Commands. */
	function getCommands() {
		return $this->commands;
	}

	/** setCommands sets Commands. */
	function setCommands($value) {
		$this->commands = $value;

		return $this;
	}
	/** getSources is Sources. */
	function getSources() {
		return $this->sources;
	}

	/** setSources sets Sources. */
	function setSources($value) {
		$this->sources = $value;

		return $this;
	}
	/** getJobAllow is Job Allow. */
	function getJobAllow() {
		return $this->job_allow;
	}

	/** setJobAllow sets Job Allow. */
	function setJobAllow($value) {
		$this->job_allow = $value;

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
	/** getShallow is Shallow. */
	function getShallow() {
		return $this->shallow;
	}

	/** setShallow sets Shallow. */
	function setShallow($value) {
		$this->shallow = $value;

		return $this;
	}
	/** getSidebarHidden is Sidebar Hidden. */
	function getSidebarHidden() {
		return $this->sidebar_hidden;
	}

	/** setSidebarHidden sets Sidebar Hidden. */
	function setSidebarHidden($value) {
		$this->sidebar_hidden = $value;

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
			$out[] = new Project($row);
		}

		return $out;
	}
}
