<?php

// @route GET /settings/users/new

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_user_new is the form that adds a user to the password file: their name,
// their first password and the folders they see.
//
// The page reads and does not write. The form puts to the API, which gates the
// change again and hashes the password.
class mod_user_new implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "Adds a user to the password file. They can sign in right away. They start as a viewer, the lowest role; an admin sets the role and the flags on their profile.";

	/**
	 * NAME is what a user name may be, as the Name field's pattern.
	 *
	 * MiniTPL compiles any {word} into an echo, and the quantifier {0,31} is
	 * one: a template cannot write the literal, so it arrives as a field of
	 * the scope the way mod_user_edit's token placeholder does.
	 */
	const NAME = "[a-z0-9][a-z0-9\\-]{0,31}";

	/** GONE is what the page says where there is no password file to add to. */
	const GONE = "task-ui has no password file to add users to.";

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
		return array();
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("user/edit")) {
			$view->refused();

			return;
		}

		$instance = Task\Host::instance(false);
		$adds = Task\Row::text($instance, "htpasswd") !== "";

		if (!$adds) {
			$view->missing(self::GONE);

			return;
		}

		$chrome->heading("Add user", array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::USERS, "href" => "/settings/users"),
			array("text" => "Add user", "href" => ""),
		));

		$head = new Task\Head("Add user", self::LEAD);
		$steps = Task\Setup::steps($chrome, Task\Setup::USERS);
		$choices = Task\Folders::choices($chrome->all(), array());

		$view->page(array(
			"head" => $head->data(),
			"pattern" => self::NAME,
			"cancel" => Task\Route::href("/settings/users"),
			"pick" => array(
				"all" => true,
				"label" => true,
				"choices" => $choices,
			),
			"trail" => Task\Setup::trail($steps),
		), "user_form.tpl");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_user_new;

	$mod->serve();
}
