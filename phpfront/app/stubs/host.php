<?php

namespace Task;

// Host declares the host calls for the linter and for the test fixtures.
// Nothing requires this at runtime: the real calls are Go functions whose
// registered names hold `::`, which is what a static call resolves first.
//
// Each answer is a static a fixture writes, so a fixture drives the host it
// runs without one.
//
// Keep it in step with phpfront/bind.go.
class Host {
	public static $base = "";
	public static $version = "";
	public static $assetVersion = "";
	public static $whoami = "";
	public static $folders = array();
	public static $running = array();
	public static $dirty = array();
	public static $rules = array();
	public static $published = array();
	public static $artifacts = array();

	/** The path the router is mounted at, "" once the prefix is gone. */
	static function base() {
		return self::$base;
	}

	/** The instance's version. */
	static function version() {
		return self::$version;
	}

	/** Where the static files answer, outside the base. */
	static function assets() {
		return "/assets";
	}

	/** The content hash of the static files, which their URLs carry as ?v=. */
	static function assetVersion() {
		return self::$assetVersion;
	}

	/** The name the Guard resolved, "" when nobody is signed in. */
	static function whoami() {
		return self::$whoami;
	}

	/** The folder catalogue: each folder and the jobs it declares. */
	static function folders() {
		return self::$folders;
	}

	/** The runs in progress, by the key their job is listed under. */
	static function running() {
		return self::$running;
	}

	/** The changed and untracked files under a checkout directory. */
	static function dirty($dir) {
		return isset(self::$dirty[$dir]) ? self::$dirty[$dir] : 0;
	}

	/** The permission titles and path lists of config/permissions.yml. */
	static function rules() {
		return self::$rules;
	}

	/** Sends an event on the stream every live label reads. */
	static function publish($topic, $payload) {
		self::$published[] = array("topic" => $topic, "payload" => $payload);

		return true;
	}

	/** Markdown as HTML, through the renderer that owns the content policy. */
	static function markdown($src, $folder = "") {
		return "<p>" . htmlspecialchars($src, ENT_QUOTES) . "</p>";
	}

	/**
	 * A note's text: the same renderer with a line break kept and @name
	 * marked. A name in link becomes a link under prefix.
	 */
	static function message($src, $prefix = "", $link = array()) {
		return "<p>" . htmlspecialchars($src, ENT_QUOTES) . "</p>";
	}

	/** The files under one folder of the artifact tree, "" for the whole tree. */
	static function artifacts($folder = "") {
		return isset(self::$artifacts[$folder]) ? self::$artifacts[$folder] : array();
	}

	/** The text of one artifact, from the top of the tree, or "". */
	static function artifact($path) {
		return "";
	}
}
