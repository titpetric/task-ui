<?php

// @route GET /settings/keys/{name}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_key_edit sets whose runs get one instance SSH key and whether git pulls
// with it.
//
// The page reads and does not write, and it takes key/view: who holds a key is
// what the Keys page already says. Saving is a PUT to /api/keys/<name>, which
// takes key/edit and gates it again.
class mod_key_edit implements Module {
	/** PAGE is the Keys page, where the form goes by itself. */
	const PAGE = "/settings/keys";

	/** LEAD is the rest of the sentence under the heading, after the key's own line. */
	const LEAD = ". Choose whose runs get it: runs started by a checked user, or by anyone with a checked role, have it in their ssh agent. Users see their keys under SSH keys in their menu.";

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

		if (!$chrome->allowed("key/view")) {
			$view->refused();

			return;
		}

		$asked = Task\Route::param("name");
		$folder = Task\Key::folder();
		$key = self::of($folder, $asked);

		if (count($key) === 0) {
			$view->missing("There is nothing at /" . Task\Route::path() . ".");

			return;
		}

		$name = $key["name"];
		$next = self::next();

		$chrome->heading("Edit " . $name, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::KEYS, "href" => self::PAGE),
			array("text" => $name, "href" => self::PAGE . "#key-" . $name),
			array("text" => "Edit", "href" => ""),
		));

		$head = new Task\Head("Edit " . $name, self::lead($key));
		$steps = Task\Setup::steps($chrome, Task\Setup::KEYS);
		$users = new Task\User($chrome->db);
		$everyone = $users->all();
		$field = Task\Key::policies($key["roles"], $key["users"], $everyone);

		$view->page(array(
			"head" => $head->data(),
			"api" => "/api/keys/" . $name,
			"saved" => $next . "?notice=Links+saved#key-" . $name,
			"cancel" => Task\Route::href($next),
			"git" => $key["git"],
			"who" => $field->data(),
			"trail" => Task\Setup::trail($steps),
		), "key_edit.tpl");
	}

	/** of is the key of the folder a name asks for, an empty row for none. */
	static function of($folder, $name) {
		foreach ($folder["keys"] as $key) {
			if ($key["name"] === $name) {
				return $key;
			}
		}

		return array();
	}

	/**
	 * next is the page a save returns to: the one that opened the form, as a
	 * path of this instance, and the Keys page for anything else.
	 * frontend/key.go:179.
	 */
	static function next() {
		$asked = Task\Route::param("next");

		if (strpos($asked, "/") !== 0 || strpos($asked, "//") === 0) {
			return self::PAGE;
		}

		return $asked;
	}

	/**
	 * lead is the sentence under the heading, as HTML. It is built here and not
	 * in the template: a template writes the newline that sits before a tag, so
	 * a sentence cut across branches arrives cut. See docs/phpfront.md.
	 *
	 * The comment is what a reader typed, so it is escaped where the markup
	 * around it is not.
	 */
	static function lead($key) {
		$said = "";
		$comment = $key["comment"];

		if ($comment !== "") {
			$said = htmlspecialchars($comment, ENT_QUOTES) . ", ";
		}

		$kind = htmlspecialchars(Task\Key::kind($key), ENT_QUOTES);
		$print = htmlspecialchars($key["fingerprint"], ENT_QUOTES);

		return $said . '<span class="dim">' . $kind . "</span> " . $print . self::LEAD;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_key_edit;

	$mod->serve();
}
