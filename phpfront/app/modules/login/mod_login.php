<?php

// @route GET /login

require "ext/autoload.php";

// mod_login is the sign in page: the form, and the path to open once the
// credentials are read.
//
// The page is what moves here. Reading the credentials, minting the session
// and dropping it stay in Go: the Guard sits in front of this application and
// of /api alike, so POST /login and /logout answer at the root and the form
// posts there. This route answers GET and writes nothing.
class mod_login implements Module {
	function name() {
		return "login";
	}

	function title() {
		return "Sign in";
	}

	function icon() {
		return "logout";
	}

	function tables() {
		return array("user");
	}

	function serve() {
		$chrome = new Task\Chrome;

		if ($chrome->user !== null) {
			Task\Route::seeOther("/");

			return;
		}

		$chrome->heading("Sign in");

		$view = new Task\View($this, $chrome);
		$asked = Task\Route::input("back");
		$back = Task\Route::safeBack($asked);

		// POST /login answers a wrong password with a redirect back here:
		// the check is Go's and the page that says so is this one.
		$error = "";
		if (Task\Route::input("error") !== "") {
			$error = "Wrong username or password.";
		}

		$view->page(array("back" => $back, "error" => $error));
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_login;

	$mod->serve();
}
