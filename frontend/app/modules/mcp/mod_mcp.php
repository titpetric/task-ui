<?php

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_mcp owns the tokens a client signs in with over MCP: the kind column
// that tells them from the ones `task-ui token` writes, the uniqueness that
// lets a user hold several of them, and the reads a page makes of them.
//
// It draws no page of its own and carries no @route. The block a reader sees
// is MCP access on the user edit page, which is the profile module's, and the
// writes are the API's: a POST to /api/users/{id}/mcp issues one and a DELETE
// of /api/users/{id}/mcp/{token} removes one. The module exists because the
// data model does: a schema folder belongs to the module that owns the table's
// shape, not to the Go tree beside it.
class mod_mcp implements Module {
	function name() {
		return "mcp";
	}

	function title() {
		return "MCP tokens";
	}

	function icon() {
		return "plug";
	}

	function tables() {
		return array("user_auth_token", "user");
	}

	/**
	 * serve answers nothing: no route reaches this module, so the only way
	 * here is a request that asked for a path the router never bound.
	 */
	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		$view->missing("MCP access is a block of a user's page, not a page of its own.");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_mcp;

	$mod->serve();
}
