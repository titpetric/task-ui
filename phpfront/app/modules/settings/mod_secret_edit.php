<?php

// @route GET /settings/secrets/edit

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_secret_edit is the form that changes a variable of the run environment:
// what it is called, which workspace its runs are in, its value and whether
// the value is kept encrypted.
//
// A secret's value is not drawn: the form says blank keeps it, and the API
// leaves the stored value alone where the field is empty. An env value is
// drawn, which the vault opens: Task\Host::variable() answers it by row ID,
// and a value the vault will not open is the shared 404 rather than a form
// that would save armour over armour.
class mod_secret_edit implements Module {
	/** TITLE is what the heading and the tab say before the variable's name. */
	const TITLE = "Edit ";

	/** LEAD is the sentence under the heading, which both forms share. */
	const LEAD = "A variable for the environment of runs. A secret's value is encrypted with the instance's encryption key and never shown again; an env value is listed on the Secrets page. One applies to every workspace, or to one workspace. Saving an existing name replaces its value.";

	/** NEXT is where the script goes once the variable is saved. */
	const NEXT = "/settings/secrets?notice=Variable+saved";

	/** SUBMIT is what the button of the form says. */
	const SUBMIT = "Save";

	/** HINT is what the Value field says of a value that is drawn. */
	const HINT = "what runs get; may be empty";

	/** KEEPS is what it says of a secret's, which the form never draws. */
	const KEEPS = "blank keeps the current value";

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
		return array("secret", "workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		// The gate is secret/view and not secret/edit: the form is a page of
		// the variable, and what saves it is the API, which gates again.
		if (!$chrome->allowed("secret/view")) {
			$view->refused();

			return;
		}

		$id = Task\Route::input("id");
		$scope = Task\Route::input("scope");
		$name = Task\Route::input("name");
		$secrets = new Task\Secret($chrome->db);
		$row = $secrets->edited($id, $scope, $name);

		if ($row === null) {
			$view->missing(Task\Secret::says($name, $scope));

			return;
		}

		$opened = Task\Host::variable($row->getId());

		if (Task\Row::text($opened, "error") !== "") {
			$view->missing(Task\Secret::says($name, $scope));

			return;
		}

		$secret = $row->getIsSecret() !== 0;
		$title = self::TITLE . $row->getName();

		$chrome->heading($title, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::SECRETS, "href" => Task\Secret::PAGE),
			array("text" => $row->getName(), "href" => ""),
		));

		$workspaces = new Task\Workspace($chrome->db);
		$head = new Task\Head($title, self::LEAD);
		$steps = Task\Setup::steps($chrome, Task\Setup::SECRETS);

		$view->page(array(
			"head" => $head->data(),
			"id" => $row->getId(),
			"name" => $row->getName(),
			"next" => self::NEXT,
			"submit" => self::SUBMIT,
			"hint" => $secret ? self::KEEPS : self::HINT,
			"blank" => $secret,
			"secret" => $secret,
			"content" => Task\Row::text($opened, "content"),
			"options" => Task\Secret::options($row->getScope(), $workspaces->all()),
			"cancel" => Task\Route::href(Task\Secret::PAGE),
			"trail" => Task\Setup::trail($steps),
		), "secret_form.tpl");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_secret_edit;

	$mod->serve();
}
