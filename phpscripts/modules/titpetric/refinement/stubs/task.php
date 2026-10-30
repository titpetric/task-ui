<?php

// Task is a stub of the host static calls, for the linter alone. Nothing
// requires this file at runtime.
//
// The three calls are Go functions whose registered names hold `::`, which is
// the first thing a static call resolves, so they reach Go without any class
// called Task declaring them. The constants are the one part of this that is
// real: task-ui registers a `Task` class holding them.
//
// Keep it in step with phpscript/bind.go.
class Task {
	/** The instance's version. */
	const VERSION = "";

	/** Where every module is served from. */
	const PREFIX = "/m";

	/** Records a module and its root, and adds its sidebar row. */
	static function register($module) {
	}

	/** Applies a directory of *.up.sql into task-ui's database. */
	static function migrate($dir) {
	}

	/** Starts a Claude run over a comment and an action file, returning its id. */
	static function dispatch($comment, $action) {
		return "";
	}

	/** The artifact tree's files, by path. */
	static function artifacts() {
		return array();
	}

	/** The text of one artifact, or "". */
	static function artifact($path) {
		return "";
	}

	/** Renders markdown as the instance's own pages do. */
	static function markdown($src) {
		return "";
	}
}
