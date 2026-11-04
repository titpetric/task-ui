<?php

// @route GET /settings/keys/generate

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_key_generate is the form that makes a key: its name, its algorithm, what
// it is for and whose runs get it.
//
// The page reads and does not write. The form posts to /api/keys/generate,
// which gates it again and is the only place the private half is made; nothing
// of it is here.
class mod_key_generate implements Module {
	/** TITLE is the heading, the crumb and what the browser tab says. */
	const TITLE = "Generate key";

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
		), "key_generate.tpl");
	}

	/**
	 * lead is the sentence under the heading, as HTML. It is built here and not
	 * in the template: a template writes the newline that sits before a tag, so
	 * a sentence cut across branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($dir) {
		$where = htmlspecialchars($dir, ENT_QUOTES);

		return "Generates a key in " . $where . ", as <code>&lt;name&gt;_&lt;algorithm&gt;.key</code> and <code>.pub</code>. A name is unique across the folder.";
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_key_generate;

	$mod->serve();
}
