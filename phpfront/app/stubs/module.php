<?php

namespace Task;

// Module declares the registration a modinfo.php builds, for the linter and
// for the test fixtures. Nothing requires this at runtime: the constructor is
// a Go function registered as `Task\Module`.
//
// Keep it in step with phpfront/module.go.
class Module {
	public $vendor = "";
	public $plugin = "";
	public $title = "";
	public $description = "";
	public $icon = "";
	public $version = "";
	public $tables = array();

	function setName($name) {
		$half = explode("/", $name);
		$this->vendor = $half[0];
		$this->plugin = count($half) > 1 ? $half[1] : "";
	}

	function setTitle($title) {
		$this->title = trim($title);
	}

	function setDescription($description) {
		$this->description = trim($description);
	}

	function setIcon($icon) {
		$this->icon = trim($icon);
	}

	function setVersion($version) {
		$this->version = trim($version);
	}

	function setTables($tables) {
		$this->tables = $tables;
	}

	function name() {
		return $this->vendor . "/" . $this->plugin;
	}

	function label() {
		return $this->title != "" ? $this->title : $this->plugin;
	}
}
