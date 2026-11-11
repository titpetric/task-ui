<?php

// Task declares the host statics a module calls, for the linter and for the
// test fixtures: `registered` is what a fixture reads back.
//
// Keep it in step with phpfront/module_bind.go.
class Task {
	const VERSION = "";
	const PREFIX = "/m";

	public static $registered = array();

	public static function register($module) {
		self::$registered[] = $module;
		return true;
	}

	public static function dispatch($comment, $action, $folder) {
		return "";
	}

	public static function markdown($src, $folder) {
		return $src;
	}

	public static function run($note) {
		return "";
	}

	public static function artifacts($folder) {
		return array();
	}

	public static function artifact($rel) {
		return "";
	}
}
