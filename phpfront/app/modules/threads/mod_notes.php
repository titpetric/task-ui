<?php

// @route GET /notes

require "ext/autoload.php";

// mod_notes answers the address the board used to hold: the board is the
// Threads timeline now, and a link written before the move lands on it.
//
// A note's own anchor rides along, so /notes#<id> opens the timeline at the
// thread it names.
class mod_notes implements Module {
	function name() {
		return "threads";
	}

	function title() {
		return "Threads";
	}

	function icon() {
		return "note";
	}

	function tables() {
		return array("note");
	}

	function serve() {
		Task\Route::go("/threads");
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_notes;

	$mod->serve();
}
