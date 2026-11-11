<?php

// @route GET /settings/secrets/new

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_secret_new is the form that adds a variable to the environment of runs:
// a secret, whose value is encrypted with the instance's key and never shown
// again, or an env entry, whose value the Secrets page lists.
//
// The page reads and does not write. The form posts to /api/secrets, which
// gates it again and encrypts the value; the key does not enter PHP and
// nothing of that is here.
class mod_secret_new implements Module {
	/** TITLE is the heading, the last crumb and what the browser tab says. */
	const TITLE = "Add variable";

	/** LEAD is the sentence under the heading. */
	const LEAD = "A variable for the environment of runs. A secret's value is encrypted with the instance's encryption key and never shown again; an env value is listed on the Secrets page. One applies to every workspace, or to one workspace. Saving an existing name replaces its value.";

	/** NEXT is where the script goes once the variable is added. */
	const NEXT = "/settings/secrets?notice=Variable+added";

	/** SUBMIT is what the button of the form says. */
	const SUBMIT = "Add variable";

	/** HINT is what the Value field says the value is for. */
	const HINT = "what runs get; may be empty";

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
		return array("workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("secret/edit")) {
			$view->refused();

			return;
		}

		$chrome->heading(self::TITLE, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::SECRETS, "href" => Task\Secret::PAGE),
			array("text" => self::TITLE, "href" => ""),
		));

		// ?scope= opens the form on one workspace, which the Secrets page
		// links with while it is filtered to that workspace. A scope that
		// names no workspace is kept and the picker marks nothing, as the Go
		// page keeps what it was given.
		$scope = Task\Route::input("scope");

		if ($scope === "") {
			$scope = Task\Workspace::EVERY;
		}

		$workspaces = new Task\Workspace($chrome->db);
		$head = new Task\Head(self::TITLE, self::LEAD);
		$steps = Task\Setup::steps($chrome, Task\Setup::SECRETS);

		// The form is shared with mod_secret_edit, so the keys of a row it
		// opens are answered empty here: no ID to post back, no name to start
		// from, no value to draw, and Secret checked, which is what a variable
		// a reader adds is.
		$view->page(array(
			"head" => $head->data(),
			"id" => "",
			"name" => "",
			"next" => self::NEXT,
			"submit" => self::SUBMIT,
			"hint" => self::HINT,
			"blank" => false,
			"secret" => true,
			"content" => "",
			"options" => Task\Secret::options($scope, $workspaces->all()),
			"cancel" => Task\Route::href(Task\Secret::PAGE),
			"trail" => Task\Setup::trail($steps),
		), "secret_form.tpl");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_secret_new;

	$mod->serve();
}
