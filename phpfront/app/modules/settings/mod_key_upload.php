<?php

// @route GET /settings/keys/upload

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_key_upload is the form that saves a key that already exists: the private
// half a reader pastes or picks, what it is for and whose runs get it.
//
// The page reads and does not write. The private half goes to
// /api/keys/upload, which gates it again and is the only thing that ever sees
// it; this page only draws the form.
class mod_key_upload implements Module {
	/** TITLE is the heading, the crumb and what the browser tab says. */
	const TITLE = "Upload key";

	/** PAGE is the page a key that was added returns to. */
	const PAGE = "/settings/keys";

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

		if (!$chrome->allowed("key/edit")) {
			$view->refused();

			return;
		}

		$chrome->heading(self::TITLE, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::KEYS, "href" => self::PAGE),
			array("text" => self::TITLE, "href" => ""),
		));

		$folder = Task\Key::folder();
		$dir = Task\Row::text($folder, "dir");
		$head = new Task\Head(self::TITLE, self::lead($dir));
		$steps = Task\Setup::steps($chrome, Task\Setup::KEYS);
		$users = new Task\User($chrome->db);
		$everyone = $users->all();
		$for = Task\Key::forUser();
		$field = Task\Key::policies(array(), array($for), $everyone);
		$open = Task\Key::openName($for);

		$view->page(array(
			"head" => $head->data(),
			"name" => Task\Key::nameField($open),
			"who" => $field->data(),
			"cancel" => Task\Route::href(self::PAGE),
			"trail" => Task\Setup::trail($steps),
		), "key_upload.tpl");
	}

	/**
	 * lead is the sentence under the heading, as HTML. It is built here and not
	 * in the template: a template writes the newline that sits before a tag, so
	 * a sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($dir) {
		$where = htmlspecialchars($dir, ENT_QUOTES);

		return "Saves a private key you already use in " . $where . ". task-ui reads its algorithm and size and works out the public key; the name gets the algorithm, as <code>deploy_rsa4096</code>. The folder refuses a key it holds, under any name: a key has one purpose.";
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_key_upload;

	$mod->serve();
}
