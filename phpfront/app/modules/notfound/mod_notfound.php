<?php

// @route GET /{rest...}
// @route POST /{rest...}

require "ext/autoload.php";

// mod_notfound is the catch-all: the page for a path no other route claimed.
//
// chi matches a literal and a parameter ahead of a wildcard, so this answers
// only what nothing else did. The Go router draws the same page from
// r.NotFound, behind the same guard, and the reason reads the same because
// Task\Route::path() drops the mount prefix: the cutover changes no string.
//
// Without it an unrouted path under the prefix answers chi's own plain text,
// which is no page at all.
class mod_notfound implements Module {
	function name() {
		return "notfound";
	}

	function title() {
		return "Not found";
	}

	function icon() {
		return "file";
	}

	function tables() {
		return array();
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);
		$path = Task\Route::path();

		$view->missing("There is nothing at /" . $path . ".");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_notfound;

	$mod->serve();
}
