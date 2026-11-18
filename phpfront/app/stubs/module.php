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
	public $tables = [];
	public $sidebar = false;
	public $sidebarHref = "";
	public $sidebarRule = "module";

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

	function enableSideBar() {
		$this->sidebar = true;
	}

	function setSideBarHref($href) {
		$this->sidebarHref = trim($href);
	}

	function setSideBarRule($rule) {
		$this->sidebarRule = trim($rule);
	}

	function getSidebar() {
		return $this->sidebar;
	}

	function getSideBarHref() {
		return $this->sidebarHref !== "" ? $this->sidebarHref : $this->href();
	}

	function getSideBarRule() {
		return $this->sidebarRule;
	}

	function getTitle() {
		return $this->label();
	}

	function getDescription() {
		return $this->description;
	}

	function getIcon() {
		return $this->icon;
	}

	function href() {
		return "/m/" . $this->name();
	}

	function name() {
		return $this->vendor . "/" . $this->plugin;
	}

	function label() {
		return $this->title != "" ? $this->title : $this->plugin;
	}
}
