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

	/**
	 * Starts a Claude run over a comment and an action file, returning its
	 * id. The folder is where in the artifact tree the run works, and the
	 * only part of it the prompt names; empty is the top of the tree.
	 */
	static function dispatch($comment, $action, $folder = "") {
		return "";
	}

	/**
	 * The files under one folder of the artifact tree, by path relative to
	 * that folder. Empty is the whole tree.
	 */
	static function artifacts($folder = "") {
		return array();
	}

	/** The text of one artifact, from the top of the tree, or "". */
	static function artifact($path) {
		return "";
	}

	/**
	 * The run a dispatch turned into, by the note it answered with, or
	 * "" for one that has not started. It is what a status label names:
	 * <span class="state" data-run-id="..."></span>.
	 */
	static function run($note) {
		return "";
	}

	/**
	 * Renders markdown as the instance's own pages do. The folder is what a
	 * relative link resolves under, for a report written inside one.
	 */
	static function markdown($src, $folder = "") {
		return "";
	}
}
