<?php

namespace Task;

// Chrome is what every page shares: who is reading, what they may do, the
// sidebar, the crumbs and what the page reports.
//
// It is built once per request and handed to the theme as one scope, so a
// component reads its own slice of it and nothing else.
class Chrome {
	var $db;
	var $user = null;
	var $access = null;

	/** title is the page's own, before the instance's name. */
	var $title = "";

	/** crumbs are the trail above the content: text, and a href or "". */
	var $crumbs = array();

	var $users = null;
	var $people = null;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
		$this->users = new User($this->db);

		$name = Host::whoami();

		if ($name !== "") {
			$this->user = $this->users->byName($name);
		}

		$this->access = $this->users->access($this->user);
	}

	/** heading sets the page's title and its trail in one call. */
	function heading($title, $crumbs = array()) {
		$this->title = $title;
		$this->crumbs = $crumbs;

		return $this;
	}

	/** allowed reports whether the reader may do what a path names. */
	function allowed($path) {
		return $this->access->allowed($path);
	}

	/** name is the reader's name, "" when nobody is signed in. */
	function name() {
		return $this->user === null ? "" : $this->user->getName();
	}

	/** who draws any user: the avatar, the name and the menu beside them. */
	function who() {
		if ($this->people === null) {
			$this->people = new Who($this->access, $this->users->roles());
		}

		return $this->people;
	}

	/** message renders a note's text, linking the mentions the reader may open. */
	function message($text) {
		return Host::message($text, Route::href("/users/"), $this->who()->linkable());
	}

	/**
	 * data is the scope the theme reads.
	 *
	 * folders and recents come back empty: the folder tree arrives with the
	 * folders module and the rail with the runs module, and both components
	 * draw their own empty line until then.
	 */
	function data() {
		$path = "/" . Route::path();
		$sidebar = new Sidebar($this->db);

		return array(
			"base" => Host::base(),
			"version" => Host::version(),
			"assets" => Host::assets(),
			"assetVersion" => Host::assetVersion(),
			"title" => $this->title,
			"path" => $path,
			"crumbs" => $this->crumbs,
			"notice" => Route::input("notice"),
			"error" => Route::input("error"),
			"user" => $this->reader(),
			"can" => $this->access->effective(),
			"settings" => Sidebar::settings($this->access),
			"sidebar" => $sidebar->items($this->access, $path),
			"folders" => array(),
			"recents" => array("href" => "/runs", "groups" => array()),
			"running" => count(Host::running()) > 0,
		);
	}

	/** reader is the signed in user, as the header draws them. */
	function reader() {
		if ($this->user === null) {
			return null;
		}

		return array(
			"name" => $this->user->getName(),
			"role" => User::role($this->user),
			"initials" => User::initials($this->user),
		);
	}
}
