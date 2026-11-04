<?php

// @route GET /settings/users/{id}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_user_by_id sends a link that names a user by ID to the page that holds
// their settings, which is their edit page.
//
// The permissions page links a member chip by ID, because that is what it has;
// the reader lands on the name. Your own ID opens without user/view, the way
// your profile does, and an ID no user holds answers Not found.
class mod_user_by_id implements Module {
	/** GONE is what the page says for an ID no user holds. */
	const GONE = "No user with that ID.";

	function name() {
		return "settings";
	}

	function title() {
		return "Settings";
	}

	function icon() {
		return "settings";
	}

	function tables() {
		return array("user");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);
		$id = Task\Route::param("id");
		$reader = $chrome->user;
		$own = $reader !== null && $reader->getId() === $id;

		if (!$own && !$chrome->allowed("user/view")) {
			$view->refused();

			return;
		}

		$users = new Task\User($chrome->db);
		$user = $users->get($id);

		if ($user === null) {
			$view->missing(self::GONE);

			return;
		}

		$name = Task\Route::pathEscape($user->getName());

		Task\Route::found("/users/" . $name . "/edit");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_user_by_id;

	$mod->serve();
}
