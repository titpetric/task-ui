<?php

// Database is a stub of the connection binding, for the linter alone.
// Nothing requires this file at runtime: the real class is phpscript's
// own, and the connection it opens is the one task-ui's provider
// answers, "task-ui".
//
// The methods are the ones a module calls. phpscript resolves a method
// name three ways, so get_all here is GetAll in Go.
class Database {
	/** Opens the named connection. */
	function __construct($name) {
	}

	/** Runs a statement that returns no rows. */
	function query($sql) {
	}

	/** Returns one row, or null. */
	function get($sql) {
		return null;
	}

	/** Returns every row. */
	function get_all($sql) {
		return array();
	}

	/** Writes a row, the array keyed by column. */
	function insert($table, $value) {
	}
}
